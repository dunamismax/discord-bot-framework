#!/bin/bash

# Discord Bot Framework - Start All Bots Script
# This script starts both Clippy and Music bots

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

echo "🚀 Starting Discord Bot Framework"
echo "================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    print_error ".env file not found. Please create one with your bot tokens."
    print_error "You can copy .env.example if it exists."
    exit 1
fi

# Check if pyproject.toml exists (verify we're in the right directory)
if [ ! -f "pyproject.toml" ]; then
    print_error "pyproject.toml not found. Please run this script from the project root directory."
    exit 1
fi

# Check if uv is installed
if ! command -v uv &> /dev/null; then
    print_error "uv package manager not found. Please install it first:"
    print_error "curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
fi

# Create necessary directories
print_header "Creating directories"
mkdir -p data logs
print_status "Created data and logs directories"

# Load environment variables
set -a
source .env
set +a

print_header "Environment loaded"
if [ -n "$CLIPPY_BOT_TOKEN" ] && [ "$CLIPPY_BOT_TOKEN" != "your_clippy_bot_token_here" ]; then
    print_status "Clippy bot token configured"
else
    print_warning "Clippy bot token not configured or using placeholder"
fi

if [ -n "$MUSIC_BOT_TOKEN" ] && [ "$MUSIC_BOT_TOKEN" != "your_music_bot_token_here" ]; then
    print_status "Music bot token configured"
else
    print_warning "Music bot token not configured or using placeholder"
fi

# Function to start a bot in background
start_bot() {
    local bot_name=$1
    local bot_path=$2
    local log_file="logs/${bot_name}.log"
    
    print_status "Starting $bot_name..."
    
    # Start the bot in background and redirect output to log file
    nohup uv run python -m "$bot_path" > "$log_file" 2>&1 &
    local pid=$!
    
    # Store PID for later reference
    echo $pid > "logs/${bot_name}.pid"
    
    print_status "$bot_name started with PID $pid (log: $log_file)"
    
    # Give it a moment to start
    sleep 2
    
    # Check if process is still running
    if kill -0 $pid 2>/dev/null; then
        print_status "$bot_name is running successfully"
    else
        print_error "$bot_name failed to start. Check $log_file for details."
        return 1
    fi
}

# Start Clippy bot
if [ -f "apps/clippy_bot/main.py" ]; then
    start_bot "clippy-bot" "apps.clippy_bot.main"
else
    print_warning "Clippy bot not found at apps/clippy_bot/main.py"
fi

# Start Music bot
if [ -f "apps/music_bot/main.py" ]; then
    start_bot "music-bot" "apps.music_bot.main"
else
    print_warning "Music bot not found at apps/music_bot/main.py"
fi

echo ""
print_header "Startup Summary"
echo "==============="

# Check health endpoints
sleep 5  # Wait a bit more for bots to fully initialize

print_status "Checking bot health endpoints..."

# Check Clippy bot health
if curl -s http://localhost:8081/health > /dev/null 2>&1; then
    print_status "✅ Clippy bot health endpoint responding"
else
    print_warning "⚠️  Clippy bot health endpoint not responding yet"
fi

# Check Music bot health
if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    print_status "✅ Music bot health endpoint responding"
else
    print_warning "⚠️  Music bot health endpoint not responding yet"
fi

echo ""
print_status "Discord bots are starting up!"
echo ""
echo "Useful commands:"
echo "  • Check logs: tail -f logs/clippy-bot.log"
echo "  • Check logs: tail -f logs/music-bot.log"
echo "  • Health check: curl http://localhost:8081/health"
echo "  • Health check: curl http://localhost:8082/health"
echo "  • Stop bots: ./scripts/stop-all.sh"
echo ""
print_status "Bots are running in the background. Check logs for any issues."