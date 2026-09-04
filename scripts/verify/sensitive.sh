head_ "Rien de sensible en clair (ADR-016)"
LISTE="${SENSIBLE_LIST:-$HOME/.config/dotfiles/sensible.txt}"
motifs_ip='(^|[^0-9.])((1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)\.){3}(1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)'

if [ ! -r "$LISTE" ]; then
  ko "$LISTE absent ou illisible : le contrôle des noms sensibles n'a pas été fait"
  motifs=
elif ! motifs=$(awk '!/^\s*(#|$)/ { printf "%s%s", separator, $0; separator="|" } END { print "" }' "$LISTE"); then
  ko "$LISTE illisible : le contrôle des noms sensibles n'a pas été fait"
  motifs=
elif [ -z "$motifs" ]; then
  ko "$LISTE vide : le contrôle des noms sensibles n'a pas été fait"
fi

tracked_hits=0
if [ -n "$motifs" ]; then
  status=0
  git grep -IlE "$motifs" -- . ':!docs/adr/016-*' > "$tmp/sensitive-tracked" 2>/dev/null || status=$?
  if [ "$status" -gt 1 ]; then
    ko "arbre : lecture interrompue"
  elif [ "$status" -eq 0 ]; then
    tracked_hits=$((tracked_hits + $(wc -l < "$tmp/sensitive-tracked" | tr -d ' ')))
  fi

  status=0
  git grep -InE "$motifs_ip" -- . > "$tmp/sensitive-ip-raw" 2>/dev/null || status=$?
  if [ "$status" -gt 1 ]; then
    ko "arbre : lecture des adresses interrompue"
  elif [ "$status" -eq 0 ]; then
    status=0
    grep -vE '127\.0\.0\.1|0\.0\.0\.0' "$tmp/sensitive-ip-raw" > "$tmp/sensitive-ip" || status=$?
    if [ "$status" -gt 1 ]; then
      ko "arbre : filtrage des adresses interrompu"
    elif [ "$status" -eq 0 ]; then
      tracked_hits=$((tracked_hits + $(wc -l < "$tmp/sensitive-ip" | tr -d ' ')))
    fi
  fi
fi
if [ "$tracked_hits" -gt 0 ]; then
  ko "arbre : $tracked_hits occurrence(s)"
else
  okif "arbre : aucune occurrence"
fi

if [ -n "$motifs" ]; then
  untracked=0
  unreadable=0
  if ! git ls-files -z --others --exclude-standard > "$tmp/sensitive-untracked"; then
    ko "contenu des fichiers non suivis : inventaire interrompu"
  else
    while IFS= read -r -d '' file; do
      status=0
      grep -IqiE "$motifs" "$file" 2>/dev/null || status=$?
      if [ "$status" -gt 1 ]; then
        unreadable=$((unreadable + 1))
        continue
      elif [ "$status" -eq 0 ]; then
        untracked=$((untracked + 1))
        continue
      fi
      status=0
      grep -IinE "$motifs_ip" "$file" > "$tmp/sensitive-untracked-ip" 2>/dev/null || status=$?
      if [ "$status" -gt 1 ]; then
        unreadable=$((unreadable + 1))
      elif [ "$status" -eq 0 ] && grep -qvE '127\.0\.0\.1|0\.0\.0\.0' "$tmp/sensitive-untracked-ip"; then
        untracked=$((untracked + 1))
      fi
    done < "$tmp/sensitive-untracked"
  fi
  if [ "$unreadable" -gt 0 ]; then
    ko "contenu des fichiers non suivis : $unreadable lecture(s) impossible(s)"
  elif [ "$untracked" -gt 0 ]; then
    ko "contenu des fichiers non suivis : $untracked occurrence(s)"
  else
    okif "contenu des fichiers non suivis : aucune occurrence"
  fi

  if ! git ls-files --cached --others --exclude-standard > "$tmp/sensitive-paths"; then
    ko "noms de fichiers : inventaire interrompu"
  else
    status=0
    grep -inE "$motifs" "$tmp/sensitive-paths" > "$tmp/sensitive-path-hits" || status=$?
    if [ "$status" -gt 1 ]; then
      ko "noms de fichiers : contrôle interrompu"
    elif [ "$status" -eq 0 ]; then
      ko "noms de fichiers suivis ou non ignorés : $(wc -l < "$tmp/sensitive-path-hits" | tr -d ' ') occurrence(s)"
    else
      okif "noms de fichiers suivis ou non ignorés : aucune occurrence"
    fi
  fi

  branch=$(git symbolic-ref --quiet --short HEAD 2>/dev/null || true)
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
  if ! git log origin/main..HEAD --format='%B' > "$tmp/sensitive-commits"; then
    ko "messages de commit non poussés : lecture interrompue"
  else
    status=0
    grep -inE "$motifs" "$tmp/sensitive-commits" > "$tmp/sensitive-commit-hits" || status=$?
    if [ "$status" -gt 1 ]; then
      ko "messages de commit non poussés : contrôle interrompu"
    elif [ "$status" -eq 0 ]; then
      ko "messages de commit non poussés : $(wc -l < "$tmp/sensitive-commit-hits" | tr -d ' ') occurrence(s)"
    else
      okif "messages de commit non poussés : aucune occurrence"
    fi
  fi
fi
