// Package websocket provides WebSocket client functionality
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"datafeed/pkg/config"
)

// Message represents a WebSocket message
type Message struct {
	Type string          `json:"type,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
	// Add other fields as needed based on your WebSocket server's message format
}

// Client handles WebSocket connections and messages
type Client struct {
	// Connection parameters
	url     string
	token   string
	headers http.Header
	conn    *websocket.Conn

	// Channels for message passing
	sendChan    chan []byte
	receiveChan chan Message

	// Connection state
	mu          sync.Mutex
	isConnected bool

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Handlers for specific message types
	handlers map[string][]func([]byte)

	// Logging
	logger *log.Logger

	// Reconnection settings
	reconnectWait time.Duration
	maxRetries    int
}

// NewClient creates a new WebSocket client
func NewClient(cfg *config.WebSocketConfig, token string) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		url:           cfg.URL,
		token:         token,
		headers:       make(http.Header),
		sendChan:      make(chan []byte, 100),
		receiveChan:   make(chan Message, 100),
		handlers:      make(map[string][]func([]byte)),
		ctx:           ctx,
		cancel:        cancel,
		logger:        log.New(os.Stdout, "[WebSocket] ", log.LstdFlags),
		reconnectWait: 2 * time.Second,
		maxRetries:    10,
	}

	// Set default headers
	client.headers.Set("Authorization", "Bearer "+token)

	// Add any additional headers from config
	if cfg.Headers != nil {
		for key, value := range cfg.Headers {
			client.headers.Set(key, value)
		}
	}

	// Set subprotocol if specified
	if cfg.Protocol != "" {
		client.headers.Set("Sec-WebSocket-Protocol", cfg.Protocol)
	}

	return client
}

// Connect establishes a WebSocket connection
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	c.logger.Printf("Connecting to WebSocket server: %s", c.url)

	// Parse URL
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	// Establish connection
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), c.headers)
	if err != nil {
		if resp != nil {
			c.logger.Printf("Connection failed with status: %d", resp.StatusCode)
		}
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	c.logger.Printf("Connected to WebSocket server")

	// Start goroutines for reading and writing
	go c.readPump()
	go c.writePump()

	// Start connection monitor for automatic reconnection
	go c.monitorConnection()

	return nil
}

// On registers a handler function for a specific message type
func (c *Client) On(messageType string, handler func(data []byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handlers[messageType] == nil {
		c.handlers[messageType] = make([]func([]byte), 0)
	}

	c.handlers[messageType] = append(c.handlers[messageType], handler)
	c.logger.Printf("Registered handler for message type: %s", messageType)
}

// Send sends a message to the WebSocket server
func (c *Client) Send(data []byte) error {
	if !c.isConnected {
		return fmt.Errorf("not connected to WebSocket server")
	}

	select {
	case c.sendChan <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

// SendJSON sends a JSON message to the WebSocket server
func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.Send(data)
}

// Receive returns a channel that receives WebSocket messages
func (c *Client) Receive() <-chan Message {
	return c.receiveChan
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Println("Closing WebSocket connection")

	// Signal all goroutines to stop
	c.cancel()

	// Close the connection
	if c.conn != nil {
		// Send close message
		err := c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			c.logger.Printf("Error sending close message: %v", err)
		}

		c.conn.Close()
		c.conn = nil
	}

	c.isConnected = false

	// Close channels
	close(c.sendChan)
	close(c.receiveChan)

	c.logger.Println("WebSocket connection closed")
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.isConnected
}

// readPump pumps messages from the WebSocket connection to the receiveChan
func (c *Client) readPump() {
	defer func() {
		c.logger.Println("Read pump exiting")
		c.handleDisconnect()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Printf("WebSocket read error: %v", err)
				}
				return
			}

			c.processMessage(message)
		}
	}
}

// writePump pumps messages from the sendChan to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.logger.Println("Write pump exiting")
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.sendChan:
			if !ok {
				// Channel closed
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping message
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Printf("WebSocket ping error: %v", err)
				return
			}
		}
	}
}

// processMessage processes a message from the WebSocket server
func (c *Client) processMessage(data []byte) {
	c.logger.Printf("Received message: %s", truncateString(string(data), 100))

	// Try to parse as JSON
	var message Message
	err := json.Unmarshal(data, &message)
	if err != nil {
		c.logger.Printf("Failed to parse message as JSON: %v", err)

		// For non-JSON messages, use a default message type
		message = Message{
			Type: "raw",
			Data: data,
		}
	}

	// Call handlers for this message type
	if message.Type != "" {
		c.mu.Lock()
		handlers := c.handlers[message.Type]
		c.mu.Unlock()

		for _, handler := range handlers {
			go handler(message.Data)
		}
	}

	// Send to receive channel
	select {
	case c.receiveChan <- message:
	default:
		c.logger.Println("Receive channel full, dropping message")
	}
}

// handleDisconnect handles a disconnection
func (c *Client) handleDisconnect() {
	c.mu.Lock()
	wasConnected := c.isConnected
	c.isConnected = false
	c.conn = nil
	c.mu.Unlock()

	if wasConnected {
		c.logger.Println("Disconnected from WebSocket server")
	}
}

// monitorConnection monitors the connection and reconnects if needed
func (c *Client) monitorConnection() {
	retries := 0

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(c.reconnectWait):
			// Check if we need to reconnect
			c.mu.Lock()
			needsReconnect := !c.isConnected
			c.mu.Unlock()

			if needsReconnect {
				if retries >= c.maxRetries {
					c.logger.Printf("Max reconnection attempts (%d) reached, giving up", c.maxRetries)
					return
				}

				retries++
				c.logger.Printf("Attempting to reconnect (attempt %d of %d)", retries, c.maxRetries)

				if err := c.Connect(); err != nil {
					c.logger.Printf("Reconnection failed: %v", err)
					// Use exponential backoff
					c.reconnectWait = time.Duration(float64(c.reconnectWait) * 1.5)
					if c.reconnectWait > 60*time.Second {
						c.reconnectWait = 60 * time.Second
					}
				} else {
					c.logger.Println("Reconnection successful")
					retries = 0
					c.reconnectWait = 2 * time.Second // Reset to initial value
				}
			}
		}
	}
}

// Helper function to truncate long strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
