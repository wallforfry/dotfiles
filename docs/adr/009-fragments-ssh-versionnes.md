# ADR-009 - Fragments SSH versionnés, `~/.ssh/config` non

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `b11059c` (fragment `nas.conf`)

## Contexte

La configuration ssh du NAS doit suivre les machines : `RemoteCommand` y est le
seul point d'accroche pour forcer zsh ([ADR-008](008-dsm-cible-de-premier-rang.md)).

Mais `~/.ssh/config` n'est pas un fichier dont on est propriétaire :
`coder config-ssh` et les extensions VS Code et Cursor y réécrivent des blocs en
permanence. Le versionner produirait un conflit à chaque `chezmoi diff`, et le
premier `chezmoi apply` distrait effacerait la configuration écrite par ces outils.

## Décision

`~/.ssh/config` reste **non versionné**. Seuls des fragments le sont, sous
`private_dot_ssh/private_config.d/`, déployés vers `~/.ssh/config.d/*.conf` avec
des permissions restreintes.

`~/.ssh/config` les charge par une ligne `Include ~/.ssh/config.d/*.conf`, ajoutée
à la main une fois, **avant tout bloc `Host`** : ssh retient la première valeur
rencontrée pour chaque option, un `Include` placé après serait sans effet sur les
options déjà définies.

## Conséquences

- Ce qui nous appartient est versionné, ce qui appartient à d'autres outils est
  laissé tranquille. Aucun conflit récurrent au diff.
- La ligne `Include` est une étape manuelle sur chaque machine neuve, non
  automatisable sans reprendre la propriété du fichier - c'est-à-dire sans annuler
  la décision.
- L'ordre du `Include` est une condition de correction silencieuse : placé trop
  bas, les fragments sont chargés mais ignorés, sans erreur.
- Un nouveau besoin de configuration ssh se traduit par un nouveau fragment, pas
  par une modification d'un fragment existant : `nas.conf` reste lisible comme la
  configuration d'un hôte.

## Alternatives écartées

- **Versionner `~/.ssh/config` entier** : conflit à chaque diff, et risque
  d'effacer les blocs écrits par `coder` et les extensions d'éditeur.
- **`~/.ssh/config` en template avec un bloc figé pour les outils tiers** :
  suppose de connaître et de reproduire ce que ces outils écrivent, qui change sans
  préavis.
- **Passer les options en ligne de commande ou par alias shell** : `RemoteCommand`
  ne servirait alors que les invocations passant par l'alias, pas `ssh` direct ni
  les clients tiers.
