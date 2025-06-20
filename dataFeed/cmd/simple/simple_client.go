package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/philippseith/signalr"

	"datafeed/pkg/auth"
	"datafeed/pkg/config"
)

// SimpleReceiver implements a basic SignalR receiver based on documentation
type SimpleReceiver struct {
	// Embed Hub to get base functionality
	signalr.Hub
}

// Example method handlers - these will be called by the server
func (r *SimpleReceiver) MarketUpdate(message string) {
	log.Printf("📈 Market Update received: %s", message)
}

func (r *SimpleReceiver) SharePriceUpdate(data string) {
	log.Printf("💰 Share Price Update: %s", data)
}

// Catch-all method for any other server calls
func (r *SimpleReceiver) OnConnected(connectionID string) {
	log.Printf("🔗 Connected with ID: %s", connectionID)
}

func (r *SimpleReceiver) OnDisconnected() {
	log.Println("❌ Disconnected from hub")
}

func main() {
	log.Println("🧪 Starting Simple SignalR Test Client (Documentation-based)")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Authenticate and get token
	log.Println("🔐 Authenticating...")
	token, err := auth.Login(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	log.Println("✅ Authentication successful")

	// Create context for the connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create HTTP connection as per documentation
	log.Printf("🔌 Creating HTTP connection to: %s", cfg.SignalRURL)
	conn, err := signalr.NewHTTPConnection(ctx, cfg.SignalRURL,
		signalr.WithHTTPHeaders(func() http.Header {
			headers := make(http.Header)
			headers.Set("Authorization", "Bearer "+token)
			headers.Set("User-Agent", "Go-SignalR-Simple-Test/1.0")
			return headers
		}),
		signalr.WithHTTPClient(&http.Client{
			Timeout: 30 * time.Second,
		}),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create HTTP connection: %v", err)
	}
	log.Println("✅ HTTP connection created")

	// Create simple receiver
	receiver := &SimpleReceiver{}

	// Create SignalR client as per documentation
	log.Println("🔄 Creating SignalR client...")
	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.TransferFormat(signalr.TransferFormatText), // JSON over text
		signalr.WithReceiver(receiver),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create SignalR client: %v", err)
	}
	log.Println("✅ SignalR client created")

	// Start the client as per documentation
	log.Println("🚀 Starting SignalR client...")
	client.Start()
	log.Println("✅ SignalR client started")

	// Wait a moment for connection to establish
	time.Sleep(3 * time.Second)

	// Test basic subscription
	log.Println("📡 Testing basic subscription...")
	go func() {
		// Simple subscription test
		result := <-client.Send("SubscribeToMarketStatusUpdatedEvent", "DSE")
		if result == nil {
			log.Println("✅ Subscription successful")
		} else {
			log.Printf("⚠️ Subscription result: %v", result)
		}
	}()

	// Test ping functionality
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("📶 Sending ping...")
				go func() {
					result := <-client.Send("ping")
					if result == nil {
						log.Println("🏓 Ping successful")
					} else {
						log.Printf("⚠️ Ping failed: %v", result)
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Simple connection monitoring
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("💓 Connection alive")
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("✅ Simple SignalR client running")
	log.Println("📋 This client tests:")
	log.Println("   - Basic HTTP connection with auth headers")
	log.Println("   - Simple SignalR client creation")
	log.Println("   - Text transfer format (JSON)")
	log.Println("   - Basic receiver with method handlers")
	log.Println("   - Simple subscription")
	log.Println("   - Ping functionality")
	log.Println("")
	log.Println("🎯 Expected server methods that will be handled:")
	log.Println("   - MarketUpdate(message)")
	log.Println("   - SharePriceUpdate(data)")
	log.Println("   - OnConnected(connectionID)")
	log.Println("   - OnDisconnected()")
	log.Println("")
	log.Println("Press Ctrl+C to exit...")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down simple client...")
	client.Stop()
	log.Println("✅ Simple client stopped")
}
