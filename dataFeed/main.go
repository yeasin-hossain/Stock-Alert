package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"datafeed/pkg/auth"
	"datafeed/pkg/config"
	"datafeed/pkg/signalr"
)

func main() {
	log.Println("Starting data feed service...")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Authenticate and get token
	log.Println("Authenticating...")
	token, err := auth.Login(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	log.Println("Authentication successful")

	// Create and connect SignalR client with enhanced error handling
	client := signalr.NewClient(cfg, token)

	// Register custom handler for special character method names
	client.RegisterCustomHandler("MarketStatusUpdated^^DSE~", func(msg signalr.Message) {
		log.Printf("🎯 SPECIAL CHAR METHOD: MarketStatusUpdated^^DSE~ received: %v", msg.Data)
	})

	// Add handlers for connection events
	client.RegisterCustomHandler("ConnectionEvent", func(msg signalr.Message) {
		log.Printf("🔗 CONNECTION EVENT: %v", msg.Data)
	})

	client.RegisterCustomHandler("Error", func(msg signalr.Message) {
		log.Printf("❌ SERVER ERROR: %v", msg.Data)
	})

	log.Println("Connecting to SignalR...")
	if err := client.Connect(); err != nil {
		log.Printf("❌ SignalR connection failed: %v", err)
		log.Println("Retrying connection in 5 seconds...")
		time.Sleep(5 * time.Second)

		// Try once more with fresh token
		log.Println("Getting fresh token for retry...")
		freshToken, authErr := auth.Login(cfg)
		if authErr != nil {
			log.Fatalf("Failed to get fresh token: %v", authErr)
		}

		client.UpdateToken(freshToken)
		if err := client.Connect(); err != nil {
			log.Fatalf("SignalR connection failed on retry: %v", err)
		}
	}

	log.Println("✅ SignalR connected successfully")

	// Create a message processor
	processor := signalr.NewMessageProcessor()

	// Process messages in a goroutine
	go func() {
		log.Println("Starting message processor...")
		for msg := range client.Messages() {
			log.Printf("📨 Received message: Method=%s", msg.Method)
			processor.Process(msg)
		}
		log.Println("Message processor stopped")
	}()

	// Monitor connection status and statistics with enhanced logging
	go func() {
		ticker := time.NewTicker(15 * time.Second) // More frequent monitoring
		defer ticker.Stop()

		for {
			<-ticker.C
			stats := client.GetConnectionStats()
			status := stats["status"]
			attempts := stats["reconnectAttempts"]
			subscriptions := stats["subscriptions"]

			// More detailed status logging
			switch status {
			case signalr.ConnectionStatusConnected:
				log.Printf("� CONNECTED - Attempts: %v, Subscriptions: %v", attempts, subscriptions)
			case signalr.ConnectionStatusReconnecting:
				log.Printf("🟡 RECONNECTING - Attempt: %v, Subscriptions: %v", attempts, subscriptions)
			case signalr.ConnectionStatusDisconnected:
				log.Printf("🔴 DISCONNECTED - Last attempts: %v, Subscriptions: %v", attempts, subscriptions)
				if lastErr := client.LastError(); lastErr != nil {
					log.Printf("   Last error: %v", lastErr)
				}
			case signalr.ConnectionStatusConnecting:
				log.Printf("🟡 CONNECTING - Attempts: %v", attempts)
			default:
				log.Printf("❓ UNKNOWN STATUS: %v - Attempts: %v, Subscriptions: %v", status, attempts, subscriptions)
			}
		}
	}()

	// Setup token refresh
	go refreshTokenPeriodically(cfg, client)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	log.Println("Application running. Press Ctrl+C to exit.")
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	client.Close()
	log.Println("Application terminated")
}

// refreshTokenPeriodically refreshes the authentication token periodically
func refreshTokenPeriodically(cfg *config.Config, client *signalr.Client) {
	// Refresh token every 50 minutes (assuming a 1-hour token lifetime)
	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		log.Println("Refreshing authentication token...")

		token, err := auth.Login(cfg)
		if err != nil {
			log.Printf("WARNING: Token refresh failed: %v", err)
			continue
		}

		// Update the token in the client
		// You'll need to add this method to the SignalR client
		if err := client.UpdateToken(token); err != nil {
			log.Printf("WARNING: Failed to update token: %v", err)
		} else {
			log.Println("Token refreshed successfully")
		}
	}
}
