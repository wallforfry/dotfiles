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

fail=0
section=0
ok()   { printf '  ✅  %s\n' "$1"; }
ko()   { printf '  ❌  %s\n' "$1" >&2; fail=1; section=1; }
head_() { section=0; printf '\n== %s\n' "$1"; }
# Le message de succès d'une section ne dépend que de cette section.
okif() { [ "$section" -eq 0 ] && ok "$1"; }

render() { # render <config> <fichier>
  chezmoi execute-template --config "$1" --source "$PWD" < "$2"
}

# --- configurations de rendu, une par combinaison à couvrir -------------------
# Jamais sous /tmp : monté noexec sur DSM, et la skill scripts l'interdit à
# tout script du dépôt, y compris à cette barrière.
mkdir -p "$HOME/.cache"
tmp=$(mktemp -d "$HOME/.cache/verify.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
# La configuration vivante sert de base : execute-template n'a pas de
# --promptDefaults, et promptChoiceOnce ne peut pas être forcé depuis un flag.
live="${CHEZMOI_CONFIG:-$HOME/.config/chezmoi/chezmoi.toml}"
if [ ! -f "$live" ]; then
  echo "❌  $live absent : lancer chezmoi init d'abord" >&2
  exit 1
fi
base=$(cat "$live")
for combo in pro:true perso:true perso:false; do
  p=${combo%:*}; g=${combo#*:}
  printf '%s\n' "$base" |
    sed -e "s/^    profile = .*/    profile = \"$p\"/" \
        -e "s/^    gui = .*/    gui = $g/" > "$tmp/$p-$g.toml"
done
configs=("$tmp"/*.toml)

head_ "Syntaxe des scripts"
n=0
for f in run_*.sh.tmpl; do
  for c in "${configs[@]}"; do
    if ! render "$c" "$f" | sh -n 2>/dev/null; then
      ko "$f ne passe pas sh -n avec $(basename "$c")"; continue 2
    fi
  done
  n=$((n + 1))
done
for f in scripts/*.sh dot_claude/hooks/executable_* dot_local/bin/executable_*; do
  [ -f "$f" ] || continue
  if head -1 "$f" | grep -q 'bash'; then checker=bash; else checker=sh; fi
  if $checker -n "$f" 2>/dev/null; then n=$((n + 1)); else ko "$f ne passe pas $checker -n"; fi
done
okif "$n scripts, syntaxe valide sur ${#configs[@]} combinaisons de profil"

head_ "Rendu des templates"
n=0
for f in $(git ls-files '*.tmpl'); do
  case "$f" in .chezmoi.toml.tmpl) continue ;; esac
  for c in "${configs[@]}"; do
    if ! render "$c" "$f" >/dev/null 2>&1; then
      ko "$f ne se rend pas avec $(basename "$c")"; continue 2
    fi
  done
  n=$((n + 1))
done
okif "$n templates rendus sur ${#configs[@]} combinaisons"

head_ "Skills"
n=0
readme=dot_claude/skills/README.md
for d in dot_claude/skills/*/; do
  s="$d/SKILL.md"; slug=$(basename "$d")
  [ -f "$s" ] || { ko "$slug : SKILL.md manquant"; continue; }
  fm=$(awk 'NR==1 && $0=="---"{f=1;next} f && $0=="---"{exit} f' "$s")
  name=$(printf '%s' "$fm" | sed -n 's/^name: *//p')
  desc=$(printf '%s' "$fm" | sed -n 's/^description: *//p')
  [ "$name" = "$slug" ] || ko "$slug : frontmatter name=« $name » ne correspond pas au répertoire"
  [ -n "$fm" ] || ko "$slug : frontmatter absent"
  # La description peut être un bloc « description: > » sur plusieurs lignes.
  if [ -z "$desc" ] && ! printf '%s' "$fm" | grep -q '^description:'; then
    ko "$slug : pas de description"
  fi
  grep -q 'Use when' "$s" || ko "$slug : la description ne dit pas quand l'utiliser"
  # Le tableau du README est dérivé du frontmatter : la section qui liste un
  # skill doit être sa catégorie, sinon l'index n'est plus une projection.
  # Indentation et guillemets sont laissés à YAML : « category: dev »,
  # « category: "dev" » et quatre espaces sont le même document, et une
  # barrière rouge sur un fichier valide use la confiance qu'elle demande.
  categ=$(printf '%s' "$fm" |
    awk '/^metadata:/{m=1;next}
         m && /^[ \t]+category:/{sub(/^[ \t]+category:[ \t]*/,""); gsub(/["'\'']/,""); print; exit}
         m && /^[^ \t]/{exit}')
  case "$categ" in
    dev|ops) ;;
    *) ko "$slug : metadata.category=« ${categ:-absent} » hors de {dev, ops}" ;;
  esac
  sec=$(awk -v s="$slug" '/^## /{c=tolower($2)} $0 ~ "^\\| `" s "` \\|"{print c; exit}' "$readme")
  if [ -z "$sec" ]; then
    ko "$slug : absent du tableau de $readme"
  elif [ "$sec" != "$categ" ]; then
    ko "$slug : listé sous « $sec » dans le README pour une catégorie « $categ »"
  fi
  # Le README se dit dérivé du frontmatter : sans ce contrôle, seule
  # l'appartenance était vérifiée, et une description modifiée sans
  # régénération laissait la barrière verte sur un index qui mentait.
  attendu=$(awk '/^description:/{d=1; sub(/^description:[ \t]*>?-?[ \t]*/,""); s=$0; next}
                 d && /^[^ \t]/{d=0}
                 d {sub(/^[ \t]+/,""); s=(s=="" ? $0 : s " " $0)}
                 END{sub(/^[ \t]+/,"",s); print s}' <<<"$fm")
  attendu=${attendu%%. *}.
  cellule=$(awk -v s="$slug" -F' \\| ' '$0 ~ "^\\| `" s "` \\|"{print $2; exit}' "$readme")
  cellule=${cellule% |}
  if [ -n "$sec" ]; then
    if [ -z "$cellule" ]; then
      ko "$slug : ligne du README illisible, dérivation non vérifiée"
    elif [ "$cellule" != "$attendu" ]; then
      ko "$slug : la ligne du README ne dérive plus de la description du frontmatter"
      printf '      README : %s\n      SKILL  : %s\n' "$cellule" "$attendu"
    fi
  fi
  # Les sous-répertoires optionnels sont une liste fermée : un nom hors liste
  # est du contenu qu'aucune convention ne décrit, et que rien ne charge.
  for sub in "$d"*/; do
    [ -d "$sub" ] || continue
    case "$(basename "$sub")" in
      references|assets|scripts) ;;
      *) ko "$slug : sous-répertoire « $(basename "$sub") » hors de {references, assets, scripts}" ;;
    esac
  done
  n=$((n + 1))
done
listed=$(grep -cE '^\| `[a-z-]+` \|' "$readme")
[ "$listed" -eq "$n" ] || ko "le tableau du README annonce $listed skills pour $n répertoires"
for slug in $(grep -oE '^\| `[a-z-]+`' "$readme" | tr -d '|` '); do
  [ -d "dot_claude/skills/$slug" ] || ko "$slug : ligne du README sans répertoire"
done
okif "$n skills, frontmatter et tableau cohérents"

head_ "Subagents"
n=0
for f in dot_claude/agents/*.md; do
  [ -f "$f" ] || continue
  slug=$(basename "$f" .md)
  fm=$(awk 'NR==1 && $0=="---"{f=1;next} f && $0=="---"{exit} f' "$f")
  name=$(printf '%s' "$fm" | sed -n 's/^name: *//p')
  [ "$name" = "$slug" ] || ko "$slug : frontmatter name=« $name » ne correspond pas au fichier"
  printf '%s' "$fm" | grep -q '^description:' || ko "$slug : pas de description"
  printf '%s' "$fm" | grep -q 'Use when' || ko "$slug : « Use when » absent de la description"
  printf '%s' "$fm" | grep -q '^tools:' || ko "$slug : pas de champ tools"
  n=$((n + 1))
done
if [ -d dot_claude/agents ] && [ ! -L dot_claude_pro/symlink_agents ] && [ ! -f dot_claude_pro/symlink_agents ]; then
  ko "dot_claude/agents existe sans symlink_agents dans dot_claude_pro : le profil pro ne les verrait pas"
fi
okif "$n subagents, frontmatter cohérent et partagé avec le profil pro"

head_ "Workflows"
# Les logs d'un dépôt public sont lisibles par tous, et un fragment déchiffré
# y sort aussi sûrement que d'un fichier versionné. Trois commandes émettent le
# contenu rendu d'une cible : « apply --verbose », « chezmoi diff », et un
# « chezmoi cat » sans redirection. Aucune n'est nécessaire à la CI, donc
# aucune n'y est admise. Les lignes de commentaire sont exclues : elles
# nomment ces commandes pour expliquer pourquoi elles sont absentes.
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

head_ "Index des ADR"
if diff <(ls docs/adr | grep -oE '^[0-9]{3}') \
        <(grep -oE '^\| \[[0-9]{3}\]' docs/adr/README.md | grep -oE '[0-9]{3}') >/dev/null; then
  ok "$(ls docs/adr | grep -cE '^[0-9]{3}') ADR, index et répertoire coïncident"
else
  ko "l'index des ADR ne coïncide pas avec le répertoire"
fi

head_ "Rien de sensible en clair (ADR-016)"
# La liste des noms interdits est elle-même la donnée à protéger : elle vient
# d'un fragment chiffré déployé, jamais du script. Sans elle, le contrôle des
# noms est annoncé comme non fait plutôt que déclaré vert (ADR-016).
LISTE="${SENSIBLE_LIST:-$HOME/.config/dotfiles/sensible.txt}"
motifs=$(grep -vE '^\s*(#|$)' "$LISTE" 2>/dev/null | paste -sd'|' -)

# 127.0.0.1 et 0.0.0.0 sont documentés à dessein dans ce dépôt : ce sont des
# constantes, pas des adresses de machine.
motifs_ip='(^|[^0-9.])((1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)\.){3}(1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)'
# Tester le contenu et non le code de sortie : celui du groupe est celui de son
# dernier grep, qui vaut 1 quand seul le premier motif a trouvé quelque chose.
# La version précédente annonçait « aucune occurrence » en tenant la preuve.
if [ -z "$motifs" ]; then
  ko "$LISTE absent ou vide : le contrôle des noms sensibles n'a pas été fait"
fi
# Sans liste, ne pas chercher : un motif de repli vide s'apparie à chaque ligne
# blanche et rapportait des occurrences imaginaires - un compte faux sur un
# contrôle déjà rouge use la confiance qu'il demande. Le contrôle des adresses
# ne dépend pas de la liste et reste fait.
hits=$( { [ -n "$motifs" ] && git grep -inE "$motifs" -- . ':!docs/adr/016-*'
          git grep -inE "$motifs_ip" -- . | grep -vE '127\.0\.0\.1|0\.0\.0\.0'; } 2>/dev/null || true)
if [ -n "$hits" ]; then
  ko "arbre : $(printf '%s\n' "$hits" | wc -l | tr -d ' ') occurrence(s)"
  printf '%s\n' "$hits" | head -3 | sed 's/^/      /'
else
  ok "arbre : aucune occurrence"
fi
if [ -n "$motifs" ] && git rev-parse origin/main >/dev/null 2>&1; then
  hits=$(git log origin/main..HEAD --format='%B' | grep -inE "$motifs" || true)
  if [ -n "$hits" ]; then
    ko "messages de commit non poussés : $(printf '%s\n' "$hits" | wc -l | tr -d ' ') occurrence(s)"
  else
    ok "messages de commit non poussés : aucune occurrence"
  fi
fi

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

head_ "État vivant préservé"
# exact_ fait supprimer par chezmoi tout ce que la source ne porte pas.
# ~/.claude tient les sessions, les projets et les plugins ; ~/.gnupg le
# porte-clés et les sockets de l'agent. Un seul attribut détruirait les deux,
# et aucun contrôle ne le disait.
n=0
for d in dot_claude dot_claude_pro private_dot_gnupg; do
  [ -d "$d" ] || continue
  n=$((n + 1))
  if [ -n "$(find "$d" -name 'exact_*' -print -quit)" ]; then
    ko "$d : attribut exact_ présent, chezmoi supprimerait l'état vivant"
  fi
done
# La racine elle-même peut être renommée : exact_dot_claude/ est un répertoire
# distinct, que la boucle ci-dessus ne visite pas.
for d in exact_dot_claude exact_dot_claude_pro exact_private_dot_gnupg; do
  [ -e "$d" ] && ko "$d : racine préfixée exact_, chezmoi supprimerait l'état vivant"
done
okif "$n racines d'état vivant, aucun attribut exact_"

printf '\n'
if [ "$fail" -eq 0 ]; then
  echo "🎉  barrière verte"
else
  echo "💥  barrière rouge" >&2
fi
exit "$fail"
