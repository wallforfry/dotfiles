fail=0
section=0
ok() { printf '  ✅  %s\n' "$1"; }
ko() { printf '  ❌  %s\n' "$1" >&2; fail=1; section=1; }
head_() { section=0; printf '\n== %s\n' "$1"; }
okif() { [ "$section" -eq 0 ] && ok "$1"; }

render() {
  chezmoi execute-template --config "$1" --source "$PWD" < "$2"
}

mkdir -p "$HOME/.cache"
tmp=$(mktemp -d "$HOME/.cache/verify.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
live="${CHEZMOI_CONFIG:-$HOME/.config/chezmoi/chezmoi.toml}"
if [ ! -f "$live" ]; then
  echo "❌  $live absent : lancer chezmoi init d'abord" >&2
  exit 1
fi
base=$(cat "$live")
for combo in pro:true perso:true perso:false; do
  p=${combo%:*}
  g=${combo#*:}
  printf '%s\n' "$base" |
    sed -e "s/^    profile = .*/    profile = \"$p\"/" \
        -e "s/^    gui = .*/    gui = $g/" > "$tmp/$p-$g.toml"
done
configs=("$tmp"/*.toml)
