head_ "Workflows"
n=0
for f in $(git ls-files '.github/*'); do
  vu=$(grep -nE '^[^#]*(--verbose|chezmoi[[:space:]]+diff)' "$f" || true)
  cat_nu=$(grep -nE '^[^#]*chezmoi[^#]*[[:space:]]cat[[:space:]]' "$f" |
    grep -vE '>[[:space:]]*("?\$|/dev/null)' || true)
  if [ -n "$vu" ] || [ -n "$cat_nu" ]; then
    ko "$f : commande qui écrirait le contenu rendu d'une cible dans un log public"
    printf '%s\n' "$vu" "$cat_nu" | grep -v '^$' | head -3 | sed 's/^/      /'
  fi
  n=$((n + 1))
done
okif "$n fichiers de CI, aucune sortie de contenu rendu"
