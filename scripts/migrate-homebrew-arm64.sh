#!/usr/bin/env bash
# Migre Homebrew de /usr/local (Intel, sous Rosetta) vers /opt/homebrew (arm64).
#
# bash et non sh : ce script ne tourne jamais au bootstrap, il se lance à la
# main sur un mac déjà outillé. Do-nothing script - chaque étape non automatisée
# imprime la commande exacte puis attend, afin que la procédure cesse de vivre
# dans une tête. Voir blog.danslimmon.com/2019/07/15/do-nothing-scripting.
#
# Usage à sens unique et à usage unique : supprimer ce fichier du dépôt une fois
# la machine migrée.

set -euo pipefail

INTEL=/usr/local
ARM=/opt/homebrew
BREWFILE="$HOME/.cache/Brewfile.intel"
FORMULES="$HOME/.cache/Brewfile.intel.formules"

wait_for_enter() { read -rp "⏎  Entrée pour continuer... "; }

step_verifier_prerequis() {
  if [ "$(uname -m)" != arm64 ]; then
    echo "❌  Cette machine n'est pas arm64 : rien à migrer." >&2
    exit 1
  fi
  for b in "$INTEL/bin/brew" "$ARM/bin/brew"; do
    if [ ! -x "$b" ]; then
      echo "❌  $b absent." >&2
      exit 1
    fi
  done
  echo "✅  arm64, et les deux préfixes répondent."
}

step_inventorier_intel() {
  mkdir -p "$(dirname "$BREWFILE")"
  # L'avertissement « circular dependency: libtiff, webp » vient de tabs de
  # kegs périmés côté Intel. Il n'affecte pas le dump.
  "$INTEL/bin/brew" bundle dump --force --file="$BREWFILE"

  # « brew bundle install » n'a pas de commutateur --formula : les casks se
  # traitent à part (step_reenregistrer_les_casks), donc le fichier servant à
  # l'installation ne contient que les formules.
  grep '^brew ' "$BREWFILE" > "$FORMULES"

  printf '📋  %s : %s taps, %s formules, %s casks\n' \
    "$BREWFILE" \
    "$(grep -c '^tap ' "$BREWFILE" || true)" \
    "$(grep -c '^brew ' "$BREWFILE" || true)" \
    "$(grep -c '^cask ' "$BREWFILE" || true)"
}

step_signaler_les_formules_hors_core() {
  echo "Trois formules du brew Intel ne sont plus dans homebrew/core et ne se"
  echo "réinstalleront pas telles quelles. Décide de chacune avant de continuer :"
  echo
  echo "  terraform 1.5.7  retirée de core au changement de licence. Soit"
  echo "                   « hashicorp/tap/terraform » et sa BSL, soit tenv."
  echo "  python@3.8       en fin de vie, retirée de core. À abandonner."
  echo "  linear           vient de schpet/tap, retapé à l'étape précédente."
  echo
  echo "Retire de $FORMULES ce que tu ne veux pas voir échouer :"
  echo
  echo "  \$EDITOR $FORMULES"
  wait_for_enter
}

step_arreter_les_services() {
  echo "Arrête les services du brew Intel, sinon deux copies se disputeront"
  echo "les mêmes ports et les mêmes répertoires d'état :"
  echo
  echo "  $INTEL/bin/brew services stop sleepwatcher"
  echo "  $INTEL/bin/brew services stop asimov"
  echo
  echo "Puis vérifie qu'il ne reste rien de démarré :"
  echo
  echo "  $INTEL/bin/brew services list"
  wait_for_enter
}

step_sauvegarder_postgres() {
  echo "PostgreSQL 16 a un répertoire de données sous $INTEL/var/postgresql@16."
  echo "Un dump logique traverse le changement d'architecture sans surprise ;"
  echo "une copie du répertoire suppose une version majeure identique."
  echo
  echo "  $INTEL/bin/brew services start postgresql@16"
  echo "  $INTEL/opt/postgresql@16/bin/pg_dumpall -U \"\$(id -un)\" > ~/pg16-avant-migration.sql"
  echo "  $INTEL/bin/brew services stop postgresql@16"
  echo
  echo "Vérifie que le fichier n'est pas vide avant de continuer :"
  echo
  echo "  ls -lh ~/pg16-avant-migration.sql"
  wait_for_enter
}

step_retaper_les_taps() {
  # homebrew/services est intégré à brew depuis 4.1 et n'est plus tapable.
  grep '^tap ' "$BREWFILE" | sed -E 's/^tap "([^"]+)".*/\1/' |
    grep -v '^homebrew/services$' | while read -r t; do
      "$ARM/bin/brew" tap "$t" 2>/dev/null || echo "⚠️   tap $t non ajouté" >&2
    done
  printf '🚰  %s taps dans %s\n' "$("$ARM/bin/brew" tap | wc -l | tr -d ' ')" "$ARM"
}

step_installer_les_formules() {
  printf 'Installation de %s formules dans %s. Hors les trois signalées, toutes\n' \
    "$(grep -c '^brew ' "$FORMULES" || true)" "$ARM"
  echo "ont une bouteille arm64_tahoe ou « all » : aucune compilation."
  "$ARM/bin/brew" bundle install --no-upgrade --file="$FORMULES" ||
    echo "⚠️   certaines formules ont échoué, voir au-dessus" >&2
  printf '🍺  %s formules explicites dans %s\n' \
    "$("$ARM/bin/brew" leaves | wc -l | tr -d ' ')" "$ARM"
}

step_comparer_les_inventaires() {
  local manquantes
  manquantes=$(comm -23 \
    <("$INTEL/bin/brew" leaves | sort) \
    <("$ARM/bin/brew" leaves | sort))
  if [ -z "$manquantes" ]; then
    echo "✅  Toutes les formules explicites du brew Intel existent en arm64."
  else
    echo "⚠️   Absentes de $ARM :" >&2
    echo "$manquantes" | sed 's/^/     /' >&2
  fi
}

step_donner_la_priorite_a_arm() {
  echo "/usr/local/bin passe avant /opt/homebrew/bin parce qu'il vient de"
  echo "/etc/paths, que path_helper place avant /etc/paths.d/homebrew."
  echo
  echo "Le dépôt corrige l'ordre dans .zprofile. Applique-le, puis ouvre un"
  echo "shell neuf et vérifie :"
  echo
  echo "  chezmoi apply"
  echo "  exec zsh -l"
  echo "  brew --prefix    # doit répondre $ARM"
  wait_for_enter
}

step_restaurer_postgres() {
  echo "Recharge le dump dans le PostgreSQL arm64 :"
  echo
  echo "  $ARM/bin/brew services start postgresql@16"
  echo "  $ARM/opt/postgresql@16/bin/psql -U \"\$(id -un)\" -d postgres -f ~/pg16-avant-migration.sql"
  echo
  echo "Puis contrôle la liste des bases :"
  echo
  echo "  $ARM/opt/postgresql@16/bin/psql -U \"\$(id -un)\" -d postgres -c '\\l'"
  wait_for_enter
}

step_reinstaller_les_venvs_pipx() {
  echo "Les venvs pipx (requests, tinytuya) pointent sur le python de $INTEL"
  echo "et cesseront de fonctionner à sa suppression :"
  echo
  echo "  pipx reinstall-all"
  wait_for_enter
}

step_reenregistrer_les_casks() {
  printf 'Les %s casks sont enregistrés auprès du brew Intel. Les applications\n' \
    "$(grep -c '^cask ' "$BREWFILE" || true)"
  echo "elles-mêmes ne bougent pas - elles vivent dans ~/Applications - mais"
  echo "leur suivi de version disparaîtra avec lui. Réenregistre-les :"
  echo
  echo "  grep '^cask ' $BREWFILE | sed -E 's/^cask \"([^\"]+)\".*/\\1/' |"
  echo "    xargs -n1 $ARM/bin/brew install --cask --appdir=\"\$HOME/Applications\" --force"
  echo
  echo "Deux cas à surveiller : macfuse installe une extension noyau et exige"
  echo "une autorisation dans Réglages Système puis un redémarrage ;"
  echo "wireshark-app et gqrx demandent aussi des droits."
  wait_for_enter
}

step_supprimer_le_brew_intel() {
  printf '⚠️   ÉTAPE IRRÉVERSIBLE - %s, %s formules installées, %s casks.\n' \
    "$(du -sh "$INTEL" 2>/dev/null | cut -f1 | tr -d ' ')" \
    "$("$INTEL/bin/brew" list --formula | wc -l | tr -d ' ')" \
    "$(grep -c '^cask ' "$BREWFILE" || true)"
  echo
  echo "Ne la fais qu'une fois les étapes précédentes vérifiées, et de"
  echo "préférence après quelques jours d'usage du brew arm64."
  echo
  echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/uninstall.sh)\" -- --path=$INTEL"
  echo
  echo "L'installeur liste ce qu'il va supprimer et demande confirmation."
  echo "$BREWFILE reste sur le disque : c'est le seul inventaire de ce qui"
  echo "existait. Ne le supprime pas avant d'être sûr."
  wait_for_enter
}

step_relancer_les_services() {
  echo "Redémarre les services sous le brew arm64 :"
  echo
  echo "  $ARM/bin/brew services start sleepwatcher"
  echo
  echo "asimov était en erreur avant la migration : vérifie s'il te sert"
  echo "encore avant de le relancer."
  wait_for_enter
}

main() {
  step_verifier_prerequis
  step_inventorier_intel
  step_signaler_les_formules_hors_core
  step_arreter_les_services
  step_sauvegarder_postgres
  step_retaper_les_taps
  step_installer_les_formules
  step_comparer_les_inventaires
  step_donner_la_priorite_a_arm
  step_restaurer_postgres
  step_reinstaller_les_venvs_pipx
  step_reenregistrer_les_casks
  step_supprimer_le_brew_intel
  step_relancer_les_services
  echo "🎉  Migration terminée. Retire ce script du dépôt."
}

main "$@"
