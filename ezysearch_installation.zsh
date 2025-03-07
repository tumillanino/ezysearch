#!/usr/bin/env zsh

# ezysearch installer script

set -e

INSTALL_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/ezysearch"
SCRIPT_PATH="$INSTALL_DIR/ezysearch.zsh"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/ezysearch"
ZSH_PLUGINS_DIR="${ZDOTDIR:-$HOME}/.zsh/plugins"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Print banner
echo "${GREEN}ezysearch installer${NC}"
echo "A universal package manager and GitHub search tool"
echo

# Create directories
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$ZSH_PLUGINS_DIR"

# Check dependencies
check_dependencies() {
  local missing=()
  
  # Required dependencies
  for cmd in fzf zsh; do
    if ! command -v $cmd &> /dev/null; then
      missing+=($cmd)
    fi
  done
  
  # Optional dependencies
  for cmd in gh bat fd; do
    if ! command -v $cmd &> /dev/null; then
      echo "${YELLOW}Optional dependency $cmd not found. Some features may be limited.${NC}"
    fi
  done
  
  if [[ ${#missing[@]} -gt 0 ]]; then
    echo "${RED}Missing required dependencies: ${missing[*]}${NC}"
    echo "Please install them first and run this installer again."
    exit 1
  fi
}

# Install main script
install_script() {
  # Download or copy the script to the install directory
  cat > "$SCRIPT_PATH" << 'EOF'
# ezysearch script content will be placed here by the installer
EOF

  # Set execute permissions
  chmod +x "$SCRIPT_PATH"
  
  # Create symbolic link to make it available in PATH
  mkdir -p "$HOME/.local/bin"
  ln -sf "$SCRIPT_PATH" "$HOME/.local/bin/ezysearch"
  
  echo "${GREEN}Installed ezysearch to $SCRIPT_PATH${NC}"
}

# Setup as zsh plugin
setup_zsh_plugin() {
  # Create plugin directory
  local plugin_dir="$ZSH_PLUGINS_DIR/ezysearch"
  mkdir -p "$plugin_dir"
  
  # Create plugin file
  cat > "$plugin_dir/ezysearch.plugin.zsh" << EOF
# ezysearch zsh plugin
source "$SCRIPT_PATH"
EOF

  echo "${GREEN}Created zsh plugin at $plugin_dir/ezysearch.plugin.zsh${NC}"
  
  # Check if plugin is already in .zshrc
  if ! grep -q "ezysearch" "${ZDOTDIR:-$HOME}/.zshrc"; then
    echo "${YELLOW}Please add the plugin to your .zshrc:${NC}"
    
    # Check for common plugin managers
    if grep -q "zplug" "${ZDOTDIR:-$HOME}/.zshrc"; then
      echo "zplug \"$plugin_dir\", from:local"
    elif grep -q "zinit" "${ZDOTDIR:-$HOME}/.zshrc"; then
      echo "zinit load \"$plugin_dir\""
    elif grep -q "oh-my-zsh" "${ZDOTDIR:-$HOME}/.zshrc"; then
      echo "Add ezysearch to the plugins array in your .zshrc"
      echo "plugins=(... ezysearch)"
    else
      echo "Add this line to your .zshrc:"
      echo "source \"$plugin_dir/ezysearch.plugin.zsh\""
    fi
  fi
}

# Main installation process
echo "Checking dependencies..."
check_dependencies

echo "Installing ezysearch..."
install_script

echo "Setting up as zsh plugin..."
setup_zsh_plugin

echo "${GREEN}Installation complete!${NC}"
echo "Run 'source ~/.zshrc' to start using ezysearch, or restart your terminal."
echo
echo "Usage:"
echo "  • Press Ctrl+P to search and install packages"
echo "  • Press Ctrl+G to search GitHub repositories" 
echo "  • Press Ctrl+T to search directories"
echo
echo "Configuration file is located at: $CONFIG_DIR/config.zsh"
