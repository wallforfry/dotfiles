copy_untracked() {
  local source=$1 destination=$2 list=$3 path
  while IFS= read -r -d '' path; do
    mkdir -p -- "$destination/$(dirname "$path")" || return
    cp -Pp -- "$source/$path" "$destination/$path" || return
  done < "$list"
}

probe_capture_contract() {
  local scratch=$1
  mkdir -p "$scratch/probe-source" "$scratch/probe-destination" || return
  printf 'missing-entry\0' > "$scratch/probe-list" || return
  if copy_untracked "$scratch/probe-source" "$scratch/probe-destination" "$scratch/probe-list" \
    >/dev/null 2>&1
  then
    return 1
  fi
  ln -s target "$scratch/probe-source/link" || return
  printf 'link\0' > "$scratch/probe-list" || return
  copy_untracked "$scratch/probe-source" "$scratch/probe-destination" "$scratch/probe-list" || return
  [ -L "$scratch/probe-destination/link" ] &&
    [ "$(readlink "$scratch/probe-destination/link")" = target ]
}

capture_worktree() {
  local source=$1 repository=$2 scratch=$3
  git -C "$source" diff --binary HEAD > "$scratch/worktree.patch" || return
  if [ -s "$scratch/worktree.patch" ]; then
    git -C "$repository" apply "$scratch/worktree.patch" || return
  fi
  git -C "$source" ls-files -z --others --exclude-standard > "$scratch/untracked" || return
  copy_untracked "$source" "$repository" "$scratch/untracked" || return
  git -C "$repository" config user.name harness-audit || return
  git -C "$repository" config user.email harness-audit@invalid || return
  git -C "$repository" config gc.auto 0 || return
  git -C "$repository" config maintenance.auto false || return
  git -C "$repository" add -A || return
  git -C "$repository" commit -qm 'test: capture audit baseline' --allow-empty || return
  git -C "$repository" rev-parse HEAD
}
