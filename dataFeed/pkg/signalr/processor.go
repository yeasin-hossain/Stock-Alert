package signalr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/andybalholm/brotli"
)

// MessageProcessor handles processing and parsing of SignalR messages
type MessageProcessor struct {
	logger *log.Logger
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		logger: log.New(os.Stdout, "[MsgProcessor] ", log.LstdFlags),
	}
}

// Process processes a SignalR message
func (p *MessageProcessor) Process(msg Message) {
	p.logger.Printf("Processing message: method=%s with data type: %T", msg.Method, msg.Data)

	// Log more details about the data
	if args, ok := msg.Data.([]interface{}); ok {
		p.logger.Printf("Message contains %d arguments", len(args))
		for i, arg := range args {
			p.logger.Printf("Arg[%d] type: %T, value: %v", i, arg, arg)
		}
	}

	switch msg.Method {
	case "SharePriceUpdated", "sharePriceUpdated":
		p.logger.Printf("Handling SharePriceUpdated event")
		p.processSharePriceUpdate(msg.Data)
	case "MarketStatusUpdated^^DSE~", "marketStatusUpdated^^dse~":
		p.logger.Printf("Handling MarketStatusUpdated event")
		p.processMarketStatusUpdate(msg.Data)
	case "Ping":
		p.logger.Printf("Handling ping message (type 6)")
		p.processPing()
	default:
		p.logger.Printf("Unknown method received: %s", msg.Method)
	}
}

// processSharePriceUpdate handles share price update messages
func (p *MessageProcessor) processSharePriceUpdate(data interface{}) {
	p.logger.Printf("Processing share price update with data type: %T", data)

	// Try different ways to extract the data based on the structure
	var dataStr string

	// Case 1: Direct string - this is what we should get now with our custom handler
	if str, ok := data.(string); ok {
		p.logger.Printf("Found direct string data of length: %d", len(str))
		dataStr = str
	} else if args, ok := data.([]interface{}); ok && len(args) > 0 {
		// Case 2: Array of arguments, take the first one
		p.logger.Printf("Found array data with %d elements", len(args))
		switch v := args[0].(type) {
		case string:
			dataStr = v
		case []byte:
			dataStr = string(v)
		case map[string]interface{}:
			// Try to extract from JSON object
			if jsonData, ok := v["data"].(string); ok {
				dataStr = jsonData
			}
		default:
			// For unexpected types, try JSON serialization
			jsonBytes, _ := json.Marshal(args[0])
			dataStr = string(jsonBytes)
		}
	} else {
		// Case 3: Unknown format, try JSON serialization to see the structure
		jsonBytes, _ := json.Marshal(data)
		dataStr = string(jsonBytes)
	}

	// If we have a data string, try to decompress and parse it
	if dataStr != "" {
		p.processDataString(dataStr)
	}
}

// processDataString tries multiple methods to decompress and parse data
func (p *MessageProcessor) processDataString(dataStr string) {
	// Try to parse as JSON first
	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &jsonObj); err == nil {
		// Check for data field in JSON
		if dataField, ok := jsonObj["data"].(string); ok {
			p.decompressAndProcess(dataField)
		}
		return
	}

	// If not JSON, try direct decompression
	p.decompressAndProcess(dataStr)
}

// decompressAndProcess attempts to decompress data and process it
func (p *MessageProcessor) decompressAndProcess(data string) {
	// Try multiple decompression strategies

	// Strategy 1: Direct decompression
	if decompressed, err := p.decompressBrotli(data); err == nil {
		p.logger.Printf("Decompression succeeded, processing data...")
		p.processDecompressedData(decompressed)
		return
	}

	// Strategy 2: Base64 decode first, then decompress
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		if decompressed, err := p.decompressBrotliBytes(decoded); err == nil {
			p.logger.Printf("Base64+Brotli decompression succeeded, processing data...")
			p.processDecompressedData(string(decompressed))
			return
		}
	}

	// Strategy 3: Check if data is already in the expected format (not compressed)
	if strings.Contains(data, "~") {
		p.processDecompressedData(data)
		return
	}
}

// decompressBrotli decompresses Brotli-compressed data
func (p *MessageProcessor) decompressBrotli(input string) (string, error) {
	decompressed, err := p.decompressBrotliBytes([]byte(input))
	if err != nil {
		return "", err
	}
	return string(decompressed), nil
}

// decompressBrotliBytes decompresses Brotli-compressed bytes
func (p *MessageProcessor) decompressBrotliBytes(input []byte) ([]byte, error) {
	br := brotli.NewReader(bytes.NewReader(input))
	decompressed, err := io.ReadAll(br)
	if err != nil {
		return nil, fmt.Errorf("brotli decompression error: %w", err)
	}
	return decompressed, nil
}

// processDecompressedData processes the final decompressed data
func (p *MessageProcessor) processDecompressedData(data string) {
	// Parse delimited data
	if strings.Contains(data, "~") {
		fields := strings.Split(data, "~")
		p.logger.Printf("Share price data received: %d fields", len(fields))

		// Log only a few sample fields to avoid flooding the console
		if len(fields) > 5 {
			p.logger.Printf("First few fields: [%s, %s, %s, ...]",
				fields[0], fields[1], fields[2])
		}
	} else {
		// Try to parse as JSON
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(data), &jsonObj); err == nil {
			p.logger.Printf("Processed JSON data successfully")
		}
	}
}

// processMarketStatusUpdate handles market status update messages
func (p *MessageProcessor) processMarketStatusUpdate(data interface{}) {
	p.logger.Printf("Processing market status update with data type: %T", data)

	var dataStr string

	// Process string data
	if str, ok := data.(string); ok {
		p.logger.Printf("Found direct string market status data of length: %d", len(str))
		dataStr = str
	} else if args, ok := data.([]interface{}); ok && len(args) > 0 {
		// Process array data
		p.logger.Printf("Found array data with %d elements", len(args))
		if str, ok := args[0].(string); ok {
			dataStr = str
		} else {
			// Try JSON conversion for other types
			jsonBytes, _ := json.Marshal(args[0])
			dataStr = string(jsonBytes)
		}
	}

	if dataStr != "" {
		// Log the market status data
		if len(dataStr) < 200 {
			p.logger.Printf("Market status data: %s", dataStr)
		} else {
			p.logger.Printf("Market status data (truncated): %s...", dataStr[:200])
		}

		// Parse the data if it's in a known format
		// For example, if it's JSON:
		var marketStatus map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &marketStatus); err == nil {
			p.logger.Printf("Parsed market status: %v", marketStatus)
		}
	}
}

// processPing handles ping messages from the server (type 6)
func (p *MessageProcessor) processPing() {
	p.logger.Printf("Processing ping message from server")

	// In SignalR protocol, ping messages (type 6) don't require a response from the client
	// The server sends these as a keep-alive mechanism
	// Simply log the ping for tracking connection health

	p.logger.Printf("Ping received - connection is active")
}

// Helper function to truncate long strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
