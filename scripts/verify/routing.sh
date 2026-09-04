head_ "Routage des skills"
if output=$(sh scripts/validate-skill-routing.sh 2>&1); then
  ok "${output#skill-routing: }"
else
  ko "contrats de routage invalides"
  printf '%s\n' "$output" | head -3 | sed 's/^/      /'
fi
