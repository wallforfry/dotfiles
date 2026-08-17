# USER.md — Contexte et attentes de l'utilisateur

Préférences de travail et contexte technique, sans information personnelle.
Identité de l'agent → `SOUL.md`, règles techniques → `AGENTS.md`.

## Acquis

TypeScript strict et l'écosystème web moderne (React, Vite, Vitest, Tailwind,
monorepos Turborepo avec pnpm) ; Node et GraphQL ; PostgreSQL, Prisma,
Convex ; Docker et Compose, Helm et Kubernetes ; CI Bitbucket Pipelines et
GitHub Actions ; shell, git et Unix ; chezmoi, `age`, GPG sur YubiKey.

Aucun rappel de fondamentaux sur ces sujets. Aller au fait.

## Contextes

- **REDACTED** (profil `pro`) — `REDACTED`, `REDACTED`, `REDACTED` :
  monorepos TypeScript. Session isolée via `claude-REDACTED`
  (`CLAUDE_CONFIG_DIR=~/.claude_REDACTED`).
- **Forges** — GitHub en pro comme en perso ; Bitbucket uniquement en pro.
  Déduire la forge de `git remote get-url origin`, jamais du profil de la
  machine : les deux coexistent côté pro.
- **REDACTED** — React, Node, PostgreSQL, Prisma, TypeScript strict.
  Environnements via `doppler` (`doppler-api-prod`, `-dev`, `-local`, `-feat-0x`).
- **REDACTED** — `REDACTED`, `REDACTED`.
- **Perso** — `REDACTED`, `REDACTED`, ces dotfiles.
- **Postes** — macOS au quotidien, Linux, NAS Synology DSM (`/tmp` monté
  `noexec`, shell forcé par un fragment SSH versionné).

## Biais techniques

- **Source unique de vérité.** Types, constantes, schémas de validation, règles
  métier : un seul emplacement autoritaire, dont tout le reste dérive. Jamais de
  définition recopiée d'un service à l'autre. C'est la règle la plus structurante
  des dépôts de l'utilisateur — la respecter avant toute autre considération de
  style.
- **Vérifier l'existant avant d'ajouter une dépendance.** D'abord les paquets
  internes du monorepo, puis la bibliothèque standard et les API natives. Une
  bibliothèque de dates ou d'utilitaires installée alors que l'équivalent maison
  existe est un défaut, pas un raccourci.
- **Style fonctionnel et immuable** en TypeScript : éviter la mutation et les
  boucles impératives, préférer les types en lecture seule, proscrire les
  assertions `as` au profit du rétrécissement de type, exports nommés.
- **Rien de cassé n'est toléré.** Un test intermittent est un bug jusqu'à preuve
  du contraire : ni `retry`, ni `skip`, on cherche la cause.
- **Ne jamais déclarer un commentaire de revue corrigé sans l'avoir vérifié dans
  le code.** Quand plusieurs commentaires sont traités, tous sont à revérifier
  pour la même erreur.
- **Boy scout, jamais mimétisme.** Le voisinage fixe l'idiome, pas le niveau de
  qualité. Corriger un défaut voisin s'il est sur le chemin, dans un commit
  séparé.
- **Le coût compte.** Contexte et tokens sont une ressource : déléguer
  l'exploration à des subagents, ne pas relire un fichier déjà lu, ne pas
  produire un rapport de dix paragraphes pour une réponse d'une ligne.

## Mode de travail

- **Un worktree par tâche.** Le travail se fait dans un worktree git isolé, pas
  sur la branche courante du dépôt principal.
- **Ampleur d'abord.** Petite tâche : agir directement. Plusieurs fichiers ou
  décision d'architecture : trois à cinq lignes de plan, validation, exécution.
- **Conventional Commits**, imposés par `commitlint` sur les dépôts qui en
  disposent. Le message suit la langue du dépôt — et **un message français est
  accentué**, sujet comme corps, les identifiants et sorties citées restant
  verbatim.
- **Périmètre.** Nettoyage adjacent bienvenu, mais dans un commit séparé du
  commit fonctionnel, afin que chacun se révoque indépendamment.
- **Désaccord.** Livrer ce qui est demandé *et* l'alternative proposée, pour
  comparaison sur pièces. Ne pas se contenter d'objecter, ne pas bloquer.
- **Barre de vérification.** Lint, types, tests et CI verts sont la barre par
  défaut, pas une revue. Avant de solliciter un relecteur sur une PR, passer
  `merge-verdict` sur sa propre PR et corriger ses constats bloquants.
- **Décisions durables.** Les choix structurants se consignent : `adr/` là où le
  dépôt en a un, la description de la PR sinon. Ne jamais contredire une décision
  enregistrée en silence — la suivre, ou énoncer le conflit.

## Points de vigilance

Signaler systématiquement, même sans être sollicité, les cas limites et les
conséquences non voulues d'un changement — et en priorité :

1. **Gestion d'erreur incomplète** — cas d'échec, délais d'attente, retours
   vides et nuls, valeur externe dégradée en défaut permissif par un ternaire au
   lieu d'être validée par un schéma.
2. **Nommage et lisibilité** — proposer directement de meilleurs noms plutôt que
   de signaler le problème.
3. **Dépendance de trop** — vérifier ce qui est déjà installé avant tout ajout.
   Si le besoin réel tient en quelques dizaines de lignes sans cas limite
   sérieux, proposer ces lignes et laisser la décision.
4. **Secret ou donnée personnelle** — rien de sensible dans un fichier versionné,
   une URL, un log ou un message de commit. Les secrets de ce dépôt passent par
   `encrypted_private_dot_secrets.age`.
