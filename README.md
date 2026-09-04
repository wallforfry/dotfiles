# Dotfiles

Mes dotfiles macOS, Linux et NAS Synology, gérés avec
[chezmoi](https://www.chezmoi.io).

## Installation

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply https://github.com/wallforfry/dotfiles.git
```

chezmoi s'installe, clone ce repo, demande le profil de la machine (`perso` ou
`pro`), puis applique la configuration. Une seule saisie de secret : la
passphrase qui déverrouille la clé `age` de la machine, réclamée une fois par
`run_before_unlock-age-key.sh.tmpl` et plus jamais ensuite
([ADR-018](docs/adr/018-chiffrement-par-paire-de-cles.md)). Au passage, un
script `run_onchange` installe les outils manquants - voir
[Outillage](#outillage).

Sur une machine neuve, `age` n'est pas encore là au moment où la clé devrait
être déverrouillée : le script le signale sans échouer, et un second
`chezmoi apply` suffit, l'outillage étant installé entre-temps.

Le repo est public : le clone ne demande aucune authentification, ni jeton, ni
clé, ni compte GitHub ([ADR-017](docs/adr/017-clone-anonyme-gh-en-ecriture.md)).
Seule l'écriture en exige une, par `gh` :

```bash
gh auth login
```

`gh auth login` installe un helper d'identifiants git global ; `.gitconfig` le
déclare avec le chemin de `gh` résolu à l'apply, et omet le bloc là où `gh`
n'est pas installé. Une machine qui n'a jamais authentifié `gh` peut appliquer
et mettre à jour, pas pousser.

Cloner en SSH plutôt qu'en HTTPS fonctionne, mais suppose une clé déjà
utilisable - ce que la YubiKey ne permet pas avant le premier apply, l'agent qui
l'expose étant configuré par `.zprofile`. ADR-017 détaille cette dépendance
circulaire.

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
| `neovim` (`nvim`) | éditeur, `~/.config/nvim` est déployé partout | LazyVim reste inerte, alias `vim` absent |
| `fd` | recherche de fichiers | telescope retombe sur `find` |
| `duf` | affichage disque | `.zshrc` n'aliase pas `df` |
| `uv` | lance `scrapling-mcp` | le MCP scrapling ne démarre pas |

Sur macOS il passe par Homebrew, et y ajoute `thefuck` (aliasé par `.zshrc`),
`pinentry-mac` (saisie du PIN de la YubiKey,
[ADR-010](docs/adr/010-gpg-agent-ssh-sous-condition.md)), `ykman` et
`bitwarden-cli` (`bw`).

En profil `pro` uniquement, il ajoute `bkt` - la CLI Bitbucket réclamée par la
skill `merge-verdict`, Bitbucket ne servant que là. Sa formule vient d'un tap
tiers, dont Homebrew exige la confiance : le script tente l'installation et,
s'il échoue, imprime le `brew trust --formula` à passer. Accorder cette
confiance appartient à l'opérateur, pas au dépôt.

Ailleurs il pose des binaires statiques dans `~/bin`, que `.zshenv` met en tête
du `PATH` - `.zshenv` et pas `.zprofile`, pour que les shells non interactifs
(`ssh nas '...'`, planificateur DSM) les trouvent aussi. Les versions sont
épinglées en tête du script ; les modifier suffit à déclencher une
réinstallation.

`nvim` fait exception au « un binaire dans `~/bin` » : il lui faut son
`VIMRUNTIME` à côté, donc l'archive amont va dans `~/.local/nvim` et `~/bin`
n'en reçoit qu'un lien. Cette archive est liée à la glibc : sur une Alpine elle
s'extrait et ne s'exécute pas.

### Applications macOS

Sur un poste graphique, le script installe aussi cinq applications par
Homebrew Cask - Arc, Bitwarden, Claude, OrbStack, Warp - dans
`~/Applications` et non `/Applications`. Chacune est sautée si son `.app`
existe déjà dans l'un ou l'autre répertoire, ce qui laisse intacte une
installation posée hors de Homebrew.

« Poste graphique » est une donnée à part, `gui`, demandée à l'init sur macOS
seulement ([ADR-019](docs/adr/019-poste-graphique-et-apps-macos.md)) : la
question ne se pose ni sur le NAS ni dans un conteneur. Une machine installée
avant cette décision n'a pas la clé et n'installe donc rien tant qu'un
`chezmoi init` n'a pas régénéré sa configuration.

### Ce que ce repo public ne contient pas

Le repo est public, et rien de sensible n'y figure - ni en clair, ni en prose.
Ce qui l'est et reste nécessaire au déploiement est chiffré par `age` :
`~/.ssh/config.d/nas.conf`, `~/.config/zsh/pro.zsh`, `~/.config/zsh/pro.zprofile`,
`~/.config/git/pro.gitconfig`, `~/.claude/CONTEXT.md` - ce dernier déployé sur
les deux profils, puisqu'il porte aussi les projets personnels. Les fichiers
publics les chargent sans les nommer, et `.chezmoiignore` écarte les fragments
du profil `pro` hors de ce profil.

La contrepartie est entière : `encrypted_*.age` est téléchargeable par
quiconque, donc attaquable hors ligne sans limite de temps. La passphrase est la
seule barrière - voir [ADR-016](docs/adr/016-depot-public-sensible-chiffre.md).

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

### Vérifier avant de pousser

```bash
bash scripts/verify.sh
```

Syntaxe des scripts, rendu des templates sur les trois combinaisons de profil,
cohérence des skills et de l'index des ADR, absence de nom sensible en clair,
chiffrement des fragments, absence d'attribut `exact_` sur les répertoires
d'état vivant, et sous-répertoires de skill limités à `references/`, `assets/`
et `scripts/`. Sort en 1 au premier contrôle rouge.

```bash
bash scripts/harness-audit.sh
```

Le pendant mesuré, jamais joué en CI : coût du contexte toujours chargé, retard
de la source de déploiement sur `origin/main`, activation réelle de chaque skill
et subagent, taux de violation des deux règles observables avant et après leur
introduction, et pouvoir de détection de `verify.sh` par injection de onze
défauts dans un clone. Sort en 1 si un défaut passe ou si une mesure n'a pas pu
être faite.

`.github/workflows/verify.yml` rejoue ce script sur `push` vers `main`, sur
`pull_request` vers `main` et à la demande, puis fait un vrai `chezmoi apply`
sur `ubuntu-latest` et `macos-latest`, pour les deux profils. Il exige un secret
`AGE_KEY` portant la clé privée `age` : sans elle, l'action de préparation sort
en 1 avant tout contrôle, plutôt que de laisser passer une barrière amputée.

Aucune commande de ce workflow n'écrit le contenu rendu d'une cible : les logs
d'un dépôt public en publieraient le clair. `verify.sh` le vérifie. DSM n'a pas
de runner et reste vérifié à la main
([ADR-020](docs/adr/020-verification-en-ci.md)).

## Secrets

Sept fichiers sont chiffrés par [age](https://age-encryption.org) et versionnés
ici : `~/.secrets`, le fragment ssh du NAS, les trois fragments du profil `pro`,
les contextes nommés de `~/.claude/CONTEXT.md`, et la clé privée elle-même.

Le chiffrement se fait vers une **paire de clés**, pas par passphrase : la clé
privée vit en `~/.config/chezmoi/key.txt`, et le dépôt en porte une copie
chiffrée par passphrase sous `age-key.txt.age`. Un `apply` ne demande donc rien
sur une machine déjà déverrouillée, quel que soit le nombre de fichiers chiffrés
([ADR-018](docs/adr/018-chiffrement-par-paire-de-cles.md)).

**Deux secrets à ne pas perdre** : la passphrase d'`age-key.txt.age` et la clé
`key.txt`. L'un des deux suffit à reconstituer l'accès ; les deux perdus, les
fichiers chiffrés sont définitivement illisibles.

Le repo étant public, `age-key.txt.age` est téléchargeable par quiconque : la
passphrase est la seule barrière, et elle est attaquable hors ligne. Elle doit
être longue et propre à cet usage.

## Profils

`chezmoi init` demande le profil de la machine :

| Profil | Contenu |
|---|---|
| `perso` | base commune uniquement |
| `pro` | fragments chiffrés `~/.config/zsh/pro.zsh`, `pro.zprofile`, `~/.config/git/pro.gitconfig`, `~/.claude_pro`, et `bkt` |

Sur macOS, une seconde question indépendante - `gui` - décide de l'installation
des applications ([Applications macOS](#applications-macos)). Les deux
dimensions se croisent : un poste `pro` et un poste `perso` sont l'un comme
l'autre graphiques, un conteneur ni l'un ni l'autre.

Les blocs macOS (Arc, Homebrew, pnpm, Coursier) ne sont rendus que sur darwin.

L'identité git n'est pas templatée : `~/.gitconfig` bascule déjà entre les deux
identités par répertoire via un `includeIf` porté par le fragment chiffré
`~/.config/git/pro.gitconfig`.

## Décisions d'architecture

`docs/adr/` consigne pourquoi ce dépôt est ainsi et pas autrement : chezmoi plutôt
qu'un bare repo, le dépôt public au sensible chiffré, DSM comme cible de premier rang, les
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
| `harness/AGENTS.md` | règles techniques : analyse critique, délégation aux subagents, exigences de vérification, rédaction et occultation, portée des gardes conditionnels, style |
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
`~/.claude/skills` - voir `dot_claude/skills/README.md` pour l'index et
l'origine des skills reprises d'un dépôt tiers.

### Profil pro

`dot_claude_pro/` ne contient que des liens relatifs vers `~/.claude`, pour
que la session isolée du profil `pro` partage la même source au lieu d'une seconde
copie à synchroniser. Le répertoire est ignoré hors profil `pro`.

### Récupération web

La skill `web-fetching` décrit une escalade à trois paliers : fetch intégré,
Firecrawl auto-hébergé, puis CloakBrowser pour l'anti-bot. `harness/AGENTS.md` n'en
garde que la règle d'usage, qui vaut pour toute tâche. Les deux derniers sont pilotés par des
scripts de `~/.local/bin`, tous bâtis sur le même principe - **un conteneur nommé,
réutilisé par toutes les sessions**.

| Script | Rôle | Enregistrement MCP |
|---|---|---|
| `firecrawl-mcp` | palier 2, pile de `~/.config/firecrawl/compose.yml` | oui |
| `cloak` | palier 3, navigateur exposé en CDP | non - consommé par Scrapling |
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

Aucune pile ne redémarre avec le daemon docker, volontairement - elles ne servent
qu'à la demande, et CloakBrowser s'arrête seul après cinq minutes d'inactivité.
L'API Firecrawl écoute sur `127.0.0.1` uniquement : elle tourne sans
authentification.

`~/.cursor/mcp.json` n'est **pas** versionné - il contient des mots de passe et des
jetons en clair. C'est pourtant lui qui enregistre `postgres-mcp` ; la bascule vers
le script à conteneur nommé y est faite à la main.

### Ce qui n'est pas versionné

`~/.claude/settings.json` reste local, et le restera : il mêle des chemins
absolus écrits par des installeurs tiers, l'état des plugins, une `statusLine`
et des hooks venus d'ailleurs. Le déployer l'écraserait.

Le hook `Stop` `agent-handoff` y est en revanche **enregistré
automatiquement** par `run_onchange_after_register-claude-hooks.sh.tmpl`, qui
fusionne cette seule entrée dans le fichier vivant et laisse le reste intact.
L'idempotence porte sur le chemin du hook, pas sur sa position : Claude Code
réordonne les entrées et d'autres outils en insèrent. Une sauvegarde
`settings.json.bak` est déposée avant toute écriture, et un rendu jq invalide
laisse le fichier inchangé.

Ne jamais ajouter l'attribut `exact_` à `dot_claude/` : `~/.claude` contient
l'état vivant des sessions, que chezmoi supprimerait.

## ssh

`~/.ssh/config` n'est **pas** versionné : `coder config-ssh` et les extensions
VS Code/Cursor y réécrivent quatre blocs automatiquement, ce qui produirait un
conflit à chaque `chezmoi diff`. Seuls des fragments le sont, sous
`private_dot_ssh/private_config.d/`.

Ils supposent une ligne présente en tête de `~/.ssh/config`, à ajouter une
fois par machine - en tête, parce que ssh retient la première valeur vue pour
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

## Dépannage

### Un fichier chiffré reste illisible

`~/.config/chezmoi/key.txt` est absent : sans la clé privée, rien ne se
déchiffre. Sur une machine neuve, c'est le cas normal du premier `apply`, où
`age` n'était pas encore installé quand `run_before_unlock-age-key.sh.tmpl` a
voulu déverrouiller la clé - le script le signale sans échouer. Vérifier
`command -v age`, puis relancer `chezmoi apply` : la passphrase est demandée,
la clé posée, et l'apply reprend.

Si la passphrase est refusée, c'est celle d'`age-key.txt.age` qui est attendue,
pas celle d'un autre coffre. Le déchiffrement manuel donne le même diagnostic
sans passer par chezmoi :

```bash
age --decrypt -o ~/.config/chezmoi/key.txt "$(chezmoi source-path)/age-key.txt.age"
```

### `fork/exec /tmp/....sh: permission denied`

`/tmp` est monté `noexec` (cas de DSM), donc chezmoi ne peut pas exécuter les
scripts `run_*` qu'il y extrait. La config générée définit `scriptTempDir`
pour les sortir de `/tmp`. Sur une machine installée avant ce réglage, la
config n'est pas régénérée par `chezmoi update` - `chezmoi init` s'en charge,
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
sur une erreur de script, aucun fichier n'a été écrit - les scripts
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
