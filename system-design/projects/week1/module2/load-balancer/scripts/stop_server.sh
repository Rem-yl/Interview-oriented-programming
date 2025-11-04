#!/bin/bash

# Load Balancer Test - Server Stop Script
# Purpose: Gracefully stop all backend server instances

# ============ Configuration ============
# Server port list (must match start_server.sh)
PORTS=(8180 8181 8182)

# PID file directory
PID_DIR="./pids"

# Build directory
BUILD_DIR="./bin"
# =======================================

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  Stopping Backend Servers${NC}"
echo -e "${YELLOW}========================================${NC}"

# Check if PID directory exists
if [ ! -d "$PID_DIR" ]; then
    echo -e "${YELLOW}No running servers found (PID directory doesn't exist)${NC}"
    exit 0
fi

# Check if there are any PID files
if [ -z "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}No running servers found (no PID files)${NC}"
    exit 0
fi

STOPPED_COUNT=0
TOTAL_COUNT=0

# Stop servers using PID files
echo -e "\n${GREEN}[1/3] Stopping servers by PID...${NC}"
for pid_file in "$PID_DIR"/server-*.pid; do
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        PORT=$(basename "$pid_file" .pid | sed 's/server-//')
        ((TOTAL_COUNT++))

        # Check if process is running
        if ps -p $PID > /dev/null 2>&1; then
            echo -e "  Stopping server on port $PORT (PID: $PID)..."

            # Graceful shutdown: SIGTERM first
            kill $PID 2>/dev/null || true

            # Wait up to 5 seconds for graceful shutdown
            for i in {1..10}; do
                if ! ps -p $PID > /dev/null 2>&1; then
                    echo -e "  ${GREEN}✓ Port $PORT stopped gracefully${NC}"
                    ((STOPPED_COUNT++))
                    break
                fi
                sleep 0.5
            done

            # Force kill if still running
            if ps -p $PID > /dev/null 2>&1; then
                echo -e "  ${YELLOW}⚠ Forcing shutdown for port $PORT${NC}"
                kill -9 $PID 2>/dev/null || true
                sleep 0.5
                if ! ps -p $PID > /dev/null 2>&1; then
                    echo -e "  ${GREEN}✓ Port $PORT force stopped${NC}"
                    ((STOPPED_COUNT++))
                else
                    echo -e "  ${RED}✗ Failed to stop port $PORT${NC}"
                fi
            fi
        else
            echo -e "  ${YELLOW}⚠ Port $PORT: Process already stopped (PID: $PID)${NC}"
            ((STOPPED_COUNT++))
        fi

        # Remove PID file
        rm -f "$pid_file"
    fi
done

# Additional cleanup: stop servers by port
echo -e "\n${GREEN}[2/3] Verifying ports are released...${NC}"
for port in "${PORTS[@]}"; do
    # Try multiple lsof syntaxes for compatibility
    pids=""
    pids=$(lsof -ti :$port 2>/dev/null || true)

    if [ -z "$pids" ]; then
        pids=$(lsof -ti TCP:$port 2>/dev/null || true)
    fi

    if [ -n "$pids" ]; then
        echo -e "  ${YELLOW}⚠ Port $port still occupied (PID: $pids), cleaning up...${NC}"
        for pid in $pids; do
            # Check if it's our server process
            cmd=$(ps -p $pid -o command= 2>/dev/null || echo "")
            if [[ "$cmd" == *"server"* ]] || [[ "$cmd" == *"$BUILD_DIR"* ]]; then
                kill -9 $pid 2>/dev/null || true
                echo -e "    ${GREEN}✓ Killed orphaned process (PID: $pid)${NC}"
            else
                echo -e "    ${RED}⚠ Port occupied by other process: $cmd${NC}"
            fi
        done
    else
        echo -e "  ${GREEN}✓ Port $port released${NC}"
    fi
done

# Clean up any remaining server processes
echo -e "\n${GREEN}[3/3] Cleaning up any orphaned server processes...${NC}"
orphaned_pids=$(ps aux | grep "$BUILD_DIR/server" | grep -v grep | awk '{print $2}' || true)

if [ -n "$orphaned_pids" ]; then
    for pid in $orphaned_pids; do
        echo -e "  Killing orphaned server process (PID: $pid)"
        kill -9 $pid 2>/dev/null || true
    done
    echo -e "  ${GREEN}✓ Orphaned processes cleaned${NC}"
else
    echo -e "  ${GREEN}✓ No orphaned processes found${NC}"
fi

# Display summary
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${GREEN}Shutdown complete! Stopped: $STOPPED_COUNT/$TOTAL_COUNT servers${NC}"
echo -e "${YELLOW}========================================${NC}"

# Final verification
echo -e "\n${YELLOW}Final status:${NC}"
any_running=false
for port in "${PORTS[@]}"; do
    if lsof -ti :$port >/dev/null 2>&1 || lsof -ti TCP:$port >/dev/null 2>&1; then
        echo -e "  ${RED}✗ Port $port: Still occupied${NC}"
        any_running=true
    else
        echo -e "  ${GREEN}✓ Port $port: Free${NC}"
    fi
done

if $any_running; then
    echo -e "\n${YELLOW}Some ports are still occupied. You may need to:${NC}"
    echo -e "  1. Check what's using the ports: ${GREEN}lsof -i :PORT${NC}"
    echo -e "  2. Manually kill processes: ${GREEN}kill -9 PID${NC}"
    exit 1
else
    echo -e "\n${GREEN}All servers stopped successfully!${NC}"
    echo -e "\n${YELLOW}To restart servers:${NC}"
    echo -e "  ${GREEN}./start_server.sh${NC}"
fi
