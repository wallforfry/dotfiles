head_ "Skills"
n=0
readme=dot_config/agent-skills/README.md
for d in dot_config/agent-skills/*/; do
  s="$d/SKILL.md"
  slug=$(basename "$d")
  [ -f "$s" ] || { ko "$slug : SKILL.md manquant"; continue; }
  fm=$(awk 'NR==1 && $0=="---"{f=1;next} f && $0=="---"{exit} f' "$s")
  name=$(printf '%s' "$fm" | sed -n 's/^name: *//p')
  desc=$(printf '%s' "$fm" | sed -n 's/^description: *//p')
  [ "$name" = "$slug" ] || ko "$slug : frontmatter name=« $name » ne correspond pas au répertoire"
  [ -n "$fm" ] || ko "$slug : frontmatter absent"
  if [ -z "$desc" ] && ! printf '%s' "$fm" | grep -q '^description:'; then
    ko "$slug : pas de description"
  fi
  grep -q 'Use when' "$s" || ko "$slug : la description ne dit pas quand l'utiliser"
  categ=$(printf '%s' "$fm" |
    awk '/^metadata:/{m=1;next}
         m && /^[ \t]+category:/{sub(/^[ \t]+category:[ \t]*/,""); gsub(/["'"'"']/ ,""); print; exit}
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
  for sub in "$d"*/; do
    [ -d "$sub" ] || continue
    case "$(basename "$sub")" in
      references|assets|scripts) ;;
      *) ko "$slug : sous-répertoire « $(basename "$sub") » hors de la liste" ;;
    esac
  done
  n=$((n + 1))
done
listed=$(grep -cE '^\| `[a-z-]+` \|' "$readme")
[ "$listed" -eq "$n" ] || ko "le tableau du README annonce $listed skills pour $n répertoires"
for slug in $(grep -oE '^\| `[a-z-]+`' "$readme" | tr -d '|` '); do
  [ -d "dot_config/agent-skills/$slug" ] || ko "$slug : ligne du README sans répertoire"
done
okif "$n skills, frontmatter et tableau cohérents"
