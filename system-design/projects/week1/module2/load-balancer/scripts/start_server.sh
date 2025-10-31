#!/bin/bash

# Load Balancer Test - Multi-server Startup Script
# Purpose: Start multiple backend server instances simultaneously for testing

set -e  # Exit immediately on error

# ============ Configuration ============
# Server port list (modify as needed)
PORTS=(8180 8181 8182)

# Log directory
LOG_DIR="./logs"

# PID file directory
PID_DIR="./pids"

# Build output directory
BUILD_DIR="./bin"
# =======================================

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create necessary directories
mkdir -p "$LOG_DIR"
mkdir -p "$PID_DIR"
mkdir -p "$BUILD_DIR"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  Load Balancer Test Environment${NC}"
echo -e "${YELLOW}========================================${NC}"

# Compile server program
echo -e "\n${GREEN}[1/3] Compiling server program...${NC}"
go build -o "$BUILD_DIR/server" ./cmd/server/main.go

if [ ! -f "$BUILD_DIR/server" ]; then
    echo -e "${RED}Error: Compilation failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Compilation successful${NC}"

# Start all server instances
echo -e "\n${GREEN}[2/3] Starting server instances...${NC}"
for port in "${PORTS[@]}"; do
    # Check if port is already in use
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        echo -e "${YELLOW}⚠ Warning: Port $port already in use, skipping${NC}"
        continue
    fi

    # Start server (in background)
    nohup "$BUILD_DIR/server" -port "$port" > "$LOG_DIR/server-$port.log" 2>&1 &

    # Save process PID
    echo $! > "$PID_DIR/server-$port.pid"

    echo -e "${GREEN}✓ Server started: Port $port (PID: $!)${NC}"

    # Brief delay to avoid port conflicts
    sleep 0.2
done

# Verify server status
echo -e "\n${GREEN}[3/3] Verifying server status...${NC}"
sleep 1  # Wait for servers to fully start

RUNNING_COUNT=0
for port in "${PORTS[@]}"; do
    # Check if process is running
    if [ -f "$PID_DIR/server-$port.pid" ]; then
        PID=$(cat "$PID_DIR/server-$port.pid")
        if ps -p $PID > /dev/null 2>&1; then
            # Try health check
            if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
                echo -e "${GREEN}✓ Port $port: Running (PID: $PID) - Health check passed${NC}"
                ((RUNNING_COUNT++))
            else
                echo -e "${YELLOW}⚠ Port $port: Running (PID: $PID) - Health check failed (may still be starting)${NC}"
            fi
        else
            echo -e "${RED}✗ Port $port: Process exited${NC}"
        fi
    fi
done

# Display summary
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${GREEN}Startup complete! Running servers: $RUNNING_COUNT/${#PORTS[@]}${NC}"
echo -e "${YELLOW}========================================${NC}"

echo -e "\n${YELLOW}Management commands:${NC}"
echo -e "  View all servers: ${GREEN}./status_server.sh${NC}"
echo -e "  Stop all servers: ${GREEN}./stop_server.sh${NC}"
echo -e "  View server logs: ${GREEN}tail -f $LOG_DIR/server-<port>.log${NC}"

echo -e "\n${YELLOW}Test commands:${NC}"
echo -e "  Health check: ${GREEN}curl http://localhost:8081/health${NC}"
echo -e "  Access service: ${GREEN}curl http://localhost:8081/${NC}"

echo -e "\n${YELLOW}Server list:${NC}"
for port in "${PORTS[@]}"; do
    echo -e "  - http://localhost:$port"
done
