# ADR-017 - Clone anonyme, `gh` comme helper d'écriture

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : à renseigner après la réécriture d'historique, qui change tous les SHA

## Contexte

ADR-003 - que celle-ci remplace, donc supprimée du répertoire, relisible par
`git log -- docs/adr/` - organisait
l'installation en deux temps parce que le dépôt était privé : un jeton d'accès
personnel tiré de Bitwarden servait au clone, puis `gh` prenait le relais comme
helper d'identifiants. La clé SSH ne pouvait pas servir au bootstrap : elle vit
sur une YubiKey exposée par l'agent GPG, lui-même configuré par `.zprofile`, que
chezmoi n'a pas encore déployé - la dépendance est circulaire.

Le dépôt est désormais public ([ADR-016](016-depot-public-sensible-chiffre.md)) :
`chezmoi init` en HTTPS réussit sans aucune authentification. Le jeton n'a plus de
rôle au clone. L'écriture, elle, en exige toujours une.

## Décision

Le clone initial est anonyme :

```sh
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply https://github.com/wallforfry/dotfiles.git
```

Aucun jeton, aucune étape de réécriture de remote. L'authentification n'intervient
qu'à la première écriture, par `gh` comme helper d'identifiants, déclaré dans
`dot_gitconfig.tmpl` avec le chemin de `gh` résolu à l'apply par `lookPath` - un
chemin codé en dur fait échouer git sur chaque opération réseau des machines où le
binaire est ailleurs. Le bloc entier est omis là où `gh` est absent.

## Conséquences

- Une machine neuve n'a besoin de rien : ni jeton, ni navigateur, ni copier-coller.
  Le NAS et les conteneurs y gagnent le plus.
- Le PAT `Dotfiles token` de Bitwarden n'a plus d'usage ici, et la documentation
  de bootstrap perd son étape la plus fragile - celle où le jeton apparaissait en
  clair dans la remote du clone.
- Une machine qui n'a jamais authentifié `gh` peut appliquer et mettre à jour, mais
  pas pousser. C'est le comportement voulu : les machines de consultation n'ont
  pas à pouvoir écrire.
- `gh` reste une dépendance des écritures, d'où sa présence dans le script
  d'outillage ([ADR-007](007-outillage-run-onchange.md)).

## Alternatives écartées

- **Cloner en SSH** (`git@github.com:…`) : suppose une clé déjà utilisable, que la
  YubiKey et l'agent GPG rendent indisponible avant le premier apply. C'est la
  dépendance circulaire d'origine, inchangée par le passage au public.
- **Conserver le bootstrap par PAT**, la décision précédente,
  ADR-003 : un secret de longue durée manipulé à
  la main pour un clone qui n'en demande plus.
- **Conserver le PAT comme helper permanent**, reprise d'ADR-003 : un secret de
  longue durée en clair dans la configuration git, sans rotation, pour un service
  que `gh` rend déjà.
- **`gh auth login` avant le clone**, reprise d'ADR-003 : reviendrait à installer
  `gh` à la main d'abord, alors que c'est justement ce que le dépôt installe.
