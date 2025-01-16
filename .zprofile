# Added by OrbStack: command-line tools and integration
source ~/.orbstack/shell/init.zsh 2>/dev/null || :

# >>> coursier install directory >>>
export PATH="$PATH:/Users/delevacw/Library/Application Support/Coursier/bin"
# <<< coursier install directory <<<

# Created by `pipx` on 2024-10-10 17:33:49
export PATH="$PATH:/Users/delevacw/.local/bin"

# Flutter
if [[ -s "$HOME/bin/flutter/bin/flutter" ]]; then
  export PATH="$PATH:$HOME/bin/flutter/bin"
  export PATH="$PATH:$HOME/platform-tools/"
  export PATH="$PATH":"$HOME/bin/flutter/.pub-cache/bin"
  echo "🚀  Flutter path added"
fi

# Volta
if command -v volta &> /dev/null; then
  export VOLTA_HOME="$HOME/.volta"
  export PATH="$VOLTA_HOME/bin:$PATH"
  export VOLTA_FEATURE_PNPM=1
  echo "🚀  Volta path added"
fi

# Golang
if [[ -s "/usr/local/bin/go" ]]; then
  export GOPATH="$HOME/go"
  export PATH="$GOPATH/bin:$PATH"
  echo "🚀  Golang path added"
fi

# GPG
if command -v gpg-agent &> /dev/null; then
  export GPG_TTY=$(tty)
  export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
  gpgconf --launch gpg-agent
  echo "🔑  GPG agent started"
fi

# Bellman
if [[ -s "$HOME/Projects/bellman/.env" ]]; then
  export BELLMAN_PROJECT_PATH=~/Projects/bellman/
  export DOPPLER_LOCAL_CONFIG=local_wall
  export $(cat ~/Projects/bellman/.env | xargs)
  echo "🔔  Bellman environment variables loaded"
fi

# Deno
if [[ -s "$HOME/.deno/env" ]]; then
  export DENO_INSTALL="$HOME/.deno"
  export PATH="$DENO_INSTALL/bin:$PATH"
  . ~/.deno/env
  echo "🚀  Deno path added"
fi

# Bun
if [[ -s "$HOME/.bun/_bun" ]]; then
  export BUN_INSTALL="$HOME/.bun"
  export PATH="$BUN_INSTALL/bin:$PATH"
  echo "🚀  Bun path added"
fi

# Java
if [[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ]]; then
  export SDKMAN_DIR="$HOME/.sdkman"
  echo "🚀  SDKMAN path added"
fi

# PNPM
export PNPM_HOME="/Users/delevacw/Library/pnpm"
case ":$PATH:" in
  *":$PNPM_HOME:"*) ;;
  *) export PATH="$PNPM_HOME:$PATH" ;;
esac

# Postgres
if [[ -s "/usr/local/opt/postgresql@16/bin" ]]; then
  export PATH="/usr/local/opt/postgresql@16/bin:$PATH"
  echo "🚀  Postgres path added"
fi

# HomeBrew
export HOMEBREW_NO_AUTO_UPDATE=1
