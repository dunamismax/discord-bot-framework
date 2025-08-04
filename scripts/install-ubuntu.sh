#!/bin/bash

# Discord Bot Framework - Ubuntu Installation Script
# This script sets up the Discord bot framework on Ubuntu Linux or WSL2

set -e

echo "🚀 Discord Bot Framework - Ubuntu Installation"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if running on Ubuntu or WSL
print_header "Checking system compatibility"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
    print_status "Detected: $OS $VER"
else
    print_error "Cannot determine OS version"
    exit 1
fi

# Check if WSL
if grep -qi microsoft /proc/version; then
    print_status "Running on WSL (Windows Subsystem for Linux)"
    IS_WSL=true
else
    IS_WSL=false
fi

# Update package lists
print_header "Updating package lists"
sudo apt update

# Install required system packages
print_header "Installing system dependencies"
PACKAGES=(
    "python3"
    "python3-pip" 
    "python3-venv"
    "python3-dev"
    "git"
    "curl"
    "wget"
    "ffmpeg"
    "build-essential"
    "pkg-config"
    "libffi-dev"
    "libssl-dev"
)

for package in "${PACKAGES[@]}"; do
    if dpkg -l | grep -q "^ii  $package "; then
        print_status "$package is already installed"
    else
        print_status "Installing $package"
        sudo apt install -y "$package"
    fi
done

# Check Python version
print_header "Checking Python version"
PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
MAJOR=$(echo $PYTHON_VERSION | cut -d'.' -f1)
MINOR=$(echo $PYTHON_VERSION | cut -d'.' -f2)

if [ "$MAJOR" -eq 3 ] && [ "$MINOR" -ge 8 ]; then
    print_status "Python $PYTHON_VERSION detected - Compatible!"
else
    print_error "Python 3.8+ is required. Found: $PYTHON_VERSION"
    exit 1
fi

# Install uv package manager if not already installed
print_header "Installing uv package manager"
if command -v uv &> /dev/null; then
    print_status "uv is already installed"
else
    print_status "Installing uv package manager"
    curl -LsSf https://astral.sh/uv/install.sh | sh
    export PATH="$HOME/.cargo/bin:$PATH"
    
    # Add to bashrc if not already there
    if ! grep -q 'export PATH="$HOME/.cargo/bin:$PATH"' ~/.bashrc; then
        echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
        print_status "Added uv to PATH in ~/.bashrc"
    fi
fi

# Verify uv installation
if command -v uv &> /dev/null; then
    UV_VERSION=$(uv --version)
    print_status "uv installed: $UV_VERSION"
else
    print_error "Failed to install uv package manager"
    exit 1
fi

# Install project dependencies
print_header "Installing project dependencies"
if [ -f "pyproject.toml" ]; then
    print_status "Installing dependencies with uv"
    uv sync --all-extras
    print_status "Dependencies installed successfully"
else
    print_warning "pyproject.toml not found. Please run this script from the project root directory"
fi

# Create environment file if it doesn't exist
print_header "Setting up configuration"
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        cp .env.example .env
        print_status "Created .env file from .env.example"
        print_warning "Please edit .env file with your Discord bot tokens"
    else
        print_status "Creating basic .env file"
        cat > .env << EOF
# Discord Bot Tokens
CLIPPY_BOT_TOKEN=your_clippy_bot_token_here
MUSIC_BOT_TOKEN=your_music_bot_token_here

# Optional: Guild IDs for testing (removes global command sync delay)
CLIPPY_GUILD_ID=
MUSIC_GUILD_ID=

# Debug mode (set to true for development)
CLIPPY_DEBUG=false
MUSIC_DEBUG=false

# Music Bot Configuration
MUSIC_MAX_PLAYLIST_SIZE=100
MUSIC_AUTO_DISCONNECT_TIMEOUT=300
EOF
        print_warning "Created .env file. Please edit it with your Discord bot tokens"
    fi
else
    print_status ".env file already exists"
fi

# Create data directory for SQLite databases
print_header "Setting up data directory"
if [ ! -d "data" ]; then
    mkdir -p data
    print_status "Created data directory for SQLite databases"
fi

# Create logs directory
if [ ! -d "logs" ]; then
    mkdir -p logs
    print_status "Created logs directory"
fi

# Set appropriate permissions
chmod +x scripts/*.sh 2>/dev/null || true

# Run validation
print_header "Running validation tests"
if python3 validate.py; then
    print_status "Validation tests passed!"
else
    print_warning "Some validation tests failed. Check the output above."
fi

# WSL-specific setup
if [ "$IS_WSL" = true ]; then
    print_header "WSL-specific setup"
    print_status "Detected WSL environment"
    print_warning "Note: If you want to access the bots from Windows, you may need to:"
    print_warning "1. Configure Windows Firewall to allow the ports (8081, 8082)"
    print_warning "2. Use the WSL IP address instead of localhost"
    print_warning "3. Consider using Caddy reverse proxy for easier access"
fi

# Installation complete
echo ""
echo "=============================================="
print_status "Installation completed successfully!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "1. Edit the .env file with your Discord bot tokens:"
echo "   nano .env"
echo ""
echo "2. Start the bots manually:"
echo "   ./scripts/start-all.sh"
echo ""
echo "3. Or install as systemd services:"
echo "   sudo ./systemd/install-services.sh"
echo ""
echo "4. Install and configure Caddy reverse proxy:"
echo "   Follow the README instructions for Caddy setup"
echo ""
echo "5. Check bot health:"
echo "   curl http://localhost:8081/health  # Clippy bot"
echo "   curl http://localhost:8082/health  # Music bot"
echo ""
print_status "Happy Discord bot hosting! 🎉"