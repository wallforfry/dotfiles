#!/usr/bin/env bash
# Mesures du harness agentique, lancées à la main ou depuis la skill harness-audit.
#
# bash et non sh POSIX : ne tourne jamais au bootstrap, et python3 est requis
# pour lire les transcripts JSONL.
#
# Quatre mesures : le coût du contexte toujours chargé, l'activation réelle des
# skills et des subagents, l'adhérence aux deux règles observables, et le pouvoir
# de détection de scripts/verify.sh par injection de défauts dans un clone.
#
# N'imprime jamais un chemin de projet : ils portent des noms de clients (ADR-016).
# Sort en 1 si la barrière laisse passer un défaut, ou si une mesure n'a pas pu
# être faite : une mesure absente n'est pas une mesure verte.

set -uo pipefail

cd "$(git rev-parse --show-toplevel)"

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
total=0
for f in harness/AGENTS.md harness/SOUL.md harness/USER.md AGENTS.md; do
  b=$(wc -c <"$f" | tr -d ' ')
  total=$((total + b))
  printf '  %-20s %6s o  ~%5s jetons\n' "$(basename "$f")" "$b" "$((b / 4))"
done
ok "$total octets, ~$((total / 4)) jetons sur chaque tâche de ce dépôt"

# --- 2. activation réelle -----------------------------------------------------
head_ "Activation des skills et des subagents"
if [ ! -d "$PROJECTS" ]; then
  ko "$PROJECTS absent : activation non mesurée"
else
  python3 - "$PROJECTS" <<'PY'
import json, os, sys, glob, collections
root = sys.argv[1]
skills = collections.Counter(); agents = collections.Counter()
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
            if d.get('timestamp'):
                days.append(d['timestamp'][:10])
            c = (d.get('message') or {}).get('content')
            if not isinstance(c, list):
                continue
            for b in c:
                if not isinstance(b, dict) or b.get('type') != 'tool_use':
                    continue
                i = b.get('input') or {}
                if b.get('name') == 'Skill':
                    skills[str(i.get('skill'))] += 1
                elif b.get('name') in ('Task', 'Agent'):
                    agents[str(i.get('subagent_type'))] += 1
print(f'  {sessions} sessions, {min(days)} au {max(days)}' if days else f'  {sessions} sessions')
local = sorted(os.listdir('dot_claude/skills'))
local = [s for s in local if os.path.exists(f'dot_claude/skills/{s}/SKILL.md')]
print('  skills de ce dépôt :')
for s in local:
    print(f'    {skills.get(s, 0):5d}  {s}')
print('  autres skills activées, top 5 :')
for n, c in [(n, c) for n, c in skills.most_common() if n not in local][:5]:
    print(f'    {c:5d}  {n}')
print('  subagents :')
for n, c in agents.most_common(6):
    print(f'    {c:5d}  {n}')
PY
  ok "activation mesurée sur les transcripts disponibles"
fi

# --- 3. adhérence aux règles observables --------------------------------------
head_ "Adhérence observable (règles introduites le $SINCE)"
if [ ! -d "$PROJECTS" ]; then
  ko "$PROJECTS absent : adhérence non mesurée"
else
  python3 - "$PROJECTS" "$SINCE" <<'PY'
import json, os, sys, glob, collections
root, since = sys.argv[1], sys.argv[2]
EXT = ('.ts', '.tsx', '.js', '.py', '.lua', '.rs', '.go', '.sh')
blocks = collections.Counter(); dash = collections.Counter()
lines = collections.Counter(); comments = collections.Counter()
for f in glob.glob(os.path.join(root, '**', '*.jsonl'), recursive=True):
    with open(f, encoding='utf-8', errors='replace') as fh:
        for line in fh:
            if not line.strip():
                continue
            try:
                d = json.loads(line)
            except ValueError:
                continue
            era = 'avant' if (d.get('timestamp') or '')[:10] < since else 'depuis'
            m = d.get('message') or {}
            c = m.get('content')
            if not isinstance(c, list):
                continue
            for b in c:
                if not isinstance(b, dict):
                    continue
                if b.get('type') == 'text' and m.get('role') == 'assistant':
                    blocks[era] += 1
                    dash[era] += b.get('text', '').count('—')
                elif b.get('type') == 'tool_use' and b.get('name') == 'Write':
                    i = b.get('input') or {}
                    if str(i.get('file_path', '')).endswith(EXT):
                        for L in str(i.get('content', '')).split('\n'):
                            s = L.strip()
                            if not s:
                                continue
                            lines[era] += 1
                            if s.startswith(('//', '#', '--', '/*', '*')):
                                comments[era] += 1
def rate(a, b):
    return a / b if b else 0.0
print('  tiret cadratin par bloc de texte assistant :')
for era in ('avant', 'depuis'):
    print(f'    {era:6s}  {blocks[era]:6d} blocs  {dash[era]:6d} occurrences  {rate(dash[era], blocks[era]):.2f}')
print('  lignes de commentaire dans le code écrit :')
for era in ('avant', 'depuis'):
    print(f'    {era:6s}  {lines[era]:6d} lignes  {comments[era]:6d} commentaires  {rate(comments[era], lines[era]):.3f}')
PY
  ok "adhérence mesurée ; corrélationnelle, le modèle a changé sur la même période"
fi

# --- 4. pouvoir de détection de la barrière -----------------------------------
head_ "Pouvoir de détection de scripts/verify.sh"
tmp=$(mktemp -d "$HOME/.cache/harness-audit.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
if ! git clone -q "$PWD" "$tmp/rep" 2>/dev/null; then
  ko "clone impossible : détection non mesurée"
else
  (cd "$tmp/rep" && git checkout -q "$(git -C "$PWD" rev-parse --abbrev-ref HEAD)" 2>/dev/null) || true
  if ! (cd "$tmp/rep" && bash scripts/verify.sh >/dev/null 2>&1); then
    ko "le clone est rouge avant toute mutation : détection non mesurable"
  else
    detected=0
    count=0
    while IFS='|' read -r label code; do
      [ -z "$label" ] && continue
      count=$((count + 1))
      (cd "$tmp/rep" && git checkout -q -- . && git clean -qfd && python3 -c "$code")
      if (cd "$tmp/rep" && bash scripts/verify.sh >/dev/null 2>&1); then
        ko "mutation non détectée : $label"
      else
        detected=$((detected + 1))
      fi
    done <<'MUT'
name d'une skill différent du répertoire|p='dot_claude/skills/adr/SKILL.md';s=open(p).read();open(p,'w').write(s.replace('name: adr','name: adrx',1))
skill retirée de l'index|import re;p='dot_claude/skills/README.md';s=open(p).read();open(p,'w').write('\n'.join(l for l in s.split('\n') if not re.match(r'^\| `adr`',l)))
metadata.category hors de {dev, ops}|p='dot_claude/skills/adr/SKILL.md';s=open(p).read();open(p,'w').write(s.replace('category: ops','category: misc').replace('category: dev','category: misc'))
erreur de syntaxe shell|p='scripts/verify.sh';s=open(p).read();open(p,'w').write(s+'\nif true; then\n')
template invalide|open('dot_gitconfig.tmpl','a').write('{{ .absent.champ }}\n')
ADR présente hors index|import glob,shutil;shutil.copy(glob.glob('docs/adr/001-*.md')[0],'docs/adr/099-fantome.md')
fragment .age en clair|open('age-key.txt.age','w').write('texte en clair\n')
chezmoi diff introduit en CI|p='.github/workflows/verify.yml';s=open(p).read();open(p,'w').write(s.replace('chezmoi status','chezmoi diff'))
skill sans ligne d'index|import os;os.makedirs('dot_claude/skills/fantome',exist_ok=True);open('dot_claude/skills/fantome/SKILL.md','w').write('---\nname: fantome\ndescription: Rien. Use when jamais.\nmetadata:\n  category: dev\n---\n# Fantome\n')
sous-répertoire de skill hors liste|import os;os.makedirs('dot_claude/skills/adr/evals',exist_ok=True);open('dot_claude/skills/adr/evals/x.json','w').write('{}')
attribut exact_ sur l'état vivant|open('dot_claude/exact_zzz','w').write('')
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
