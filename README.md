# Dotfiles

These are my NixOS and macOS dotfiles

## Installation

```bash
curl https://raw.githubusercontent.com/wallforfry/dotfiles/main/scripts/config-init | bash
```

It will move your existing files in `~/.dotfiles-backup` directory


## Usage

### Edit dotfiles

```bash
dotfiles add ~/.vimrc
dotfiles commit -m "Edit .vimrc"
dotfiles push
```

### Pull dotfiles

```bash
dotfiles pull
```

### Show dotfiles status
```bash
dotfiles status
```

### Don't want to see untracked files
```bash
dotfiles config status.showUntrackedFiles no
```
