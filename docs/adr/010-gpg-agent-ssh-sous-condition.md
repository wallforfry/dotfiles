# ADR-010 — GPG comme agent SSH, sous condition

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `b59ac4e` (garde sur `SSH_AUTH_SOCK`), `9247893` (alias `switch-card`)

## Contexte

La clé SSH vit sur une YubiKey, exposée par l'agent GPG. `.zprofile` doit donc
pointer `SSH_AUTH_SOCK` sur le socket ssh de `gpg-agent` — et les commits signés
(`commit.gpgsign = true`) dépendent du même agent.

Sur le NAS, il n'y a pas d'agent GPG utilisable. La version initiale exportait
malgré tout le socket : `gpgconf` échouait, mais l'export avait lieu, ce qui
**coupait l'agent ssh transmis par `ForwardAgent`** — le seul qui fonctionnait là —
en plus d'afficher deux erreurs `gpgconf` à chaque login (`b59ac4e`).

## Décision

L'export de `SSH_AUTH_SOCK` est conditionné à trois vérifications successives, dans
cet ordre :

1. `gpg-agent` est présent (`command -v`) ;
2. son lancement réussit (`gpgconf --launch gpg-agent`) ;
3. le socket retourné existe et est bien un socket (`[[ -S "$gpg_sock" ]]`).

La règle générale qui en découle : **ne jamais remplacer un mécanisme fonctionnel
par un mécanisme supposé.** Un export qui écrase une variable déjà utile doit
d'abord prouver que son remplaçant marche.

L'alias `switch-card` (`gpg-connect-agent 'scd serialno' 'learn --force' /bye`)
gère le changement de YubiKey. Ses arguments doivent être entourés de guillemets
englobants, sans quoi zsh ne retient que le premier mot et traite le reste comme
des demandes d'affichage d'alias (`9247893`).

## Conséquences

- `ForwardAgent` continue de servir sur les machines sans GPG, et le login n'affiche
  plus d'erreur.
- Sur une machine où l'agent GPG est cassé plutôt qu'absent, la dégradation est
  silencieuse : pas d'agent GPG, pas de message. C'est le comportement voulu au
  login, mais il faut y penser au diagnostic.
- Toute variable d'environnement qui remplace un mécanisme existant relève de la
  même règle : la vérification vient avant l'export, pas après.

## Alternatives écartées

- **Exporter `SSH_AUTH_SOCK` inconditionnellement** : l'état initial, qui cassait
  `ForwardAgent` sur le NAS.
- **Conditionner par le système ou le profil** plutôt que par la capacité : une
  machine Linux avec YubiKey ou un macOS sans agent tomberaient du mauvais côté.
  Le test porte sur ce qui répond, pas sur ce que la machine est censée être.
- **Une clé SSH en fichier plutôt que sur YubiKey** : supprime le problème et la
  protection matérielle avec lui.
