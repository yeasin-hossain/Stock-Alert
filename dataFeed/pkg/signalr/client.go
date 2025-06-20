// Package signalr provides SignalR client functionality
package signalr

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/philippseith/signalr"

	"datafeed/pkg/config"
)

// Message represents a SignalR message
type Message struct {
	Method string
	Data   interface{}
}

// ConnectionStatus represents the current state of the connection
type ConnectionStatus int

const (
	// ConnectionStatusDisconnected indicates the client is disconnected
	ConnectionStatusDisconnected ConnectionStatus = iota
	// ConnectionStatusConnecting indicates the client is connecting
	ConnectionStatusConnecting
	// ConnectionStatusConnected indicates the client is connected
	ConnectionStatusConnected
	// ConnectionStatusReconnecting indicates the client is reconnecting
	ConnectionStatusReconnecting
)

// ClientConfig holds configuration options for the SignalR client
type ClientConfig struct {
	// Connection settings
	ConnectionTimeout    time.Duration
	ReconnectDelay       time.Duration
	MaxReconnectDelay    time.Duration
	MaxReconnectAttempts int

	// Message handling
	MessageBufferSize int
	EnableHeartbeat   bool
	HeartbeatInterval time.Duration

	// HTTP settings
	UserAgent         string
	AdditionalHeaders map[string]string
	HTTPTimeout       time.Duration
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ConnectionTimeout:    30 * time.Second,
		ReconnectDelay:       2 * time.Second,
		MaxReconnectDelay:    2 * time.Minute,
		MaxReconnectAttempts: 20,
		MessageBufferSize:    100,
		EnableHeartbeat:      true,
		HeartbeatInterval:    30 * time.Second,
		UserAgent:            "Go-SignalR-Client/1.0",
		HTTPTimeout:          30 * time.Second,
		AdditionalHeaders:    make(map[string]string),
	}
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

	// Connection management
	connMu        sync.Mutex
	connStatus    ConnectionStatus
	connError     error
	reconnectChan chan struct{}

	// Reconnection settings
	baseReconnectDelay   time.Duration
	maxReconnectDelay    time.Duration
	maxReconnectAttempts int
	reconnectAttempts    int

	// Subscriptions to reapply on reconnection
	subscriptionsMu sync.RWMutex
	subscriptions   map[string][]interface{}
}

// Messages returns the channel that receives SignalR messages
func (c *Client) Messages() <-chan Message {
	return c.messagesChan
}

// Status returns the current connection status
func (c *Client) Status() ConnectionStatus {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.connStatus
}

// LastError returns the last connection error
func (c *Client) LastError() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.connError
}

// Subscribe subscribes to a SignalR event with the provided arguments
func (c *Client) Subscribe(method string, args ...interface{}) error {
	if c.Status() != ConnectionStatusConnected {
		return fmt.Errorf("not connected (status: %v)", c.Status())
	}

	// Store subscription for reconnect
	c.storeSubscription(method, args...)

	// Debug the subscription
	c.logger.Printf("Subscribing to method %s with %d arguments", method, len(args))

	// Use Invoke as per documentation
	go func() {
		c.logger.Printf("Starting Invoke for method: %s", method)
		result := <-c.client.Send(method, args...)
		c.logger.Printf("Subscription result for %s: %v", method, result)

		// Check for errors in the result
		if result != nil {
			c.logger.Printf("Subscription completed with result: %v", result)
		}
	}()

	return nil
}

// storeSubscription stores a subscription for reapplication after reconnect
func (c *Client) storeSubscription(method string, args ...interface{}) {
	c.subscriptionsMu.Lock()
	defer c.subscriptionsMu.Unlock()

	// Create a copy of args to store
	argsCopy := make([]interface{}, len(args))
	copy(argsCopy, args)

	c.subscriptions[method] = argsCopy
}

// MessageHandler is a function type for handling SignalR messages
type MessageHandler func(msg Message)

// MessageReceiver implements signalr.Receiver for handling server callbacks
// The key is to embed signalr.Hub and implement the Receive method to catch all calls
type MessageReceiver struct {
	messagesChan chan<- Message
	logger       *log.Logger
	client       *Client // Reference back to the client
	signalr.Hub          // Embed Hub - this provides the base receiver functionality

	// Handler registry
	handlersMu sync.RWMutex
	handlers   map[string]MessageHandler
}

// The SignalR library will call Receive for ANY method that doesn't exist on the receiver
// This is our universal handler for all server-to-client method calls
// including those with special characters like "MarketStatusUpdated^^DSE~"

// RegisterHandler registers a custom handler for a specific method name
func (r *MessageReceiver) RegisterHandler(methodName string, handler MessageHandler) {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()

	lowerMethod := strings.ToLower(methodName)
	r.logger.Printf("Registering handler for method: %s (stored as: %s)", methodName, lowerMethod)
	r.handlers[lowerMethod] = handler
}

// Receive handles incoming SignalR messages and sends them to the message channel
// This is the core function that gets called by the SignalR library for all server-to-client methods
func (r *MessageReceiver) Receive(method string, args ...interface{}) {
	// Log every received message with details for debugging
	if r.logger != nil {
		r.logger.Printf("===> ENTRY POINT: Receive method called with method=%s and %d arguments", method, len(args))

		// If we have arguments, log their types to help with debugging
		if len(args) > 0 {
			for i, arg := range args {
				typeName := fmt.Sprintf("%T", arg)
				r.logger.Printf("  Arg[%d] type: %s", i, typeName)

				// For string args, log a preview of content
				if str, ok := arg.(string); ok {
					preview := str
					if len(str) > 100 {
						preview = str[:100] + "..."
					}
					r.logger.Printf("  Arg[%d] content: %s", i, preview)
				}
			}
		}
	} else {
		log.Printf("WARNING: Logger is nil in MessageReceiver.Receive")
	}

	// Normalize method name for case-insensitive matching
	normalizedMethod := strings.ToLower(method)

	// Check for registered handler first
	r.handlersMu.RLock()
	handler, exists := r.handlers[normalizedMethod]
	r.handlersMu.RUnlock()

	if exists {
		r.logger.Printf("Found registered handler for method: %s", method)
		msg := Message{
			Method: method,
			Data:   args,
		}
		handler(msg)
		return
	}

	// Route to specific handler methods based on the method name
	switch normalizedMethod {
	case "sharepriceupdated":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				r.logger.Printf("Routing to SharePriceUpdated handler")
				r.SharePriceUpdated(str)
				return
			}
		}
	case "marketstatusupdated^^dse~":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				r.logger.Printf("Routing to MarketStatusUpdated^^DSE~ handler")
				r.MarketStatusUpdated__DSE_(str)
				return
			}
		}
	}

	// For non-routed messages or if routing failed, send to the general channel
	r.logger.Printf("No specific handler found for method: %s, using general handler", method)
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

	// Notify the client of the error
	if r.client != nil {
		r.client.handleServerError(errorMessage)
	}
}

// HandleConnectionEvent processes connection state change notifications from the server
func (r *MessageReceiver) HandleConnectionEvent(eventType string, data interface{}) {
	r.logger.Printf("Connection event received: %s, data: %v", eventType, data)

	// Notify the client of connection events
	if r.client != nil {
		switch strings.ToLower(eventType) {
		case "connected":
			r.client.handleConnected()
		case "disconnected":
			r.client.handleDisconnected(fmt.Errorf("server sent disconnect: %v", data))
		case "reconnecting":
			r.client.handleReconnecting()
		}
	}

	// Forward the event to the message channel
	r.messagesChan <- Message{
		Method: "ConnectionEvent",
		Data: map[string]interface{}{
			"type": eventType,
			"data": data,
		},
	}
}

// NewClient creates a new SignalR client
func NewClient(cfg *config.Config, token string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	messagesChan := make(chan Message, 100)

	client := &Client{
		hubURL:               cfg.SignalRURL,
		token:                token,
		messagesChan:         messagesChan,
		logger:               log.New(os.Stdout, "[_________SignalR_________] ", log.LstdFlags),
		ctx:                  ctx,
		cancel:               cancel,
		reconnectChan:        make(chan struct{}, 1),
		connStatus:           ConnectionStatusDisconnected,
		baseReconnectDelay:   2 * time.Second,
		maxReconnectDelay:    2 * time.Minute,
		maxReconnectAttempts: 20,
		subscriptions:        make(map[string][]interface{}),
	}

	// Create message receiver with proper handlers map and client reference
	client.receiver = &MessageReceiver{
		messagesChan: messagesChan,
		logger:       log.New(os.Stdout, "[***********SignalR Receiver***********] ", log.LstdFlags),
		client:       client,
		handlers:     make(map[string]MessageHandler),
	}

	return client
}

// NewClientWithConfig creates a new SignalR client with custom configuration
func NewClientWithConfig(cfg *config.Config, token string, clientCfg *ClientConfig) *Client {
	if clientCfg == nil {
		clientCfg = DefaultClientConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	messagesChan := make(chan Message, clientCfg.MessageBufferSize)

	client := &Client{
		hubURL:               cfg.SignalRURL,
		token:                token,
		messagesChan:         messagesChan,
		logger:               log.New(os.Stdout, "[_________SignalR_________] ", log.LstdFlags),
		ctx:                  ctx,
		cancel:               cancel,
		reconnectChan:        make(chan struct{}, 1),
		connStatus:           ConnectionStatusDisconnected,
		baseReconnectDelay:   clientCfg.ReconnectDelay,
		maxReconnectDelay:    clientCfg.MaxReconnectDelay,
		maxReconnectAttempts: clientCfg.MaxReconnectAttempts,
		subscriptions:        make(map[string][]interface{}),
	}

	// Create message receiver with proper handlers map and client reference
	client.receiver = &MessageReceiver{
		messagesChan: messagesChan,
		logger:       log.New(os.Stdout, "[***********SignalR Receiver***********] ", log.LstdFlags),
		client:       client,
		handlers:     make(map[string]MessageHandler),
	}

	return client
}

// handleConnected updates the connection state when connected
func (c *Client) handleConnected() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	wasReconnecting := c.connStatus == ConnectionStatusReconnecting
	c.connStatus = ConnectionStatusConnected
	c.reconnectAttempts = 0
	c.connError = nil

	c.logger.Printf("SignalR connection established")

	// If we were reconnecting, reapply all subscriptions
	if wasReconnecting {
		go c.reapplySubscriptions()
	}
}

// handleDisconnected updates the connection state when disconnected
func (c *Client) handleDisconnected(err error) {
	c.connMu.Lock()
	if c.connStatus == ConnectionStatusDisconnected {
		// Already disconnected
		c.connMu.Unlock()
		return
	}

	c.connStatus = ConnectionStatusDisconnected
	c.connError = err
	c.connMu.Unlock()

	c.logger.Printf("SignalR disconnected: %v", err)

	// Trigger reconnection if not explicitly closed
	select {
	case <-c.ctx.Done():
		// Context canceled, don't reconnect
		return
	default:
		// Trigger reconnection
		select {
		case c.reconnectChan <- struct{}{}:
		default:
			// Channel already has a pending reconnect signal
		}
	}
}

// handleReconnecting updates the connection state when reconnecting
func (c *Client) handleReconnecting() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	c.connStatus = ConnectionStatusReconnecting
	c.logger.Printf("SignalR reconnecting...")
}

// handleServerError processes errors from the server
func (c *Client) handleServerError(errorMessage string) {
	c.logger.Printf("Server error: %s", errorMessage)

	// We don't disconnect on all errors, but log them
	// Some errors might indicate need for reconnection
	if strings.Contains(strings.ToLower(errorMessage), "unauthorized") ||
		strings.Contains(strings.ToLower(errorMessage), "auth") {
		c.logger.Printf("Authentication error detected, will reconnect with fresh token")
		c.handleDisconnected(fmt.Errorf("auth error: %s", errorMessage))
	}
}

// Connect establishes a connection to the SignalR hub
func (c *Client) Connect() error {
	c.connMu.Lock()

	// Check if already connecting or connected
	if c.connStatus == ConnectionStatusConnecting || c.connStatus == ConnectionStatusConnected {
		c.connMu.Unlock()
		return nil
	}

	// Update status
	c.connStatus = ConnectionStatusConnecting
	c.connMu.Unlock()

	c.logger.Println("Connecting to SignalR hub:", c.hubURL)

	// Create HTTP connection with configurable options
	// Use a timeout for the initial connection
	creationCtx, creationCancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer creationCancel()

	// Configurable HTTP connection with proper headers
	conn, err := signalr.NewHTTPConnection(creationCtx, c.hubURL,
		signalr.WithHTTPHeaders(func() http.Header {
			h := make(http.Header)
			h.Set("Authorization", "Bearer "+c.token)
			h.Set("User-Agent", "Go-SignalR-Client/1.0")
			h.Set("Accept", "application/json")
			h.Set("Content-Type", "application/json")

			// Add any additional headers if we have a client config
			// This is a placeholder for future enhancement
			return h
		}),
		// Add keep-alive and timeout settings
		signalr.WithHTTPClient(&http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}),
	)
	if err != nil {
		c.handleConnectionError(err)
		return fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	// Create signalr client with receiver that can handle special character method names
	c.client, err = signalr.NewClient(
		c.ctx,
		signalr.WithConnection(conn),
		signalr.TransferFormat(signalr.TransferFormatText), // Use text format for JSON
		signalr.WithReceiver(c.receiver),                   // Use our receiver with Hub embedding
	)
	if err != nil {
		c.handleConnectionError(err)
		return fmt.Errorf("failed to create SignalR client: %w", err)
	}

	// Start the client
	c.client.Start()
	c.logger.Println("SignalR client started")

	// Handle connection state manually since we don't have callbacks
	c.handleConnected()

	// Start connection monitor
	go c.monitorConnection()

	// Start heartbeat to detect broken connections
	c.startHeartbeat()

	// Subscribe to default events automatically after connection
	c.SubscribeToDefaultEvents()

	return nil
}

// handleConnectionError processes connection errors
func (c *Client) handleConnectionError(err error) {
	c.connMu.Lock()
	c.connStatus = ConnectionStatusDisconnected
	c.connError = err
	c.connMu.Unlock()
}

// SubscribeToDefaultEvents subscribes to all the required events
// after the connection is established
func (c *Client) SubscribeToDefaultEvents() {
	c.logger.Println("Subscribing to default events...")

	// Wait a moment for the connection to stabilize
	time.Sleep(2 * time.Second)

	// Subscribe to market status updates with retry logic
	go func() {
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			c.logger.Printf("Subscribing to market status updates (attempt %d/%d)...", attempt, maxRetries)

			if c.Status() != ConnectionStatusConnected {
				c.logger.Printf("Not connected, skipping subscription attempt %d", attempt)
				time.Sleep(5 * time.Second)
				continue
			}

			if err := c.Subscribe("SubscribeToMarketStatusUpdatedEvent", "DSE"); err != nil {
				c.logger.Printf("Warning: market status subscription failed (attempt %d): %v", attempt, err)
				if attempt < maxRetries {
					time.Sleep(5 * time.Second)
					continue
				}
			} else {
				c.logger.Println("✅ Successfully subscribed to market status updates")
				break
			}
		}
	}()

	// Subscribe to share price updates
	// Commented out but stored for reference
	/*
		go func() {
			c.logger.Println("Subscribing to share price updates...")
			if err := c.Subscribe("SubscribeToSharePriceUpdatedEvent", "500$1$$Asc", "DSE", nil, "", "", "", []interface{}{}, "", nil, false, nil); err != nil {
				c.logger.Printf("Warning: share price subscription failed: %v", err)
			} else {
				c.logger.Println("Successfully subscribed to share price updates")
			}
		}()
	*/
}

// monitorConnection continuously monitors the connection status and
// attempts reconnection if necessary
func (c *Client) monitorConnection() {
	c.logger.Println("Starting connection monitor")

	// Wait for reconnect signals or cancellation
	for {
		select {
		case <-c.ctx.Done():
			// Client is closing
			c.logger.Println("Connection monitor shutting down: context canceled")
			return

		case <-c.reconnectChan:
			// Need to reconnect
			c.attemptReconnect()
		}
	}
}

// attemptReconnect tries to reconnect with exponential backoff
func (c *Client) attemptReconnect() {
	c.connMu.Lock()

	// Check if we've already reconnected somehow
	if c.connStatus == ConnectionStatusConnected {
		c.connMu.Unlock()
		return
	}

	// Check if we've exceeded the maximum number of attempts
	if c.reconnectAttempts >= c.maxReconnectAttempts {
		c.logger.Printf("Giving up on reconnection after %d attempts", c.reconnectAttempts)
		c.connMu.Unlock()
		return
	}

	// Calculate backoff time
	c.reconnectAttempts++
	attempt := c.reconnectAttempts
	backoff := time.Duration(float64(c.baseReconnectDelay) * (1.5 * float64(attempt-1)))
	if backoff > c.maxReconnectDelay {
		backoff = c.maxReconnectDelay
	}

	c.connStatus = ConnectionStatusReconnecting
	c.connMu.Unlock()

	// Log the reconnection attempt
	c.logger.Printf("Reconnection attempt #%d after %v", attempt, backoff)

	// Wait for backoff period
	select {
	case <-time.After(backoff):
		break
	case <-c.ctx.Done():
		return
	}

	// Attempt reconnection
	c.logger.Println("Executing reconnection attempt")

	// Close existing client if any
	if c.client != nil {
		c.client = nil
	}

	// Reconnect
	if err := c.Connect(); err != nil {
		c.logger.Printf("Reconnection attempt #%d failed: %v", attempt, err)

		// Schedule another attempt
		select {
		case c.reconnectChan <- struct{}{}:
		default:
		}
	} else {
		c.logger.Printf("Reconnection successful after %d attempts", attempt)
	}
}

// reapplySubscriptions reapplies all stored subscriptions after reconnection
func (c *Client) reapplySubscriptions() {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()

	c.logger.Printf("Reapplying %d stored subscriptions", len(c.subscriptions))

	for method, args := range c.subscriptions {
		c.logger.Printf("Resubscribing to %s with %d arguments", method, len(args))
		if err := c.Subscribe(method, args...); err != nil {
			c.logger.Printf("Error resubscribing to %s: %v", method, err)
		}
	}
}

// UpdateToken updates the authentication token and reconnects if necessary
func (c *Client) UpdateToken(newToken string) error {
	c.connMu.Lock()
	oldToken := c.token
	c.token = newToken
	needsReconnect := c.isConnected() && oldToken != newToken
	c.connMu.Unlock()

	c.logger.Println("Authentication token updated")

	// If connected with the old token, reconnect with the new one
	if needsReconnect {
		c.logger.Println("Token changed, reconnecting with new token...")

		// Close existing connection
		if c.client != nil {
			c.client.Stop()
			c.client = nil
		}

		// Trigger reconnection
		select {
		case c.reconnectChan <- struct{}{}:
		default:
		}
	}

	return nil
}

// isConnected returns true if the client is currently connected
// This function assumes the lock is already held
func (c *Client) isConnected() bool {
	return c.connStatus == ConnectionStatusConnected
}

// Close cleanly closes the SignalR connection
func (c *Client) Close() {
	c.logger.Println("Closing SignalR client")

	// Cancel context to stop all operations
	if c.cancel != nil {
		c.cancel()
	}

	// Update status
	c.connMu.Lock()
	c.connStatus = ConnectionStatusDisconnected
	c.connMu.Unlock()

	// Close the client if it exists
	if c.client != nil {
		// Try to send close message gracefully
		c.logger.Println("Stopping SignalR client")
		c.client.Stop()
		c.client = nil
	}

	// Close message channel last
	close(c.messagesChan)
	c.logger.Println("SignalR client closed")
}

// startHeartbeat starts a heartbeat to detect broken connections
func (c *Client) startHeartbeat() {
	c.logger.Println("Starting connection heartbeat")

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check if we're connected
				if c.Status() != ConnectionStatusConnected {
					continue
				}

				// Send a ping to check the connection
				c.logger.Println("Sending heartbeat ping")
				go func() {
					// Try to invoke a ping method
					// Create a channel to receive the result with timeout
					resultChan := c.client.Send("ping")

					select {
					case result := <-resultChan:
						if result != nil {
							c.logger.Printf("WARNING: Heartbeat ping failed: %v", result)

							// The connection might be broken
							c.logger.Println("Heartbeat failed, triggering reconnection")
							select {
							case c.reconnectChan <- struct{}{}:
							default:
							}
						} else {
							c.logger.Println("Heartbeat ping successful")
						}
					case <-time.After(10 * time.Second):
						// Ping timeout - connection might be broken
						c.logger.Println("Heartbeat ping timeout, triggering reconnection")
						select {
						case c.reconnectChan <- struct{}{}:
						default:
						}
					}
				}()

			case <-c.ctx.Done():
				c.logger.Println("Heartbeat stopping due to context cancellation")
				return
			}
		}
	}()
}

// RegisterCustomHandler allows registering custom handlers for specific methods
func (c *Client) RegisterCustomHandler(methodName string, handler MessageHandler) {
	if c.receiver != nil {
		c.receiver.RegisterHandler(methodName, handler)
		c.logger.Printf("Registered custom handler for method: %s", methodName)
	}
}

// GetConnectionStats returns connection statistics
func (c *Client) GetConnectionStats() map[string]interface{} {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	stats := map[string]interface{}{
		"status":            c.connStatus,
		"reconnectAttempts": c.reconnectAttempts,
		"lastError":         c.connError,
		"subscriptions":     len(c.subscriptions),
	}

	return stats
}

// Ping sends a ping message to test the connection
func (c *Client) Ping() error {
	if c.Status() != ConnectionStatusConnected {
		return fmt.Errorf("not connected")
	}

	c.logger.Println("Sending ping to server")
	go func() {
		result := <-c.client.Send("ping")
		if result == nil {
			c.logger.Println("Ping successful")
		} else {
			c.logger.Printf("Ping failed: %v", result)
		}
	}()

	return nil
}

// We need to handle the method name mapping issue
// The server calls "MarketStatusUpdated^^DSE~" but Go can't have such method names
// Solution: Use interface{} and implement our own method dispatch

// MethodDispatcher handles method calls with special characters
type MethodDispatcher struct {
	receiver *MessageReceiver
}

// Implement a universal method handler using reflection workaround
func (md *MethodDispatcher) CallMethod(method string, args ...interface{}) {
	md.receiver.logger.Printf("MethodDispatcher: Calling method %s", method)

	// Map problematic method names to our handlers
	switch method {
	case "MarketStatusUpdated^^DSE~":
		md.receiver.logger.Printf("Mapping MarketStatusUpdated^^DSE~ to our handler")
		md.receiver.Receive("MarketStatusUpdated^^DSE~", args...)
	case "SharePriceUpdated":
		md.receiver.logger.Printf("Mapping SharePriceUpdated to our handler")
		md.receiver.Receive("SharePriceUpdated", args...)
	default:
		md.receiver.logger.Printf("Unknown method: %s, using default handler", method)
		md.receiver.Receive(method, args...)
	}
}
