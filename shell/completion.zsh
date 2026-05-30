#compdef ezysearch

local -a opts
opts=(
  '--help[Show help message]'
  '--version[Show version information]'
  '--install[Install ezysearch]'
  '--config[Write the default configuration file]'
  '--config-path[Print the configuration file path]'
  '--print-config[Print the active configuration]'
  '--default-config[Print the default configuration]'
  '--doctor[Check configuration and optional tools]'
  '--check[Check configuration and optional tools]'
  '--completion[Print shell completions]:shell:(bash zsh)'
  '--completions[Print shell completions]:shell:(bash zsh)'
  '--list-package-managers[List supported package managers]'
  '--package-manager[Use a package manager instead of auto-detect]:manager:(auto yay pacman apt brew homebrew dnf zypper)'
  '--manager[Alias for --package-manager]:manager:(auto yay pacman apt brew homebrew dnf zypper)'
  '--auto[Auto-detect package manager]'
  '--yay[Use yay packages]'
  '--pacman[Use pacman packages]'
  '--apt[Use apt packages]'
  '--brew[Use Homebrew packages]'
  '--homebrew[Use Homebrew packages]'
  '--hombrew[Use Homebrew packages]'
  '--dnf[Use dnf packages]'
  '--zypper[Use zypper packages]'
  '-h[Show help message]'
  '-v[Show version information]'
  '-V[Show version information]'
)

_arguments -s $opts
