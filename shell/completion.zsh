#compdef ezysearch

local -a opts
opts=(
  '--help[Show help message]'
  '--version[Show version information]'
  '--install[Install ezysearch]'
  '-h[Show help message]'
  '-v[Show version information]'
)

_arguments -s $opts