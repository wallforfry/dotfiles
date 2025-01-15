# Added by OrbStack: command-line tools and integration
source ~/.orbstack/shell/init.zsh 2>/dev/null || :

# >>> coursier install directory >>>
export PATH="$PATH:/Users/delevacw/Library/Application Support/Coursier/bin"
# <<< coursier install directory <<<

# Created by `pipx` on 2024-10-10 17:33:49
export PATH="$PATH:/Users/delevacw/.local/bin"

# Flutter
export PATH="$PATH:$HOME/bin/flutter/bin"
export PATH="$PATH:$HOME/platform-tools/"
export PATH="$PATH":"$HOME/bin/flutter/.pub-cache/bin"

# Volta
export VOLTA_HOME="$HOME/.volta"
export PATH="$VOLTA_HOME/bin:$PATH"
export VOLTA_FEATURE_PNPM=1

# Golang
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"

# GPG
export GPG_TTY=$(tty)
export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
gpgconf --launch gpg-agent


# Bellman
export DOPPLER_LOCAL_CONFIG=local_wall
export BELLMAN_PROJECT_PATH=~/Projects/bellman/

# Deno
export DENO_INSTALL="$HOME/.deno"
export PATH="$PATH:$DENO_INSTALL/bin"
. ~/.deno/env

# Bun
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"

# Java
export SDKMAN_DIR="$HOME/.sdkman"

# PNPM
export PNPM_HOME="/Users/delevacw/Library/pnpm"
case ":$PATH:" in
  *":$PNPM_HOME:"*) ;;
  *) export PATH="$PNPM_HOME:$PATH" ;;
esac


