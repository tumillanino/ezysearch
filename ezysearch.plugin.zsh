#!/usr/bin/env zsh

# ezysearch - A universal package manager search utility
# Supports: pacman+yay, apt, homebrew, dnf

# Configuration variables
EZY_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/ezysearch"
EZY_CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/ezysearch"

# Create config and cache directories if they don't exist
mkdir -p "$EZY_CONFIG_DIR" "$EZY_CACHE_DIR"

# Default configuration file
EZY_CONFIG_FILE="$EZY_CONFIG_DIR/config.zsh"

# Create default config if it doesn't exist
if [[ ! -f "$EZY_CONFIG_FILE" ]]; then
  cat > "$EZY_CONFIG_FILE" <<EOL
# ezysearch configuration
# Keybindings
bindkey '^P' ezy_package_search  # Ctrl+P for package search
bindkey '^G' ezy_github_search   # Ctrl+G for GitHub repo search
bindkey '^T' ezy_directory_search # Ctrl+T for directory search

# Directory search settings
export EZY_DIR_COMMAND="fd --hidden --strip-cwd-prefix --exclude .git"
export EZY_DIR_PREVIEW="bat --color=always -n --line-range :500 {}"

# GitHub search settings
export EZY_GITHUB_LIMIT=50
EOL
fi

# Source configuration
source "$EZY_CONFIG_FILE"

# Detect package manager
detect_package_manager() {
  if command -v yay &> /dev/null; then
    echo "yay"
  elif command -v pacman &> /dev/null; then
    echo "pacman"
  elif command -v apt &> /dev/null; then
    echo "apt"
  elif command -v brew &> /dev/null; then
    echo "brew"
  elif command -v dnf &> /dev/null; then
    echo "dnf"
  else
    echo "unknown"
  fi
}

# Package manager search function
ezy_package_search() {
  local selected_package
  local pkg_manager=$(detect_package_manager)
  
  case "$pkg_manager" in
    yay)
      selected_package=$(yay -Slq | fzf --preview 'yay -Si {}' --height=97% --layout=reverse)
      ;;
    pacman)
      selected_package=$(pacman -Sl | awk '{print $2}' | sort -u | fzf --height=97% --layout=reverse --preview 'pacman -Si {}')
      ;;
    apt)
      selected_package=$(apt-cache pkgnames | fzf --preview 'apt-cache show {}' --height=97% --layout=reverse)
      ;;
    brew)
      selected_package=$(brew formulae | fzf --preview 'brew info {}' --height=97% --layout=reverse)
      ;;
    dnf)
      # Cache the package list to avoid slow dnf searches
      local dnf_cache="$EZY_CACHE_DIR/dnf_packages"
      if [[ ! -f "$dnf_cache" ]] || [[ $(find "$dnf_cache" -mtime +1) ]]; then
        # Cache doesn't exist or is older than 1 day
        dnf list available | awk '{if(NR>1)print $1}' > "$dnf_cache"
      fi
      selected_package=$(cat "$dnf_cache" | fzf --preview 'dnf info {}' --height=97% --layout=reverse)
      ;;
    *)
      echo "No supported package manager found"
      return 1
      ;;
  esac

  if [[ -n "$selected_package" ]]; then
    LBUFFER+=" $selected_package"
  fi
  zle reset-prompt
}

# GitHub repository search function
ezy_github_search() {
  # Check if GitHub CLI is installed
  if ! command -v gh &> /dev/null; then
    echo "GitHub CLI (gh) is not installed. Please install it first."
    return 1
  fi

  # Ensure the user is authenticated with GitHub CLI
  if ! gh auth status &> /dev/null; then
    echo "You need to authenticate with GitHub CLI first. Run 'gh auth login'"
    return 1
  fi

  local query=""
  local limit=${EZY_GITHUB_LIMIT:-50}
  local selected_repo
  
  # If we have a partial repo name in the command line, use it as search query
  if [[ -n "$LBUFFER" && "$LBUFFER" =~ "git " ]]; then
    # Extract any text after "git " as the query
    query=$(echo "$LBUFFER" | sed 's/git \(.*\)/\1/')
  fi
  
  selected_repo=$(gh search repos "$query" --limit $limit --json nameWithOwner,url --jq '.[] | "\(.nameWithOwner) \(.url)"' | \
    fzf --height=97% --layout=reverse --preview 'echo {} | cut -d" " -f2 | xargs gh repo view')

  if [[ -n "$selected_repo" ]]; then
    # Get just the URL from the selection
    local repo_url=$(echo "$selected_repo" | awk '{print $NF}')
    
    # If the command already starts with "git clone", just append the URL
    if [[ "$LBUFFER" =~ "git" ]]; then
      # Clear any text after "git clone "
      LBUFFER=$(echo "$LBUFFER" | sed 's/git .*/git /')
      LBUFFER+="$repo_url"
    else
      # Otherwise just add the URL
      LBUFFER+="$repo_url"
    fi
  fi
  zle reset-prompt
}

# Directory search function
ezy_directory_search() {
  local selected_dir
  local dir_command=${EZY_DIR_COMMAND:-"fd --hidden --strip-cwd-prefix --exclude .git"}
  
  # If fd is not available, fall back to find
  if ! command -v fd &> /dev/null; then
    dir_command="find . -type f -not -path '*/\.git/*'"
  fi
  
  local preview=${EZY_DIR_PREVIEW:-"bat --color=always -n --line-range :500 {}"}
  
  # If bat is not available, fall back to cat or less
  if ! command -v bat &> /dev/null; then
    if command -v less &> /dev/null; then
      preview="less {}"
    else
      preview="cat {}"
    fi
  fi
  
  selected_dir=$(eval $dir_command | fzf --height=97% --layout=reverse --preview "$preview")
  
  if [[ -n "$selected_dir" ]]; then
    LBUFFER+=" $selected_dir"
  fi
  zle reset-prompt
}

# Register ZLE widgets
zle -N ezy_package_search
zle -N ezy_github_search
zle -N ezy_directory_search

# Initialize keybindings
# These will be overridden by any settings in the config file
bindkey '^P' ezy_package_search
bindkey '^G' ezy_github_search
bindkey '^T' ezy_directory_search

