#!/bin/bash

# OpenAI Quota Proxy - macOS Installation Script
# This script installs and sets up the OpenAI Quota Proxy on macOS

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="/usr/local/bin"
APP_NAME="openai-quota"

echo -e "${BLUE}🚀 OpenAI Quota Proxy - macOS Installation${NC}"
echo "=================================================="

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This script is designed for macOS only."
    exit 1
fi

print_status "Detected macOS system"

# Check if Go is installed
print_info "Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_warning "Go is not installed. Installing Go via Homebrew..."
    
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        print_info "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        
        # Add Homebrew to PATH for this session
        if [[ -f "/opt/homebrew/bin/brew" ]]; then
            eval "$(/opt/homebrew/bin/brew shellenv)"
        elif [[ -f "/usr/local/bin/brew" ]]; then
            eval "$(/usr/local/bin/brew shellenv)"
        fi
    fi
    
    # Install Go
    brew install go
    print_status "Go installed successfully"
else
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status "Go is already installed: $GO_VERSION"
fi

# Verify Go version (require 1.21+)
GO_VERSION_NUMBER=$(go version | cut -d' ' -f3 | sed 's/go//' | cut -d'.' -f1,2)
REQUIRED_VERSION="1.21"

if [[ $(echo "$GO_VERSION_NUMBER $REQUIRED_VERSION" | awk '{print ($1 >= $2)}') == 1 ]]; then
    print_status "Go version $GO_VERSION_NUMBER meets requirements (>= $REQUIRED_VERSION)"
else
    print_error "Go version $GO_VERSION_NUMBER is too old. Please update to Go $REQUIRED_VERSION or later."
    exit 1
fi

# Check if we're in the project directory
if [[ ! -f "main.go" ]] || [[ ! -f "go.mod" ]]; then
    print_error "This script must be run from the openai-quota project directory."
    print_info "Make sure you're in the directory containing main.go and go.mod files."
    exit 1
fi

print_status "Found project files in current directory"

# Install dependencies
print_info "Installing Go dependencies..."
go mod tidy
go mod download
print_status "Dependencies installed"

# Build the application
print_info "Building the application..."
make build
print_status "Application built successfully"

# Run quick tests to verify build
print_info "Running quick tests to verify installation..."
if make test-quick > /dev/null 2>&1; then
    print_status "All tests passed"
else
    print_warning "Some tests failed, but continuing with installation"
fi

# Ask for installation preference
echo ""
print_info "Choose installation option:"
echo "1) Install globally to $INSTALL_DIR (requires sudo)"
echo "2) Install locally to ~/.local/bin"
echo "3) Skip installation (just build)"
read -p "Enter choice (1-3): " choice

case $choice in
    1)
        print_info "Installing globally to $INSTALL_DIR..."
        sudo cp $APP_NAME $INSTALL_DIR/
        sudo chmod +x $INSTALL_DIR/$APP_NAME
        print_status "Installed globally to $INSTALL_DIR/$APP_NAME"
        INSTALLED_PATH="$INSTALL_DIR/$APP_NAME"
        ;;
    2)
        LOCAL_BIN="$HOME/.local/bin"
        mkdir -p $LOCAL_BIN
        cp $APP_NAME $LOCAL_BIN/
        chmod +x $LOCAL_BIN/$APP_NAME
        print_status "Installed locally to $LOCAL_BIN/$APP_NAME"
        INSTALLED_PATH="$LOCAL_BIN/$APP_NAME"
        
        # Check if ~/.local/bin is in PATH
        if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
            print_warning "~/.local/bin is not in your PATH"
            print_info "Add this to your ~/.zshrc or ~/.bash_profile:"
            echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
        fi
        ;;
    3)
        print_info "Skipping installation. Binary available as: ./$APP_NAME"
        INSTALLED_PATH="./$APP_NAME"
        ;;
    *)
        print_error "Invalid choice. Exiting."
        exit 1
        ;;
esac

# Create sample configuration if it doesn't exist
if [[ ! -f "config/app.env" ]]; then
    print_info "Creating sample configuration..."
    mkdir -p config
    cat > config/app.env << 'EOF'
# OpenAI Quota Proxy Configuration
# Copy this file and customize for your deployment

# Server Configuration
PORT=8081
QUOTA=2.0
PRICING_FILE=config/model_pricing.csv

# OpenAI API Configuration  
# Set your OpenAI API key as environment variable:
# export OPENAI_API_KEY=your_api_key_here

# Logging
LOG_LEVEL=info

# Security
# ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
EOF
    print_status "Created sample configuration: config/app.env"
fi

# Create launchd service file for auto-start (optional)
read -p "Do you want to create a LaunchAgent for auto-start? (y/n): " create_service

if [[ $create_service == "y" || $create_service == "Y" ]]; then
    SERVICE_DIR="$HOME/Library/LaunchAgents"
    SERVICE_FILE="$SERVICE_DIR/com.openai-quota.proxy.plist"
    
    mkdir -p $SERVICE_DIR
    
    cat > $SERVICE_FILE << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.openai-quota.proxy</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALLED_PATH</string>
        <string>-quota</string>
        <string>2.0</string>
        <string>-port</string>
        <string>8081</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$(pwd)</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/Library/Logs/openai-quota.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/Library/Logs/openai-quota-error.log</string>
</dict>
</plist>
EOF
    
    print_status "Created LaunchAgent: $SERVICE_FILE"
    print_info "To start the service: launchctl load $SERVICE_FILE"
    print_info "To stop the service: launchctl unload $SERVICE_FILE"
    print_info "Logs will be written to: $HOME/Library/Logs/openai-quota.log"
fi

echo ""
print_status "🎉 Installation completed successfully!"
echo ""
print_info "Usage examples:"
echo "  # Run with default settings"
echo "  $APP_NAME"
echo ""
echo "  # Run with custom quota and port"
echo "  $APP_NAME -quota 5.0 -port 8080"
echo ""
echo "  # Check server status"
echo "  curl http://localhost:8081/v1/chat/completions"
echo ""
echo "  # Test API call (replace YOUR_API_KEY)"
echo "  curl -X POST http://localhost:8081/v1/chat/completions \\"
echo "       -H \"Authorization: Bearer YOUR_API_KEY\" \\"
echo "       -H \"Content-Type: application/json\" \\"
echo "       -d '{\"model\":\"gpt-4o\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}'"
echo ""
print_info "Configuration file: config/app.env"
print_info "Documentation: README.md"
print_info "For help: $APP_NAME -help"

if [[ $choice == 1 || $choice == 2 ]]; then
    echo ""
    print_info "The application has been installed and is ready to use!"
fi