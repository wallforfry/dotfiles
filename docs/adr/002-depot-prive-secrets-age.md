# ADR-002 - Dépôt privé et secrets chiffrés par `age`

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `b286cad` (initial commit), `16c9f98` (installation sur repo privé)

## Contexte

`~/.secrets` contient des jetons et des variables d'environnement sensibles, et
doit suivre les machines. Le versionner en clair est exclu ; ne pas le versionner
ramène à la saisie manuelle sur chaque poste, c'est-à-dire à la divergence.

Deux informations distinctes plaident pour le privé, et une seule ne suffirait
pas :

- Un dépôt public exposerait `encrypted_private_dot_secrets.age` au
  téléchargement, donc à une attaque hors ligne sans limite de temps contre la
  seule passphrase.
- `dot_gitconfig.tmpl` et le bloc `pro` de `dot_zshrc.tmpl` portent des
  informations d'employeur - noms de projets et de configurations Doppler - qui
  n'ont pas à être publiées, chiffrement ou non.

## Décision

Le dépôt reste privé. `~/.secrets` est chiffré par `age` en passphrase symétrique
et versionné sous `encrypted_private_dot_secrets.age`. La passphrase est saisie à
`chezmoi apply` et n'est stockée nulle part dans le dépôt.

Le binaire `age` est requis, et non l'implémentation intégrée à chezmoi : celle-ci
ne gère pas les passphrases et échoue sur « no identities specified ».

## Conséquences

- **Passphrase perdue, secrets irrécupérables.** Aucune récupération n'est prévue,
  par construction.
- `age` devient une dépendance d'installation, d'où sa présence dans le script
  d'outillage ([ADR-007](007-outillage-run-onchange.md)).
- Le dépôt privé impose une authentification avant le premier clone, d'où le
  bootstrap par PAT ([ADR-003](003-bootstrap-pat-puis-gh.md)).
- Toute commande chezmoi qui touche un fichier chiffré devient interactive. En
  contexte non interactif - CI, agent, script - il faut restreindre la portée
  (`chezmoi apply ~/.chemin`) ou passer `--exclude=encrypted`.

## Alternatives écartées

- **Dépôt public avec secrets chiffrés seuls** : le fichier chiffré resterait
  téléchargeable, et les informations d'employeur publiées.
- **`age` avec paire de clés** plutôt qu'une passphrase : déplace le secret vers
  un fichier de clé qu'il faut alors transporter hors bande sur chaque machine
  neuve, sans rien gagner sur le poste déjà installé.
- **Gestionnaire de secrets en ligne** (1Password, Bitwarden, Vault) comme source
  de `~/.secrets` : ajoute une dépendance réseau et un CLI à installer sur le NAS,
  là où un fichier chiffré suffit. Bitwarden est déjà utilisé, mais pour le seul
  jeton de bootstrap, hors du chemin critique.
