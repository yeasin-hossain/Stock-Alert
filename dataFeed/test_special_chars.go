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

func testSpecialChars() {
	log.Println("Testing SignalR client with special character method names...")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Authenticate
	token, err := auth.Login(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Create SignalR client
	client := signalr.NewClient(cfg, token)

	// Register a custom handler for the special character method
	client.RegisterCustomHandler("MarketStatusUpdated^^DSE~", func(msg signalr.Message) {
		log.Printf("🎯 CAPTURED: MarketStatusUpdated^^DSE~ message: %v", msg.Data)
	})

	// Connect
	if err := client.Connect(); err != nil {
		log.Fatalf("SignalR connection failed: %v", err)
	}
	defer client.Close()

	// Create a message processor
	processor := signalr.NewMessageProcessor()

	// Process all messages
	go func() {
		log.Println("Starting message processor...")
		for msg := range client.Messages() {
			log.Printf("📨 Received message: Method=%s", msg.Method)
			processor.Process(msg)
		}
	}()

	// Monitor connection and stats
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			stats := client.GetConnectionStats()
			log.Printf("📊 Connection Stats: Status=%v, Attempts=%v, Subscriptions=%v",
				stats["status"], stats["reconnectAttempts"], stats["subscriptions"])
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✅ SignalR client running and ready to receive messages with special characters!")
	log.Println("   Watching for: MarketStatusUpdated^^DSE~, SharePriceUpdated, and other methods")
	log.Println("   Press Ctrl+C to exit")

	<-sigChan
	log.Println("🛑 Shutting down...")
}
