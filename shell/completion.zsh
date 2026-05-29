#compdef ezysearch

local -a opts
opts=(
  '--help[Show help message]'
  '--version[Show version information]'
  '--install[Install ezysearch]'
  '--config[Generate default configuration file]'
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
)

_arguments -s $opts
