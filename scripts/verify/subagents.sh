head_ "Subagents"
n=0
for f in dot_claude/agents/*.md; do
  [ -f "$f" ] || continue
  slug=$(basename "$f" .md)
  fm=$(awk 'NR==1 && $0=="---"{f=1;next} f && $0=="---"{exit} f' "$f")
  name=$(printf '%s' "$fm" | sed -n 's/^name: *//p')
  [ "$name" = "$slug" ] || ko "$slug : frontmatter name=« $name » ne correspond pas au fichier"
  printf '%s' "$fm" | grep -q '^description:' || ko "$slug : pas de description"
  printf '%s' "$fm" | grep -q 'Use when' || ko "$slug : « Use when » absent de la description"
  printf '%s' "$fm" | grep -q '^tools:' || ko "$slug : pas de champ tools"
  n=$((n + 1))
done
if [ -d dot_claude/agents ] && [ ! -L dot_claude_pro/symlink_agents ] && [ ! -f dot_claude_pro/symlink_agents ]; then
  ko "dot_claude/agents existe sans symlink_agents dans dot_claude_pro : le profil pro ne les verrait pas"
fi
okif "$n subagents, frontmatter cohérent et partagé avec le profil pro"
