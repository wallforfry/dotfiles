# ADR-003 — Bootstrap par PAT puis bascule sur `gh`

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `16c9f98`, `91c2a25` (documentation), `1a50cb4` (chemin de `gh` non codé en dur)

## Contexte

Le dépôt est privé ([ADR-002](002-depot-prive-secrets-age.md)) : `chezmoi init`
échoue sans authentification. Sur une machine neuve, rien n'est encore installé —
ni clé SSH utilisable, ni `gh`, ni helper d'identifiants.

La clé SSH n'est pas une option de bootstrap : elle vit sur une YubiKey exposée
par l'agent GPG, lui-même configuré par `.zprofile`, que chezmoi n'a pas encore
déployé. La dépendance est circulaire.

## Décision

L'installation se fait en deux temps.

1. Un jeton d'accès personnel, tiré de Bitwarden (entrée `github.com`, champ
   personnalisé `Dotfiles token`), sert au clone initial : `chezmoi init --apply
   "https://${GH_PAT}@github.com/…"`.
2. Le jeton est ensuite remplacé par `gh` comme helper d'identifiants, et la
   remote réécrite sans lui.

`dot_gitconfig.tmpl` déclare le helper avec le chemin de `gh` résolu à l'apply par
`lookPath`, et omet le bloc entier là où `gh` est absent — un chemin codé en dur
fait échouer git sur chaque opération réseau des machines où le binaire est
ailleurs.

## Conséquences

- Une machine neuve n'exige qu'un navigateur et un copier-coller, pas une clé
  préinstallée.
- Le jeton apparaît en clair dans la remote du clone jusqu'à l'étape 2. C'est le
  coût assumé du bootstrap ; l'étape 2 n'est pas optionnelle.
- `gh` devient une dépendance des mises à jour, d'où sa présence dans le script
  d'outillage ([ADR-007](007-outillage-run-onchange.md)).
- Le champ personnalisé Bitwarden n'est accessible qu'avec `bw get item` et `jq`,
  pas `bw get password` — donc pas sur une machine nue, où le passage par le
  coffre web est la seule voie.

## Alternatives écartées

- **`chezmoi init git@github.com:…` en SSH** : suppose une clé déjà utilisable,
  que la YubiKey et l'agent GPG rendent indisponible avant le premier apply.
- **`gh auth login` avant le clone** : reviendrait à installer `gh` à la main
  d'abord, alors que c'est justement ce que le dépôt installe.
- **Conserver le PAT comme helper permanent** : un secret de longue durée en clair
  dans la configuration git, sans rotation, pour un service que `gh` rend déjà.
