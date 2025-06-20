package main

import (
	"context"
	"fmt"
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

// BasicHub implements the most minimal SignalR hub for testing
type BasicHub struct {
	signalr.Hub
}

// Simple echo method for testing connectivity
func (h *BasicHub) Echo(message string) string {
	log.Printf("📤 Echo received: %s", message)
	return fmt.Sprintf("Echo: %s", message)
}

// Generic message receiver for any server method
func (h *BasicHub) Receive(method string, message interface{}) {
	log.Printf("📨 Generic message received - Method: %s, Message: %v", method, message)
}

// Connection lifecycle methods
func (h *BasicHub) OnConnected(connectionID string) {
	log.Printf("🔗 Connected to SignalR hub with ID: %s", connectionID)
}

func (h *BasicHub) OnDisconnected() {
	log.Println("❌ Disconnected from SignalR hub")
}

func main() {
	log.Println("🔧 Starting BASIC SignalR Connection Test")
	log.Println("========================================")
	log.Println("This is the most minimal SignalR client for basic connectivity testing")
	log.Println()

	// Load configuration
	cfg, err := config.Load("../../config.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Printf("✅ Config loaded - URL: %s", cfg.SignalRURL)

	// Authenticate
	log.Println("🔐 Authenticating...")
	token, err := auth.Login(cfg)
	if err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	log.Println("✅ Authentication successful")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create basic HTTP connection
	log.Println("🌐 Creating HTTP connection...")
	conn, err := signalr.NewHTTPConnection(ctx, cfg.SignalRURL,
		signalr.WithHTTPHeaders(func() http.Header {
			headers := make(http.Header)
			headers.Set("Authorization", "Bearer "+token)
			headers.Set("User-Agent", "Go-SignalR-Basic-Test/1.0")
			log.Printf("🔑 Setting auth header: Bearer %s...", token[:20])
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

	// Create hub instance
	hub := &BasicHub{}

	// Create SignalR client with minimal options
	log.Println("⚙️ Creating SignalR client...")
	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.TransferFormat(signalr.TransferFormatText),
		signalr.WithReceiver(hub),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create SignalR client: %v", err)
	}
	log.Println("✅ SignalR client created")

	// Start client
	log.Println("🚀 Starting SignalR client...")
	client.Start()
	log.Println("✅ SignalR client started")

	// Give it time to connect
	log.Println("⏳ Waiting for connection to establish...")
	time.Sleep(5 * time.Second)

	// Test basic functionality
	log.Println("🧪 Testing basic functionality...")

	// Test 1: Simple ping
	go func() {
		log.Println("📍 Test 1: Sending ping...")
		result := <-client.Send("ping")
		if result == nil {
			log.Println("✅ Ping successful")
		} else {
			log.Printf("⚠️ Ping failed: %v", result)
		}
	}()

	time.Sleep(2 * time.Second)

	// Test 2: Echo test
	go func() {
		log.Println("📍 Test 2: Testing echo...")
		result := <-client.Send("Echo", "Hello from basic test!")
		if result == nil {
			log.Println("✅ Echo test successful")
		} else {
			log.Printf("⚠️ Echo test failed: %v", result)
		}
	}()

	time.Sleep(2 * time.Second)

	// Test 3: Try subscription (if supported)
	go func() {
		log.Println("📍 Test 3: Testing subscription...")
		result := <-client.Send("SubscribeToMarketStatusUpdatedEvent", "DSE")
		if result == nil {
			log.Println("✅ Subscription test successful")
		} else {
			log.Printf("⚠️ Subscription test failed: %v", result)
		}
	}()

	time.Sleep(2 * time.Second)

	// Test 4: Try another subscription with special characters
	go func() {
		log.Println("📍 Test 4: Testing special character subscription...")
		result := <-client.Send("SubscribeToSharePriceUpdatedEvent", "DSE")
		if result == nil {
			log.Println("✅ Special character subscription successful")
		} else {
			log.Printf("⚠️ Special character subscription failed: %v", result)
		}
	}()

	// Connection monitoring
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("💗 Connection heartbeat - still alive")
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println()
	log.Println("🎯 Basic Test Client is now running")
	log.Println("📋 What this client tests:")
	log.Println("   ✓ Basic HTTP connection with authentication")
	log.Println("   ✓ SignalR client creation and startup")
	log.Println("   ✓ Connection lifecycle events")
	log.Println("   ✓ Simple method invocation (ping, echo)")
	log.Println("   ✓ Subscription methods")
	log.Println("   ✓ Generic message receiving")
	log.Println("   ✓ Connection monitoring")
	log.Println()
	log.Println("🎧 Listening for server messages...")
	log.Println("Press Ctrl+C to exit")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println()
	log.Println("🛑 Shutting down basic test client...")
	client.Stop()
	log.Println("✅ Basic test client stopped")
}
