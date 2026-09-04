head_ "Fragments chiffrés"
n=0
for f in $(git ls-files '*.age'); do
  if head -c 40 "$f" | grep -qE 'BEGIN AGE ENCRYPTED|age-encryption\.org'; then
    n=$((n + 1))
  else
    ko "$f n'est pas chiffré"
  fi
done
okif "$n fragments, tous chiffrés"
