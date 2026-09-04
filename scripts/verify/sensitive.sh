head_ "Rien de sensible en clair (ADR-016)"
LISTE="${SENSIBLE_LIST:-$HOME/.config/dotfiles/sensible.txt}"
motifs=$(grep -vE '^\s*(#|$)' "$LISTE" 2>/dev/null | paste -sd'|' -)
motifs_ip='(^|[^0-9.])((1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)\.){3}(1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)'

if [ -z "$motifs" ]; then
  ko "$LISTE absent ou vide : le contrôle des noms sensibles n'a pas été fait"
fi

hits=$( { [ -n "$motifs" ] && git grep -inE "$motifs" -- . ':!docs/adr/016-*'
          git grep -inE "$motifs_ip" -- . | grep -vE '127\.0\.0\.1|0\.0\.0\.0'; } 2>/dev/null || true)
if [ -n "$hits" ]; then
  ko "arbre : $(printf '%s\n' "$hits" | wc -l | tr -d ' ') occurrence(s)"
  printf '%s\n' "$hits" | head -3 | sed 's/^/      /'
else
  ok "arbre : aucune occurrence"
fi

if [ -n "$motifs" ]; then
  untracked=0
  while IFS= read -r -d '' file; do
    if grep -IqiE "$motifs" "$file" 2>/dev/null; then
      untracked=$((untracked + 1))
      continue
    fi
    ip_hits=$(grep -IinE "$motifs_ip" "$file" 2>/dev/null | grep -vE '127\.0\.0\.1|0\.0\.0\.0' || true)
    [ -z "$ip_hits" ] || untracked=$((untracked + 1))
  done < <(git ls-files -z --others --exclude-standard)
  if [ "$untracked" -gt 0 ]; then
    ko "contenu des fichiers non suivis : $untracked occurrence(s)"
  else
    ok "contenu des fichiers non suivis : aucune occurrence"
  fi

  hits=$(git ls-files --cached --others --exclude-standard | grep -inE "$motifs" || true)
  if [ -n "$hits" ]; then
    ko "noms de fichiers suivis ou non ignorés : $(printf '%s\n' "$hits" | wc -l | tr -d ' ') occurrence(s)"
  else
    ok "noms de fichiers suivis ou non ignorés : aucune occurrence"
  fi

  branch=$(git symbolic-ref --quiet --short HEAD || true)
  branch=${branch:-${GITHUB_HEAD_REF:-${GITHUB_REF_NAME:-}}}
  if [ -z "$branch" ]; then
    ok "nom de branche : aucun sur HEAD détachée"
  elif printf '%s\n' "$branch" | grep -qiE "$motifs"; then
    ko "nom de branche : occurrence sensible"
  else
    ok "nom de branche : aucune occurrence"
  fi
fi

if [ -n "$motifs" ] && git rev-parse origin/main >/dev/null 2>&1; then
  hits=$(git log origin/main..HEAD --format='%B' | grep -inE "$motifs" || true)
  if [ -n "$hits" ]; then
    ko "messages de commit non poussés : $(printf '%s\n' "$hits" | wc -l | tr -d ' ') occurrence(s)"
  else
    ok "messages de commit non poussés : aucune occurrence"
  fi
fi
