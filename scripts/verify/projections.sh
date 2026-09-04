head_ "Projections d'instructions"
n=0
for f in harness/*.md; do
  imports=$(grep -nE '^@[A-Za-z]' "$f" || true)
  if [ -n "$imports" ]; then
    ko "$f : import « @ » propre à un seul hôte dans une source agnostique"
    printf '%s\n' "$imports" | head -3 | sed 's/^/      /'
  fi
  n=$((n + 1))
done
for base in AGENTS SOUL USER; do
  f="dot_claude/$base.md.tmpl"
  [ -f "$f" ] || { ko "$f absent : projection Claude manquante"; continue; }
  [ "$(wc -l < "$f" | tr -d ' ')" = 1 ] ||
    ko "$f : une projection ne porte qu'une ligne d'include"
  grep -q "include \"harness/$base.md\"" "$f" ||
    ko "$f : n'inclut pas harness/$base.md"
  grep -q "@$base.md" dot_claude/CLAUDE.md ||
    ko "dot_claude/CLAUDE.md : @$base.md absent, Claude ne chargerait pas $base"
  grep -q "include \"harness/$base.md\"" dot_codex/AGENTS.md.tmpl ||
    ko "dot_codex/AGENTS.md.tmpl : n'inclut pas harness/$base.md"
done
for h in dot_claude dot_codex; do
  [ -e "$h/symlink_skills" ] &&
    ko "$h/symlink_skills : un lien sur le répertoire entier détruirait l'état vivant"
done
for d in dot_config/agent-skills/*/; do
  slug=$(basename "$d")
  for h in dot_claude dot_codex; do
    l="$h/skills/symlink_$slug"
    if [ ! -f "$l" ]; then
      ko "$l absent : $h ne verrait pas la skill $slug"
    elif [ "$(tr -d '\n' < "$l")" != "../../.config/agent-skills/$slug" ]; then
      ko "$l : cible hors de la source unique de skills"
    fi
  done
done
for h in dot_claude dot_codex; do
  for l in "$h"/skills/symlink_*; do
    [ -f "$l" ] || continue
    slug=$(basename "$l" | sed 's/^symlink_//')
    [ -d "dot_config/agent-skills/$slug" ] || ko "$l : lien sans skill correspondante"
  done
done
[ "$(tr -d '\n' < dot_claude_pro/symlink_skills)" = "../.claude/skills" ] ||
  ko "dot_claude_pro/symlink_skills : cible hors de ~/.claude/skills"
if grep -q 'CONTEXT.md' dot_codex/AGENTS.md.tmpl && [ ! -f dot_codex/symlink_CONTEXT.md ]; then
  ko "dot_codex : l'adaptateur renvoie vers CONTEXT.md sans symlink_CONTEXT.md"
fi
okif "$n sources canoniques sans import, projections et liens de skills cohérents"
