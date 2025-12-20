#!/bin/bash
# LlamaGate Installer for Unix/Linux/macOS
# This script installs dependencies and sets up LlamaGate

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Parse arguments
SKIP_GO_CHECK=false
SKIP_OLLAMA_CHECK=false
SILENT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-go-check)
            SKIP_GO_CHECK=true
            shift
            ;;
        --skip-ollama-check)
            SKIP_OLLAMA_CHECK=true
            shift
            ;;
        --silent)
            SILENT=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Function to print colored output
print_info() {
    echo -e "${CYAN}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to prompt user
prompt_user() {
    local prompt="$1"
    local default="$2"
    local required="${3:-false}"
    local input
    
    if [ "$SILENT" = true ] && [ -n "$default" ]; then
        echo "$default"
        return
    fi
    
    while true; do
        if [ -n "$default" ]; then
            read -p "$prompt [$default]: " input
            if [ -z "$input" ]; then
                input="$default"
            fi
        else
            read -p "$prompt: " input
        fi
        
        if [ -n "$input" ] || [ "$required" = false ]; then
            echo "$input"
            return
        fi
        print_error "This field is required. Please enter a value."
    done
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command_exists apt-get; then
            echo "debian"
        elif command_exists yum; then
            echo "rhel"
        elif command_exists pacman; then
            echo "arch"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

print_info "========================================"
print_info "LlamaGate Installer"
print_info "========================================"
echo ""

# Step 1: Check Go installation
print_info "[1/6] Checking Go installation..."
if [ "$SKIP_GO_CHECK" = false ]; then
    if command_exists go; then
        GO_VERSION=$(go version)
        print_success "Go is installed: $GO_VERSION"
    else
        print_error "Go is not installed"
        INSTALL_GO=$(prompt_user "Would you like to install Go? (Y/n)" "Y")
        
        if [[ "$INSTALL_GO" =~ ^[Yy]$ ]] || [ -z "$INSTALL_GO" ]; then
            print_info "Installing Go..."
            
            if [ "$OS" = "macos" ]; then
                if command_exists brew; then
                    print_info "Using Homebrew to install Go..."
                    brew install go
                else
                    print_error "Homebrew not found. Please install Homebrew first:"
                    echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                    exit 1
                fi
            elif [ "$OS" = "debian" ]; then
                print_info "Installing Go via apt..."
                sudo apt-get update
                sudo apt-get install -y golang-go
            elif [ "$OS" = "rhel" ]; then
                print_info "Installing Go via yum..."
                sudo yum install -y golang
            elif [ "$OS" = "arch" ]; then
                print_info "Installing Go via pacman..."
                sudo pacman -S --noconfirm go
            else
                print_error "Automatic Go installation not supported for this OS."
                print_info "Please install Go manually from https://go.dev/dl/"
                print_info "After installing, restart this script."
                exit 1
            fi
            
            # Verify installation
            if command_exists go; then
                print_success "Go installed successfully"
            else
                print_error "Go installation may have failed. Please install manually."
                exit 1
            fi
        else
            print_warning "Skipping Go installation. Please install Go manually and restart this script."
            exit 1
        fi
    fi
else
    print_warning "Skipping Go check (--skip-go-check)"
fi

# Step 2: Check Ollama installation
echo ""
print_info "[2/6] Checking Ollama installation..."
if [ "$SKIP_OLLAMA_CHECK" = false ]; then
    if command_exists ollama; then
        print_success "Ollama is installed"
        OLLAMA_RUNNING=false
        
        if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
            OLLAMA_RUNNING=true
            print_success "Ollama is running"
        else
            print_warning "Ollama is installed but not running"
            START_OLLAMA=$(prompt_user "Would you like to start Ollama now? (Y/n)" "Y")
            if [[ "$START_OLLAMA" =~ ^[Yy]$ ]] || [ -z "$START_OLLAMA" ]; then
                print_info "Starting Ollama..."
                ollama serve >/dev/null 2>&1 &
                sleep 3
                print_success "Ollama started"
            fi
        fi
    else
        print_error "Ollama is not installed"
        INSTALL_OLLAMA=$(prompt_user "Would you like to install Ollama? (Y/n)" "Y")
        
        if [[ "$INSTALL_OLLAMA" =~ ^[Yy]$ ]] || [ -z "$INSTALL_OLLAMA" ]; then
            print_info "Installing Ollama..."
            
            if [ "$OS" = "macos" ]; then
                if command_exists brew; then
                    print_info "Installing Ollama via Homebrew..."
                    brew install ollama
                else
                    print_error "Homebrew not found. Please install Ollama manually:"
                    print_info "Visit: https://ollama.com/download"
                    if command_exists open; then
                        open "https://ollama.com/download" 2>/dev/null || true
                    fi
                fi
            elif [ "$OS" = "debian" ] || [ "$OS" = "rhel" ] || [ "$OS" = "linux" ]; then
                print_info "Installing Ollama via official installer..."
                curl -fsSL https://ollama.com/install.sh | sh
            else
                print_info "Opening Ollama download page..."
                print_info "Please install Ollama from: https://ollama.com/download"
                if command_exists xdg-open; then
                    xdg-open "https://ollama.com/download" 2>/dev/null || true
                elif command_exists open; then
                    open "https://ollama.com/download" 2>/dev/null || true
                fi
            fi
            
            CONTINUE=$(prompt_user "Press Enter after you have installed Ollama to continue...")
            
            # Verify installation
            if command_exists ollama; then
                print_success "Ollama installed successfully"
            else
                print_error "Ollama installation not detected. Please restart this script after installing."
                exit 1
            fi
        else
            print_warning "Skipping Ollama installation. LlamaGate requires Ollama to function."
        fi
    fi
else
    print_warning "Skipping Ollama check (--skip-ollama-check)"
fi

# Step 3: Install Go dependencies
echo ""
print_info "[3/6] Installing Go dependencies..."
cd "$PROJECT_ROOT"
go mod download
print_success "Dependencies installed"

# Step 4: Build LlamaGate
echo ""
print_info "[4/6] Building LlamaGate..."
go build -o llamagate ./cmd/llamagate
if [ -f "llamagate" ]; then
    chmod +x llamagate
    print_success "Build successful"
else
    print_error "Build failed - llamagate not found"
    exit 1
fi

# Step 5: Create configuration file
echo ""
print_info "[5/6] Setting up configuration..."
ENV_FILE="$PROJECT_ROOT/.env"
if [ -f "$ENV_FILE" ]; then
    print_warning ".env file already exists"
    OVERWRITE=$(prompt_user "Would you like to overwrite it? (y/N)" "N")
    if [[ ! "$OVERWRITE" =~ ^[Yy]$ ]]; then
        print_warning "Keeping existing .env file"
        CREATE_ENV=false
    else
        CREATE_ENV=true
    fi
else
    CREATE_ENV=true
fi

if [ "$CREATE_ENV" = true ]; then
    print_info "Creating .env file..."
    
    OLLAMA_HOST=$(prompt_user "Ollama host" "http://localhost:11434")
    API_KEY=$(prompt_user "API key (leave empty to disable authentication)" "")
    RATE_LIMIT=$(prompt_user "Rate limit (requests per second)" "10")
    DEBUG=$(prompt_user "Enable debug logging? (true/false)" "false")
    PORT=$(prompt_user "Server port" "8080")
    LOG_FILE=$(prompt_user "Log file path (leave empty for console only)" "")
    
    cat > "$ENV_FILE" <<EOF
# LlamaGate Configuration
# Generated by installer on $(date '+%Y-%m-%d %H:%M:%S')

# Ollama server URL
OLLAMA_HOST=$OLLAMA_HOST

# API key for authentication (leave empty to disable authentication)
API_KEY=$API_KEY

# Rate limit (requests per second)
RATE_LIMIT_RPS=$RATE_LIMIT

# Enable debug logging (true/false)
DEBUG=$DEBUG

# Server port
PORT=$PORT

# Log file path (leave empty to log only to console)
LOG_FILE=$LOG_FILE
EOF
    
    print_success "Configuration file created"
fi

# Step 6: Create systemd service (Linux only, optional)
echo ""
print_info "[6/6] Post-installation setup..."

if [ "$OS" != "macos" ] && [ "$OS" != "unknown" ] && [ "$OS" != "macos" ]; then
    CREATE_SERVICE=$(prompt_user "Would you like to create a systemd service? (y/N)" "N")
    if [[ "$CREATE_SERVICE" =~ ^[Yy]$ ]]; then
        SERVICE_FILE="/etc/systemd/system/llamagate.service"
        INSTALL_DIR="$PROJECT_ROOT"
        
        print_info "Creating systemd service..."
        sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=LlamaGate - OpenAI-compatible proxy for Ollama
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/llamagate
Restart=always
RestartSec=5
EnvironmentFile=$INSTALL_DIR/.env

[Install]
WantedBy=multi-user.target
EOF
        
        sudo systemctl daemon-reload
        print_success "Systemd service created"
        print_info "To start the service: sudo systemctl start llamagate"
        print_info "To enable on boot: sudo systemctl enable llamagate"
    fi
fi

# Summary
echo ""
print_info "========================================"
print_success "Installation Complete!"
print_info "========================================"
echo ""
print_success "LlamaGate has been installed successfully!"
echo ""
print_info "Quick Start:"
echo "  1. Run: ./llamagate"
echo "  2. Or use: ./scripts/unix/run.sh"
echo ""
print_info "Configuration:"
echo "  Edit .env file to change settings"
echo ""
print_info "Documentation:"
echo "  See README.md for full documentation"
echo "  See TESTING.md for testing instructions"
echo ""
print_info "Test the installation:"
echo "  Run: ./scripts/unix/test.sh"
echo ""

