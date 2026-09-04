head_ "État vivant préservé"
n=0
for d in dot_claude dot_claude_pro dot_codex private_dot_gnupg; do
  [ -d "$d" ] || continue
  n=$((n + 1))
  if [ -n "$(find "$d" -name 'exact_*' -print -quit)" ]; then
    ko "$d : attribut exact_ présent, chezmoi supprimerait l'état vivant"
  fi
done
for d in exact_dot_claude exact_dot_claude_pro exact_dot_codex exact_private_dot_gnupg; do
  [ -e "$d" ] && ko "$d : racine préfixée exact_, chezmoi supprimerait l'état vivant"
done
okif "$n racines d'état vivant, aucun attribut exact_"
