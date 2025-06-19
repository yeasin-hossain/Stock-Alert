# DataFeed Microservice

This Go microservice logs in to a remote service, connects to a SignalR hub, and logs all received messages.

## Features
- Authenticates to a remote service via HTTP POST (username/password)
- Connects to a SignalR hub using the authentication token/cookie
- Logs all received SignalR messages
- Configuration via `config.yaml`

## Usage
1. Set up `config.yaml` with your endpoints and credentials.
2. Run the service:
   ```sh
   go run main.go
   ```

## Dependencies
- [github.com/philippseith/signalr](https://github.com/philippseith/signalr)
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2)

## Project Structure
- `main.go` - Entry point
- `config.go` - Loads configuration
- `auth.go` - Handles authentication
- `signalr_client.go` - Handles SignalR connection and message logging
