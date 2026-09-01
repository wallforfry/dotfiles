# ADR-015 - Serveurs MCP lourds en conteneurs nommés

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `e50783c`, `c3a6825` (les quatre pilotes)

## Contexte

Un serveur MCP livré en image Docker s'enregistre naturellement comme
`docker run -i --rm <image>`. C'est la forme que donnent les documentations amont, et
c'est celle qui était en place dans `~/.cursor/mcp.json` pour `crystaldba/postgres-mcp`.

Elle fuit. Mesuré sur ce poste avant correction :

```
relaxed_bhabha       Up 23 hours
elated_cartwright    Up 23 hours
musing_mcnulty       Up 28 hours
affectionate_rubin   Up 2 days
distracted_newton    Up 2 days
upbeat_banach        Up 3 days
```

Six conteneurs pour un serveur, le plus ancien depuis trois jours, **tous en
`AutoRemove: true`**. C'est le cœur du piège : `--rm` retire le conteneur à la sortie
de son processus, et ce processus ne sort pas quand le client meurt. Un `--rm` donne
donc l'illusion d'un nettoyage qui n'a jamais lieu.

Ces images ne sont pas non plus installables nativement sans peine : Scrapling
embarque des navigateurs et des dépendances Python lourdes.

## Décision

Un serveur MCP livré par image Docker s'enregistre à travers un **script de
`~/.local/bin` qui `docker exec` dans un conteneur nommé unique**, démarré à la
demande. Quatre pilotes suivent cette forme : `firecrawl-mcp`, `scrapling-mcp`,
`postgres-mcp`, `cloak`.

Trois règles les gouvernent :

1. **Un conteneur par configuration distincte, nommé d'après elle.** `postgres-mcp`
   dérive son nom d'une empreinte de `DATABASE_URI` : deux projets sur deux bases ne
   partagent pas un conteneur, deux sessions sur la même base en partagent un.
2. **Les identifiants passent par l'environnement**, jamais par la ligne de commande,
   lisible par tout processus de la machine.
3. **`docker start` d'abord, `docker run` ensuite, `docker start` de nouveau en
   secours** : la même commande réutilise et relance, et le dernier appel couvre la
   course où une session concurrente crée le conteneur entre les deux premiers.

Chaque pilote sort en 69 quand le daemon docker est indisponible, plutôt que d'échouer
obscurément.

## Conséquences

- Un oubli coûte **au plus un conteneur par service**, non un par session.
- Les dépendances lourdes restent isolées du poste.
- Le conteneur survit entre les sessions, volontairement : c'est ce qui rend le
  démarrage suivant instantané. Contrepartie, il faut un `--stop` explicite, d'où
  cette option sur les quatre pilotes.
- Un conteneur qui traîne retient l'image sur laquelle il a été créé : après un
  changement de version épinglée, il faut le retirer pour que la nouvelle prenne.
- Docker devient une dépendance dure de ces serveurs, indisponible sur les cibles
  sans daemon ([ADR-008](008-dsm-cible-de-premier-rang.md)).
- Un pilote de plus à maintenir par serveur, là où une ligne de configuration
  suffisait.
- `~/.cursor/mcp.json` reste hors du dépôt - il contient des mots de passe et des
  jetons en clair ([ADR-016](016-depot-public-sensible-chiffre.md)) - donc la bascule vers
  les pilotes s'y fait à la main, sans être rejouable par `chezmoi apply`.

Vérifié contre une base jetable : trois sessions simultanées obtiennent chacune une
réponse MCP valide (`postgres-mcp` 1.6.0, mode `restricted`) en partageant un seul
conteneur.

## Alternatives écartées

- **`docker run -i --rm` par session** : l'état de départ, réfuté par la mesure
  ci-dessus. C'est l'alternative la plus tentante, puisque c'est celle que
  recommandent les documentations amont.
- **Installation native des serveurs** : les dépendances navigateur de Scrapling sont
  ingérables sur le poste. Le cas de CloakBrowser montre la limite inverse - son
  paquet npm *est* installé, mais ne fournit que le CLI de licence, pas le serveur
  CDP, qui reste dans l'image.
- **Un nettoyage périodique** des conteneurs orphelins, par `cron` ou par hook : traite
  le symptôme, et ne sait pas distinguer un conteneur orphelin d'un conteneur servant
  une session vivante - il finirait par couper une session en cours.
- **Un conteneur par session avec un hook d'arrêt côté client** : suppose que chaque
  client MCP en exécute un à la mort de la session, ce qui n'est vrai d'aucun de ceux
  utilisés ici. C'est précisément parce que rien ne s'exécute à la mort du client que
  `--rm` échoue.
