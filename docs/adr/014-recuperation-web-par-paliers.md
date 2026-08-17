# ADR-014 — Récupération web par paliers, Firecrawl auto-hébergé

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `d6dd620` (Firecrawl et escalade), `3f9c683` (Scrapling et CloakBrowser), `71ba985`

## Contexte

Un agent qui lit le web se heurte à trois murs successifs, et un seul outil ne les
franchit pas tous : une page rendue côté client retourne une coquille vide au fetch
intégré ; une page protégée par Cloudflare bloque un client HTTP quelconque ; et
certaines protections résistent même à un client furtif.

Sans règle, l'agent choisit son outil au hasard — souvent le plus lourd, parce
qu'il « marche toujours ». Le coût n'est pas théorique : la pile Firecrawl tient
cinq conteneurs et plus de 6 Gio, contre zéro pour le fetch intégré.

Le choix de l'auto-hébergement porte une seconde question : une API hébergée voit
passer toutes les URL lues, donc le contenu des recherches et des lectures d'un
poste de travail professionnel.

## Décision

Trois paliers, décrits dans `harness/AGENTS.md`, avec une règle d'usage : **ne
jamais démarrer au-dessus du premier, et n'escalader qu'après un échec constaté**,
pas sur une intuition.

| Palier | Outil | Pour |
|---|---|---|
| 1 | fetch et recherche intégrés | le cas courant |
| 2 | Firecrawl auto-hébergé (`firecrawl-mcp`) | rendu côté client, lots, crawls |
| 3 | CloakBrowser (`cloak`) à travers Scrapling | protections anti-bot |

CloakBrowser n'est pas un serveur MCP : c'est un navigateur exposé en CDP, et Scrapling
en est le client, par son paramètre `cdp_url`.

`stealthy_fetch` de Scrapling **n'est pas un palier**, alors qu'il occupait cette place
à la rédaction initiale. Il exige Camoufox, absent de l'image `pyd4vinci/scrapling` et
impossible à y installer : le dépôt amont de Camoufox publie des tags mais aucune
release, si bien que son propre téléchargeur résout zéro version. Ses outils `get`,
`fetch`, `screenshot` et de session fonctionnent, et c'est à ce titre que Scrapling
reste enregistré — comme client CDP de CloakBrowser.

Firecrawl est auto-hébergé, jamais l'API publique. Son API tourne sans
authentification et n'écoute donc que sur `127.0.0.1`.

**Le panneau navigateur n'est pas un palier.** Il répond à un autre besoin —
l'interaction : cliquer, remplir, attendre un rendu. Le ranger dans l'escalade
laisserait croire qu'on s'y replie quand un fetch échoue.

Chaque palier est livré par un script de `~/.local/bin` adossé à un conteneur nommé
([ADR-015](015-mcp-en-conteneurs-nommes.md)).

## Conséquences

- Le coût suit le besoin : la majorité des lectures ne démarre aucun conteneur.
- Aucun tiers n'apprend quelles URL sont lues.
- **Il faut arrêter ce qu'on démarre.** `firecrawl-mcp --stop`, `scrapling-mcp --stop` ;
  CloakBrowser s'arrête seul après cinq minutes d'inactivité. Aucune pile ne
  redémarre avec le daemon docker, volontairement — d'où l'écart assumé avec le
  `restart: unless-stopped` de la source.
- Le premier démarrage télécharge environ 2 Gio d'images, à faire à la main : la
  poignée de main MCP expirerait avant la fin.
- Les images suivent `:latest` pour Firecrawl et Scrapling, comme leur amont le
  recommande. Aucun lockfile ne protège d'une régression ; le recours est d'épingler
  par variable d'environnement.
- Le `cdp_url` en `host.docker.internal` suppose qu'un port publié sur la boucle
  locale de l'hôte soit joignable depuis un conteneur. Vérifié sur macOS avec
  OrbStack : depuis le conteneur Scrapling, `host.docker.internal:9222` répond 200.
  **Faux sur Linux**, où il faudrait un réseau docker commun et `http://cloak:9222`.
- Les paliers 2 et 3 exigent docker, donc ne sont pas disponibles sur les cibles qui
  n'en ont pas — le NAS notamment ([ADR-008](008-dsm-cible-de-premier-rang.md)). Les
  scripts sortent en 69 plutôt que d'échouer obscurément.
- Le nom du pilote CloakBrowser est `cloak` et non `cloakbrowser` : le paquet npm de
  ce nom est installé et son shim Volta précède `~/.local/bin` dans le `PATH`.
- **Les trois paliers sont exercés**, chacun sur `example.com` : Firecrawl rend 180
  caractères de markdown avec le bon titre ; Scrapling répond en 3 s à froid et ses
  outils `get` et `fetch` rendent un statut 200 ; le palier 3 rend un statut 200 à
  travers CloakBrowser (Chrome 146). Rien n'a été exercé contre une protection
  anti-bot réelle, seulement contre une page inerte.

## Alternatives écartées

- **L'API Firecrawl hébergée** : supprime la pile locale et ses 6 Gio, mais fait
  transiter chez un tiers toutes les URL lues depuis un poste professionnel, et se
  paie. L'auto-hébergement échange du matériel contre de la confidentialité, ce qui
  est le bon sens de l'échange ici.
- **Pas de palier 2, le panneau navigateur suffisant** : il ne fait ni lot, ni
  crawl, ni recherche rendant des corps de page, et surtout il confond récupération
  et interaction — la confusion que cette ADR défait.
- **Scrapling seul, sans Firecrawl** : sert le palier 3 mais pas le rendu côté client
  à grande échelle, et n'a pas de recherche.
- **Un seul outil « qui marche toujours »** : c'est l'état par défaut sans règle, et
  il fait payer le coût du dernier palier à chaque lecture triviale.
- **`stealthy_fetch` comme palier anti-bot**, la forme de la source et de cette ADR à
  sa rédaction. Écartée par la mesure : Camoufox est introuvable en amont, et le
  contourner supposait `uv run --with 'camoufox[geoip]'` plus un volume de cache — un
  contournement d'un manque de l'image, à revérifier à chaque mise à jour, pour un
  palier que CloakBrowser couvre déjà et qui, lui, fonctionne.
