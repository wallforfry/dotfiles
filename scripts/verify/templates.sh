head_ "Rendu des templates"
n=0
for f in $(git ls-files '*.tmpl'); do
  case "$f" in .chezmoi.toml.tmpl) continue ;; esac
  for c in "${configs[@]}"; do
    if ! render "$c" "$f" >/dev/null 2>&1; then
      ko "$f ne se rend pas avec $(basename "$c")"
      continue 2
    fi
  done
  n=$((n + 1))
done
okif "$n templates rendus sur ${#configs[@]} combinaisons"
