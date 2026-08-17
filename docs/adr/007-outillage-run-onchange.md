# ADR-007 — Outillage installé par un `run_onchange`, `~/bin` via `.zshenv`

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `ec5bd58` (installation), `473405f` (PATH via `.zshenv`), `f29360b` (zsh)

## Contexte

Trois outils conditionnent le fonctionnement des dotfiles : `age` déchiffre
`~/.secrets` ([ADR-002](002-depot-prive-secrets-age.md)), `gh` sert de helper
d'identifiants ([ADR-003](003-bootstrap-pat-puis-gh.md)), starship fournit le
prompt ([ADR-006](006-starship-comme-prompt.md)). Aucun n'est présent sur une
machine neuve, et le NAS n'a ni Homebrew ni gestionnaire de paquets — l'absence de
starship y produisait une erreur au login.

`.zprofile` avait d'abord porté l'ajout de `~/bin` au `PATH`. Le commit `473405f`
en donne la limite : « `.zprofile` n'est lu que par les shells de login : ni ssh
non interactif ni le planificateur DSM n'y trouvaient chezmoi, starship ou age ».

## Décision

`run_onchange_before_install-tools.sh.tmpl` installe ce qui manque à chaque apply
où son contenu rendu a changé. Il applique trois règles :

- **Gestionnaire de paquets d'abord.** Homebrew sur macOS ; `apt`, `apk`, `dnf` ou
  `pacman` selon ce qui répond. zsh en particulier ne s'installe pas en binaire
  statique : il lui faut son arborescence de fonctions.
- **Binaires statiques ensuite.** Là où rien ne répond, les binaires vont dans
  `~/bin`, l'architecture étant déduite de `uname -m`.
- **Dégradation, jamais échec.** Un outil manquant émet un avertissement nommant
  la conséquence et laisse l'apply se poursuivre.

`~/bin` passe en tête du `PATH` depuis **`.zshenv`**, lu par toute invocation de
zsh, et non `.zprofile`, réservé aux shells de login.

Les versions d'`age` et de `gh` sont épinglées dans le script *et* dans son
en-tête : modifier l'en-tête change le contenu rendu, donc déclenche la
réexécution.

## Conséquences

- Une machine neuve devient utilisable en un `chezmoi init --apply`.
- Un `chezmoi apply` peut télécharger depuis le réseau. Le script reste
  idempotent : chaque effet est gardé par un `command -v`.
- Les binaires posés dans `~/bin` ne sont pas mis à jour par le script tant que sa
  version épinglée ne bouge pas. C'est voulu, mais cela veut dire qu'un `age`
  ancien peut survivre longtemps.
- La règle « jamais fatal » a un revers : une installation ratée ne se remarque
  qu'en lisant les avertissements de l'apply.
- `.zshenv` étant lu par tout shell zsh, y compris non interactif, il ne doit
  contenir que des ajustements de `PATH` et des gardes d'existence — jamais de
  sortie sur stdout, qui casserait `scp` et `rsync`.

## Alternatives écartées

- **Documenter l'installation manuelle** : reporte sur l'opérateur ce qui est
  automatisable, et la première machine oubliée casse au login.
- **Un `run_once_`** : ne rejouerait pas à l'ajout d'un outil ou au changement
  d'une version épinglée.
- **Rendre l'absence d'un outil fatale** : abandonner l'apply au milieu laisse un
  `$HOME` à moitié configuré, pire que l'absence de starship.
- **`~/bin` dans `.zprofile`** : essayé, insuffisant — les shells non interactifs
  et le planificateur DSM n'y voient rien (`473405f`).
