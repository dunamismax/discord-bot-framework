#!/bin/bash

# Discord Bot Framework - Systemd Service Installation Script
# This script installs and configures systemd services for the Discord bots

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

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

echo "🔧 Installing Discord Bot Framework Systemd Services"
echo "==================================================="

# Get the actual user who ran sudo
REAL_USER=${SUDO_USER:-$USER}
REAL_HOME=$(eval echo ~$REAL_USER)
PROJECT_DIR="$REAL_HOME/discord-bot-framework"

print_status "Real user: $REAL_USER"
print_status "Project directory: $PROJECT_DIR"

# Check if project directory exists
if [ ! -d "$PROJECT_DIR" ]; then
    print_error "Project directory not found: $PROJECT_DIR"
    print_error "Please ensure the Discord bot framework is installed in the user's home directory"
    exit 1
fi

# Check if service files exist
print_header "Checking service files"
SERVICE_FILES=("clippy-bot.service" "music-bot.service")

for service_file in "${SERVICE_FILES[@]}"; do
    if [ -f "systemd/$service_file" ]; then
        print_status "Found $service_file"
    else
        print_error "Service file not found: systemd/$service_file"
        exit 1
    fi
done

# Create systemd service directory if it doesn't exist
print_header "Setting up systemd services"
mkdir -p /etc/systemd/system

# Copy service files and update paths
for service_file in "${SERVICE_FILES[@]}"; do
    print_status "Installing $service_file"
    
    # Update the service file with correct paths
    sed "s|/home/ubuntu|$REAL_HOME|g" "systemd/$service_file" | \
    sed "s|User=ubuntu|User=$REAL_USER|g" | \
    sed "s|Group=ubuntu|Group=$REAL_USER|g" > "/etc/systemd/system/$service_file"
    
    print_status "Installed /etc/systemd/system/$service_file"
done

# Set proper permissions
chmod 644 /etc/systemd/system/clippy-bot.service
chmod 644 /etc/systemd/system/music-bot.service

# Create necessary directories
print_header "Creating directories"
mkdir -p "$PROJECT_DIR/data"
mkdir -p "$PROJECT_DIR/logs"
chown -R $REAL_USER:$REAL_USER "$PROJECT_DIR/data"
chown -R $REAL_USER:$REAL_USER "$PROJECT_DIR/logs"

# Reload systemd daemon
print_header "Reloading systemd daemon"
systemctl daemon-reload
print_status "Systemd daemon reloaded"

# Check .env file
print_header "Checking configuration"
if [ -f "$PROJECT_DIR/.env" ]; then
    print_status ".env file found"
    
    # Check if tokens are configured
    if grep -q "your_.*_token_here" "$PROJECT_DIR/.env"; then
        print_warning "Bot tokens appear to be placeholder values in .env file"
        print_warning "Please edit $PROJECT_DIR/.env with your actual Discord bot tokens"
    else
        print_status "Bot tokens appear to be configured"
    fi
else
    print_warning ".env file not found at $PROJECT_DIR/.env"
    print_warning "Please create and configure the .env file before starting services"
fi

# Enable services
print_header "Enabling services"
systemctl enable clippy-bot.service
systemctl enable music-bot.service
print_status "Services enabled for automatic startup"

# Service management instructions
echo ""
print_header "Installation completed successfully!"
echo "================================================="
echo ""
echo "Service management commands:"
echo ""
echo "Start services:"
echo "  sudo systemctl start clippy-bot music-bot"
echo ""
echo "Stop services:"
echo "  sudo systemctl stop clippy-bot music-bot"
echo ""
echo "Restart services:"
echo "  sudo systemctl restart clippy-bot music-bot"
echo ""
echo "Check status:"
echo "  sudo systemctl status clippy-bot music-bot"
echo ""
echo "View logs:"
echo "  sudo journalctl -u clippy-bot -f"
echo "  sudo journalctl -u music-bot -f"
echo ""
echo "Disable auto-start:"
echo "  sudo systemctl disable clippy-bot music-bot"
echo ""

# Check if services should be started now
echo -n "Would you like to start the services now? [y/N]: "
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    print_header "Starting services"
    
    systemctl start clippy-bot music-bot
    
    # Wait a moment for services to start
    sleep 3
    
    # Check status
    print_status "Service status:"
    systemctl --no-pager status clippy-bot music-bot
    
    echo ""
    print_status "Services started! Check the status above for any issues."
    print_status "Logs can be viewed with: sudo journalctl -u clippy-bot -f"
else
    print_status "Services installed but not started."
    print_status "Start them manually when ready: sudo systemctl start clippy-bot music-bot"
fi

echo ""
print_status "Systemd service installation complete! 🎉"