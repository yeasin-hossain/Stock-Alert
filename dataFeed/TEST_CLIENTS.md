# SignalR Test Clients Documentation

This document describes the different SignalR test clients available in this project and when to use each one.

## Overview

We have created multiple test clients to help diagnose and validate SignalR connectivity at different levels of complexity:

1. **Basic Test Client** (`cmd/basic/basic_client.go`) - Minimal connectivity testing
2. **Simple Client** (`cmd/simple/simple_client.go`) - Documentation-based implementation
3. **Special Characters Test** (`cmd/test/test_special_chars.go`) - Tests special character method names
4. **Main Application** (`main.go`) - Production-ready robust client

## Test Clients

### 1. Basic Test Client (`./run.sh basic`)

**Purpose**: Fundamental connectivity testing with minimal functionality

**When to use**:
- First-time setup validation
- Debugging basic connection issues
- Verifying authentication works
- Testing server availability

**Features**:
- ✅ HTTP connection with authentication
- ✅ Basic SignalR client creation
- ✅ Connection lifecycle events (connect/disconnect)
- ✅ Simple method invocation (ping, echo)
- ✅ Basic subscription testing
- ✅ Generic message receiving
- ✅ Connection heartbeat monitoring

**Expected output**:
```
🔧 Starting BASIC SignalR Connection Test
✅ Config loaded - URL: wss://example.com/signalrhub
✅ Authentication successful
✅ HTTP connection created
✅ SignalR client created
✅ SignalR client started
📍 Test 1: Sending ping...
✅ Ping successful
📍 Test 2: Testing echo...
✅ Echo test successful
```

### 2. Simple Client (`./run.sh simple`)

**Purpose**: Documentation-based SignalR implementation following standard patterns

**When to use**:
- Learning SignalR implementation patterns
- Validating standard SignalR features
- Reference implementation for basic use cases
- Debugging message handling

**Features**:
- ✅ Standard SignalR receiver pattern
- ✅ Predefined method handlers
- ✅ Text transfer format (JSON)
- ✅ Basic subscription management
- ✅ Ping functionality
- ✅ Connection monitoring

**Method handlers**:
- `MarketUpdate(message string)`
- `SharePriceUpdate(data string)`
- `OnConnected(connectionID string)`
- `OnDisconnected()`

### 3. Special Characters Test (`./run.sh test`)

**Purpose**: Testing method names with special characters and edge cases

**When to use**:
- Validating special character handling
- Testing method names with symbols (^, ~, etc.)
- Debugging method name parsing issues
- Verifying robust message handling

**Features**:
- ✅ Universal `Receive` method for all server messages
- ✅ Special character method name handling
- ✅ Robust message parsing
- ✅ Comprehensive logging

**Tested method names**:
- `MarketStatusUpdated^^DSE~`
- `SharePriceUpdated`
- Methods with special characters

### 4. Main Application (`./run.sh main`)

**Purpose**: Production-ready robust SignalR client

**When to use**:
- Production deployment
- Long-running data feed operations
- Maximum reliability requirements
- Full feature implementation

**Features**:
- ✅ Modular architecture
- ✅ Automatic reconnection
- ✅ Subscription reapplication
- ✅ Comprehensive error handling
- ✅ Connection monitoring
- ✅ Message processing pipeline
- ✅ Configurable retry logic

## Troubleshooting Guide

### Connection Issues

1. **Start with Basic Test** (`./run.sh basic`)
   - Validates fundamental connectivity
   - Tests authentication
   - Confirms server availability

2. **Authentication Failures**
   - Check `config.yaml` credentials
   - Verify API endpoints
   - Check token validity

3. **Message Handling Issues**
   - Use Simple Client for standard patterns
   - Use Special Characters Test for complex method names
   - Check server method signatures

### Testing Workflow

```bash
# 1. Test basic connectivity
./run.sh basic

# 2. If basic works, test standard patterns
./run.sh simple

# 3. If you have special character methods
./run.sh test

# 4. Run production client
./run.sh main
```

### Build and Verify

```bash
# Build all clients
./run.sh build

# This creates executables:
# - datafeed (main application)
# - cmd/test/test_special_chars
# - cmd/simple/simple_client  
# - cmd/basic/basic_client
```

## Configuration

All clients use the same `config.yaml` file:

```yaml
signalr_url: "wss://your-server.com/signalrhub"
api_url: "https://your-server.com/api"
username: "your-username"
password: "your-password"
```

## Expected Server Methods

Based on the test clients, your SignalR server should implement:

### Standard Methods
- `ping` - Connection testing
- `Echo(string)` - Echo testing
- `SubscribeToMarketStatusUpdatedEvent(string)`
- `SubscribeToSharePriceUpdatedEvent(string)`

### Server-to-Client Methods
- `MarketUpdate(string)`
- `SharePriceUpdate(string)`
- `MarketStatusUpdated^^DSE~(data)` - Special character example
- `SharePriceUpdated(data)`

## Logging and Debugging

### Log Levels
- 🔧 System operations
- ✅ Success operations
- ❌ Errors
- ⚠️ Warnings
- 📍 Test operations
- 📨 Message reception
- 💗 Connection heartbeat

### Debug Mode
```bash
./run.sh debug  # Enables detailed logging to file
```

## Performance Considerations

### Basic Test Client
- **Memory**: ~5MB
- **CPU**: Minimal
- **Network**: Low (heartbeat only)
- **Use case**: Quick connectivity tests

### Simple Client
- **Memory**: ~8MB
- **CPU**: Low
- **Network**: Medium (subscriptions + heartbeat)
- **Use case**: Standard SignalR operations

### Main Application
- **Memory**: ~15MB
- **CPU**: Medium (message processing)
- **Network**: High (full data feed)
- **Use case**: Production data processing

## Best Practices

1. **Start Simple**: Always begin with the basic test client
2. **Incremental Testing**: Progress from basic → simple → special chars → main
3. **Monitor Logs**: Each client provides detailed logging
4. **Check Configuration**: Ensure `config.yaml` is properly configured
5. **Network Considerations**: Test on the same network as production when possible

## Support

If you encounter issues:

1. Check the logs from each test client
2. Verify network connectivity to the SignalR server
3. Confirm authentication credentials
4. Test with the basic client first
5. Refer to `TROUBLESHOOTING.md` for common issues

---

**Note**: This documentation assumes you have a SignalR server running and accessible. All test clients require proper configuration in `config.yaml` and network access to the SignalR server.
