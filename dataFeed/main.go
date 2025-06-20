package main

import (
	"log"
	"time"

	"datafeed/pkg/auth"
	"datafeed/pkg/config"
	"datafeed/pkg/signalr"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Authenticate and get token
	token, err := auth.Login(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Create and connect SignalR client
	client := signalr.NewClient(cfg, token)
	log.Println("Connecting to SignalR and subscribing to events...")
	if err := client.Connect(); err != nil {
		log.Fatalf("SignalR connection failed: %v", err)
	}

	// Note: Subscriptions are now handled automatically in the Connect method

	// Create a message processor
	processor := signalr.NewMessageProcessor()

	// Process messages
	go func() {
		for msg := range client.Messages() {
			processor.Process(msg)
		}
	}()

	// Keep the application running
	log.Println("Application running. Press Ctrl+C to exit.")
	for {
		time.Sleep(1 * time.Hour)
	}
}
