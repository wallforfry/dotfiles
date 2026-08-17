# Dotfiles

Mes dotfiles macOS, Linux et NAS Synology, gérés avec
[chezmoi](https://www.chezmoi.io).

## Installation

Ce repo est **privé** : il faut s'authentifier auprès de GitHub avant de
pouvoir le cloner. L'authentification se fait en deux temps — un PAT tiré de
Bitwarden pour le clone initial, puis `gh` pour tous les `chezmoi update`
suivants.

### 1. Récupérer le token

Le PAT est dans Bitwarden, entrée **`github.com`**, champ personnalisé
**`Dotfiles token`**. Si le CLI Bitwarden est disponible :

```bash
export GH_PAT=$(bw get item github.com | jq -r '.fields[] | select(.name == "Dotfiles token").value')
```

`bw get password` ne renverrait que le mot de passe du compte : les champs
personnalisés ne sont accessibles que via `bw get item`, d'où le passage par
`jq`. Déverrouiller le coffre au préalable (`bw unlock`) si nécessaire.

Sinon, copie le champ depuis le coffre web et `export GH_PAT=...` à la main —
c'est la seule option sur une machine sans `bw` ni `jq`, typiquement un NAS.

### 2. Installer

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply "https://${GH_PAT}@github.com/wallforfry/dotfiles.git"
```

chezmoi s'installe, clone ce repo, demande le profil de la machine (`perso`
ou `pro`) et la passphrase de déchiffrement des secrets, puis applique la
configuration. Au passage, un script `run_onchange` installe les outils
manquants (`age`, `gh`, `jq`, `ripgrep`, `starship`) — voir [Outillage](#outillage).

### 3. Basculer sur gh

Le token est pour l'instant inscrit en clair dans la remote du clone. On le
remplace par `gh`, qui gère ensuite les identifiants pour les mises à jour :

```bash
echo "$GH_PAT" | gh auth login --with-token
chezmoi git -- remote set-url origin https://github.com/wallforfry/dotfiles.git
unset GH_PAT
```

`gh auth login` installe un helper d'identifiants git global ; `.gitconfig`
le déclare avec le chemin de `gh` résolu à l'apply, et omet le bloc là où
`gh` n'est pas installé. `chezmoi update` fonctionne dès lors sans rien
saisir.

### Pourquoi pas en SSH ?

`chezmoi init git@github.com:wallforfry/dotfiles.git` fonctionne aussi, mais
suppose une clé SSH déjà utilisable sur la machine neuve. Avec une clé sur
YubiKey, l'agent GPG qui l'expose est configuré par `.zprofile` — que chezmoi
n'a pas encore appliqué à ce stade. Le chemin par token évite cette
dépendance circulaire.

## Outillage

`run_onchange_before_install-tools.sh.tmpl` installe ce qui manque, à chaque
apply où le script a changé :

| Outil | Rôle | Si absent |
|---|---|---|
| `age` | déchiffre `~/.secrets` | secrets inaccessibles (voir dépannage) |
| `gh` | helper d'identifiants git | mot de passe demandé à chaque pull |
| `jq` | filtrage JSON | scripts et agents qui lisent du JSON échouent |
| `ripgrep` (`rg`) | recherche de code | agents et telescope retombent sur `grep` |
| `starship` | prompt | `.zshrc` retombe sur le prompt zsh par défaut |

Sur macOS il passe par Homebrew. Ailleurs il pose des binaires statiques dans
`~/bin`, que `.zshenv` met en tête du `PATH` — `.zshenv` et pas `.zprofile`,
pour que les shells non interactifs (`ssh nas '...'`, planificateur DSM) les
trouvent aussi. Les versions d'`age`, `gh`, `jq` et `ripgrep` sont épinglées en
tête du script ; les modifier suffit à déclencher une réinstallation.

### Pourquoi le repo reste privé

Le rendre public exposerait `encrypted_private_dot_secrets.age` au
téléchargement, donc à une attaque hors ligne sans limite de temps contre la
seule passphrase. `dot_gitconfig.tmpl` et le bloc pro de `dot_zshrc.tmpl`
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
| `pro` | aliases Doppler/REDACTED, `cdb`/`cds`, `claude-REDACTED`, `REDACTED_PROJECT_PATH`, `~/.claude_REDACTED` |

Les blocs macOS (Arc, Homebrew, pnpm, Coursier) ne sont rendus que sur darwin.

L'identité git n'est pas templatée : `~/.gitconfig` bascule déjà entre les deux
identités par répertoire via `includeIf "gitdir:~/Projects/REDACTED/"`.

## Décisions d'architecture

`docs/adr/` consigne pourquoi ce dépôt est ainsi et pas autrement : chezmoi plutôt
qu'un bare repo, `age` en passphrase, DSM comme cible de premier rang, les
fragments ssh, `harness/` comme source unique. Seules les décisions **en vigueur**
y figurent.

[`docs/adr/README.md`](docs/adr/README.md) porte l'index, le format, et surtout le
test qui dit **quand** une ADR se justifie et quand elle ne se justifie pas. La
skill `adr` en applique la procédure.

## Agents IA

Les instructions données aux agents de code sont versionnées ici, plus dans un
`~/.claude` recopié à la main sur chaque machine.

### Source unique

`harness/` contient les trois fichiers d'instructions, agnostiques de l'agent :

| Fichier | Rôle |
|---|---|
| `harness/AGENTS.md` | règles techniques : analyse critique, délégation aux subagents, exigences de vérification, style |
| `harness/SOUL.md` | voix de l'agent : langue, registre, priorités |
| `harness/USER.md` | contexte et attentes de l'utilisateur : acquis, biais techniques, mode de travail |

`dot_claude/{AGENTS,SOUL,USER}.md.tmpl` ne sont que des projections d'une ligne
(`{{ include "harness/…" }}`) et `dot_claude/CLAUDE.md` l'adaptateur Claude qui
les importe. **Éditer `harness/`, jamais les projections.**

`AGENTS.md` à la racine porte les règles propres à ce dépôt (chezmoi, langues,
vérification, secrets) ; `CLAUDE.md` s'y réduit à un `@AGENTS.md`. Aucun des
deux n'est déployé.

### Skills

`dot_claude/skills/<slug>/SKILL.md`, une skill par répertoire, déployées dans
`~/.claude/skills` — voir `dot_claude/skills/README.md` pour l'index et
l'origine des skills reprises d'un dépôt tiers.

### Profil pro

`dot_claude_REDACTED/` ne contient que des liens relatifs vers `~/.claude`, pour
que la session `claude-REDACTED` partage la même source au lieu d'une seconde
copie à synchroniser. Le répertoire est ignoré hors profil `pro`.

### Récupération web

`harness/AGENTS.md` décrit une escalade à trois paliers : fetch intégré, Firecrawl
auto-hébergé, puis CloakBrowser pour l'anti-bot. Les deux derniers sont pilotés par des
scripts de `~/.local/bin`, tous bâtis sur le même principe — **un conteneur nommé,
réutilisé par toutes les sessions**.

| Script | Rôle | Enregistrement MCP |
|---|---|---|
| `firecrawl-mcp` | palier 2, pile de `~/.config/firecrawl/compose.yml` | oui |
| `cloak` | palier 3, navigateur exposé en CDP | non — consommé par Scrapling |
| `scrapling-mcp` | client CDP de `cloak` ; aussi `get`, `fetch`, `screenshot` | oui |
| `postgres-mcp` | hors escalade, même discipline de conteneur | oui |

`stealthy_fetch` de Scrapling est inutilisable : Camoufox est absent de l'image et son
dépôt amont ne publie aucune release, donc son téléchargeur ne résout aucune version.
Voir [ADR-014](docs/adr/014-recuperation-web-par-paliers.md).

Chacun accepte `--stop` et `--status` ; `cloak` ajoute `--start` et `--url`,
`firecrawl-mcp` ajoute `--start`. Sans argument, les trois serveurs MCP parlent
stdio : c'est cette forme qu'on enregistre.

```bash
firecrawl-mcp --start    # une fois : environ 2 Gio d'images à télécharger
```

```bash
claude mcp add --scope user firecrawl -- "$HOME/.local/bin/firecrawl-mcp"
```

```bash
claude mcp add --scope user scrapling -- "$HOME/.local/bin/scrapling-mcp"
```

Aucune pile ne redémarre avec le daemon docker, volontairement — elles ne servent
qu'à la demande, et CloakBrowser s'arrête seul après cinq minutes d'inactivité.
L'API Firecrawl écoute sur `127.0.0.1` uniquement : elle tourne sans
authentification.

`~/.cursor/mcp.json` n'est **pas** versionné — il contient des mots de passe et des
jetons en clair. C'est pourtant lui qui enregistre `postgres-mcp` ; la bascule vers
le script à conteneur nommé y est faite à la main.

### Ce qui n'est pas versionné

`~/.claude/settings.json` reste local : il mêle des chemins absolus écrits par
des installeurs (OpenIsland) et l'état des plugins. Le hook `Stop`
`~/.claude/hooks/agent-handoff` est donc déployé mais **pas enregistré** —
l'ajouter à la main dans `settings.json` :

```bash
jq --arg c "$HOME/.claude/hooks/agent-handoff" '.hooks.Stop += [{hooks: [{type: "command", command: $c}]}]' ~/.claude/settings.json > ~/.cache/settings.json && mv ~/.cache/settings.json ~/.claude/settings.json
```

Ne jamais ajouter l'attribut `exact_` à `dot_claude/` : `~/.claude` contient
l'état vivant des sessions, que chezmoi supprimerait.

## ssh

`~/.ssh/config` n'est **pas** versionné : `coder config-ssh` et les extensions
VS Code/Cursor y réécrivent quatre blocs automatiquement, ce qui produirait un
conflit à chaque `chezmoi diff`. Seuls des fragments le sont, sous
`private_dot_ssh/private_config.d/`.

Ils supposent une ligne présente en tête de `~/.ssh/config`, à ajouter une
fois par machine — en tête, parce que ssh retient la première valeur vue pour
chaque option :

```bash
Include ~/.ssh/config.d/*.conf
```

## Prompt

[starship](https://starship.rs), configuré par `dot_config/starship.toml`. Le
fichier ne redéfinit aucun réglage de module : il réécrit seulement `format`
pour ne garder que les modules utiles (git, langages présents dans
`~/Projects`, `cmd_duration`, `docker_context`, `terraform`…). Le reste du
prompt par défaut est écarté, et l'apparence des modules gardés continue de
suivre les évolutions amont. Ajouter un module = l'ajouter à la chaîne
`format`.

`username`/`hostname` ne s'affichent qu'en SSH ou sous un autre utilisateur,
ce qui donne le nom de la machine distante dans le prompt.

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

## Dépannage

### `chezmoi: .secrets: no identities specified`

`age` n'est pas dans le `PATH`. chezmoi bascule alors sur son implémentation
intégrée, qui ne déchiffre que par identité et jamais par passphrase — d'où
un message qui parle d'identités alors que le problème est l'absence du
binaire. `command -v age` pour confirmer, puis relancer un `chezmoi apply`
une fois `~/bin` dans le `PATH`.

### `fork/exec /tmp/....sh: permission denied`

`/tmp` est monté `noexec` (cas de DSM), donc chezmoi ne peut pas exécuter les
scripts `run_*` qu'il y extrait. La config générée définit `scriptTempDir`
pour les sortir de `/tmp`. Sur une machine installée avant ce réglage, la
config n'est pas régénérée par `chezmoi update` — `chezmoi init` s'en charge,
ou ajouter la clé à la main **avant toute section** du fichier TOML :

```bash
mkdir -p ~/.cache/chezmoi
cd ~/.config/chezmoi && { printf 'scriptTempDir = "%s/.cache/chezmoi"\n' "$HOME"; grep -v '^scriptTempDir' chezmoi.toml; } > chezmoi.toml.new && mv chezmoi.toml.new chezmoi.toml
```

Une clé ajoutée en fin de fichier atterrit dans la table `[data]` et est
ignorée sans avertissement.

### `/usr/local/bin/gh ... : No such file or directory`

Reliquat d'une version du `.gitconfig` où le chemin de `gh` était codé en
dur. Récupérer la version courante du repo, ou neutraliser le helper le temps
de le faire :

```bash
git config --global --unset-all credential."https://github.com".helper
git config --global --unset-all credential."https://gist.github.com".helper
```

### Le PATH ne contient pas `~/bin`

`.zshenv` s'en charge, mais n'est lu qu'au démarrage d'un shell : ouvrir un
nouveau shell après le premier `chezmoi apply`. Si l'apply s'est interrompu
sur une erreur de script, aucun fichier n'a été écrit — les scripts
`run_*_before` s'exécutent avant l'écriture des dotfiles.

### zsh n'est pas le shell de login

Sur DSM, `chsh` n'est pas utilisable et le shell de `/etc/passwd` est de toute
façon réinitialisé aux mises à jour système. Pire, le shell qu'ouvre ssh n'est
**pas** un shell de login (`shopt -q login_shell` renvoie 1, `$0` vaut `/bin/sh`
et non `-sh`) : `~/.profile` n'est donc jamais lu, et aucun dotfile ne peut
servir de point d'accroche.

La bascule se fait donc côté client, par `~/.ssh/config.d/nas.conf` :
`RemoteCommand exec /usr/local/bin/zsh -l`.

Contrepartie du `RemoteCommand` : `ssh nas <commande>`, `scp` et `rsync`
échouent sur ces hôtes. Les contourner avec `-o RemoteCommand=none`.

### Synology (DSM)

Pas de gestionnaire de paquets : `age`, `gh`, `jq`, `ripgrep` et `starship` viennent des
binaires statiques posés dans `~/bin`. Vérifier aussi que le service des
répertoires personnels est activé, sans quoi `$HOME` n'est pas
`/var/services/homes/<user>` et la config chezmoi atterrit ailleurs que là où
elle est relue.
