# ADR-021 - Codex second hôte, skills hors de `dot_claude/`

- **Statut** : accepté
- **Date** : 2026-09
- **Commits** : à compléter au commit qui porte ce fichier

## Contexte

L'installation de Codex - l'application ChatGPT de bureau, qui lit `CODEX_HOME=~/.codex` -
a montré que le montage d'[ADR-012](012-instructions-ia-versionnees.md) ne tenait pas sa
promesse. Elle annonçait qu'« ajouter un agent (Codex, Cursor) coûte une projection, pas
une copie ». Mesures faites sur la machine avant tout changement :

- `~/.codex/AGENTS.md` absent, aucune skill visible : ni règle, ni voix, ni procédure.
- `harness/AGENTS.md` portait deux lignes `@SOUL.md` et `@USER.md`. Cet import Markdown
  est propre à Claude Code. Déployée telle quelle, la source canonique aurait donné à
  Codex les règles techniques sans la voix ni les préférences, et les deux lignes
  auraient été rendues comme du texte.
- Les skills vivaient sous `dot_claude/skills/`, chemin propre à un hôte, alors que leur
  contenu ne l'est pas. Un lien symbolique de chezmoi ne peut pas viser l'arbre source :
  `harness/` est dans `.chezmoiignore` et n'a pas de destination.
- **Les deux répertoires de skills portent de l'état que ce dépôt ne possède pas** :
  `~/.codex/skills/.system/` porte six skills livrées par l'application et son marqueur,
  réinstallés par elle ; `~/.claude/skills/` porte un `find-skills` installé hors de ce
  dépôt. Un lien symbolique sur le répertoire entier les supprime : mesuré en bac à sable
  sur un `HOME` factice, chezmoi remplace le répertoire d'un bloc, sans invite, sans
  avertissement et sans sauvegarde.

Le binaire porte par ailleurs les preuves de ce que Codex sait faire, relevées par
`strings` faute de documentation locale : moteur de hooks (événements `Stop`,
`PreCompact`, `SessionStart`…, un seul type de gestionnaire opérant, `command`), et
sous-agents (`spawn_agent`, rôles déclarés en TOML sous `~/.codex/agents`).

## Décision

Codex est un hôte de premier rang, servi par la même source que Claude :

| Chemin | Rôle |
|---|---|
| `dot_codex/AGENTS.md.tmpl` | adaptateur Codex : inline les trois sources dans `~/.codex/AGENTS.md` |
| `dot_codex/symlink_CONTEXT.md` | lien vers `~/.claude/CONTEXT.md`, fragment chiffré partagé |
| `dot_config/agent-skills/` | les skills, agnostiques, déployées dans `~/.config/agent-skills` |
| `dot_{claude,codex}/skills/symlink_<slug>` | **un lien par skill**, jamais sur le répertoire |

Les sources de `harness/` ne portent plus aucune syntaxe propre à un hôte : charger les
trois fichiers est le travail de l'adaptateur. Chez Claude, `dot_claude/CLAUDE.md`
importe `@AGENTS.md`, `@SOUL.md`, `@USER.md` et `@CONTEXT.md` ; chez Codex, le template
les inline, car Codex n'expanse aucun import et désigne `CONTEXT.md` par une consigne de
lecture.

`scripts/verify.sh` gagne une section « Projections d'instructions » : elle refuse un
import `@` dans `harness/`, exige que chaque adaptateur charge les trois sources, exige
un lien par skill et par hôte vers la source unique, refuse un lien orphelin, et
**refuse un `symlink_skills` sur le répertoire entier** - le défaut qui détruirait l'état
vivant mesuré ci-dessus.

## Conséquences

- Ajouter un troisième hôte coûte un répertoire `dot_<hôte>/` avec un adaptateur et un
  lien par skill. La promesse d'ADR-012 est tenue, cette fois vérifiée par la barrière.
- Le chemin canonique des skills change : `dot_config/agent-skills/`, et non plus
  `dot_claude/skills/`. C'est l'amendement d'ADR-012 le plus visible ; toute référence à
  l'ancien chemin dans une skill, la barrière ou l'audit est fausse.
- **Ajouter une skill coûte désormais deux liens**, un par hôte. C'est le prix de ne pas
  toucher au répertoire vivant ; la barrière le réclame, elle ne le pose pas.
- `~/.claude/skills/README.md` n'est plus déployé : l'index vit avec la source, dans
  `~/.config/agent-skills`. Le répertoire d'un hôte ne porte plus que des skills.
- `~/.codex` tient de l'état vivant - bases SQLite, `auth.json` - donc l'interdiction
  d'`exact_` d'ADR-012 s'y étend, et la barrière la contrôle. Elle ne suffit pas seule :
  un lien nommant un chemin que l'application possède détruit tout de même, d'où le
  contrôle dédié.
- `CONTEXT.md` n'est pas conditionné au profil, comme chez Claude : les deux hôtes savent
  la même chose de la machine, et il n'y a pas de garde à laisser pourrir.
- **Le hook `agent-handoff` reste propre à Claude.** Codex a bien un événement `Stop`,
  mais son gestionnaire n'est actif qu'une fois son empreinte approuvée à la main
  (`hooks.state.<clé>.trusted_hash`), et cette empreinte change à chaque édition : un
  `chezmoi apply` ne peut pas rendre un hook vivant. Le script lit de surcroît un
  transcript au format Claude et un `CLAUDE_CODE_AUTO_COMPACT_WINDOW`. Le porter sans
  pouvoir l'exercer produirait une garantie non vérifiée ; la skill `handoff` reste
  invocable à la main sur Codex.
- La règle de délégation d'`harness/AGENTS.md` fonctionne sur Codex par chance et non par
  construction : `spawn_agent` refuse de se déclencher sauf si un `AGENTS.md` ou une skill
  demande explicitement la délégation, ce que cette règle fait. La supprimer y
  désactiverait les sous-agents, pas seulement l'incitation à s'en servir.
- **Ce que Codex fait de ces fichiers n'est pas exercé.** Que `~/.codex/AGENTS.md` soit
  bien chargé et qu'un `~/.codex/skills/<slug>` lien soit bien suivi se déduit du binaire,
  pas d'une session observée. La première session Codex est la vérification qui manque.

## Alternatives écartées

- **Un lien sur le répertoire de skills entier**, un par hôte : trois fichiers au lieu de
  vingt-quatre, et c'est ce qui avait été écrit d'abord. Écarté sur mesure : il supprime
  les six skills système de Codex et le `find-skills` de `~/.claude`.
- **Laisser les skills sous `dot_claude/skills/` et y faire pointer `~/.codex/skills`** :
  un diff minimal, mais le chemin canonique reste celui d'un hôte, et le prochain agent
  hérite de la même dette.
- **Les skills sous `harness/skills/`** : le chemin le plus cohérent avec ADR-012.
  Impossible : `harness/` n'est pas déployé, et un `symlink_` de chezmoi ne peut viser que
  la destination, jamais l'arbre source. Faire viser le clone chezmoi lierait la
  configuration déployée à la présence du clone.
- **Copier `harness/` dans `dot_codex/`** : deux copies à synchroniser, exactement le
  défaut qu'ADR-012 a supprimé pour `~/.claude_pro`.
- **Garder `@SOUL.md` dans `harness/AGENTS.md` et laisser Codex l'ignorer** : la source
  la plus chargée du harness serait alors incomplète chez un hôte sur deux, sans que rien
  ne le dise.
- **Porter le hook `agent-handoff` sur Codex dès maintenant** : le format d'entrée par
  événement n'est pas déterminable depuis le binaire, et l'approbation d'empreinte est
  manuelle. Un hook livré sans être exercé est une barrière annoncée verte sans preuve.
