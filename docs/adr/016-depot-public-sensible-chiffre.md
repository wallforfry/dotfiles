# ADR-016 - Dépôt public, sensible chiffré et hors prose

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : `07d72fc` (sortie du sensible), `1f8595f` (documentation), `b26883d` (fragments chiffrés), `c63f6be` (pile technique)

## Contexte

Ce dépôt était privé, et ADR-002 - que celle-ci remplace, donc supprimée du
répertoire, relisible par `git log -- docs/adr/` - donnait deux raisons de le laisser ainsi : un dépôt public exposerait
`encrypted_private_dot_secrets.age` au téléchargement, donc à une attaque hors
ligne sans limite de temps contre la seule passphrase ; et plusieurs fichiers
portaient des informations d'employeur, publiables ni en clair ni chiffrées.

Le dépôt est aussi le support de `harness/` et de `dot_config/agent-skills/`, dont
l'intérêt est nul s'ils restent illisibles. La publication est demandée, et ce que
le dépôt portait de sensible a été inventorié :

- `private_dot_ssh/private_config.d/nas.conf` - nom d'hôte du NAS, adresses IP
  locale et VPN, nom d'utilisateur.
- le bloc `pro` de `dot_zshrc.tmpl` et de `dot_zprofile.tmpl` - noms de projets,
  configurations Doppler, chemins de travail.
- l'`includeIf` de `dot_gitconfig.tmpl` et son `excludesfile` codé en dur sur un
  chemin `/Users/<login>`.
- la section « Contextes » de `harness/USER.md`, la prose du `README.md` et
  d'[ADR-004](004-profils-perso-pro.md), [ADR-011](011-identite-git-par-repertoire.md)
  et [ADR-012](012-instructions-ia-versionnees.md).

Nom, courriel et identifiant de clé GPG de `dot_gitconfig.tmpl` sont d'une autre
nature : tout commit signé poussé sur un dépôt public les expose déjà. Les
chiffrer coûterait une saisie de passphrase par `apply` pour dissimuler une donnée
publiée ailleurs.

## Décision

Le dépôt est public. Rien de sensible n'y figure, ni en clair ni en prose :

1. **Ce qui est sensible et nécessaire au déploiement est chiffré** par `age`,
   selon le mécanisme d'[ADR-018](018-chiffrement-par-paire-de-cles.md) :
   `~/.secrets`, `nas.conf`, `~/.config/zsh/pro.zsh`,
   `~/.config/zsh/pro.zprofile`, `~/.config/git/pro.gitconfig`,
   `~/.claude/CONTEXT.md`, `~/.config/dotfiles/sensible.txt`. Ce dernier est la
   liste des noms que `scripts/verify.sh` interdit : l'énumération est la donnée
   à protéger, donc elle ne peut pas vivre dans le script qui la lit.
2. **Les fichiers publics chargent ces fragments sans les nommer** : un
   `[ -f … ] && source …` en zsh, un `[include]` en git - git ignore silencieusement
   un chemin absent - et `@CONTEXT.md` dans `dot_claude/CLAUDE.md`.
3. **Les fragments propres au profil `pro` sont ignorés hors de ce profil**
   (`.chezmoiignore`). `~/.claude/CONTEXT.md` fait exception et se déploie
   partout : il porte aussi les projets personnels, dont une machine perso a
   besoin.
4. **La prose ne porte aucun nom de projet, sauf le sien.** Employeur, clients,
   projets professionnels et personnels sont décrits dans
   `~/.claude/CONTEXT.md`, où `harness/USER.md` renvoie. Un dépôt public dit
   déjà à qui il appartient : la liste de ce sur quoi son auteur travaille est
   une information de plus, qu'il n'a pas à donner.

   **La pile technique relève de la même règle** : la liste des outils tiers
   installés sur une machine - CLI de PaaS, de CI, de gestion de secrets - dit
   sur quelle infrastructure travaille son propriétaire, et un lecteur qui sait
   déjà pour qui il travaille en tire une carte de ce qu'il faudrait attaquer.
   Seuls les outils dont ce dépôt dépend sont nommés, parce qu'un fichier
   déployé les réclame ([ADR-007](007-outillage-run-onchange.md)) ; ce qui est
   installé par ailleurs n'a pas à figurer ici. La règle couvre le message de
   commit au même titre que le template, la prose et le journal.
5. **L'historique est réécrit avant publication** (`git-filter-repo`), le dépôt
   distant est basculé en public, et la passphrase `age` est changée : la
   précédente a protégé un fichier désormais téléchargeable par quiconque.

## Conséquences

- **La passphrase devient la seule barrière, et elle est attaquable hors ligne.**
  Elle doit être longue et propre à cet usage. C'est le coût central de la
  décision, et il ne se rattrape pas après publication : une passphrase faible
  aujourd'hui reste cassable sur une copie faite aujourd'hui. Le nombre de
  fichiers qu'elle protège et le nombre de saisies qu'elle coûte relèvent
  d'[ADR-018](018-chiffrement-par-paire-de-cles.md).
- La réécriture d'historique change tous les SHA. Les champs **Commits** de
  toutes les ADR deviennent faux et sont à reprendre après la bascule.
- Le clone initial n'exige plus d'authentification, ce qui rend le bootstrap par
  PAT sans objet ([ADR-017](017-clone-anonyme-gh-en-ecriture.md)).
- Un fragment chiffré ne se relit pas d'un coup d'œil : modifier `pro.zsh` demande
  `chezmoi edit`, non un éditeur sur le fichier source.
- Toute addition future doit se demander où elle atterrit. La règle est dans
  l'`AGENTS.md` racine, pas seulement ici.

## Alternatives écartées

- **Rester privé** - la décision précédente, ADR-002.
  Elle protégeait le fichier chiffré du téléchargement, mais rendait `harness/` et
  les skills inutilisables comme référence publique, ce qui était leur intérêt.
- **Second dépôt public curé**, alimenté depuis celui-ci : n'expose aucun fichier
  chiffré et ne demande aucune réécriture d'historique, mais impose une
  synchronisation manuelle entre deux dépôts, donc une divergence garantie.
- **Sortir le sensible dans un petit dépôt privé tiré en external chezmoi** :
  garde `apply` non interactif et n'expose rien de chiffré, au prix d'un second
  dépôt à cloner, authentifier et maintenir sur chaque machine - la dépendance
  circulaire du bootstrap que [ADR-017](017-clone-anonyme-gh-en-ecriture.md)
  cherche justement à éviter.
- **`age` avec paire de clés** plutôt qu'une passphrase, reprise d'ADR-002 : sur un
  dépôt public l'argument s'inverse en partie, une clé de 32 octets n'étant pas
  attaquable par force brute là où une passphrase l'est. Écartée quand même : il
  faut alors transporter le fichier de clé hors bande sur chaque machine neuve,
  avant tout `apply`, ce qui recrée la dépendance circulaire du bootstrap.
- **Gestionnaire de secrets en ligne** (1Password, Bitwarden, Vault) comme source
  de `~/.secrets`, reprise d'ADR-002 : ajoute une dépendance réseau et un CLI à
  installer sur le NAS, là où un fichier chiffré suffit.
- **Publier sans réécrire l'historique** : les commits passés porteraient encore
  le nom d'hôte du NAS, ses adresses et les configurations Doppler. Nettoyer
  l'arbre de travail seul ne cache rien.
