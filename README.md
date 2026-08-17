# Dotfiles

Mes dotfiles NixOS et macOS, gérés avec [chezmoi](https://www.chezmoi.io).

## Installation

Ce repo est **privé** : il faut s'authentifier auprès de GitHub avant de
pouvoir le cloner.

```bash
# 1. authentification (macOS ; sur Linux, installer gh via le gestionnaire local)
brew install gh && gh auth login

# 2. installation
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply wallforfry
```

`gh auth login` configure le helper d'identifiants git globalement, ce qui
suffit à chezmoi pour cloner en HTTPS.

chezmoi s'installe, clone ce repo, demande le profil de la machine
(`perso` ou `pro`) et la passphrase de déchiffrement des secrets, puis
applique la configuration.

### Pourquoi pas en SSH ?

`chezmoi init git@github.com:wallforfry/dotfiles.git` fonctionne aussi, mais
suppose une clé SSH déjà utilisable sur la machine neuve. Avec une clé sur
YubiKey, l'agent GPG qui l'expose est configuré par `.zprofile` — que chezmoi
n'a pas encore appliqué à ce stade. Le chemin `gh` évite cette dépendance
circulaire.

### Pourquoi le repo reste privé

Le rendre public exposerait `encrypted_private_dot_secrets.age` au
téléchargement, donc à une attaque hors ligne sans limite de temps contre la
seule passphrase. `dot_gitconfig` et le bloc pro de `dot_zshrc.tmpl`
contiennent par ailleurs des informations d'employeur (noms de projets et de
configurations Doppler) qui n'ont pas à être publiées.

## Usage

### Modifier un dotfile

```bash
chezmoi edit ~/.zshrc     # édite la source, applique à la sortie de l'éditeur
chezmoi cd                # ouvre un shell dans la source
git commit -am "..." && git push
```

### Ajouter un fichier

```bash
chezmoi add ~/.config/foo/config.toml
chezmoi add --encrypt ~/.un-fichier-sensible
```

### Récupérer les changements des autres machines

```bash
chezmoi update
```

### Voir ce qui va changer

```bash
chezmoi diff
chezmoi status
```

## Secrets

`~/.secrets` est chiffré avec [age](https://age-encryption.org) (passphrase
symétrique) et versionné sous `encrypted_private_dot_secrets.age`. La
passphrase n'est stockée nulle part dans ce repo : elle est saisie à
`chezmoi apply`. **Perdue, les secrets sont irrécupérables.**

## Profils

`chezmoi init` demande le profil de la machine :

| Profil | Contenu |
|---|---|
| `perso` | base commune uniquement |
| `pro` | aliases Doppler/REDACTED, `cdb`/`cds`, `claude-REDACTED`, `REDACTED_PROJECT_PATH` |

Les blocs macOS (Arc, Homebrew, pnpm, Coursier) ne sont rendus que sur darwin.

L'identité git n'est pas templatée : `~/.gitconfig` bascule déjà entre les deux
identités par répertoire via `includeIf "gitdir:~/Projects/REDACTED/"`.

## oh-my-zsh

Fourni par `.chezmoiexternal.toml` sous forme d'archive, rafraîchi tous les
7 jours par `chezmoi update`. Son auto-update interne est désactivé
(`zstyle ':omz:update' mode disabled`) pour éviter la désynchronisation.

## Migration depuis l'ancien bare repo

Les machines encore sur l'ancien montage (bare repo `~/.dotfiles` + alias
`dotfiles`) basculent avec :

```bash
./scripts/migrate-from-bare.sh --dry-run   # inspection, ne modifie rien
./scripts/migrate-from-bare.sh             # bascule
```

Le bare repo n'est supprimé qu'après validation du démarrage du shell, et un
backup horodaté est conservé dans `~/.dotfiles-backup-<stamp>`.
