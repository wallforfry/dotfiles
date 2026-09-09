# ADR-023 - Outillage applicatif en Go, shell limité au bootstrap

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : `71cf7b9`

## Contexte

Le dépôt comptait 28 scripts ou composants shell pour 2 016 lignes et sept
modules Python pour 1 153 lignes. Trois catégories y cohabitaient : le bootstrap
`chezmoi`, des lanceurs de processus déployés, et une barrière de vérification
avec télémétrie et injection de défauts. Les deux dernières portent des modèles
de données, des invariants et des tests ; leur répartition entre shell, `awk`,
`jq` et Python rendait leurs contrats implicites et multipliait les chemins
d'erreur.

La télémétrie Python disposait de 20 tests et parcourait 390 760 enregistrements
en 1,3 seconde avec son cache chaud lors de la mesure précédant cette décision.
La vitesse seule ne justifiait donc pas une réécriture. Le besoin est une source
unique de vérité pour les types, les erreurs, les commandes et les contrôles,
avec des tests capables d'exercer les branches sans lancer Docker ni modifier
le poste.

[ADR-008](008-dsm-cible-de-premier-rang.md) impose cependant `sh` POSIX avant
l'installation de l'outillage : DSM n'a pas de gestionnaire de paquets garanti
et `/tmp` y est monté `noexec`. Un binaire ne peut pas installer la chaîne qui
sert à le construire. Le déverrouillage `age`, l'installation de Go et la
construction initiale restent donc une frontière de bootstrap distincte.

L'implémentation de la migration utilise uniquement la bibliothèque standard de
Go. Elle vérifie quatre compilations sans CGO, `darwin/linux` et
`amd64/arm64`, ce qui établit le périmètre réellement exercé. Aucun besoin du
dépôt ne réclame le contrôle manuel de la mémoire ou les abstractions bas niveau
qui compenseraient le coût supplémentaire d'une chaîne Rust.

## Décision

La logique post-bootstrap du dépôt est implémentée dans un CLI Go unique nommé
`dotfiles`, sans dépendance tierce. Ses sous-commandes possèdent la vérification,
l'audit du harness, la validation du routage, l'enregistrement du hook Claude et
les cinq lanceurs déployés. Les noms publics de ces lanceurs restent des liens
chezmoi vers le binaire unique ; leur contrat de ligne de commande, leurs codes
de sortie et leur politique de conteneurs nommés ne changent pas.

Le bootstrap conserve exactement trois scripts `sh` POSIX : déverrouiller la
clé `age`, installer les prérequis dont Go, puis construire le CLI. La version
Go téléchargée hors Homebrew est épinglée et son archive vérifiée par SHA-256 ;
la construction utilise `CGO_ENABLED=0` et n'écrit aucun exécutable sous `/tmp`.

Les anciennes implémentations shell et Python sont supprimées avec leurs
appelants dans le même changement. Un outil ponctuel devenu obsolète est
supprimé plutôt que traduit. La barrière compile, teste, analyse statiquement et
construit le CLI pour les quatre cibles avant d'exécuter ses contrôles de dépôt.

## Conséquences

- Les types JSON, la normalisation, les sorties, les mutations et l'exécution
  des processus deviennent testables sans dépendre de pipelines shell.
- Une machine neuve télécharge aussi la chaîne Go, environ 67 à 71 Mo selon
  l'architecture, puis compile le CLI après le premier apply. Chaque changement
  du dépôt peut déclencher une nouvelle construction locale.
- Le dépôt conserve du shell au seul endroit où le supprimer créerait une
  dépendance circulaire. Il conserve aussi les fragments zsh, qui doivent
  modifier l'environnement du shell parent et ne sont pas des exécutables
  applicatifs.
- Les liens déployés partagent une panne possible : si la construction Go
  échoue, les commandes pointent vers un binaire absent. Le hook de construction
  le signale sans interrompre le reste du bootstrap, et l'apply suivant peut le
  réparer.
- La barrière locale nécessite désormais Go. La CI et chaque poste doivent donc
  nommer et exercer cette dépendance avant de publier un résultat vert.
- La matrice de mutations reste dominée par les sous-processus. Ses contrôles
  ciblés mesurent 43 mutants en 22,31 secondes sur macOS ; l'audit complet avec
  cache chaud prend 24,17 secondes.
- DSM reste sans runner CI. Les builds Linux `arm64` et `amd64` prouvent la
  compilation, pas l'exécution sur son noyau et sa libc ; un changement du
  bootstrap y reste une vérification manuelle.

## Alternatives écartées

- **Tout migrer, bootstrap compris** : impossible sans binaire préinstallé, et
  contraire à la panne documentée par ADR-008. Le shell POSIX est la condition
  d'entrée, pas une implémentation applicative concurrente.
- **Rust** : ses garanties mémoire et ses binaires compacts sont réelles, mais
  l'outillage est dominé par les entrées-sorties et les sous-processus. Go couvre
  JSON, fichiers, empreintes, HTTP et processus dans sa bibliothèque standard,
  avec une compilation croisée plus directe pour les quatre cibles retenues.
- **Conserver Python pour la télémétrie et Go pour le reste** : évite une partie
  du portage, mais maintient deux chaînes, deux modèles d'erreur et une frontière
  d'intégration dans le composant qui partage le plus de données avec l'audit.
- **Conserver tous les scripts existants** : leur performance était suffisante,
  mais cette option ne répond pas au besoin de contrats typés, d'injection des
  processus et de source unique de vérité. Elle conserve aussi des composants
  ponctuels après la fin de leur usage.
