# ADR-012 - Instructions IA versionnées, `harness/` source unique

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `45f0998` (harness et projections), `64f0ccb` (skills et hook)

## Contexte

`~/.claude` n'était pas versionné. Son `CLAUDE.md` se réduisait à un `@RTK.md` de
huit octets, aucune skill personnelle n'existait, et le second répertoire de
configuration `~/.claude_pro` - ouvert par un alias du profil `pro` - était une
copie à maintenir à la main.

Les instructions données à un agent sont de la configuration comme une autre : leur
perte au changement de machine coûte le même travail que celle d'un `.zshrc`, et
leur divergence entre deux répertoires produit deux comportements différents pour
la même demande.

Les dépôts de travail utilisent déjà le motif `CLAUDE.md` → `@AGENTS.md` avec un
`AGENTS.md` comme base partagée. Rien ne justifiait un autre montage à l'échelle de
l'utilisateur.

## Décision

`harness/` est la source unique des instructions, agnostique de l'agent :

| Fichier | Rôle |
|---|---|
| `harness/AGENTS.md` | règles techniques valables sur tout projet |
| `harness/SOUL.md` | voix de l'agent : langue, registre, priorités |
| `harness/USER.md` | contexte et attentes de l'utilisateur |

`dot_claude/{AGENTS,SOUL,USER}.md.tmpl` n'en sont que des projections d'une ligne,
par la fonction `include` de chezmoi, et `dot_claude/CLAUDE.md` l'adaptateur Claude
qui les importe. `dot_claude_pro/` ne contient que des liens relatifs vers
`~/.claude`.

Les skills vivent sous `dot_claude/skills/<slug>/SKILL.md`. Ce qui est
inconditionnel va dans `harness/AGENTS.md`, toujours chargé ; ce qui est une
procédure conditionnelle va dans une skill, chargée au besoin.

## Conséquences

- Une instruction s'écrit une fois et vaut pour les deux répertoires de
  configuration, sans synchronisation.
- Ajouter un agent (Codex, Cursor) coûte une projection, pas une copie.
- Deux endroits à ne pas confondre : éditer une projection au lieu de la source
  produit un changement qu'un `chezmoi apply` écrasera. Le rappel est dans
  `AGENTS.md`.
- `~/.claude/settings.json` reste hors du dépôt : il mêle des chemins absolus
  écrits par des installeurs tiers, de l'état de plugins, une `statusLine`. Le
  déployer l'écraserait.
  **Amendement** : la conséquence « le hook `agent-handoff` doit être enregistré
  à la main » ne tient plus. `run_onchange_after_register-claude-hooks.sh.tmpl`
  fusionne cette entrée dans le fichier vivant, sans toucher au reste. Le fichier
  demeure hors du dépôt : c'est la fusion qui est versionnée, pas la
  configuration.
- **Ne jamais ajouter l'attribut `exact_` à `dot_claude/`** : `~/.claude` contient
  l'état vivant des sessions et des projets, que chezmoi supprimerait.

## Alternatives écartées

- **Laisser `~/.claude` non versionné** : l'état de départ, une perte sèche au
  changement de machine.
- **Un outil de projection dédié**, comme l'`arnes` du dépôt dont ce montage
  s'inspire : chezmoi assure déjà le déploiement, un second moteur serait redondant.
- **Instructions directement dans `dot_claude/`, sans `harness/`** : plus court d'un
  répertoire, mais range des règles agnostiques dans un chemin propre à un agent, et
  rend le prochain agent plus coûteux à ajouter.
- **Versionner `settings.json`** : le fichier est réécrit par des installeurs et
  porte des chemins absolus, donc un conflit à chaque diff.
