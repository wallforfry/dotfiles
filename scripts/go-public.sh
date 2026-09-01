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
REPLACEMENTS="$WORK/replacements.txt"
REMOTE="$(git -C "$SOURCE_DIR" remote get-url origin 2>/dev/null || echo '<remote>')"

wait_for_enter() { read -rp "→ Entrée pour continuer... "; }

step_preconditions() {
  local missing=0 tool
  for tool in chezmoi git age gh git-filter-repo; do
    command -v "$tool" >/dev/null 2>&1 || { echo "ERREUR: $tool absent." >&2; missing=1; }
  done
  [ "$missing" = 0 ] || exit 1
  mkdir -p "$WORK"
  chmod 700 "$WORK"
  echo "Préconditions OK. Source chezmoi : $SOURCE_DIR"
  echo "Répertoire de travail hors dépôt : $WORK"
}

step_write_patterns_file() {
  echo "Écris dans $PATTERNS un motif sensible par ligne, littéral, sans commentaire :"
  echo "  - nom d'hôte du NAS, adresses IP locale et VPN, nom d'utilisateur"
  echo "  - noms d'employeur, de clients, de projets internes"
  echo "  - noms de configurations d'environnement"
  echo
  echo "Ce fichier ne doit jamais entrer dans le dépôt : il vit sous $WORK."
  echo "Commande : \$EDITOR $PATTERNS && chmod 600 $PATTERNS"
  wait_for_enter
  [ -s "$PATTERNS" ] || { echo "ERREUR: $PATTERNS vide." >&2; exit 1; }
  echo "$(wc -l < "$PATTERNS" | tr -d ' ') motifs lus."
}

step_encrypt_fragments() {
  echo "Chiffre les fragments. Chaque commande demande la passphrase age :"
  echo
  echo "  chmod 600 ~/.ssh/config.d/nas.conf ~/.config/zsh/pro.zsh \\"
  echo "            ~/.config/zsh/pro.zprofile ~/.config/git/pro.gitconfig ~/.claude/CONTEXT.md"
  echo "  chezmoi add --encrypt ~/.ssh/config.d/nas.conf"
  echo "  chezmoi add --encrypt ~/.config/zsh/pro.zsh"
  echo "  chezmoi add --encrypt ~/.config/zsh/pro.zprofile"
  echo "  chezmoi add --encrypt ~/.config/git/pro.gitconfig"
  echo "  chezmoi add --encrypt ~/.claude/CONTEXT.md"
  echo
  echo "Le chmod 600 d'abord : c'est lui qui fait poser l'attribut private_ par chezmoi."
  wait_for_enter
}

step_verify_source_tree() {
  echo "Recherche des motifs dans l'arbre de travail de $SOURCE_DIR..."
  if git -C "$SOURCE_DIR" grep -n -I -F -f "$PATTERNS" -- . ':!.oh-my-zsh'; then
    echo "ERREUR: motifs sensibles encore présents en clair. Corrige avant de continuer." >&2
    exit 1
  fi
  echo "Arbre de travail propre."
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
  awk 'NF { print $0 "==>REDACTED" }' "$PATTERNS" > "$REPLACEMENTS"
  chmod 600 "$REPLACEMENTS"
  echo "Réécriture d'historique. À faire sur un clone frais, jamais sur $SOURCE_DIR :"
  echo
  echo "  git clone $REMOTE $WORK/rewrite && cd $WORK/rewrite"
  echo "  git filter-repo --invert-paths --path private_dot_ssh/private_config.d/nas.conf"
  echo "  git filter-repo --replace-text $REPLACEMENTS"
  echo
  echo "Ordre imposé : la suppression de chemin d'abord, le remplacement de texte ensuite."
  echo "Tous les SHA changent - c'est le but, et c'est ce qui périme les champs Commits des ADR."
  wait_for_enter
}

step_verify_history() {
  echo "Vérifie l'historique réécrit :"
  echo
  echo "  cd $WORK/rewrite"
  echo "  git grep -I -F -f $PATTERNS \$(git rev-list --all) -- . ':!.oh-my-zsh' ; echo \"code \$?\""
  echo
  echo "Attendu : aucune ligne, code 1. Une seule correspondance annule la publication."
  wait_for_enter
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

main() {
  step_preconditions
  step_write_patterns_file
  step_encrypt_fragments
  step_verify_source_tree
  step_check_both_profiles
  step_rewrite_history
  step_verify_history
  step_refresh_adr_commits
  step_force_push
  step_switch_visibility
  step_rotate_passphrase
  step_cleanup
  echo "Bascule terminée."
}

main "$@"
