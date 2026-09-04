# Architecture Decision Records

Décisions structurantes de ce dépôt. Format inspiré de
[MADR](https://adr.github.io/madr/), réduit à cinq sections.

Une ADR répond à une seule question : **pourquoi est-ce ainsi et pas autrement ?**
Ce que fait le code se lit dans le code ; ce qui a été écarté, et pourquoi, ne se
lit nulle part.

## Quand en créer une

Les trois conditions doivent être réunies :

1. **La décision contraint la suite.** Choisir autrement obligerait à retoucher
   plusieurs fichiers, à changer une habitude, ou à migrer des machines déjà
   installées.
2. **Une alternative crédible existe.** Quelqu'un - toi dans six mois, un agent -
   pourrait raisonnablement proposer l'inverse. Sans alternative, il n'y a pas de
   décision, juste un fait.
3. **La justification n'est pas déjà lisible.** Si un commentaire dans le fichier
   la porte suffisamment, l'ADR ne fait que dupliquer.

**Déclencheur pratique** : si tu rédiges un corps de message de commit de plus
d'une dizaine de lignes pour justifier un choix, ou si tu y écris « plutôt que »,
c'est une ADR qui cherche à sortir.

### Quand ne pas en créer

- Ajouter un outil au script d'installation, bouger une version épinglée,
  éditer une skill : routine, aucune ADR.
- Un correctif, même subtil : son message de commit est le bon support. Le
  correctif devient une ADR seulement s'il révèle une contrainte durable de la
  plateforme - c'est le cas d'[ADR-008](008-dsm-cible-de-premier-rang.md).
- Une préférence de style : elle va dans `AGENTS.md`, pas ici.
- Un sujet déjà couvert par une ADR en vigueur : l'amender, ou la remplacer.

### Le cas du retrait

Renoncer à un outil est une décision comme une autre, et la plus utile à
consigner : sans trace, il sera réadopté. La décision en vigueur est alors « ne
pas utiliser », et le contexte porte la mesure qui l'a motivée.

## Comment en créer une

1. **Numéro** = le plus grand existant + 1. Jamais réutilisé, jamais renuméroté :
   l'index, les autres ADR et l'historique y renvoient.
2. **Fichier** `NNN-titre-en-kebab-case.md`, titre nominal et français.
3. **Cinq sections, dans cet ordre** :

   | Section | Contenu |
   |---|---|
   | En-tête | statut, date au mois, commits qui portent la décision |
   | `## Contexte` | le problème observé, avec ses preuves : message d'erreur, mesure, citation de commit |
   | `## Décision` | une phrase testable, à l'indicatif présent. Ce qu'on fait, pas ce qu'on aimerait |
   | `## Conséquences` | ce que ça coûte autant que ce que ça gagne. Une ADR sans coût est une publicité |
   | `## Alternatives écartées` | au moins une, avec la raison du rejet |

4. **Citer les commits** (`git log --oneline`) qui portent la décision. C'est ce
   qui permet à un lecteur de vérifier sans te croire.
5. **Ajouter la ligne à l'index**, en ordre numérique.
6. **Ne jamais réécrire une ADR en vigueur pour changer d'avis.** En écrire une
   nouvelle qui la remplace, puis supprimer l'ancienne en reportant ses
   alternatives dans la section « Alternatives écartées » de la nouvelle.

### Vérification

L'index et le répertoire doivent coïncider :

```bash
diff <(ls docs/adr | grep -oE '^[0-9]{3}') <(grep -oE '^\| \[[0-9]{3}\]' docs/adr/README.md | grep -oE '[0-9]{3}')
```

## Portée

Seules les décisions **en vigueur** sont enregistrées. Une décision remplacée ne
garde pas son fichier : elle réapparaît dans la section « Alternatives écartées »
de celle qui l'a remplacée.

C'est un choix, contre l'usage courant qui conserve les ADR périmées avec un
statut `superseded`. Il garde le répertoire lisible et rend l'index digne de
confiance - tout ce qui y figure s'applique. Il perd l'historique de la
délibération, récupérable par `git log -- docs/adr/`.

## Langue

Les ADR sont rédigées en français, par exception à la règle « instructions
d'agent en anglais » de l'`AGENTS.md` racine : elles consignent un raisonnement
personnel, pas une interface. Identifiants, commandes, chemins et corps de commit
cités restent verbatim.

## Fiabilité des motivations

Ce dépôt commence à l'adoption de chezmoi, en 2026-08 ; l'historique du bare repo
qui l'a précédé n'a pas été repris. Les ADR antérieures à cette bascule
reconstituent donc une motivation depuis l'état du code et le contenu des
commits, sans historique pour l'étayer. Le champ **Contexte** est une
reconstitution a posteriori, sauf lorsqu'il cite le corps d'un commit - dans ce
cas la citation est explicite.

## Index

| # | Titre | Date |
|---|---|---|
| [001](001-chezmoi-gestionnaire.md) | chezmoi comme gestionnaire de dotfiles | 2026-08 |
| [004](004-profils-perso-pro.md) | Profils `perso` et `pro` demandés à l'init | 2026-08 |
| [005](005-oh-my-zsh-en-external.md) | oh-my-zsh en external, auto-update désactivé | 2026-08 |
| [006](006-starship-comme-prompt.md) | starship comme prompt unique | 2026-08 |
| [007](007-outillage-run-onchange.md) | Outillage installé par un `run_onchange`, `~/bin` via `.zshenv` | 2026-08 |
| [008](008-dsm-cible-de-premier-rang.md) | Synology DSM comme cible de premier rang | 2026-08 |
| [009](009-fragments-ssh-versionnes.md) | Fragments SSH versionnés, `~/.ssh/config` non | 2026-08 |
| [010](010-gpg-agent-ssh-sous-condition.md) | GPG comme agent SSH, sous condition | 2026-08 |
| [011](011-identite-git-par-repertoire.md) | Identité git par répertoire via `includeIf` | 2026-08 |
| [012](012-instructions-ia-versionnees.md) | Instructions IA versionnées, `harness/` source unique | 2026-08 |
| [013](013-conventional-commits.md) | Conventional Commits | 2026-08 |
| [014](014-recuperation-web-par-paliers.md) | Récupération web par paliers, Firecrawl auto-hébergé | 2026-08 |
| [015](015-mcp-en-conteneurs-nommes.md) | Serveurs MCP lourds en conteneurs nommés | 2026-08 |
| [016](016-depot-public-sensible-chiffre.md) | Dépôt public, sensible chiffré et hors prose | 2026-09 |
| [017](017-clone-anonyme-gh-en-ecriture.md) | Clone anonyme, `gh` comme helper d'écriture | 2026-09 |
| [018](018-chiffrement-par-paire-de-cles.md) | Chiffrement par paire de clés `age`, clé déverrouillée une fois | 2026-09 |
| [019](019-poste-graphique-et-apps-macos.md) | Applications macOS derrière une donnée `gui`, posées dans `~/Applications` | 2026-09 |
| [020](020-verification-en-ci.md) | Barrière rejouée en CI, clé `age` en secret | 2026-09 |
| [021](021-codex-second-hote.md) | Codex second hôte, skills hors de `dot_claude/` | 2026-09 |
