# Enhanced SignalR Client for Go

This project provides a robust, production-ready SignalR client implementation in Go with advanced features for reliable real-time data reception.

## Features

### 🔄 Automatic Reconnection
- Exponential backoff strategy for reconnection attempts
- Configurable retry limits and delays  
- Subscription preservation and reapplication after reconnection
- Connection state monitoring

### 💓 Connection Health Monitoring
- Built-in heartbeat mechanism to detect broken connections
- Automatic connection testing with ping messages
- Connection status tracking and reporting
- Connection statistics and metrics

### 🔐 Authentication Management
- Automatic token refresh to prevent authentication timeouts
- Bearer token authentication with configurable headers
- Graceful handling of authentication errors

### 📨 Message Processing
- Support for all SignalR message types including special characters
- Custom message handler registration
- Fallback message handling for unknown methods
- Message buffering with configurable buffer sizes

### ⚙️ Configurable Options
- Customizable connection timeouts and retry settings
- HTTP client configuration with keep-alive settings
- Transfer format selection (Text/Binary)
- User-agent and custom header support

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "datafeed/pkg/auth"
    "datafeed/pkg/config"
    "datafeed/pkg/signalr"
)

func main() {
    // Load configuration
    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Authenticate
    token, err := auth.Login(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Create SignalR client
    client := signalr.NewClient(cfg, token)
    
    // Connect
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Process messages
    for msg := range client.Messages() {
        log.Printf("Received: %s - %v", msg.Method, msg.Data)
    }
}
```

### Advanced Configuration

```go
// Create custom client configuration
clientCfg := signalr.DefaultClientConfig()
clientCfg.MessageBufferSize = 500
clientCfg.ReconnectDelay = 5 * time.Second
clientCfg.MaxReconnectAttempts = 50
clientCfg.HeartbeatInterval = 15 * time.Second

// Create client with custom config
client := signalr.NewClientWithConfig(cfg, token, clientCfg)

// Register custom handlers
client.RegisterCustomHandler("SharePriceUpdated", func(msg signalr.Message) {
    // Handle share price updates
    log.Printf("Share price update: %v", msg.Data)
})

client.RegisterCustomHandler("MarketStatusUpdated^^DSE~", func(msg signalr.Message) {
    // Handle market status updates with special characters
    log.Printf("Market status: %v", msg.Data)
})
```

### HTTP Connection Configuration

The client uses a configurable HTTP connection with optimal settings:

```go
// Configurable HTTP connection
conn, err := signalr.NewHTTPConnection(ctx, hubURL, 
    signalr.WithHTTPHeaders(func() http.Header {
        h := make(http.Header)
        h.Set("Authorization", "Bearer " + token)
        h.Set("User-Agent", "Go-SignalR-Client/1.0")
        h.Set("Accept", "application/json")
        h.Set("Content-Type", "application/json")
        return h
    }),
    signalr.WithHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            IdleConnTimeout:       90 * time.Second,
            TLSHandshakeTimeout:   10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        },
    }),
)

// Client with JSON encoding
client, err := signalr.NewClient(ctx,
    signalr.WithConnection(conn),
    signalr.TransferFormat(signalr.TransferFormatText),
    signalr.WithReceiver(receiver))

client.Start()
```

## Connection Management

### Connection Status

```go
// Check connection status
status := client.Status()
switch status {
case signalr.ConnectionStatusConnected:
    log.Println("Connected")
case signalr.ConnectionStatusReconnecting:
    log.Println("Reconnecting...")
case signalr.ConnectionStatusDisconnected:
    log.Println("Disconnected")
}

// Get connection statistics
stats := client.GetConnectionStats()
log.Printf("Status: %v, Attempts: %v, Subscriptions: %v", 
    stats["status"], stats["reconnectAttempts"], stats["subscriptions"])
```

### Token Management

```go
// Refresh token periodically
go func() {
    ticker := time.NewTicker(50 * time.Minute)
    defer ticker.Stop()
    
    for {
        <-ticker.C
        newToken, err := auth.Login(cfg)
        if err == nil {
            client.UpdateToken(newToken)
        }
    }
}()
```

## Message Handling

### Built-in Handlers

The client includes built-in handlers for common message types:
- `SharePriceUpdated`
- `MarketStatusUpdated^^DSE~` (handles special characters)
- Connection events (connect/disconnect/reconnect)
- Error messages

### Custom Handlers

Register custom handlers for specific message types:

```go
client.RegisterCustomHandler("CustomMethod", func(msg signalr.Message) {
    // Process custom message
    data := msg.Data.([]interface{})
    log.Printf("Custom data: %v", data[0])
})
```

### Fallback Handling

All unhandled messages are sent to the main message channel:

```go
processor := signalr.NewMessageProcessor()
for msg := range client.Messages() {
    processor.Process(msg)
}
```

## Configuration

### Client Configuration Options

```go
type ClientConfig struct {
    ConnectionTimeout    time.Duration  // Connection timeout
    ReconnectDelay      time.Duration  // Initial reconnect delay
    MaxReconnectDelay   time.Duration  // Maximum reconnect delay
    MaxReconnectAttempts int           // Maximum reconnect attempts
    MessageBufferSize   int           // Message channel buffer size
    EnableHeartbeat     bool          // Enable heartbeat
    HeartbeatInterval   time.Duration  // Heartbeat interval
    UserAgent          string        // HTTP User-Agent
    HTTPTimeout        time.Duration  // HTTP client timeout
}
```

### Default Configuration

```go
clientCfg := signalr.DefaultClientConfig()
// ConnectionTimeout: 30s
// ReconnectDelay: 2s  
// MaxReconnectDelay: 2m
// MaxReconnectAttempts: 20
// MessageBufferSize: 100
// EnableHeartbeat: true
// HeartbeatInterval: 30s
```

## Error Handling

The client provides comprehensive error handling:

- **Connection Errors**: Automatic reconnection with exponential backoff
- **Authentication Errors**: Token refresh and reconnection
- **Message Errors**: Logged and forwarded to error handlers
- **Network Errors**: Detected via heartbeat and triggers reconnection

## Logging

Comprehensive logging for debugging and monitoring:

```
[_________SignalR_________] Connecting to SignalR hub: wss://example.com/hub
[***********SignalR Receiver***********] Receive method called with method=SharePriceUpdated
[_________SignalR_________] Heartbeat ping successful
[_________SignalR_________] Reconnection attempt #1 after 2s
```

## Best Practices

1. **Always handle the message channel**: Process messages in a separate goroutine
2. **Use custom handlers**: Register specific handlers for known message types  
3. **Monitor connection status**: Check connection health periodically
4. **Implement graceful shutdown**: Close the client properly on application exit
5. **Configure appropriate timeouts**: Set timeouts based on your network conditions
6. **Handle token refresh**: Implement automatic token refresh for long-running applications

## Thread Safety

The SignalR client is designed to be thread-safe:
- All connection state changes are protected by mutexes
- Message channels are safe for concurrent access
- Handler registration is thread-safe
- Multiple goroutines can safely call client methods

## Requirements

- Go 1.19+
- github.com/philippseith/signalr
- github.com/gorilla/websocket
- github.com/andybalholm/brotli

## License

This project is licensed under the MIT License.

## Handling Method Names with Special Characters

The SignalR protocol can send method names with special characters (like `MarketStatusUpdated^^DSE~`) that are not valid Go identifiers. Our solution handles this by:

1. **Embedding `signalr.Hub`**: Our `MessageReceiver` embeds `signalr.Hub` which provides base receiver functionality
2. **Universal `Receive` Method**: All server-to-client method calls are routed through the `Receive` method
3. **Method Name Mapping**: We handle special characters by processing all method names in the `Receive` method

### Example: Handling Special Characters

```go
// The server calls "MarketStatusUpdated^^DSE~" but Go can't have such method names
// Our Receive method handles this automatically:

func (r *MessageReceiver) Receive(method string, args ...interface{}) {
    r.logger.Printf("Received method: %s with %d arguments", method, len(args))
    
    // Handle any method name, including those with special characters
    switch strings.ToLower(method) {
    case "marketstatusupdated^^dse~":
        // Process market status with special characters
        r.processMarketStatus(args...)
    case "sharepriceupdated":
        // Process share price updates
        r.processSharePrice(args...)
    default:
        // Handle unknown methods
        r.processUnknownMethod(method, args...)
    }
}
```

### Server Method Call Flow

1. Server sends: `MarketStatusUpdated^^DSE~` with data
2. SignalR library tries to find method `MarketStatusUpdated^^DSE~` on receiver
3. Method doesn't exist (invalid Go identifier), so `Receive` method is called
4. Our `Receive` method processes the call and routes it appropriately

This approach ensures **ALL** server-to-client method calls are captured and processed, regardless of special characters in method names.
