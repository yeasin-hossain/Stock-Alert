# SignalR Go Client Test Suite

A comprehensive SignalR client implementation in Go with multiple test clients for different use cases.

## Quick Start

1. **Configure** - Edit `config.yaml` with your SignalR server details
2. **Test Connection** - Run `./run.sh basic` to verify basic connectivity  
3. **Run Main Client** - Use `./run.sh main` for production data feed

## Available Clients

| Client | Command | Purpose |
|--------|---------|---------|
| **Basic Test** | `./run.sh basic` | Minimal connectivity testing |
| **Simple Client** | `./run.sh simple` | Documentation-based implementation |
| **Special Chars Test** | `./run.sh test` | Tests special character method names |
| **Main Application** | `./run.sh main` | Production-ready robust client |

## Configuration

Create or edit `config.yaml`:

```yaml
signalr_url: "wss://your-server.com/signalrhub"
api_url: "https://your-server.com/api"  
username: "your-username"
password: "your-password"
```

## Testing Workflow

```bash
# 1. Test basic connectivity first
./run.sh basic

# 2. If basic works, test standard patterns  
./run.sh simple

# 3. Test special character handling (if needed)
./run.sh test

# 4. Run production client
./run.sh main

# Build all clients
./run.sh build
```

## Features

### ✅ Basic Test Client
- Fundamental connectivity testing
- Authentication validation
- Simple method invocation (ping, echo)
- Connection lifecycle monitoring

### ✅ Simple Client  
- Standard SignalR receiver patterns
- Predefined method handlers
- Text transfer format (JSON)
- Basic subscription management

### ✅ Special Characters Test
- Universal message receiver
- Special character method names (`MarketStatusUpdated^^DSE~`)
- Robust message parsing
- Edge case handling

### ✅ Main Application
- Production-ready robust client
- Automatic reconnection
- Subscription reapplication  
- Comprehensive error handling
- Modular architecture

## Troubleshooting

**Connection Issues**: Start with `./run.sh basic`
**Authentication**: Check credentials in `config.yaml`
**Method Handling**: Use `./run.sh simple` for standard patterns
**Special Characters**: Use `./run.sh test` for complex method names

See `TEST_CLIENTS.md` for detailed documentation and `TROUBLESHOOTING.md` for common issues.

## Architecture

```
dataFeed/
├── main.go                    # Production client
├── cmd/
│   ├── basic/basic_client.go  # Basic connectivity test
│   ├── simple/simple_client.go # Documentation-based client  
│   └── test/test_special_chars.go # Special character testing
├── pkg/
│   ├── auth/                  # Authentication module
│   ├── config/                # Configuration management
│   └── signalr/              # SignalR client implementation
└── run.sh                     # Easy run/build script
```

## Dependencies
- [github.com/philippseith/signalr](https://github.com/philippseith/signalr)
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2)

## Requirements

- Go 1.19+
- SignalR server with authentication
- Network access to SignalR endpoint

---

**Note**: All clients require proper configuration and network access to your SignalR server.
