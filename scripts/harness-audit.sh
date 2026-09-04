#!/usr/bin/env bash
# Mesures du harness agentique, lancées à la main ou depuis la skill harness-audit.
#
# bash et non sh POSIX : ne tourne jamais au bootstrap, et python3 est requis
# pour lire les transcripts JSONL.
#
# Cinq mesures : le coût du contexte toujours chargé, le retard de la source de
# déploiement, l'activation réelle des skills et des subagents, l'adhérence aux
# deux règles observables, et le pouvoir de détection de scripts/verify.sh par
# injection de défauts dans un clone.
#
# N'imprime jamais un chemin de projet : ils portent des noms de clients (ADR-016).
# Sort en 1 si la barrière laisse passer un défaut, ou si une mesure n'a pas pu
# être faite : une mesure absente n'est pas une mesure verte.

set -uo pipefail

# La racine vient de l'emplacement du script, jamais du répertoire courant :
# la skill est lancée depuis n'importe quel dépôt, et un rev-parse sur le cwd
# mesurerait cet autre dépôt.
src=${BASH_SOURCE[0]:-}
if [ -z "$src" ]; then
  echo "❌  emplacement du script inconnu : lancer bash <chemin>/harness-audit.sh" >&2
  exit 1
fi
# pwd -P résout les répertoires, jamais le fichier final : sans cette boucle, un
# lanceur en lien symbolique donne le parent du lien, donc un autre dépôt.
while [ -L "$src" ]; do
  cible=$(readlink "$src")
  case $cible in
    /*) src=$cible ;;
    *) src=$(dirname -- "$src")/$cible ;;
  esac
done
root=$(cd -- "$(dirname -- "$src")/.." && pwd -P) || exit 1
if ! git -C "$root" rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "❌  $root hors d'un dépôt git : aucune mesure possible" >&2
  exit 1
fi
cd "$root"

fail=0
ok() { printf '  ✅  %s\n' "$1"; }
ko() { printf '  ❌  %s\n' "$1" >&2; fail=1; }
head_() { printf '\n== %s\n' "$1"; }

PROJECTS="${CLAUDE_PROJECTS:-$HOME/.claude/projects}"
CODEX_SESSIONS="${CODEX_SESSIONS:-$HOME/.codex/sessions}"
# Date d'introduction des deux règles observables (ce5cc72, 45f0998).
SINCE="${HARNESS_RULES_SINCE:-2026-08-17}"

if ! command -v python3 >/dev/null; then
  echo "❌  python3 absent : aucune mesure possible" >&2
  exit 1
fi

now_ms() { python3 -c 'import time; print(time.monotonic_ns() // 1000000)'; }
duration() { printf '  Durée : %s ms\n' "$(($(now_ms) - $1))"; }
run_control() {
  local repository=$1 control=$2
  if [ "$control" = scripts/verify.sh ]; then
    (cd "$repository" && SENSIBLE_LIST="$mutation_list" bash scripts/verify.sh >/dev/null 2>&1)
  else
    (cd "$repository" && SENSIBLE_LIST="$mutation_list" bash -c \
      'source scripts/verify/support.sh; source "$1"; exit "$fail"' _ "$control" >/dev/null 2>&1)
  fi
}
audit_started=$(now_ms)

# --- 1. contexte toujours chargé ---------------------------------------------
section_started=$(now_ms)
head_ "Contexte toujours chargé"
# Les descriptions de frontmatter comptent : elles sont le routeur, donc
# toujours en contexte. Sans elles, déplacer une section vers une skill
# paraîtrait gratuit alors qu'il déplace une part du coût dans le routeur.
total=0
mesure() { # mesure <étiquette> <octets>
  total=$((total + $2))
  printf '  %-34s %6s o  ~%5s jetons\n' "$1" "$2" "$(($2 / 4))"
}
for f in harness/AGENTS.md harness/SOUL.md harness/USER.md AGENTS.md dot_claude/CLAUDE.md; do
  if [ ! -f "$f" ]; then
    ko "$f absent : coût du contexte non mesuré"
    continue
  fi
  mesure "$f" "$(wc -c <"$f" | tr -d ' ')"
done
# Une description est un bloc plié « description: > » : tout jusqu'à la
# prochaine clé de premier niveau.
descs=$(awk 'FNR==1{d=0}
             /^description:/{d=1; n+=length($0)+1; next}
             d && /^[^ \t]/{d=0}
             d {n+=length($0)+1}
             END{print n+0}' dot_config/agent-skills/*/SKILL.md dot_claude/agents/*.md)
mesure "descriptions locales des skills" "$descs"
ok "$total octets, ~$((total / 4)) jetons sur chaque tâche de ce dépôt"
echo "  Le CONTEXT.md de chaque hôte est chiffré : son coût n'est pas mesurable ici"
echo "  Les descriptions de plugins gérées par l'hôte ne sont pas mesurables depuis le dépôt"
duration "$section_started"

# --- 2. retard de la source de déploiement ------------------------------------
section_started=$(now_ms)
head_ "Source de déploiement"
# La source chezmoi est un clone distinct du dépôt de travail (ADR-001) : un
# commit fusionné ne prend effet qu'après un chezmoi update. Aucun signal ne le
# disait, chezmoi status comparant la source à la destination et jamais à la
# remote. Pas de fetch ici : la mesure reste locale, donc toujours possible, et
# se lit par rapport à la dernière récupération.
if ! command -v chezmoi >/dev/null; then
  ko "chezmoi absent : retard de la source non mesuré"
else
  src=$(chezmoi source-path)
  if [ "$src" = "$PWD" ]; then
    ok "clone unique : la source de déploiement est ce dépôt"
  elif ! git -C "$src" rev-parse --verify -q origin/main >/dev/null; then
    ko "$(basename "$src") : origin/main inconnue, retard non mesuré"
  else
    behind=$(git -C "$src" rev-list --count HEAD..origin/main)
    ok "source en retard de $behind commits sur origin/main à la dernière récupération"
  fi
fi
duration "$section_started"

# --- 3. activation réelle et adhérence observable -----------------------------
# Un seul parcours des JSONL : les deux mesures lisent les mêmes lignes, et deux
# parcours de plusieurs centaines de sessions coûtent le double pour rien.
section_started=$(now_ms)
head_ "Activation et adhérence (règles introduites le $SINCE)"
if ! PYTHONDONTWRITEBYTECODE=1 python3 -B scripts/harness_telemetry.py \
  --claude-root "$PROJECTS" \
  --codex-root "$CODEX_SESSIONS" \
  --since "$SINCE" \
  --cache "$HOME/.cache/harness-audit/telemetry-v1.json"
then
  ko "lecture des transcripts interrompue : activation et adhérence non mesurées"
else
  ok "mesuré sur les transcripts disponibles"
  echo "  l'adhérence est corrélationnelle : le modèle a changé sur la même période"
fi
duration "$section_started"

# --- 4. pouvoir de détection de la barrière -----------------------------------
section_started=$(now_ms)
head_ "Pouvoir de détection de scripts/verify.sh"
mkdir -p "$HOME/.cache"
tmp=$(mktemp -d "$HOME/.cache/harness-audit.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
mutation_list="$tmp/sensible.txt"
marker="harness-audit-$RANDOM-$RANDOM"
printf '%s\n' "$marker" > "$mutation_list"
if ! awk -F '\t' 'NF != 5 || $1 == "" || $2 == "" || $3 !~ /^(reject|accept|observe)$/ || $4 == "" { exit 1 }' \
  scripts/harness-mutation-cases.tsv
then
  ko "matrice de promesses invalide"
elif ! git clone -q --no-hardlinks "$PWD" "$tmp/rep" 2>/dev/null; then
  ko "clone impossible : détection non mesurée"
else
  git diff --binary HEAD > "$tmp/worktree.patch"
  if [ -s "$tmp/worktree.patch" ]; then
    git -C "$tmp/rep" apply "$tmp/worktree.patch"
  fi
  while IFS= read -r -d '' path; do
    mkdir -p -- "$tmp/rep/$(dirname "$path")"
    cp -p -- "$path" "$tmp/rep/$path"
  done < <(git ls-files -z --others --exclude-standard)
  git -C "$tmp/rep" config user.name harness-audit
  git -C "$tmp/rep" config user.email harness-audit@invalid
  git -C "$tmp/rep" config gc.auto 0
  git -C "$tmp/rep" config maintenance.auto false
  git -C "$tmp/rep" add -A
  git -C "$tmp/rep" commit -qm 'test: capture audit baseline' --allow-empty
  baseline=$(git -C "$tmp/rep" rev-parse HEAD)
  if ! (cd "$tmp/rep" && SENSIBLE_LIST="$mutation_list" bash scripts/verify.sh >/dev/null 2>&1); then
    ko "le clone est rouge avant toute mutation : détection non mesurable"
  else
    passed=0
    checked=0
    rejected=0
    accepted=0
    observed=0
    while IFS="$(printf '\t')" read -r _ control expectation _ _; do
      if [ "$expectation" != observe ]; then
        [ -f "$tmp/rep/$control" ] || ko "contrôle absent de la matrice : $control"
        if [ "$control" != scripts/verify.sh ] && ! grep -Fq "  $control" scripts/verify.sh; then
          ko "contrôle absent de l'orchestrateur : $control"
        fi
      fi
    done < scripts/harness-mutation-cases.tsv
    while IFS="$(printf '\t')" read -r promise control expectation label code; do
      [ -z "$promise" ] && continue
      if [ "$expectation" = observe ]; then
        observed=$((observed + 1))
        printf '  %-22s %-24s observation\n' "$promise" "$control"
        continue
      fi
      checked=$((checked + 1))
      printf '%s\n' "$marker" > "$mutation_list"
      git -C "$tmp/rep" switch -q --detach "$baseline"
      git -C "$tmp/rep" restore --source="$baseline" --staged --worktree .
      git -C "$tmp/rep" clean -qfd
      if ! (cd "$tmp/rep" && HARNESS_AUDIT_MARKER="$marker" \
        HARNESS_AUDIT_LIST="$mutation_list" python3 -c "$code")
      then
        ko "mutation inapplicable, à réécrire : $label"
        continue
      fi
      if run_control "$tmp/rep" "$control"; then
        result=accept
      else
        result=reject
      fi
      if [ "$result" = "$expectation" ]; then
        passed=$((passed + 1))
        [ "$result" = reject ] && rejected=$((rejected + 1)) || accepted=$((accepted + 1))
      else
        ko "$control ne respecte pas $promise : $label devait être $expectation"
      fi
    done < scripts/harness-mutation-cases.tsv
    if [ "$passed" -eq "$checked" ]; then
      ok "$rejected défauts rejetés, $accepted anti-mutants acceptés, $observed promesses observées"
    else
      ko "$passed/$checked comportements de barrière conformes"
    fi
  fi
fi
duration "$section_started"
printf '  Durée totale : %s ms\n' "$(($(now_ms) - audit_started))"

printf '\n'
if [ "$fail" -eq 0 ]; then
  echo "🎉  mesures complètes"
else
  echo "💥  une mesure manque ou la barrière laisse passer un défaut" >&2
fi
exit "$fail"
