#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cases=${1:-"$root/scripts/skill-routing-cases.tsv"}
cache_root=${XDG_CACHE_HOME:-"$HOME/.cache"}
mkdir -p "$cache_root"
work=$(mktemp -d "$cache_root/skill-routing.XXXXXX")
trap 'rm -rf "$work"' EXIT HUP INT TERM

fail() {
  printf 'skill-routing: %s\n' "$1" >&2
  exit 1
}

validate_shape() {
  awk -F '\t' '
    NR == 1 {
      if ($0 != "id\tkind\texpected\texcluded\tprompt") exit 1
      next
    }
    NF != 5 || $1 !~ /^[a-z0-9-]+$/ || $2 !~ /^(positive|negative|ambiguous)$/ || $5 == "" {
      exit 1
    }
  ' "$cases" || fail "corpus TSV invalide"

  duplicates=$(tail -n +2 "$cases" | cut -f1 | sort | uniq -d)
  [ -z "$duplicates" ] || fail "identifiants dupliqués: $duplicates"
}

validate_slug_list() {
  value=$1
  [ "$value" = "-" ] && return
  for slug in $(printf '%s' "$value" | tr ',' ' '); do
    [ -f "$root/dot_config/agent-skills/$slug/SKILL.md" ] || fail "skill inconnue: $slug"
  done
}

validate_cases() {
  tail -n +2 "$cases" |
    while IFS="$(printf '\t')" read -r id kind expected excluded prompt; do
      validate_slug_list "$expected"
      validate_slug_list "$excluded"
      case "$kind" in
        positive)
          [ "$expected" != "-" ] && [ "$excluded" = "-" ] || fail "$id: contrat positif invalide"
          ;;
        negative)
          [ "$expected" = "-" ] && [ "$excluded" != "-" ] || fail "$id: contrat négatif invalide"
          ;;
        ambiguous)
          [ "$excluded" != "-" ] || fail "$id: scénario ambigu sans exclusion"
          ;;
      esac
      printf '%s\t%s\n' "$kind" "$expected" >> "$work/coverage"
      printf '%s\t%s\n' "$kind" "$excluded" >> "$work/coverage"
    done
}

description_length() {
  awk '
    /^description:/ {
      reading = 1
      line = $0
      sub(/^description:[[:space:]]*>?[[:space:]]*/, "", line)
      if (line != "") text = line
      next
    }
    reading && /^[a-z][a-z-]*:/ { exit }
    reading {
      sub(/^[[:space:]]+/, "")
      if ($0 != "") text = text (text == "" ? "" : " ") $0
    }
    END { print length(text) }
  ' "$1"
}

validate_coverage() {
  positive=0
  negative=0
  for skill_file in "$root"/dot_config/agent-skills/*/SKILL.md; do
    slug=${skill_file%/SKILL.md}
    slug=${slug##*/}
    length=$(description_length "$skill_file")
    [ "$length" -lt 400 ] || fail "$slug: description de $length caractères, limite locale 399"
    grep -q "^positive[[:space:]]$slug$" "$work/coverage" || fail "$slug: scénario positif absent"
    grep -q "^negative[[:space:]]$slug$" "$work/coverage" || fail "$slug: scénario négatif absent"
    positive=$((positive + 1))
    negative=$((negative + 1))
  done
  ambiguous=$(awk -F '\t' '$2 == "ambiguous" { count++ } END { print count + 0 }' "$cases")
  [ "$ambiguous" -ge 3 ] || fail "moins de trois scénarios ambigus"
  printf 'skill-routing: %s positives, %s négatives, %s ambiguës, 0 erreur\n' \
    "$positive" "$negative" "$ambiguous"
}

[ -f "$cases" ] || fail "corpus absent: $cases"
validate_shape
validate_cases
validate_coverage
