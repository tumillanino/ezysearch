#!/bin/bash

# ezysearch installation script for Go version

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Print banner
echo -e "${GREEN}ezysearch Go installer${NC}"
echo "A universal package manager and GitHub search tool"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.21 or later and try again"
    exit 1
fi

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -f "cmd/ezysearch/main.go" ]]; then
    echo -e "${RED}Error: This script must be run from the ezysearch root directory${NC}"
    exit 1
fi

# Build the binary
echo "Building ezysearch..."
go build -o ezysearch cmd/ezysearch/main.go

# Install to ~/.local/bin
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"
cp ezysearch "${INSTALL_DIR}/"

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${GREEN}ezysearch installed to ${INSTALL_DIR}/ezysearch${NC}"
echo
echo "Run 'ezysearch' to start using it, or restart your terminal."
echo
echo "Usage:"
echo "  • Press Ctrl+P to search and install packages"
echo "  • Press Ctrl+G to search GitHub repositories" 
echo "  • Press Ctrl+T to search directories"
echo
echo "Configuration directory: ~/.config/ezysearch/"