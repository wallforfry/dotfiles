# ADR-013 — Conventional Commits

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : tout l'historique depuis `b286cad`

## Contexte

Un dépôt de configuration se lit surtout par son historique : quand une machine se
met à mal fonctionner, la question est « qu'est-ce qui a changé, et pourquoi ».
C'est aussi la source des ADR de ce répertoire — celles antérieures à l'adoption de
chezmoi n'ont rien d'autre.

L'historique suit Conventional Commits depuis le premier commit, sans qu'aucun
outil ne l'impose : `feat`, `fix`, `docs` et `revert` sont utilisés, sans portée.

## Décision

Les messages suivent Conventional Commits : `<type>: <sujet à l'impératif>`. Les
types en usage sont `feat`, `fix`, `docs`, `revert`, sans portée — l'arborescence
est trop petite pour que la portée informe.

Le **corps** porte la raison, pas le contenu du diff : le problème observé, ce qui
a été tenté, pourquoi cette voie. C'est ce corps qui rend l'historique exploitable
et qui alimente le champ « Contexte » des ADR.

Le nettoyage adjacent va dans un commit séparé du commit fonctionnel, pour que
chacun se révoque indépendamment.

Aucun outil ne vérifie la forme : la convention tient par relecture. Un
`commitlint` supposerait Node sur toutes les machines du dépôt, y compris le NAS
([ADR-008](008-dsm-cible-de-premier-rang.md)).

## Conséquences

- L'historique se lit par type, et un `git log --oneline` répond à « qu'est-ce qui
  a changé de comportement » sans ouvrir les diffs.
- Les corps de commit deviennent la matière première des ADR. Le déclencheur
  documenté dans le [README](README.md) — un corps de plus d'une dizaine de lignes
  pour justifier un choix — n'existe que parce que ces corps sont substantiels.
- Rien n'empêche mécaniquement un message mal formé.
- **Point non tranché** : l'accentuation des messages. Les premiers commits sont
  accentués (`feat: exécuter les scripts chezmoi hors de /tmp`), les plus récents
  ne le sont pas (`fix: ne pas ecraser SSH_AUTH_SOCK…`). La dérive est constatée,
  pas décidée ; à trancher dans un sens ou dans l'autre, cette ADR étant alors à
  remplacer.

## Alternatives écartées

- **`commitlint` avec un hook** : impose Node comme dépendance du dépôt, pour une
  convention déjà tenue par relecture sur un dépôt à un seul auteur.
- **Messages libres** : perdrait la lisibilité par type, et surtout l'habitude
  d'écrire le pourquoi dans le corps — d'où les ADR de ce répertoire tirent leur
  substance.
- **Portées (`feat(zsh):`)** : quatre répertoires de premier niveau, la portée
  répéterait le chemin du diff.
