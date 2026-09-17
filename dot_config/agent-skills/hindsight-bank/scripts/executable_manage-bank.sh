#!/bin/sh
set -eu

CONFIG="$HOME/.hindsight/dotfiles.json"

usage() {
  echo "usage: manage-bank.sh create <bank-id> | add <directory> <bank-id> | remove <directory>" >&2
  exit 64
}

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "manage-bank: $1 absent" >&2; exit 69; }
}

mapping_directory() {
  directory=$1
  [ -d "$directory" ] || { echo "manage-bank: dossier introuvable" >&2; exit 64; }
  cd "$directory" && pwd -P
}

write_mapping() {
  action=$1
  repository=$2
  bank=${3:-}
  [ -r "$CONFIG" ] || { echo "manage-bank: configuration Hindsight absente" >&2; exit 69; }
  [ "$action" != add ] || [ -n "$bank" ] || { echo "manage-bank: bank absente" >&2; exit 64; }
  candidate=$(mktemp "${CONFIG}.next.XXXXXX")
  trap 'rm -f "$candidate"' EXIT HUP INT TERM
  if [ "$action" = add ]; then
    jq --arg repository "$repository" --arg bank "$bank" '
      if (.registrations | type) != "array" then error("registrations must be an array") else . end
      | .registrations = ([.registrations[] | select(.repository != $repository)] + [{repository: $repository, bank: $bank}])
    ' "$CONFIG" > "$candidate"
  else
    jq --arg repository "$repository" '
      if (.registrations | type) != "array" then error("registrations must be an array") else . end
      | .registrations = [.registrations[] | select(.repository != $repository)]
    ' "$CONFIG" > "$candidate"
  fi
  chmod 600 "$candidate"
  mv "$candidate" "$CONFIG"
  trap - EXIT HUP INT TERM
  chezmoi add --encrypt "$CONFIG"
  chezmoi apply --force
}

require hindsight
require jq
require chezmoi
[ $# -ge 1 ] || usage
case $1 in
  create)
    [ $# -eq 2 ] || usage
    hindsight bank create "$2"
    ;;
  add)
    [ $# -eq 3 ] || usage
    root=$(mapping_directory "$2")
    write_mapping add "$root" "$3"
    ;;
  remove)
    [ $# -eq 2 ] || usage
    root=$(mapping_directory "$2")
    write_mapping remove "$root"
    ;;
  *) usage ;;
esac
