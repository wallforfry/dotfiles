# ADR-005 - oh-my-zsh en external, auto-update désactivé

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `97a18fc` (initial commit)

## Contexte

`.zshrc` charge oh-my-zsh depuis `~/.oh-my-zsh`. Ce répertoire est un dépôt git
amont de plusieurs milliers de fichiers. Le versionner dans les dotfiles noie
l'historique du dépôt sous du code tiers ; le laisser s'installer à la main sur
chaque machine le sort du périmètre géré.

oh-my-zsh se met à jour tout seul, en `git pull` dans son propre répertoire. Sous
gestion chezmoi, cette mise à jour modifie une destination que chezmoi croit
connaître : l'état de la source et celui de la destination divergent sans que rien
ne le signale.

## Décision

oh-my-zsh est déclaré comme external dans `.chezmoiexternal.toml` : archive de la
branche `master`, `exact = true`, `stripComponents = 1`,
`refreshPeriod = "168h"`.

Son auto-update est désactivé dans `.zshrc` par
`zstyle ':omz:update' mode disabled`. Les mises à jour passent par
`chezmoi update`, qui rafraîchit l'external selon sa période.

Le thème reste `robbyrussell` et la liste des plugins se limite à `git`.

## Conséquences

- Le dépôt reste à la taille de la configuration, pas à celle d'oh-my-zsh.
- Une seule voie de mise à jour, donc un seul endroit où quelque chose peut mal
  tourner.
- L'archive suit `master` sans épingler de version : un `chezmoi update` peut
  ramener une régression amont, sans lockfile pour revenir en arrière. La période
  de 168 h borne la fréquence d'exposition, pas le risque.
- `exact = true` fait de `~/.oh-my-zsh` un répertoire dont chezmoi supprime tout
  contenu non déclaré. Une modification locale y est perdue au prochain apply :
  personnaliser passe par `.zshrc`, jamais par le répertoire.

## Alternatives écartées

- **Versionner oh-my-zsh** (sous-module ou copie) : quelques milliers de fichiers
  tiers dans l'historique, pour un code qu'on ne modifie pas.
- **Laisser l'auto-update actif** : deux mécanismes écrivant dans le même
  répertoire, dont un que chezmoi ignore, donc une divergence silencieuse.
- **Se passer d'oh-my-zsh** : le prompt vient déjà de starship
  ([ADR-006](006-starship-comme-prompt.md)), mais la complétion et les raccourcis
  du plugin `git` restent la raison de le garder.
