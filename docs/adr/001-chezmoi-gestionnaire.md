# ADR-001 - chezmoi comme gestionnaire de dotfiles

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `97a18fc` (initial commit)

## Contexte

La configuration vivait dans un bare repo `~/.dotfiles` avec `$HOME` pour arbre
de travail. Ce montage ne sait rien faire d'autre que copier des fichiers
identiques partout : il n'a ni variables, ni conditions par machine, ni
chiffrement. Or les postes visés divergent - macOS avec Homebrew, Linux avec un
gestionnaire de paquets, NAS Synology sans ni l'un ni l'autre - et la
configuration contient des secrets.

Les contournements possibles dans un bare repo (fichiers `.local` non versionnés,
branches par machine, scripts de post-installation) déplacent le problème sans le
résoudre : ce qui diffère cesse d'être versionné, ou cesse d'être partagé.

## Décision

chezmoi gère la configuration. Le dépôt est la source, `$HOME` la destination, et
un chemin source détermine sa destination : `dot_<nom>` vers `~/.<nom>`, les
attributs `private_`, `encrypted_`, `symlink_`, `executable_`, `run_onchange_`
portant le reste.

La bascule depuis `~/.dotfiles` a été assurée par un script idempotent et
interruptible, dont l'étape destructive venait en dernier et n'était atteinte
qu'après validation du shell. Toutes les machines étant migrées, il a été retiré
du dépôt - relisible par `git log -- scripts/`.

## Conséquences

- Templates (`.tmpl`), données par machine et chiffrement natif deviennent
  disponibles : ce qui diffère reste versionné.
- Un fichier ne s'édite plus en place. `chezmoi edit` ou une édition de la source
  suivie de `chezmoi apply` - modifier la destination directement expose à la
  perdre au prochain apply.
- La source est un clone distinct (`~/.local/share/chezmoi`) du dépôt de travail :
  un commit poussé ne prend effet qu'après y avoir été récupéré.
- Deux couches à comprendre au lieu d'une : le rendu du template, puis le
  déploiement. `chezmoi diff` et `chezmoi execute-template` sont les seuls moyens
  d'inspecter la première.

## Alternatives écartées

- **Conserver le bare repo** : aucun mécanisme pour ce qui diffère par machine ni
  pour les secrets, ce qui est précisément le besoin.
- **GNU Stow** : ne gère que des symlinks, ni templates ni chiffrement.
- **yadm** : bare repo outillé, avec templates et chiffrement - mais un modèle de
  fichiers alternatifs par classe de machine moins expressif que les données
  chezmoi, pour un projet nettement moins actif.
- **Nix home-manager** : répond au besoin et à bien davantage, au prix d'un
  gestionnaire de paquets complet à installer sur chaque cible. Impraticable sur
  DSM, qui est une cible de premier rang ([ADR-008](008-dsm-cible-de-premier-rang.md)).
