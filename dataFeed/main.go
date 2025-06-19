package main

import (
	"log"
)

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	token, err := Login(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	if err := ConnectAndLogSignalR(cfg, token); err != nil {
		log.Fatalf("SignalR connection failed: %v", err)
	}
}
