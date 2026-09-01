# ADR-018 - Chiffrement par paire de clés `age`, clé déverrouillée une fois

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : `dd22a8b` (paire de clés)

## Contexte

[ADR-016](016-depot-public-sensible-chiffre.md) a fait passer le dépôt de un à six
fichiers chiffrés : `~/.secrets`, `nas.conf` et les quatre fragments du profil
`pro`. `age -p` chiffre par passphrase, et chezmoi la réclame **par fichier**, à
chaque `apply` qui en touche un.

Mesuré pendant la bascule : cinq invites consécutives pour chiffrer les cinq
fragments, cinq de plus pour les relire. La conséquence n'est pas seulement
pénible, elle est dangereuse - une invite répétée cinq fois se traite
mécaniquement, et rien ne garantissait alors que la même passphrase avait été
saisie partout. C'est ce constat qui a motivé l'ajout d'une étape de relecture.

Une saisie vide fait par ailleurs générer une passphrase par `age`, différente à
chaque fichier, sans que rien ne le signale avant le premier `apply` d'une autre
machine.

## Décision

Le chiffrement passe de la passphrase à une paire de clés `age`.

- `.chezmoi.toml.tmpl` déclare `identity = ~/.config/chezmoi/key.txt` et
  `recipientsFile`, un fichier du répertoire source qui porte la clé publique -
  publiable par nature, et ignoré au déploiement.
- La clé privée est versionnée **chiffrée par passphrase**, sous
  `age-key.txt.age`, hors déploiement elle aussi.
- `run_before_unlock-age-key.sh.tmpl` la déchiffre si `key.txt` est absent, avant
  que chezmoi n'ait à déchiffrer un fichier : **une saisie par machine**, aucune
  ensuite. Le script sort sans rien faire si la clé est là, et se contente d'un
  avertissement si `age` manque encore.

## Conséquences

- Un `apply` ne demande plus rien sur une machine déjà déverrouillée, quel que
  soit le nombre de fichiers chiffrés. Le contexte non interactif - NAS,
  planificateur, agent - cesse d'être un cas particulier.
- **Le gain n'est pas cryptographique.** La clé privée étant publiée chiffrée par
  passphrase, la surface de force brute reste une passphrase : elle porte
  désormais sur un seul fichier au lieu de six, pas sur zéro. Une clé transportée
  hors bande serait plus solide, au prix d'un transport manuel sur chaque machine
  neuve.
- **Deux secrets à perdre au lieu d'un.** `key.txt` perdue mais passphrase connue
  se rattrape ; les deux perdues, les fichiers chiffrés sont définitivement
  illisibles.
- Le fichier de clé déverrouillé vit en clair sur le disque, en `0600`. Un vol de
  poste allumé n'a plus de passphrase à deviner - c'est la contrepartie assumée
  du confort.
- Ajouter un fichier chiffré ne coûte plus rien, ce qui invite à en ajouter. La
  règle de l'`AGENTS.md` racine reste la barrière : rien de sensible qui puisse
  rester hors du dépôt.
- La bascule impose de rechiffrer les six fichiers et de refaire la réécriture
  d'historique, les blobs `.age` changeant tous.

## Alternatives écartées

- **Rester sur `age -p`**, la décision d'ADR-016 pour ce point : six invites par
  `apply` complet, et le risque de passphrases divergentes que la bascule a
  effectivement produit.
- **Paire de clés avec clé transportée hors bande**, sans copie chiffrée dans le
  dépôt : plus solide, et zéro saisie, mais la clé doit arriver sur chaque machine
  neuve avant le premier `apply` - la dépendance circulaire que
  [ADR-017](017-clone-anonyme-gh-en-ecriture.md) évite justement au bootstrap.
- **Tout reconsolider dans `~/.secrets`** pour revenir à un seul fichier chiffré :
  `nas.conf` et le fragment gitconfig ne sont pas des variables d'environnement,
  il faudrait les reconstruire par script au login. Le remède est pire.
- **Un agent de secrets qui garde la passphrase en mémoire** (`age-agent` n'existe
  pas ; un `ssh-agent` détourné, ou `pass`) : une dépendance de plus à installer
  avant le premier apply, sur le NAS comme ailleurs.
