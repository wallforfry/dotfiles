head_ "Bootstrap sans gestionnaire"

chezmoi_bin=$(command -v chezmoi)
if config=$(
  "$chezmoi_bin" execute-template --init --config "${configs[0]}" --source "$PWD" \
    --override-data '{"chezmoi":{"os":"linux"}}' < .chezmoi.toml.tmpl 2>/dev/null
) && grep -Fqx "scriptTempDir = \"$HOME/.cache/chezmoi\"" <<< "$config"; then
  config_ok=1
else
  config_ok=0
  ko ".chezmoi.toml.tmpl ne place pas les scripts sous \$HOME/.cache/chezmoi"
fi

rendered="$tmp/install-tools-linux.sh"
if ! "$chezmoi_bin" execute-template --config "${configs[0]}" --source "$PWD" \
  --override-data '{"chezmoi":{"os":"linux"},"profile":"perso","gui":false}' \
  < run_onchange_before_install-tools.sh.tmpl > "$rendered"; then
  ko "run_onchange_before_install-tools.sh.tmpl ne se rend pas pour Linux"
fi

tested=0
for arch in x86_64 aarch64; do
  scenario="$tmp/bootstrap-$arch"
  fake_bin="$scenario/bin"
  fake_home="$scenario/home"
  curl_log="$scenario/curl.log"
  stdout="$scenario/stdout"
  stderr="$scenario/stderr"
  mkdir -p "$fake_bin" "$fake_home"

  for command_name in awk cat chmod cut grep ln ls mkdir mv rm sed sh tar unzip; do
    command_path=$(command -v "$command_name")
    ln -s "$command_path" "$fake_bin/$command_name"
  done

  printf '%s\n' '#!/bin/sh' 'printf "%s\\n" "$*" >> "$BOOTSTRAP_CURL_LOG"' 'exit 22' \
    > "$fake_bin/curl"
  printf '%s\n' '#!/bin/sh' 'printf "%s\\n" "$SIMULATED_ARCH"' > "$fake_bin/uname"
  printf '%s\n' '#!/bin/sh' 'if [ "${1:-}" = -u ]; then echo 0; else exit 1; fi' > "$fake_bin/id"
  real_mktemp=$(command -v mktemp)
  printf '%s\n' '#!/bin/sh' \
    'case " $* " in *" $HOME/.cache/"*) ;; *) exit 97 ;; esac' \
    "exec \"$real_mktemp\" \"\$@\"" > "$fake_bin/mktemp"
  chmod +x "$fake_bin/curl" "$fake_bin/uname" "$fake_bin/id" "$fake_bin/mktemp"

  if ! HOME="$fake_home" PATH="$fake_bin" SIMULATED_ARCH="$arch" \
    BOOTSTRAP_CURL_LOG="$curl_log" /bin/sh "$rendered" > "$stdout" 2> "$stderr"; then
    ko "bootstrap Linux sans gestionnaire interrompu sur $arch"
    continue
  fi
  if [ -s "$stdout" ]; then
    ko "bootstrap Linux annonce un succès malgré le réseau indisponible sur $arch"
    continue
  fi
  if ! grep -Fq 'zsh absent et aucun gestionnaire de paquets connu' "$stderr" ||
     ! grep -Fq 'starship non installé, prompt zsh par défaut' "$stderr"; then
    ko "bootstrap Linux ne signale pas sa dégradation sur $arch"
    continue
  fi
  case "$arch" in
    x86_64) age_arch=amd64; rust_arch=x86_64-unknown-linux-musl ;;
    aarch64) age_arch=arm64; rust_arch=aarch64-unknown-linux-musl ;;
  esac
  if ! grep -Fq "linux-$age_arch.tar.gz" "$curl_log" ||
     ! grep -Fq "$rust_arch.tar.gz" "$curl_log"; then
    ko "bootstrap Linux ne sélectionne pas les archives attendues sur $arch"
    continue
  fi
  tested=$((tested + 1))
done

if [ "$config_ok" -eq 1 ] && [ "$tested" -eq 2 ]; then
  ok "2 architectures Linux sans gestionnaire, cache utilisateur et dégradation vérifiés"
fi
