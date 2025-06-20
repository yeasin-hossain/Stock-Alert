#!/bin/bash

# Script to run either main application or test

case "$1" in
    "main")
        echo "🚀 Starting main SignalR data feed application..."
        cd "$(dirname "$0")"
        go run main.go
        ;;
    "debug")
        echo "🐛 Starting main application with debug logging..."
        cd "$(dirname "$0")"
        export GOMAXPROCS=1
        go run main.go 2>&1 | tee -a datafeed.log
        ;;
    "test")
        echo "🧪 Starting SignalR test with special characters..."
        cd "$(dirname "$0")/cmd/test"
        go run test_special_chars.go
        ;;
    "simple")
        echo "🧪 Starting Simple SignalR client (documentation-based)..."
        cd "$(dirname "$0")/cmd/simple"
        go run simple_client.go
        ;;
    "basic")
        echo "🔧 Starting Basic SignalR connection test..."
        cd "$(dirname "$0")/cmd/basic"
        go run basic_client.go
        ;;
    "build")
        echo "🔨 Building applications..."
        cd "$(dirname "$0")"
        echo "Building main application..."
        go build -o datafeed main.go
        echo "Building test application..."
        go build -o cmd/test/test_special_chars cmd/test/test_special_chars.go
        echo "Building simple client..."
        go build -o cmd/simple/simple_client cmd/simple/simple_client.go
        echo "Building basic test..."
        go build -o cmd/basic/basic_client cmd/basic/basic_client.go
        echo "✅ Build complete!"
        ;;
    *)
        echo "Usage: $0 {main|debug|test|simple|basic|build}"
        echo "  main   - Run the main data feed application"
        echo "  debug  - Run main application with debug logging to file"
        echo "  test   - Run the special character test"
        echo "  simple - Run the simple documentation-based client"
        echo "  basic  - Run the basic connection test"
        echo "  build  - Build all applications"
        exit 1
        ;;
esac
