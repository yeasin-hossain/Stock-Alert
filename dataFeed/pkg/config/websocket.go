package config

// WebSocketConfig contains configuration for the WebSocket connection
type WebSocketConfig struct {
	URL      string            `yaml:"url"`      // WebSocket server URL
	Headers  map[string]string `yaml:"headers"`  // Additional headers to include in the connection
	Protocol string            `yaml:"protocol"` // WebSocket subprotocol (if any)
}
