head_ "Syntaxe des scripts"
n=0
for f in run_*.sh.tmpl; do
  for c in "${configs[@]}"; do
    if ! render "$c" "$f" | sh -n 2>/dev/null; then
      ko "$f ne passe pas sh -n avec $(basename "$c")"
      continue 2
    fi
  done
  n=$((n + 1))
done
for f in scripts/*.sh dot_claude/hooks/executable_* dot_local/bin/executable_*; do
  [ -f "$f" ] || continue
  if head -1 "$f" | grep -q 'bash'; then checker=bash; else checker=sh; fi
  if $checker -n "$f" 2>/dev/null; then
    n=$((n + 1))
  else
    ko "$f ne passe pas $checker -n"
  fi
done
okif "$n scripts, syntaxe valide sur ${#configs[@]} combinaisons de profil"
