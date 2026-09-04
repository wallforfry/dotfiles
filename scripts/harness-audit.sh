#!/usr/bin/env bash
# Mesures du harness agentique, lancées à la main ou depuis la skill harness-audit.
#
# bash et non sh POSIX : ne tourne jamais au bootstrap, et python3 est requis
# pour lire les transcripts JSONL.
#
# Cinq mesures : le coût du contexte toujours chargé, le retard de la source de
# déploiement, l'activation réelle des skills et des subagents, l'adhérence aux
# deux règles observables, et le pouvoir de détection de scripts/verify.sh par
# injection de défauts dans un clone.
#
# N'imprime jamais un chemin de projet : ils portent des noms de clients (ADR-016).
# Sort en 1 si la barrière laisse passer un défaut, ou si une mesure n'a pas pu
# être faite : une mesure absente n'est pas une mesure verte.

set -uo pipefail

# La racine vient de l'emplacement du script, jamais du répertoire courant :
# la skill est lancée depuis n'importe quel dépôt, et un rev-parse sur le cwd
# mesurerait cet autre dépôt.
src=${BASH_SOURCE[0]:-}
if [ -z "$src" ]; then
  echo "❌  emplacement du script inconnu : lancer bash <chemin>/harness-audit.sh" >&2
  exit 1
fi
# pwd -P résout les répertoires, jamais le fichier final : sans cette boucle, un
# lanceur en lien symbolique donne le parent du lien, donc un autre dépôt.
while [ -L "$src" ]; do
  cible=$(readlink "$src")
  case $cible in
    /*) src=$cible ;;
    *) src=$(dirname -- "$src")/$cible ;;
  esac
done
root=$(cd -- "$(dirname -- "$src")/.." && pwd -P) || exit 1
if ! git -C "$root" rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "❌  $root hors d'un dépôt git : aucune mesure possible" >&2
  exit 1
fi
cd "$root"

fail=0
ok() { printf '  ✅  %s\n' "$1"; }
ko() { printf '  ❌  %s\n' "$1" >&2; fail=1; }
head_() { printf '\n== %s\n' "$1"; }

PROJECTS="${CLAUDE_PROJECTS:-$HOME/.claude/projects}"
# Date d'introduction des deux règles observables (ce5cc72, 45f0998).
SINCE="${HARNESS_RULES_SINCE:-2026-08-17}"

if ! command -v python3 >/dev/null; then
  echo "❌  python3 absent : aucune mesure possible" >&2
  exit 1
fi

# --- 1. contexte toujours chargé ---------------------------------------------
head_ "Contexte toujours chargé"
# Les descriptions de frontmatter comptent : elles sont le routeur, donc
# toujours en contexte. Sans elles, déplacer une section vers une skill
# paraîtrait gratuit alors qu'il déplace une part du coût dans le routeur.
total=0
mesure() { # mesure <étiquette> <octets>
  total=$((total + $2))
  printf '  %-34s %6s o  ~%5s jetons\n' "$1" "$2" "$(($2 / 4))"
}
for f in harness/AGENTS.md harness/SOUL.md harness/USER.md AGENTS.md dot_claude/CLAUDE.md; do
  if [ ! -f "$f" ]; then
    ko "$f absent : coût du contexte non mesuré"
    continue
  fi
  mesure "$f" "$(wc -c <"$f" | tr -d ' ')"
done
# Une description est un bloc plié « description: > » : tout jusqu'à la
# prochaine clé de premier niveau.
descs=$(awk 'FNR==1{d=0}
             /^description:/{d=1; n+=length($0)+1; next}
             d && /^[^ \t]/{d=0}
             d {n+=length($0)+1}
             END{print n+0}' dot_config/agent-skills/*/SKILL.md dot_claude/agents/*.md)
mesure "descriptions (routeur des skills)" "$descs"
ok "$total octets, ~$((total / 4)) jetons sur chaque tâche de ce dépôt"
echo "  Le CONTEXT.md de chaque hôte est chiffré : son coût n'est pas mesurable ici"

# --- 2. retard de la source de déploiement ------------------------------------
head_ "Source de déploiement"
# La source chezmoi est un clone distinct du dépôt de travail (ADR-001) : un
# commit fusionné ne prend effet qu'après un chezmoi update. Aucun signal ne le
# disait, chezmoi status comparant la source à la destination et jamais à la
# remote. Pas de fetch ici : la mesure reste locale, donc toujours possible, et
# se lit par rapport à la dernière récupération.
if ! command -v chezmoi >/dev/null; then
  ko "chezmoi absent : retard de la source non mesuré"
else
  src=$(chezmoi source-path)
  if [ "$src" = "$PWD" ]; then
    ok "clone unique : la source de déploiement est ce dépôt"
  elif ! git -C "$src" rev-parse --verify -q origin/main >/dev/null; then
    ko "$(basename "$src") : origin/main inconnue, retard non mesuré"
  else
    behind=$(git -C "$src" rev-list --count HEAD..origin/main)
    ok "source en retard de $behind commits sur origin/main à la dernière récupération"
  fi
fi

# --- 3. activation réelle et adhérence observable -----------------------------
# Un seul parcours des JSONL : les deux mesures lisent les mêmes lignes, et deux
# parcours de plusieurs centaines de sessions coûtent le double pour rien.
head_ "Activation et adhérence (règles introduites le $SINCE)"
if [ ! -d "$PROJECTS" ]; then
  ko "$PROJECTS absent : activation et adhérence non mesurées"
elif ! python3 - "$PROJECTS" "$SINCE" <<'PY'
import json, os, sys, glob, collections
root, since = sys.argv[1], sys.argv[2]
EXT = ('.ts', '.tsx', '.js', '.py', '.lua', '.rs', '.go', '.sh')
skills = collections.Counter(); agents = collections.Counter()
blocks = collections.Counter(); dash = collections.Counter()
lines = collections.Counter(); comments = collections.Counter()
sessions = 0; days = []
for f in glob.glob(os.path.join(root, '**', '*.jsonl'), recursive=True):
    sessions += 1
    with open(f, encoding='utf-8', errors='replace') as fh:
        for line in fh:
            if not line.strip():
                continue
            try:
                d = json.loads(line)
            except ValueError:
                continue
            ts = (d.get('timestamp') or '')[:10]
            if ts:
                days.append(ts)
            era = 'avant' if ts < since else 'depuis'
            m = d.get('message') or {}
            c = m.get('content')
            if not isinstance(c, list):
                continue
            for b in c:
                if not isinstance(b, dict):
                    continue
                kind = b.get('type')
                if kind == 'text' and m.get('role') == 'assistant':
                    blocks[era] += 1
                    dash[era] += b.get('text', '').count('—')
                elif kind == 'tool_use':
                    name = b.get('name'); i = b.get('input') or {}
                    if name == 'Skill':
                        skills[str(i.get('skill'))] += 1
                    elif name in ('Task', 'Agent'):
                        agents[str(i.get('subagent_type'))] += 1
                    elif name == 'Write' and str(i.get('file_path', '')).endswith(EXT):
                        for L in str(i.get('content', '')).split('\n'):
                            t = L.strip()
                            if not t:
                                continue
                            lines[era] += 1
                            if t.startswith(('//', '#', '--', '/*', '*')):
                                comments[era] += 1
if not sessions:
    sys.exit('  aucun transcript lu')
print(f'  {sessions} sessions, {min(days)} au {max(days)}' if days else f'  {sessions} sessions')
local = sorted(d for d in os.listdir('dot_config/agent-skills')
               if os.path.exists(f'dot_config/agent-skills/{d}/SKILL.md'))
print('  skills de ce dépôt :')
for name in local:
    print(f'    {skills.get(name, 0):5d}  {name}')
print('  autres skills activées, top 5 :')
for name, count in [(n, c) for n, c in skills.most_common() if n not in local][:5]:
    print(f'    {count:5d}  {name}')
print('  subagents :')
for name, count in agents.most_common(6):
    print(f'    {count:5d}  {name}')
def rate(a, b):
    return a / b if b else 0.0
print('  tiret cadratin par bloc de texte assistant :')
for era in ('avant', 'depuis'):
    print(f'    {era:6s}  {blocks[era]:6d} blocs  {dash[era]:6d} occurrences  {rate(dash[era], blocks[era]):.2f}')
print('  lignes de commentaire dans le code écrit :')
for era in ('avant', 'depuis'):
    print(f'    {era:6s}  {lines[era]:6d} lignes  {comments[era]:6d} commentaires  {rate(comments[era], lines[era]):.3f}')
PY
then
  ko "lecture des transcripts interrompue : activation et adhérence non mesurées"
else
  ok "mesuré sur les transcripts disponibles"
  echo "  l'adhérence est corrélationnelle : le modèle a changé sur la même période"
fi

# --- 4. pouvoir de détection de la barrière -----------------------------------
head_ "Pouvoir de détection de scripts/verify.sh"
mkdir -p "$HOME/.cache"
tmp=$(mktemp -d "$HOME/.cache/harness-audit.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
branche=$(git rev-parse --abbrev-ref HEAD)
if ! git clone -q "$PWD" "$tmp/rep" 2>/dev/null; then
  ko "clone impossible : détection non mesurée"
else
  # La branche est résolue avant le cd : $PWD s'évalue après, donc dans le clone.
  (cd "$tmp/rep" && git checkout -q "$branche" 2>/dev/null) || true
  if ! (cd "$tmp/rep" && bash scripts/verify.sh >/dev/null 2>&1); then
    ko "le clone est rouge avant toute mutation : détection non mesurable"
  else
    detected=0
    count=0
    # Sur le fd 3 : sur stdin, git, python3 et verify.sh consomment le heredoc
    # et sautent des mutations, ce qui rendait le compte non reproductible.
    while IFS='|' read -r label code <&3; do
      [ -z "$label" ] && continue
      count=$((count + 1))
      # Une mutation qui ne s'applique plus - chaîne cherchée disparue - serait
      # rapportée « non détectée », donc attribuée à la barrière.
      if ! (cd "$tmp/rep" && git checkout -q -- . && git clean -qfd && python3 -c "$code"); then
        ko "mutation inapplicable, à réécrire : $label"
      elif (cd "$tmp/rep" && bash scripts/verify.sh >/dev/null 2>&1); then
        ko "mutation non détectée : $label"
      else
        detected=$((detected + 1))
      fi
    done 3<<'MUT'
name d'une skill différent du répertoire|p='dot_config/agent-skills/adr/SKILL.md';s=open(p).read();open(p,'w').write(s.replace('name: adr','name: adrx',1))
skill retirée de l'index|import re;p='dot_config/agent-skills/README.md';s=open(p).read();open(p,'w').write('\n'.join(l for l in s.split('\n') if not re.match(r'^\| `adr`',l)))
metadata.category hors de {dev, ops}|p='dot_config/agent-skills/adr/SKILL.md';s=open(p).read();open(p,'w').write(s.replace('category: ops','category: misc').replace('category: dev','category: misc'))
erreur de syntaxe shell|p='scripts/verify.sh';s=open(p).read();open(p,'w').write(s+'\nif true; then\n')
template invalide|open('dot_gitconfig.tmpl','a').write('{{ .absent.champ }}\n')
ADR présente hors index|import glob,shutil;shutil.copy(glob.glob('docs/adr/001-*.md')[0],'docs/adr/099-fantome.md')
fragment .age en clair|open('age-key.txt.age','w').write('texte en clair\n')
chezmoi diff introduit en CI|p='.github/workflows/verify.yml';s=open(p).read();open(p,'w').write(s.replace('chezmoi status','chezmoi diff'))
skill sans ligne d'index|import os;os.makedirs('dot_config/agent-skills/fantome',exist_ok=True);open('dot_config/agent-skills/fantome/SKILL.md','w').write('---\nname: fantome\ndescription: Rien. Use when jamais.\nmetadata:\n  category: dev\n---\n# Fantome\n')
sous-répertoire de skill hors liste|import os;os.makedirs('dot_config/agent-skills/adr/evals',exist_ok=True);open('dot_config/agent-skills/adr/evals/x.json','w').write('{}')
attribut exact_ sur l'état vivant|open('dot_claude/exact_zzz','w').write('')
import @ réintroduit dans une source agnostique|open('harness/AGENTS.md','a').write('@SOUL.md\n')
hôte privé d'une skill|import os;os.remove('dot_codex/skills/symlink_adr')
lien de skill vers un arbre étranger|open('dot_claude/skills/symlink_adr','w').write('../../.agents/skills/adr\n')
lien sur le répertoire de skills entier|open('dot_codex/symlink_skills','w').write('../.config/agent-skills\n')
harness absent de l'adaptateur Codex|p='dot_codex/AGENTS.md.tmpl';s=open(p).read();open(p,'w').write(s.replace('{{ include "harness/USER.md" }}\n',''))
MUT
    if [ "$detected" -eq "$count" ]; then
      ok "$detected/$count défauts injectés détectés"
    else
      ko "$detected/$count défauts injectés détectés"
    fi
  fi
fi

printf '\n'
if [ "$fail" -eq 0 ]; then
  echo "🎉  mesures complètes"
else
  echo "💥  une mesure manque ou la barrière laisse passer un défaut" >&2
fi
exit "$fail"
