# SignalR Connection Troubleshooting Guide

## Common Connection Issues

### Issue: "Connection closed with an error" immediately after subscription

**Symptoms:**
```
[SignalR] Subscription result for SubscribeToMarketStatusUpdatedEvent: <nil>
level=debug message="{\"type\":7,\"error\":\"Connection closed with an error.\",\"allowReconnect\":true}"
```

**Possible Causes & Solutions:**

#### 1. Authentication Issues
- **Problem**: Token expired or invalid
- **Solution**: 
  ```bash
  # Check token validity in config.yaml
  # Verify username/password are correct
  # Run with debug mode to see detailed logs
  ./run.sh debug
  ```

#### 2. Subscription Parameter Issues
- **Problem**: Server doesn't accept "DSE" parameter
- **Solution**: Try different parameters or check server documentation

#### 3. Server-side Validation
- **Problem**: Server rejects the connection after subscription
- **Solution**: Check if your account has permission to subscribe to market data

#### 4. Network/Firewall Issues
- **Problem**: Connection drops due to network
- **Solution**: Check firewall settings and network stability

## Debugging Steps

### 1. Enable Debug Logging
```bash
./run.sh debug
```
This will log all output to `datafeed.log`

### 2. Check Connection Status
The application now shows detailed connection status:
- 🟢 CONNECTED - All good
- 🟡 RECONNECTING - Automatic retry in progress  
- 🔴 DISCONNECTED - Check logs for errors
- ❓ UNKNOWN STATUS - Unexpected state

### 3. Monitor Reconnection Attempts
- Application automatically retries with exponential backoff
- Max 20 attempts before giving up
- Fresh token fetched on authentication errors

### 4. Verify Configuration
Check your `config.yaml`:
```yaml
login_url: "https://your-server.com/api/login"
signalr_url: "wss://your-server.com/signalr/hub"
username: "your-username"
password: "your-password"
```

### 5. Test Different Scenarios
```bash
# Test main application
./run.sh main

# Test with special character handling
./run.sh test

# Build and run standalone
./run.sh build
./datafeed
```

## Enhanced Features

### Automatic Recovery
- ✅ **Auto-reconnection** with exponential backoff
- ✅ **Token refresh** every 50 minutes
- ✅ **Subscription retry** up to 3 attempts
- ✅ **Connection health monitoring** every 15 seconds

### Special Character Support
- ✅ **Method name handling** for `MarketStatusUpdated^^DSE~`
- ✅ **Universal receiver** catches all server methods
- ✅ **Custom handlers** for specific message types

### Error Handling
- ✅ **Detailed error logging** with context
- ✅ **Connection state tracking** with visual indicators
- ✅ **Graceful shutdown** with proper cleanup

## Next Steps

If connection issues persist:

1. **Check server status** - Ensure the SignalR hub is running
2. **Verify credentials** - Test login endpoint directly
3. **Review server logs** - Check for rejection reasons
4. **Test with minimal client** - Use simple SignalR test
5. **Contact server administrator** - May need account permissions

## Log Analysis

Key log patterns to look for:

**Good Connection:**
```
✅ SignalR connected successfully
🟢 CONNECTED - Attempts: 0, Subscriptions: 1
✅ Successfully subscribed to market status updates
```

**Authentication Issues:**
```
❌ SignalR connection failed: authentication failed
WARNING: Token refresh failed: invalid credentials
```

**Network Issues:**
```
🔴 DISCONNECTED - Last attempts: 5
Connection closed with an error
🟡 RECONNECTING - Attempt: 3
```
