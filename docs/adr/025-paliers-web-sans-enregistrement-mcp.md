# ADR-025 - Paliers web lourds appelés à la demande, sans enregistrement MCP

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : à compléter au commit

## Contexte

[ADR-014](014-recuperation-web-par-paliers.md) pose la règle « ne jamais démarrer au-dessus du
premier palier », et [ADR-015](015-mcp-en-conteneurs-nommes.md) un conteneur nommé « démarré à la
demande ». Pour Firecrawl et Scrapling, enregistrés en portée utilisateur dans Claude Code (profils
par défaut et pro), « à la demande » voulait dire en réalité **à chaque ouverture de session** :

- Claude Code, Codex et Cursor lancent tous les serveurs stdio enregistrés dès l'ouverture d'une
  session. Aucun ne sait retarder ce lancement jusqu'au premier appel d'outil.
- Sans argument, `firecrawl-mcp` lançait `docker compose up --wait` si l'API ne répondait pas, puis
  `npx firecrawl-mcp`. De son côté, `scrapling-mcp` démarrait son conteneur puis y lançait par
  `docker exec` un processus `scrapling mcp` par session. Rien ne les arrêtait jamais.

Mesures du 2026-09-22, sur ce poste :

- Les journaux MCP de Claude Code comptent **102 démarrages** de chacun des deux serveurs entre le
  17 et le 22 septembre, pour **0 appel**. Depuis leur arrivée le 2026-08-17, sur 272 sessions :
  **8 appels Firecrawl, 0 appel Scrapling**. Sur la même période, `WebFetch` en compte 112.
- Au repos, avec 7 sessions ouvertes : la pile Firecrawl pèse environ 710 Mio, dont RabbitMQ à
  27 % de CPU. Le conteneur Scrapling pèse 541 Mio, soit 7 processus Python de 69 à 97 Mio. Chaque
  session ajoute 110 à 150 Mio côté hôte (chaîne `npx`/`node`, client `docker exec`).
- Une seconde cause faisait revenir Firecrawl sans aucune session. Après le redémarrage brutal du
  poste le 2026-09-18, le daemon a restauré tous les conteneurs en code 255, et
  `restart: on-failure` a relancé les cinq conteneurs dans la même seconde que les conteneurs
  `always`. Les conteneurs en `no` sont restés arrêtés. Le code de moby le confirme
  (`restartmanager.ShouldRestart` : `restart = exitCode != 0` pour on-failure, sans regarder un
  arrêt manuel). Le commentaire de la composition, le README et la skill `web-fetching` affirmaient
  le contraire.

## Décision

**Firecrawl et Scrapling ne sont plus enregistrés comme serveurs MCP.** L'agent les démarre
explicitement, après un échec constaté du palier 1, par la skill `web-fetching` :

- `firecrawl --start` lance la composition, attend que l'API réponde vraiment sur
  `/v0/health/liveness`, puis l'agent l'appelle en HTTP (`curl` sur l'adresse affichée) et finit par
  `firecrawl --stop`. Le pilote n'a plus de mode stdio.
- `scrapling <outil> [arguments-json]` démarre le conteneur nommé si besoin, y ouvre une session MCP
  ponctuelle par `docker exec`, fait un seul `tools/call`, affiche le résultat puis ferme la
  session. `scrapling --stop` arrête le conteneur.
- Tous les services de la composition Firecrawl sont en `restart: "no"`.

## Conséquences

- Une session qui ne lit pas le web ne lance plus rien : ni conteneur, ni processus `npx` ou Python.
- Firecrawl ne revient plus au démarrage du daemon. En contrepartie, un crash n'est plus réparé
  automatiquement : c'est le `firecrawl --start` lancé avant chaque usage qui répare.
- L'agent perd les outils MCP natifs. Il passe par `curl` pour Firecrawl et par un appel Bash pour
  Scrapling, dont il doit connaître les noms d'outils et les paramètres ; la skill les porte.
- Chaque appel Scrapling paie une poignée de main MCP d'environ une seconde, au lieu d'une seule par
  session. Ses sessions navigateur (`open_session`) ne survivent pas d'un appel à l'autre : chaque
  appel est un processus neuf.
- L'appel HTTP direct supprime `npx firecrawl-mcp`. Ce paquet n'était pas épinglé, et faute de
  `FIRECRAWL_API_URL` il envoyait silencieusement les URL à l'API hébergée.
- Arrêter reste la charge de la session qui a démarré. Un oubli coûte la pile jusqu'au prochain
  `--stop` ou redémarrage du poste, mais il ne se produit que dans les rares sessions qui en ont eu
  besoin, au lieu de toutes.
- L'enregistrement vivait hors du dépôt, dans `~/.claude.json` et `~/.claude_pro/.claude.json` : on
  le retire à la main, une fois par profil.

## Alternatives écartées

- **Enregistrement MCP au démarrage de chaque session**, l'état antérieur : fait payer environ 100 %
  des sessions pour en servir 1 %, mesures ci-dessus.
- **Un proxy MCP paresseux dans le pilote Go** : il servirait `initialize` et `tools/list` sans
  backend, depuis un cache pour Scrapling dont le serveur tourne dans le conteneur, puis démarrerait
  le backend au premier `tools/call`. Il garde les outils natifs, mais demande plusieurs centaines de
  lignes (horloge, exécuteur asynchrone, signaux, renumérotation des identifiants, cache indexé sur
  l'image `:latest`). Il laisse aussi le coût par session des processus hôte, et devrait suivre l'ère
  sans état de MCP (2026-07-28), qui retire `initialize`. Tout cela pour 8 appels en cinq semaines.
- **Corriger seulement la politique de redémarrage** : la pile ne reviendrait plus au redémarrage,
  mais la première session ouverte la relancerait quand même.
- **Un arrêt périodique sur inactivité**, par minuterie ou `launchd`. Déjà écarté par ADR-015 : sans
  proxy qui observe les appels, il ne distingue pas un conteneur orphelin d'un conteneur servant une
  session vivante, et arrêter Scrapling tue les `docker exec` des sessions connectées.
- **Une option native de lancement différé** : aucun des trois clients n'en a. Cursor en discute une
  (`startup_mode = "lazy"`), sans implémentation.
- **La CLI `scrapling extract`** plutôt qu'une session MCP ponctuelle : elle n'expose pas
  `cdp_url`. Or c'est par ce paramètre que Scrapling sert de client à CloakBrowser, pour le
  palier 3.
