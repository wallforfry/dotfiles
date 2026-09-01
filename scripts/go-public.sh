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

SOURCE_DIR="$(chezmoi source-path 2>/dev/null || echo "$HOME/.local/share/chezmoi")"
WORK="$HOME/.cache/go-public"
PATTERNS="$WORK/patterns.txt"
EFFECTIVE="$WORK/patterns.effective.txt"
REPLACEMENTS="$WORK/replacements.txt"
REMOTE="$(git -C "$SOURCE_DIR" remote get-url origin 2>/dev/null || echo '<remote>')"

FRAGMENTS=(
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
  for tool in chezmoi git age gh git-filter-repo; do
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

step_encrypt_fragments() {
  local f
  for f in "${FRAGMENTS[@]}"; do
    if [ ! -f "$f" ]; then
      echo "⚠️  $f absent, fragment ignoré." >&2
      continue
    fi
    chmod 600 "$f"
    echo "Chiffrement de $f - chezmoi demande la passphrase :"
    chezmoi add --encrypt "$f"
  done
  echo "Fragments chiffrés dans $SOURCE_DIR :"
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
  echo "Relecture de chaque fragment chiffré : la passphrase est redemandée."
  echo "C'est le seul contrôle qui attrape une passphrase divergente entre deux fichiers."
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
  echo "Vérifie le rendu sur les deux profils, sur cette machine :"
  echo
  echo "  chezmoi -S $SOURCE_DIR diff"
  echo "  chezmoi -S $SOURCE_DIR execute-template < $SOURCE_DIR/.chezmoiignore"
  echo "  chezmoi -S $SOURCE_DIR --config <(sed 's/\"pro\"/\"perso\"/' ~/.config/chezmoi/chezmoi.toml) diff"
  echo
  echo "Attendu en perso : aucun des fragments pro, donc aucune demande de passphrase pour eux."
  wait_for_enter
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

step_refresh_adr_commits() {
  echo "Les champs **Commits** de toutes les ADR pointent des SHA qui n'existent plus."
  echo "Reprends-les depuis le nouvel historique (git log --oneline), ADR-016 et ADR-017 incluses,"
  echo "puis commite : docs: reprendre les SHA des ADR après réécriture d'historique"
  wait_for_enter
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

step_rotate_passphrase() {
  echo "Change la passphrase age. L'ancienne a protégé des fichiers désormais téléchargeables :"
  echo
  echo "  chezmoi apply --exclude=encrypted   # rien ne dépend des fichiers chiffrés"
  echo "  # pour chaque fichier chiffré, déchiffre avec l'ancienne, rechiffre avec la nouvelle :"
  echo "  chezmoi re-add   # après avoir mis à jour la passphrase du dépôt"
  echo
  echo "Nouvelle passphrase : longue, propre à cet usage, hors de tout dépôt."
  echo "C'est la seule barrière restante, et elle est attaquable hors ligne."
  wait_for_enter
}

step_cleanup() {
  echo "Nettoyage :"
  echo
  echo "  rm -rf ~/.claude_<ancien suffixe>   # remplacé par ~/.claude_pro, plus géré par chezmoi"
  echo "  rm -rf $WORK                   # motifs et clone de réécriture"
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
    step_rotate_passphrase
    step_cleanup
  )
  check_steps_defined "${steps[@]}"
  local step
  for step in "${steps[@]}"; do "$step"; done
  echo "Bascule terminée."
}

main "$@"
