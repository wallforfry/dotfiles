#!/usr/bin/env bash
# Bascule du bare repo ~/.dotfiles vers chezmoi.
# Idempotent, interruptible. L'étape destructive est la dernière et n'est
# atteinte qu'après validation du shell.
set -euo pipefail

BARE="$HOME/.dotfiles"
STAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP="$HOME/.dotfiles-backup-$STAMP"
DRY_RUN=0
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

run() {
  if [ "$DRY_RUN" = 1 ]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

dotfiles() { git --git-dir="$BARE" --work-tree="$HOME" "$@"; }

# 1. Préconditions
if ! command -v chezmoi >/dev/null 2>&1; then
  echo "ERREUR: chezmoi n'est pas installé." >&2
  exit 1
fi

if [ ! -d "$BARE" ]; then
  echo "Le bare repo $BARE est absent : migration déjà effectuée. Rien à faire."
  exit 0
fi

if [ -n "$(dotfiles status --porcelain)" ]; then
  echo "ERREUR: le bare repo a des modifications non committées :" >&2
  dotfiles status --short >&2
  echo "Committe-les ou stashe-les avant de migrer." >&2
  exit 1
fi

# 2. Sauvegarde de tous les fichiers suivis
#
# La liste est matérialisée dans un fichier AVANT la boucle : un `ls-files |
# while read` s'exécute dans un sous-shell, et si la boucle n'itère jamais,
# la sauvegarde est silencieusement vide. C'est exactement ce qui s'est
# produit lors de la première migration réelle — d'où le décompte vérifié
# en fin d'étape, qui transforme cet échec silencieux en arrêt net.
echo "==> Sauvegarde dans $BACKUP"
run mkdir -p "$BACKUP"

LIST="$(mktemp)"
trap 'rm -f "$LIST"' EXIT
dotfiles ls-files > "$LIST"
EXPECTED=0
MISSING=0

backup_one() {
  # $1 = chemin relatif à $HOME. Copie puis vérifie que la cible existe.
  # Les entrées de submodule sont des répertoires : on ne peut pas se
  # contenter de compter les fichiers réguliers.
  local f="$1"
  EXPECTED=$((EXPECTED + 1))
  run mkdir -p "$BACKUP/$(dirname "$f")"
  run cp -a "$HOME/$f" "$BACKUP/$f"
  if [ "$DRY_RUN" = 0 ] && [ ! -e "$BACKUP/$f" ]; then
    echo "    ÉCHEC de sauvegarde: $f" >&2
    MISSING=$((MISSING + 1))
  fi
}

while IFS= read -r f; do
  [ -n "$f" ] || continue
  [ -e "$HOME/$f" ] || continue
  backup_one "$f"
done < "$LIST"

[ -f "$HOME/.secrets" ] && backup_one ".secrets"

if [ "$EXPECTED" -eq 0 ]; then
  echo "ERREUR: aucun fichier à sauvegarder. Le bare repo semble vide." >&2
  echo "Rien n'a été supprimé." >&2
  exit 1
fi

if [ "$MISSING" -ne 0 ]; then
  echo "ERREUR: sauvegarde incomplète — $MISSING entrée(s) sur $EXPECTED absente(s)." >&2
  echo "Rien n'a été supprimé. Backup partiel : $BACKUP" >&2
  exit 1
fi

echo "==> Sauvegarde terminée : $EXPECTED entrées"

# 3. Initialisation de chezmoi (prompts : profil, passphrase age)
echo "==> chezmoi init --apply"
run chezmoi init --apply wallforfry

# 4. Validation du shell AVANT toute suppression
echo "==> Vérification du shell"
if [ "$DRY_RUN" = 0 ]; then
  if ! zsh -ic 'exit' >/dev/null 2>&1; then
    echo "ERREUR: le shell ne démarre pas correctement." >&2
    echo "Rien n'a été supprimé. Backup : $BACKUP" >&2
    exit 1
  fi
fi
echo "==> Shell OK"

# 5. Suppression de l'ancien système (seule étape destructive)
echo "==> Suppression de l'ancien système"
for p in "$BARE" "$HOME/shells" "$HOME/.p10k.zsh" "$HOME/submodules.list" \
         "$HOME/.gitmodules" "$HOME/install" "$HOME/scripts"; do
  [ -e "$p" ] && run rm -rf "$p"
done

echo
echo "Migration terminée. Backup conservé dans $BACKUP"
echo "Source chezmoi : $(chezmoi source-path 2>/dev/null || echo '~/.local/share/chezmoi')"
echo "Ouvre un nouveau shell pour vérifier."
