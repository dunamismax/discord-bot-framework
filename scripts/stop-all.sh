#!/bin/bash

# Discord Bot Framework - Stop All Bots Script
# This script stops both Clippy and Music bots

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

echo "🛑 Stopping Discord Bot Framework"
echo "================================="

# Function to stop a bot
stop_bot() {
    local bot_name=$1
    local pid_file="logs/${bot_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        print_status "Stopping $bot_name (PID: $pid)..."
        
        if kill -0 $pid 2>/dev/null; then
            # Send SIGTERM for graceful shutdown
            kill -TERM $pid
            
            # Wait up to 10 seconds for graceful shutdown
            local count=0
            while kill -0 $pid 2>/dev/null && [ $count -lt 10 ]; do
                sleep 1
                count=$((count + 1))
            done
            
            # If still running, force kill
            if kill -0 $pid 2>/dev/null; then
                print_warning "$bot_name didn't stop gracefully, force killing..."
                kill -KILL $pid
                sleep 1
            fi
            
            # Verify it's stopped
            if kill -0 $pid 2>/dev/null; then
                print_error "Failed to stop $bot_name"
                return 1
            else
                print_status "$bot_name stopped successfully"
            fi
        else
            print_warning "$bot_name was not running (PID $pid not found)"
        fi
        
        # Remove PID file
        rm -f "$pid_file"
    else
        print_warning "No PID file found for $bot_name (logs/${bot_name}.pid)"
    fi
}

# Stop both bots
stop_bot "clippy-bot"
stop_bot "music-bot"

# Also try to kill any remaining Python processes that might be bots
print_header "Cleaning up any remaining bot processes"

# Find and kill any remaining Discord bot processes
REMAINING_PIDS=$(pgrep -f "apps\..*\.main" 2>/dev/null || true)
if [ -n "$REMAINING_PIDS" ]; then
    print_warning "Found remaining bot processes, stopping them..."
    echo "$REMAINING_PIDS" | xargs kill -TERM 2>/dev/null || true
    sleep 2
    echo "$REMAINING_PIDS" | xargs kill -KILL 2>/dev/null || true
    print_status "Cleaned up remaining processes"
else
    print_status "No remaining bot processes found"
fi

# Check if health endpoints are down
print_header "Verifying shutdown"

if curl -s http://localhost:8081/health > /dev/null 2>&1; then
    print_warning "⚠️  Clippy bot health endpoint still responding"
else
    print_status "✅ Clippy bot health endpoint is down"
fi

if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    print_warning "⚠️  Music bot health endpoint still responding"
else
    print_status "✅ Music bot health endpoint is down"
fi

echo ""
print_status "All Discord bots have been stopped!"
echo ""
echo "Log files are preserved in the logs/ directory:"
echo "  • logs/clippy-bot.log"
echo "  • logs/music-bot.log"
echo ""
print_status "To start the bots again, run: ./scripts/start-all.sh"