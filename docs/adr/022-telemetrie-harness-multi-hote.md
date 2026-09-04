# ADR-022 - Télémétrie du harness multi-hôte

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : à compléter au commit qui porte ce fichier

## Contexte

L'audit du harness lisait 692 sessions Claude, mais aucune session Codex. Il ne
reconnaissait que les appels Claude `Skill`, `Task`, `Agent` et `Write` : une
skill effectivement utilisée depuis Codex restait donc affichée à zéro, et les
écritures par `Edit`, patch ou outil shell échappaient au taux de commentaires.

Les deux hôtes stockent des JSONL de formes différentes. Leur absence de signal
n'a pas non plus le même sens : Codex n'émet pas d'événement explicite
d'activation de skill. Additionner directement leurs compteurs transforme ainsi
une inconnue en mesure.

Le parcours complet des transcripts prenait plusieurs dizaines de secondes alors
que leur grande majorité ne change pas entre deux audits. Enfin, les mutations
partaient du dernier commit et ignoraient le worktree en cours, précisément celui
qu'un audit avant commit doit valider.

## Décision

Chaque format de transcript est adapté vers un même événement interne, puis les
agrégats sont produits depuis ces événements. Une valeur absente reste
`<inconnu>` et une valeur explicitement nulle devient `<aucun>` ; aucun zéro n'est
déduit d'un signal que l'hôte ne fournit pas.

Codex projette aussi certains messages et appels de subagent sous
`item_completed`. Ces projections sont reconnues mais ne produisent pas un second
événement : le `message` ou `function_call` source reste l'enregistrement
autoritaire. `FileChange` et `CommandExecution`, qui n'ont pas cette projection
source, restent normalisés depuis `item_completed`.

Le cache ne conserve que des compteurs agrégés, des empreintes de fichiers et des
identifiants de chemins hachés. Il ne conserve ni chemin ni contenu brut, change
de version quand le schéma évolue et s'écrit atomiquement avec des permissions
restreintes.

La mesure de barrière capture dans son clone les modifications suivies et les
fichiers non suivis du worktree courant, puis en fait un commit local jetable.
Ses cas forment une matrice `promesse -> contrôle -> attente` : les mutants doivent
être rejetés, les anti-mutants acceptés, et les promesses comportementales sont
explicitement marquées comme observations plutôt que comptées dans le score. La
barrière complète valide d'abord le clone ; chaque cas rejoue ensuite son contrôle
ciblé, dont l'appartenance à l'orchestrateur est elle-même exigée.

Le corpus de routage est un contrat d'entrée positif, négatif et ambigu. Le shell
en valide la forme, la couverture et la taille des descriptions ; le comportement
du modèle reste une mesure distincte à rejouer sur chaque hôte.

## Conséquences

- Un chiffre Claude et un chiffre Codex ont la même sémantique ; une capacité
  absente ou non observable reste visible comme telle.
- Un second audit réutilise les sessions inchangées. Mesuré sur macOS : 6,62 s
  sans cache, puis 0,57 s avec cache pour la télémétrie multi-hôte complète.
- Le cache local contient des empreintes stables. Il ne doit jamais être versionné
  ni présenté comme une anonymisation cryptographique de données publiables.
- Une écriture shell est détectable sans que les lignes effectivement produites
  soient reconstructibles de façon fiable ; son contenu reste donc inconnu.
- Chaque nouveau format d'hôte exige un adaptateur et des fixtures. Un changement
  de format non reconnu doit rendre la mesure incomplète, jamais silencieusement
  verte.
- Les mutations restent le poste lent, mais rejouer leurs contrôles ciblés ramène
  sur macOS une matrice étendue à 29 mutants à 11,18 s et l'audit complet à
  11,92 s ; la version qui rejouait toute la barrière prenait 46,88 s pour 20 mutants.

## Alternatives écartées

- **Un rapport séparé par hôte** : plus simple à parser, mais impossible à
  comparer et condamné à redéfinir chaque métrique.
- **Compter l'absence d'événement comme zéro** : produit le faux constat qui a
  déclenché cette décision.
- **Conserver les chemins dans le cache** : facilite le diagnostic d'un fichier
  corrompu, mais publie dans un artefact persistant les noms que l'audit s'interdit
  d'afficher.
- **Relire tous les transcripts à chaque exécution** : sans état et exact, mais le
  coût répété n'apporte aucune information sur les sessions inchangées.
- **Muter seulement le dernier commit** : isole un état reproductible, mais ne
  valide pas les changements que l'utilisateur s'apprête à committer.
- **Rejouer toute la barrière pour chaque mutant** : prouve directement la commande
  publique, mais prenait 46,88 s pour la matrice. Une exécution complète du clone,
  le contrôle de l'orchestrateur puis le composant ciblé gardent la même couverture
  sans répéter les sections indépendantes.
- **Faire appeler un modèle par la CI pour le routage** : teste la sémantique, au
  prix d'un résultat non déterministe, payant et différent selon l'hôte. Le corpus
  garde ces essais possibles sans en faire une barrière mécanique.
