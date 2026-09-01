#!/usr/bin/env bash
# Bascule de ce dépôt en public : chiffrement des fragments sensibles,
# réécriture d'historique, publication, rotation de la passphrase.
#
# Do-nothing script : chaque étape non automatisée imprime la commande exacte
# puis attend. Rien n'est exécuté à l'insu de l'opérateur.
#
# bash et non sh : ce script ne tourne jamais pendant le bootstrap, seulement à
# la main sur un poste déjà outillé.
#
# Voir docs/adr/016-depot-public-sensible-chiffre.md.
set -euo pipefail

# Une étape a déjà échoué sans un mot : errexit sort en silence, et un « awk exit »
# qui ferme un tube suffit à tuer le script par SIGPIPE plus pipefail.
trap 'echo "ERREUR: interruption ligne $LINENO (code $?)." >&2' ERR

SOURCE_DIR="$(chezmoi source-path 2>/dev/null || echo "$HOME/.local/share/chezmoi")"
WORK="$HOME/.cache/go-public"
PATTERNS="$WORK/patterns.txt"
EFFECTIVE="$WORK/patterns.effective.txt"
REPLACEMENTS="$WORK/replacements.txt"
REMOTE="$(git -C "$SOURCE_DIR" remote get-url origin 2>/dev/null || echo '<remote>')"

KEY="$HOME/.config/chezmoi/key.txt"
FRAGMENTS=(
  "$HOME/.secrets"
  "$HOME/.ssh/config.d/nas.conf"
  "$HOME/.config/zsh/pro.zsh"
  "$HOME/.config/zsh/pro.zprofile"
  "$HOME/.config/git/pro.gitconfig"
  "$HOME/.claude/CONTEXT.md"
)

wait_for_enter() { read -rp "→ Entrée pour continuer... "; }

# Les motifs sont saisis avec des commentaires et des lignes vides ; grep et
# filter-repo les prendraient au premier degré.
effective_patterns() {
  grep -vE '^[[:space:]]*(#|$)' "$PATTERNS" > "$EFFECTIVE"
  chmod 600 "$EFFECTIVE"
  [ -s "$EFFECTIVE" ]
}

step_preconditions() {
  local missing=0 tool
  for tool in chezmoi git age age-keygen gh git-filter-repo; do
    command -v "$tool" >/dev/null 2>&1 || { echo "ERREUR: $tool absent." >&2; missing=1; }
  done
  [ "$missing" = 0 ] || exit 1
  mkdir -p "$WORK"
  chmod 700 "$WORK"
  if [ ! -f "$SOURCE_DIR/docs/adr/016-depot-public-sensible-chiffre.md" ]; then
    echo "ERREUR: $SOURCE_DIR ne porte pas encore la préparation au public." >&2
    echo "Fusionne la branche de préparation dans main, pousse, puis relance." >&2
    exit 1
  fi
  local f
  for f in "${FRAGMENTS[@]}"; do
    [ -f "$f" ] || { echo "ERREUR: $f absent - le rechiffrement part du clair." >&2; exit 1; }
  done
  echo "Préconditions OK. Source chezmoi : $SOURCE_DIR"
  echo "Répertoire de travail hors dépôt : $WORK"
}

step_write_patterns_file() {
  if [ ! -f "$PATTERNS" ]; then
    cat > "$PATTERNS" <<'TEMPLATE'
# Un motif littéral par ligne. Lignes vides et « # » ignorées.
# À couvrir : nom d'hôte du NAS, adresses IP locale et VPN, nom d'utilisateur,
# noms d'employeur, de clients, de projets internes, de configurations
# d'environnement. Casse comprise : filter-repo remplace littéralement.
TEMPLATE
    chmod 600 "$PATTERNS"
  fi
  echo "Complète $PATTERNS, un motif littéral par ligne."
  echo "Ce fichier ne doit jamais entrer dans le dépôt : il vit sous $WORK."
  echo "Ouverture de ${EDITOR:-vi} à la validation."
  wait_for_enter
  "${EDITOR:-vi}" "$PATTERNS"
  effective_patterns || { echo "ERREUR: aucun motif dans $PATTERNS." >&2; exit 1; }
  echo "$(wc -l < "$EFFECTIVE" | tr -d ' ') motifs retenus."
}

step_generate_age_key() {
  if [ -f "$SOURCE_DIR/age-key.txt.age" ] && [ -f "$KEY" ]; then
    echo "Paire de clés déjà en place."
    return
  fi
  mkdir -p "$(dirname "$KEY")"
  if [ ! -f "$KEY" ]; then
    age-keygen -o "$KEY" 2>/dev/null
    chmod 600 "$KEY"
    echo "Clé privée générée : $KEY"
  fi
  age-keygen -y "$KEY" > "$SOURCE_DIR/age-recipients.txt"
  echo "Clé publique : $(cat "$SOURCE_DIR/age-recipients.txt")"
  echo
  echo "Chiffrement de la clé privée pour le dépôt. Une seule passphrase, et c'est"
  echo "la dernière barrière publique : longue, propre à cet usage, hors de tout dépôt."
  echo "Attends l'invite avant de taper."
  wait_for_enter
  age --passphrase -o "$SOURCE_DIR/age-key.txt.age" "$KEY"
  echo "Clé chiffrée dans $SOURCE_DIR/age-key.txt.age"
}

step_refresh_local_config() {
  # La configuration locale est produite depuis .chezmoi.toml.tmpl à l'init, pas à
  # l'apply : sans ce passage, chezmoi continue de chiffrer par passphrase.
  chezmoi init
  local identity
  identity="$(chezmoi execute-template '{{ .chezmoi.config.age.identity }}')"
  if [ -z "$identity" ]; then
    echo "ERREUR: la configuration ne déclare pas d'identity age." >&2
    echo "Vérifie $(chezmoi execute-template '{{ .chezmoi.configFile }}')." >&2
    exit 1
  fi
  [ "$identity" = "$KEY" ] || echo "⚠️  identity = $identity, attendu $KEY" >&2
  echo "Configuration régénérée : chiffrement vers $identity."
}

step_encrypt_fragments() {
  local f
  if [ -n "$(chezmoi execute-template '{{ .chezmoi.config.age.passphrase }}' | grep -x true || true)" ]; then
    echo "ERREUR: la configuration chiffre encore par passphrase." >&2
    echo "Les fichiers partiraient chiffrés par un secret que la clé ne lira pas." >&2
    exit 1
  fi
  echo "Rechiffrement des ${#FRAGMENTS[@]} fichiers vers la clé publique."
  echo "Aucune saisie : c'est tout l'objet de la paire de clés."
  for f in "${FRAGMENTS[@]}"; do
    chmod 600 "$f"
    chezmoi add --encrypt "$f"
    echo "  $f"
  done
  echo "Fichiers chiffrés dans $SOURCE_DIR :"
  find "$SOURCE_DIR" -name '*.age' -not -path '*/.git/*' | sed "s|^$SOURCE_DIR/|  |"
}

step_verify_source_tree() {
  echo "Recherche des motifs dans l'arbre de travail de $SOURCE_DIR..."
  if git -C "$SOURCE_DIR" grep -n -I -F -f "$EFFECTIVE" -- . ':!.oh-my-zsh'; then
    echo "ERREUR: motifs sensibles encore présents en clair. Corrige avant de continuer." >&2
    exit 1
  fi
  echo "Arbre de travail propre."
}

step_verify_decryption() {
  local f rc=0
  echo "Relecture de chaque fichier chiffré, sans saisie : c'est ce qui atteste que"
  echo "la clé déchiffre bien ce qu'elle vient de chiffrer."
  for f in "${FRAGMENTS[@]}"; do
    [ -f "$f" ] || continue
    if diff -q <(chezmoi cat "$f") "$f" >/dev/null; then
      echo "  OK   $f"
    else
      echo "  ÉCHEC $f - déchiffrement impossible ou contenu divergent" >&2
      rc=1
    fi
  done
  [ "$rc" = 0 ] || {
    echo "ERREUR: rechiffre les fragments en cause avec la même passphrase." >&2
    exit 1
  }
}

step_commit_fragments() {
  git -C "$SOURCE_DIR" pull --ff-only
  git -C "$SOURCE_DIR" add -A
  if git -C "$SOURCE_DIR" diff --cached --quiet; then
    echo "Rien à commiter : les fragments chiffrés sont déjà dans l'historique."
  else
    git -C "$SOURCE_DIR" commit -m "feat: chiffrer les fragments sensibles"
  fi
  git -C "$SOURCE_DIR" push origin main
  echo "Fragments chiffrés poussés : la réécriture les verra."
}

step_check_both_profiles() {
  local perso="$WORK/chezmoi-perso.toml" rendered missing=0 path
  sed 's/profile = "pro"/profile = "perso"/' "$(chezmoi execute-template '{{ .chezmoi.configFile }}')" > "$perso"
  rendered="$(chezmoi --config "$perso" -S "$SOURCE_DIR" execute-template < "$SOURCE_DIR/.chezmoiignore")"
  # CONTEXT.md n'y est pas : il se déploie sur les deux profils, portant aussi les
  # projets personnels.
  for path in .config/git/pro.gitconfig .config/zsh/pro.zsh .config/zsh/pro.zprofile; do
    # Pas de « grep -q » derrière un tube : il sort dès la première ligne trouvée,
    # ce qui tue le producteur par SIGPIPE et fait échouer le script via pipefail.
    case $'\n'"$rendered"$'\n' in
      *$'\n'"$path"$'\n'*) ;;
      *) echo "⚠️  $path non ignoré en perso" >&2; missing=1 ;;
    esac
  done
  [ "$missing" = 0 ] || {
    echo "ERREUR: une machine perso déploierait un fragment du profil pro." >&2
    exit 1
  }
  echo "Profil perso : les trois fragments pro sont ignorés."
}

step_rewrite_history() {
  awk '{ print $0 "==>REDACTED" }' "$EFFECTIVE" > "$REPLACEMENTS"
  chmod 600 "$REPLACEMENTS"
  rm -rf "$WORK/rewrite"
  git clone "$REMOTE" "$WORK/rewrite"
  # Ordre imposé : suppression de chemin d'abord, remplacement de texte ensuite.
  git -C "$WORK/rewrite" filter-repo \
    --invert-paths --path private_dot_ssh/private_config.d/nas.conf
  git -C "$WORK/rewrite" filter-repo --replace-text "$REPLACEMENTS"
  echo "Historique réécrit dans $WORK/rewrite. Tous les SHA ont changé."
}

step_verify_history() {
  local revs
  revs="$(git -C "$WORK/rewrite" rev-list --all)"
  if git -C "$WORK/rewrite" grep -I -F -f "$EFFECTIVE" $revs -- . ':!.oh-my-zsh'; then
    echo "ERREUR: motifs sensibles encore dans l'historique réécrit." >&2
    echo "Ne publie pas. Complète $PATTERNS et relance." >&2
    exit 1
  fi
  echo "Historique réécrit propre sur $(echo "$revs" | wc -l | tr -d ' ') commits."
}

# Renseigne le champ Commits d'une ADR écrite avant la réécriture, qui ne pouvait
# pas connaître ses propres SHA. Chaque argument est « libellé:début du sujet ».
fill_commits_field() {
  local dir="$1" num="$2" file entry label subject sha field=""
  shift 2
  for file in "$dir"/docs/adr/"$num"-*.md; do
    [ -f "$file" ] && break
    return 0
  done
  for entry in "$@"; do
    label="${entry%%:*}"
    subject="${entry#*:}"
    sha="$(git -C "$dir" log --all --format='%h%x09%s' | awk -F'\t' -v s="$subject" 'index($2, s) == 1 && !seen { print $1; seen = 1 }')"
    [ -n "$sha" ] || { echo "⚠️  ADR-$num : aucun commit pour « $subject »" >&2; continue; }
    [ -z "$field" ] || field="$field, "
    field="$field\`$sha\` ($label)"
  done
  [ -n "$field" ] || return 0
  perl -pi -e "s/^- \*\*Commits\*\* : à renseigner.*\$/- **Commits** : $field/" "$file"
}

step_refresh_adr_commits() {
  local dir="$WORK/rewrite" sha subj new remapped=0
  # git-filter-repo ne conserve pas de table ancien->nouveau exploitable après deux
  # passes : l'appariement se fait par sujet de commit, unique dans ce dépôt.
  for sha in $(grep -rhoE '`[0-9a-f]{7,40}`' "$dir"/docs/adr/*.md | tr -d '`' | sort -u); do
    subj="$(git -C "$SOURCE_DIR" log -1 --format=%s "$sha" 2>/dev/null)" || continue
    [ -n "$subj" ] || continue
    new="$(git -C "$dir" log --all --format='%h%x09%s' | awk -F'\t' -v s="$subj" '$2 == s && !seen { print $1; seen = 1 }')"
    if [ -z "$new" ]; then
      echo "⚠️  aucune correspondance pour $sha - $subj" >&2
      continue
    fi
    perl -pi -e "s/\Q$sha\E/$new/g" "$dir"/docs/adr/*.md
    remapped=$((remapped + 1))
  done
  fill_commits_field "$dir" 016 "sortie du sensible:refactor!: sortir le sensible" \
    "documentation:docs: consigner le passage en public" \
    "fragments chiffrés:feat: chiffrer les fragments"
  fill_commits_field "$dir" 017 "documentation du clone anonyme:docs: consigner le passage en public"
  fill_commits_field "$dir" 018 "paire de clés:feat: chiffrer par paire de clés"
  if grep -rlq "à renseigner après la réécriture" "$dir"/docs/adr/*.md; then
    echo "⚠️  des champs Commits restent à renseigner :" >&2
    grep -rl "à renseigner après la réécriture" "$dir"/docs/adr/*.md >&2
  fi
  git -C "$dir" add -A
  if git -C "$dir" diff --cached --quiet; then
    echo "Aucune référence à remapper."
  else
    git -C "$dir" commit -q -m "docs: reprendre les SHA des ADR après réécriture d'historique"
    echo "$remapped références remappées, commitées dans le clone réécrit."
  fi
  for sha in $(grep -rhoE '`[0-9a-f]{7,40}`' "$dir"/docs/adr/*.md | tr -d '`' | sort -u); do
    git -C "$dir" cat-file -e "${sha}^{commit}" 2>/dev/null || echo "⚠️  référence orpheline : $sha" >&2
  done
}

step_force_push() {
  echo "Publication de l'historique réécrit - destructif et non réversible côté distant :"
  echo
  echo "  cd $WORK/rewrite"
  echo "  git remote add origin $REMOTE   # filter-repo a retiré la remote"
  echo "  git push --force origin main"
  echo
  echo "Tout clone existant devient incompatible : $SOURCE_DIR et les autres machines"
  echo "sont à réinitialiser (chezmoi init) après cette étape."
  wait_for_enter
}

step_switch_visibility() {
  echo "Bascule la visibilité, une fois l'historique vérifié et poussé :"
  echo
  echo "  gh repo edit $REMOTE --visibility public --accept-visibility-change-consequences"
  echo
  echo "Rien ne doit précéder la vérification d'historique : la publication est irréversible,"
  echo "un dépôt public est cloné et mis en cache par des tiers dans la minute."
  wait_for_enter
}

step_backup_key() {
  echo "Sauvegarde les deux secrets, hors de tout dépôt :"
  echo
  echo "  - la passphrase de $SOURCE_DIR/age-key.txt.age"
  echo "  - une copie de $KEY"
  echo
  echo "Les deux perdus, les fichiers chiffrés sont définitivement illisibles."
  echo "La clé seule perdue se rattrape par la passphrase, et inversement."
  wait_for_enter
}

step_cleanup() {
  echo "Nettoyage :"
  echo
  echo "  rm -rf ~/.claude_<ancien suffixe>   # remplacé par ~/.claude_pro, plus géré par chezmoi"
  echo "  rm -rf $WORK                   # motifs et clone de réécriture"
  echo
  echo "Ne supprime pas $KEY : c'est la clé de déchiffrement de cette machine."
  echo
  echo "Révoque enfin le PAT GitHub de bootstrap : le clone anonyme le rend inutile (ADR-017)."
  wait_for_enter
}

# Un « command not found » sur une étape a déjà interrompu une bascule en cours :
# main vérifie que chacune existe avant d'exécuter la première.
check_steps_defined() {
  local step missing=0
  for step in "$@"; do
    declare -F "$step" >/dev/null || { echo "ERREUR: étape $step non définie." >&2; missing=1; }
  done
  [ "$missing" = 0 ] || exit 1
}

main() {
  local steps=(
    step_preconditions
    step_write_patterns_file
    step_generate_age_key
    step_refresh_local_config
    step_encrypt_fragments
    step_verify_source_tree
    step_verify_decryption
    step_commit_fragments
    step_check_both_profiles
    step_rewrite_history
    step_verify_history
    step_refresh_adr_commits
    step_force_push
    step_switch_visibility
    step_backup_key
    step_cleanup
  )
  check_steps_defined "${steps[@]}"
  local step
  for step in "${steps[@]}"; do "$step"; done
  echo "Bascule terminée."
}

main "$@"
