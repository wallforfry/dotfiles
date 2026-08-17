# ADR-004 — Profils `perso` et `pro` demandés à l'init

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `b286cad` (initial commit)

## Contexte

Une partie de la configuration n'a de sens que sur une machine professionnelle :
alias Doppler, raccourcis `cdb`/`cds`, `BELLMAN_PROJECT_PATH`, session Claude
isolée. La déployer partout pollue les machines personnelles et publie du contexte
d'employeur là où il n'a rien à faire.

Cette distinction ne se déduit ni du système d'exploitation ni du nom d'hôte : les
deux profils tournent sur macOS, et les noms d'hôtes ne suivent aucune convention
exploitable.

## Décision

`.chezmoi.toml.tmpl` demande le profil à l'init, parmi `perso` et `pro`, avec
`promptChoiceOnce` — la valeur est mémorisée dans la configuration locale et n'est
plus redemandée.

Le profil pilote deux mécanismes :

- des blocs `{{ if eq .profile "pro" }}` dans les templates, pour un fragment de
  fichier ;
- des entrées conditionnelles dans `.chezmoiignore`, lui-même un template, pour un
  fichier ou un répertoire entier — c'est ainsi que `.claude_septeo` reste absent
  hors profil `pro`.

Les blocs propres à un système restent portés par `.chezmoi.os`, indépendamment du
profil.

## Conséquences

- Un même dépôt sert les deux mondes, sans branche ni fichier local non versionné.
- Le profil se choisit à l'init et n'est plus posé ensuite. En changer suppose
  d'éditer `~/.config/chezmoi/chezmoi.toml` à la main, ou de relancer
  `chezmoi init`.
- Tout template qui lit `.profile` doit être vérifié pour les deux valeurs, pas
  seulement celle de la machine courante : `chezmoi execute-template` avec une
  configuration alternative est le seul moyen de le faire sans changer de poste.
- Le nombre de combinaisons à vérifier est le produit des profils par les systèmes.
  Un troisième profil doublerait ce coût — raison de s'en tenir à deux.

## Alternatives écartées

- **Détection par nom d'hôte ou par domaine** : aucune convention de nommage
  fiable, et une machine renommée changerait silencieusement de configuration.
- **Répertoires `hosts/` ou `profiles/`** : catégorie supplémentaire à maintenir
  alors que les données chezmoi et `.chezmoiignore` suffisent.
- **Deux dépôts, ou deux branches** : la base commune diverge dès le premier
  commit oublié d'un côté.
- **Fichiers locaux non versionnés** (`~/.zshrc.local`) : ce qui diffère cesse
  d'être versionné, c'est-à-dire le défaut que chezmoi était censé corriger
  ([ADR-001](001-chezmoi-gestionnaire.md)).
