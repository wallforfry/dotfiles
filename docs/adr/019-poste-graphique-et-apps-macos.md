# ADR-019 - Les applications macOS derrière une donnée `gui`, dans `~/Applications`

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : à compléter au commit

## Contexte

[ADR-007](007-outillage-run-onchange.md) fixait un critère net pour le script
d'installation : un outil y entre parce qu'un fichier déployé le réclame. Un
inventaire de cette machine a montré que le critère était mal tenu -
`~/.config/nvim` est déployé partout sans que rien n'installe `nvim`,
`.zshrc` fait l'alias `df=duf` et l'`eval` de `thefuck`, `dot_local/bin/scrapling-mcp`
appelle `uv`, et [dot_zprofile.tmpl](../../dot_zprofile.tmpl) source
`~/.orbstack/shell/init.zsh` sans qu'OrbStack soit installé.

OrbStack, lui, n'est pas un binaire : c'est une application. Étendre le script
aux applications pose deux questions que le critère d'ADR-007 ne tranche pas.

**Où les installer.** L'opérateur les place dans `~/Applications`. C'est déjà
posé pour les shells interactifs par `dot_zshrc.tmpl` -
`HOMEBREW_CASK_OPTS="--appdir=~/Applications"` - mais un `run_onchange` tourne
sous `sh` depuis chezmoi et ne lit aucun fichier de démarrage.

**Sur quelles machines.** Aucune des deux dimensions existantes ne répond : un
poste `pro` et un poste `perso` sont tous deux graphiques, tandis qu'un
conteneur et le NAS ne le sont ni l'un ni l'autre. Le profil ne peut donc pas
porter cette information.

## Décision

Une donnée `gui`, distincte de `profile`, commande l'installation des
applications. Elle est demandée par `promptBoolOnce` **uniquement sur darwin**,
avec `true` par défaut, et vaut `false` partout ailleurs - une machine sans
interface graphique ne voit jamais la question.

Les casks sont installés avec `--appdir="$HOME/Applications"` passé
explicitement sur la ligne de commande, et chacun est sauté si son `.app`
existe déjà dans `~/Applications` **ou** dans `/Applications`, ce qui laisse
intacte une application posée hors de Homebrew.

La liste est courte et nommée : `arc`, `bitwarden`, `claude`, `orbstack`,
`warp`. Elle n'est pas le miroir de `brew list --cask`.

Le script lit la donnée par `get . "gui"` et non `.gui`, parce qu'une machine
installée avant cette décision a un `chezmoi.toml` sans la clé, sur laquelle
`.gui` ferait échouer l'apply entier.

## Conséquences

- Un `chezmoi init --apply` sur un mac neuf pose aussi le navigateur, le
  terminal et le gestionnaire de mots de passe, pas seulement les binaires.
- Une question de plus à l'init, sur macOS seulement.
- Les machines déjà installées gardent un `chezmoi.toml` sans `gui` : elles
  n'installent aucune application jusqu'à un `chezmoi init` qui régénère la
  configuration. C'est une dégradation silencieuse, cohérente avec la règle
  « dégradation, jamais échec » d'ADR-007, mais silencieuse tout de même.
- La liste des casks est une décision manuelle : une application installée à la
  main sur un poste ne se propage pas aux autres tant qu'elle n'est pas ajoutée
  ici.
- Le doublon de valeur `~/Applications` entre `dot_zshrc.tmpl` et ce script est
  assumé : les deux contextes ne partagent aucun fichier lisible par les deux.

## Alternatives écartées

- **Déduire de `.chezmoi.os == "darwin"`** : zéro question, mais aucune
  échappatoire pour un mac de CI ou un mac dont la politique d'entreprise gère
  ses applications elle-même, où le script installerait des doublons.
- **Ajouter un troisième profil `desktop`** : `profile` est un choix unique, ce
  qui obligerait à croiser les valeurs (`pro-desktop`, `perso-desktop`) et
  ferait grossir la liste de façon multiplicative pour deux dimensions
  indépendantes.
- **`HOMEBREW_CASK_OPTS` exporté depuis le script** : dépend du fait que
  Homebrew développe le `~` de la valeur, quand `--appdir` reçoit un chemin
  déjà absolu.
- **Un `Brewfile` et `brew bundle`** : rendrait chaque machine identique à
  celle-ci, `bochs`, `dump1090-fa` et `sl` compris, et remplacerait un choix
  explicite par un instantané.
- **Forcer `/Applications`** : la croyance qu'OrbStack l'exige est fausse - son
  cask interpole `#{appdir}` dans l'artefact `app` comme dans les binaires
  `orb` et `orbctl`. Rien ne justifie l'exception.
