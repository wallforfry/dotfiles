#!/usr/bin/env bash
# Barrière mécanique de ce dépôt, lancée à la main comme depuis
# .github/workflows/verify.yml (ADR-020).
#
# bash et non sh POSIX : ne tourne jamais au bootstrap, seulement à la main, en
# CI, ou depuis le subagent dotfiles-reviewer.
#
# Sort en 1 dès qu'un contrôle échoue, et rapporte des comptes, pas des adjectifs.

set -uo pipefail

cd "$(git rev-parse --show-toplevel)"

checks=(
  scripts/verify/syntax.sh
  scripts/verify/templates.sh
  scripts/verify/skills.sh
  scripts/verify/routing.sh
  scripts/verify/subagents.sh
  scripts/verify/projections.sh
  scripts/verify/workflows.sh
  scripts/verify/adr.sh
  scripts/verify/sensitive.sh
  scripts/verify/encryption.sh
  scripts/verify/live-state.sh
  scripts/verify/telemetry.sh
)

for check in scripts/verify/support.sh "${checks[@]}"; do
  if ! bash -n "$check"; then
    printf '❌  %s ne passe pas bash -n\n' "$check" >&2
    exit 1
  fi
done

source scripts/verify/support.sh
for check in "${checks[@]}"; do
  source "$check"
done

printf '\n'
if [ "$fail" -eq 0 ]; then
  echo "🎉  barrière verte"
else
  echo "💥  barrière rouge" >&2
fi
exit "$fail"
