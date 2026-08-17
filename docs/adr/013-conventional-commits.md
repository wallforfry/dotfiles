# ADR-013 - Conventional Commits

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : tout l'historique depuis `b286cad`

## Contexte

Un dépôt de configuration se lit surtout par son historique : quand une machine se
met à mal fonctionner, la question est « qu'est-ce qui a changé, et pourquoi ».
C'est aussi la source des ADR de ce répertoire - celles antérieures à l'adoption de
chezmoi n'ont rien d'autre.

L'historique suit Conventional Commits depuis le premier commit, sans qu'aucun
outil ne l'impose : `feat`, `fix`, `docs` et `revert` sont utilisés, sans portée.

L'accentuation, elle, a dérivé sans décision. Un seul sujet de tout l'historique
porte ses accents - `fix: exécuter les scripts chezmoi hors de /tmp` (`4a7d1e8`) -
alors que trois autres les perdent sur des mots qui en attendent : `ne pas ecraser`,
`un fragment ssh versionne`, `les biais deduits`. Les corps sont majoritairement
accentués (12 sur 17). La dérive va donc vers l'ASCII dans les sujets, sans qu'aucune
contrainte technique ne l'explique : `i18n.commitEncoding` et
`i18n.logOutputEncoding` sont aux valeurs par défaut, c'est-à-dire UTF-8.

## Décision

Les messages suivent Conventional Commits : `<type>: <sujet à l'impératif>`. Les
types en usage sont `feat`, `fix`, `docs`, `revert`, sans portée - l'arborescence
est trop petite pour que la portée informe.

Le **corps** porte la raison, pas le contenu du diff : le problème observé, ce qui
a été tenté, pourquoi cette voie. C'est ce corps qui rend l'historique exploitable
et qui alimente le champ « Contexte » des ADR.

**Un message français est accentué**, sujet et corps. La langue du message suit
celle du dépôt : française ici, anglaise dans un dépôt anglophone - et un message
anglais n'a pas de question d'accentuation. Comme partout ailleurs, l'exception
porte sur la prose seule : identifiants, commandes, chemins, noms de fichiers et
sorties citées restent verbatim.

Le nettoyage adjacent va dans un commit séparé du commit fonctionnel, pour que
chacun se révoque indépendamment.

Aucun outil ne vérifie la forme : la convention tient par relecture. Un
`commitlint` supposerait Node sur toutes les machines du dépôt, y compris le NAS
([ADR-008](008-dsm-cible-de-premier-rang.md)).

## Conséquences

- L'historique se lit par type, et un `git log --oneline` répond à « qu'est-ce qui
  a changé de comportement » sans ouvrir les diffs.
- Les corps de commit deviennent la matière première des ADR. Le déclencheur
  documenté dans le [README](README.md) - un corps de plus d'une dizaine de lignes
  pour justifier un choix - n'existe que parce que ces corps sont substantiels.
- Rien n'empêche mécaniquement un message mal formé.
- **`git log --grep` devient sensible aux accents**, et c'est le coût réel de
  l'accentuation. Mesuré sur ce dépôt : `--grep=executer` ne trouve rien,
  `--grep=exécuter` trouve le commit, et aucun radical non accentué ne s'en
  approche - `--grep=ecuter` ne trouve rien non plus, l'accent coupant le mot en
  deux. Le contournement est de chercher en expression régulière, l'accent
  remplacé par un point : `git log -E --grep='ex.cuter'` trouve les deux formes.
- **L'historique reste mixte.** La règle vaut pour les messages à venir ; les
  messages déjà écrits ne sont pas réécrits. Réécrire l'historique pour des accents
  changerait tous les hachages du dépôt pour un gain nul.

## Alternatives écartées

- **`commitlint` avec un hook** : impose Node comme dépendance du dépôt, pour une
  convention déjà tenue par relecture sur un dépôt à un seul auteur.
- **Messages libres** : perdrait la lisibilité par type, et surtout l'habitude
  d'écrire le pourquoi dans le corps - d'où les ADR de ce répertoire tirent leur
  substance.
- **Portées (`feat(zsh):`)** : quatre répertoires de premier niveau, la portée
  répéterait le chemin du diff.
- **Messages français en ASCII**, la direction vers laquelle l'historique dérivait.
  C'est l'alternative sérieuse : `git log --grep` reste littéral, et rien ne peut
  s'afficher en mojibake sur un terminal mal configuré. Rejetée parce que le reste
  du dépôt est du français accentué - README, ADR, commentaires des scripts - et
  que le sujet de commit en serait le seul endroit dégradé, pour un argument
  d'encodage qui ne tient pas : `i18n.commitEncoding` et `i18n.logOutputEncoding`
  sont à UTF-8 par défaut, et git, GitHub et les terminaux utilisés le gèrent. Le
  coût sur `--grep` est réel, mais il se contourne d'un caractère.
