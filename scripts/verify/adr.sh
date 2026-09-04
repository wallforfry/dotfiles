head_ "Index des ADR"
if diff <(ls docs/adr | grep -oE '^[0-9]{3}') \
        <(grep -oE '^\| \[[0-9]{3}\]' docs/adr/README.md | grep -oE '[0-9]{3}') >/dev/null; then
  ok "$(ls docs/adr | grep -cE '^[0-9]{3}') ADR, index et répertoire coïncident"
else
  ko "l'index des ADR ne coïncide pas avec le répertoire"
fi
