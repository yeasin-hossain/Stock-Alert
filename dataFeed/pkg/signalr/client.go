// Package signalr provides SignalR client functionality
package signalr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/philippseith/signalr"

	"datafeed/pkg/config"
)

// Message represents a SignalR message
type Message struct {
	Method string
	Data   interface{}
}

// Client handles SignalR connections and messages
type Client struct {
	hubURL       string
	token        string
	client       signalr.Client
	messagesChan chan Message
	logger       *log.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	receiver     *MessageReceiver
}

// Messages returns the channel that receives SignalR messages
func (c *Client) Messages() <-chan Message {
	return c.messagesChan
}

// Subscribe subscribes to a SignalR event with the provided arguments
func (c *Client) Subscribe(method string, args ...interface{}) error {
	if c.client == nil {
		return errors.New("client not connected")
	}

	// Debug the subscription
	c.logger.Printf("Subscribing to method %s with %d arguments", method, len(args))

	// Use Invoke as per documentation
	go func() {
		c.logger.Printf("Starting Invoke for method: %s", method)
		result := <-c.client.Send(method, args...)
		c.logger.Printf("Subscription result for %s: %v", method, result)
	}()

	return nil
}

// MessageHandler is a function type for handling SignalR messages
type MessageHandler func(msg Message)

// MessageReceiver implements signalr.Receiver for handling server callbacks
type MessageReceiver struct {
	messagesChan chan<- Message
	logger       *log.Logger
	signalr.Hub  // Embed the Hub as per documentation
}

// Receive handles incoming SignalR messages and sends them to the message channel
func (r *MessageReceiver) Receive(method string, args ...interface{}) {
	// Use both standard logger and custom logger for maximum visibility
	fmt.Printf("=== DEBUG: Receive method called with method=[%s] and %d arguments ===\n", method, len(args))
	log.Printf("=== DEBUG: Receive method called with method=[%s] and %d arguments ===", method, len(args))

	if r.logger != nil {
		r.logger.Printf("===> ENTRY POINT: Receive method called with method=%s and %d arguments", method, len(args))
	} else {
		log.Printf("WARNING: Logger is nil in MessageReceiver.Receive")
	}

	// Special case for ping message (type 6) which might come with empty method name
	if method == "" || method == "ping" {
		// Check if this might be a ping message
		if len(args) == 0 || (len(args) == 1 && args[0] == nil) {
			r.logger.Printf("Detected possible ping message (type 6)")
			r.HandlePing()
			return
		}
	}

	// Try to detect ping message from JSON format: {"type":6}
	if len(args) > 0 {
		if messageMap, ok := args[0].(map[string]interface{}); ok {
			if typeVal, ok := messageMap["type"].(float64); ok && typeVal == 6 {
				r.logger.Printf("Detected ping message from JSON payload (type 6)")
				r.HandlePing()
				return
			}
		}

		// Log the first argument if available
		r.logger.Printf("First argument type: %T", args[0])
		// Try to log string values
		if str, ok := args[0].(string); ok && len(str) < 100 {
			r.logger.Printf("First argument value (string): %s", str)

			// Check if the string contains type:6 (ping message in text format)
			if str == "{\"type\":6}" {
				r.logger.Printf("Detected ping message from string (type 6)")
				r.HandlePing()
				return
			}
		}
	}

	// Route to specific handler methods based on the method name
	switch method {
	case "SharePriceUpdated", "sharePriceUpdated":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				r.logger.Printf("Routing to SharePriceUpdated handler")
				r.SharePriceUpdated(str)
				return
			}
		}
	case "MarketStatusUpdated^^DSE~", "marketStatusUpdated^^dse~":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				r.logger.Printf("Routing to MarketStatusUpdated^^DSE~ handler")
				r.MarketStatusUpdated__DSE_(str)
				return
			}
		}
	}

	// For non-routed messages or if routing failed, send to the general channel
	r.messagesChan <- Message{
		Method: method,
		Data:   args,
	}
}

// SharePriceUpdated is called when the server sends a SharePriceUpdated event
func (r *MessageReceiver) SharePriceUpdated(data string) {
	r.logger.Printf("SharePriceUpdated specific handler called with data length: %d", len(data))
	if len(data) < 100 {
		r.logger.Printf("Data content: %s", data)
	} else {
		r.logger.Printf("Data content (truncated): %s...", data[:100])
	}

	// Process the message using a specific method
	// You can add custom processing logic here if needed

	// Send the processed message to the channel
	r.messagesChan <- Message{
		Method: "SharePriceUpdated",
		Data:   data,
	}
}

// MarketStatusUpdated^^DSE~ is called when the server sends a MarketStatusUpdated event
func (r *MessageReceiver) MarketStatusUpdated__DSE_(data string) {
	r.logger.Printf("MarketStatusUpdated^^DSE~ handler called with data length: %d", len(data))
	if len(data) < 100 {
		r.logger.Printf("Market status data content: %s", data)
	} else {
		r.logger.Printf("Market status data content (truncated): %s...", data[:100])
	}

	// Send the processed message to the channel
	r.messagesChan <- Message{
		Method: "MarketStatusUpdated^^DSE~",
		Data:   data,
	}
}

// SubscribeToSharePriceUpdatedEvent handles subscription responses
func (r *MessageReceiver) SubscribeToSharePriceUpdatedEvent(result interface{}) {
	r.logger.Printf("Subscription result received: %v", result)

	// No need to forward this to the messagesChan
	// as it's just a confirmation of the subscription
}

// HandleError handles any error messages from the server
func (r *MessageReceiver) HandleError(errorMessage string) {
	r.logger.Printf("Error received from server: %s", errorMessage)

	r.messagesChan <- Message{
		Method: "Error",
		Data:   errorMessage,
	}
}

// HandlePing handles ping messages (type 6) from the server
func (r *MessageReceiver) HandlePing() {
	r.logger.Printf("Received ping message from server (type 6)")

	// Send a ping response message to the channel for processing
	r.messagesChan <- Message{
		Method: "Ping",
		Data:   nil,
	}
}

// NewClient creates a new SignalR client
func NewClient(cfg *config.Config, token string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	messagesChan := make(chan Message, 100)
	client := &Client{
		hubURL:       cfg.SignalRURL,
		token:        token,
		messagesChan: messagesChan,
		logger:       log.New(os.Stdout, "[_________SignalR_________] ", log.LstdFlags),
		ctx:          ctx,
		cancel:       cancel,
	}

	client.receiver = &MessageReceiver{
		messagesChan: messagesChan,
		logger:       log.New(os.Stdout, "[SignalR Receiver] ", log.LstdFlags),
	}

	return client
}

// Connect establishes a connection to the SignalR hub
func (c *Client) Connect() error {
	c.logger.Println("Connecting to SignalR hub:", c.hubURL)

	// Create HTTP connection with authorization header
	// Use a timeout for the initial connection
	creationCtx, creationCancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer creationCancel()

	conn, err := signalr.NewHTTPConnection(creationCtx, c.hubURL, signalr.WithHTTPHeaders(func() http.Header {
		h := make(http.Header)
		h.Set("Authorization", "Bearer "+c.token)
		return h
	}))
	if err != nil {
		return fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	// Create signalr client with receiver
	c.client, err = signalr.NewClient(
		c.ctx,
		signalr.WithConnection(conn),
		signalr.WithReceiver(c.receiver),
	)
	if err != nil {
		return fmt.Errorf("failed to create SignalR client: %w", err)
	}

	// Start the client
	c.client.Start()
	c.logger.Println("Connected successfully to SignalR hub")

	// Subscribe to default events automatically after connection
	c.SubscribeToDefaultEvents()

	return nil
}

// SubscribeToDefaultEvents subscribes to all the required events
// after the connection is established
func (c *Client) SubscribeToDefaultEvents() {
	c.logger.Println("Subscribing to default events...")

	// Subscribe to market status updates
	go func() {
		c.logger.Println("Subscribing to market status updates...")
		if err := c.Subscribe("SubscribeToMarketStatusUpdatedEvent", "DSE"); err != nil {
			c.logger.Printf("Warning: market status subscription failed: %v", err)
		} else {
			c.logger.Println("Successfully subscribed to market status updates")
		}
	}()

	// Subscribe to share price updates
	// go func() {
	// 	c.logger.Println("Subscribing to share price updates...")
	// 	if err := c.Subscribe("SubscribeToSharePriceUpdatedEvent", "500$1$$Asc", "DSE", nil, "", "", "", []interface{}{}, "", nil, false, nil); err != nil {
	// 		c.logger.Printf("Warning: share price subscription failed: %v", err)
	// 	} else {
	// 		c.logger.Println("Successfully subscribed to share price updates")
	// 	}
	// }()
}

// Close closes the SignalR connection
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	close(c.messagesChan)
	c.logger.Println("SignalR client closed")
}
