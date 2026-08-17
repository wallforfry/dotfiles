# ADR-006 — starship comme prompt unique

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `ec5bd58` (installation et config versionnée)

## Contexte

Le prompt doit être identique sur macOS, Linux et le NAS. Un prompt écrit en zsh
suit le shell ; un prompt fourni par un thème oh-my-zsh suit oh-my-zsh, dont les
mises à jour ne sont pas maîtrisées ([ADR-005](005-oh-my-zsh-en-external.md)).

Le NAS n'avait pas starship, ce qui produisait une erreur au login — le commit
`ec5bd58` traite les deux sujets ensemble : l'installation et la forme de la
configuration.

## Décision

starship fournit le prompt, initialisé en fin de `.zshrc` sous condition de
présence :

```sh
command -v starship >/dev/null && eval "$(starship init zsh)"
```

`~/.config/starship.toml` est versionné, **réduit aux modules réellement
utilisés** plutôt que figé sur le dump complet des valeurs par défaut.

Le garde-fou est explicite : starship absent, `.zshrc` retombe sur le prompt zsh
par défaut sans erreur.

## Conséquences

- Un seul prompt, une seule configuration, quelle que soit la machine.
- Une configuration réduite continue de suivre les évolutions amont : les modules
  non déclarés prennent les défauts du jour, y compris leurs améliorations. Un
  dump complet les aurait gelés à la version d'écriture.
- Contrepartie : une évolution amont peut changer l'apparence du prompt sans
  qu'aucun fichier du dépôt n'ait bougé.
- starship devient une dépendance d'installation, mais facultative — d'où son
  traitement en dégradation dans le script d'outillage
  ([ADR-007](007-outillage-run-onchange.md)).

## Alternatives écartées

- **Un thème oh-my-zsh** (`robbyrussell` seul, powerlevel10k) : lie le prompt au
  cycle de vie d'oh-my-zsh, et powerlevel10k impose ses polices.
- **Un prompt écrit à la main en zsh** : portable, mais tout est à écrire et à
  maintenir — segments git, statut, durée.
- **Dump complet de `starship.toml`** : configuration exhaustive et auto-documentée,
  au prix du gel de tous les défauts et d'un fichier illisible au diff.
