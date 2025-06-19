package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/philippseith/signalr"
)

// ConnectAndLogSignalR connects to the SignalR hub and logs all received messages
func ConnectAndLogSignalR(cfg *Config, token string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up HTTP connection with timeout for negotiation
	creationCtx, creationCancel := context.WithTimeout(ctx, 2*time.Second)
	defer creationCancel()

	conn, err := signalr.NewHTTPConnection(creationCtx, cfg.SignalRURL, signalr.WithHTTPHeaders(func() http.Header {
		h := make(http.Header)
		h.Set("Authorization", "Bearer "+token)
		return h
	}))
	if err != nil {
		return err
	}

	receiver := &SignalRReceiver{}
	client, err := signalr.NewClient(ctx,
		signalr.WithConnection(conn),
		signalr.WithReceiver(receiver),
	)
	if err != nil {
		return err
	}

	client.Start() // Start does not return a value
	log.Println("SignalR client started. Listening for messages...")

	// Subscribe to the 'SubscribeToSharePriceUpdatedEvent' channel/event with arguments
	go func() {
		args := []interface{}{"500$1$$Asc", "DSE", nil, "", "", "", []interface{}{}, "", nil, false, nil}
		result := <-client.Invoke("SubscribeToSharePriceUpdatedEvent", args...)
		log.Printf("Subscribed to SubscribeToSharePriceUpdatedEvent, result: %v", result)
	}()

	<-ctx.Done()
	return nil
}

// SignalRReceiver implements signalr.Receiver for handling server callbacks
type SignalRReceiver struct{}

func (r *SignalRReceiver) Receive(method string, args ...interface{}) {
	if method == "SharePriceUpdated" && len(args) > 0 {
		if str, ok := args[0].(string); ok {
			log.Printf("SharePriceUpdated received: %s", str)
			// Example: parse the string if needed
			// fields := strings.Split(str, "~")
			// log.Printf("Parsed fields: %v", fields)
		} else {
			log.Printf("SharePriceUpdated received (non-string): %v", args[0])
		}
	} else {
		log.Printf("Received SignalR message: method=%s args=%v", method, args)
	}
}
