---
name: neovim
description: >
  Maintain the dotfiles repository's Neovim and LazyVim configuration. Use when editing Lua under
  dot_config/nvim/, adding or disabling a plugin, changing options, or updating keymaps. Make sure to
  use it whenever a change affects LazyVim specs or Neovim behaviour, even if Neovim is unnamed.
metadata:
  category: ops
---

# Neovim (LazyVim)

## Overview

The configuration lives at `dot_config/nvim/` in the dotfiles repository and chezmoi deploys it to
`~/.config/nvim`. Edit the repository source, never the deployed copy - `chezmoi apply` would
overwrite it.

It is close to the LazyVim starter: the customisation surface is `lua/config/` for editor behaviour
and `lua/plugins/` for specs. `lua/plugins/example.lua` is the starter's commented catalogue, not
active configuration; do not treat its contents as this setup's conventions.

## Steps

1. Identify the owner of the change before editing:
   - startup and LazyVim bootstrap: `dot_config/nvim/init.lua`, `dot_config/nvim/lua/config/lazy.lua`,
     `dot_config/nvim/lazyvim.json`
   - editor behaviour: `dot_config/nvim/lua/config/{options,keymaps,autocmds}.lua`
   - plugin specs and pins: `dot_config/nvim/lua/plugins/`, `dot_config/nvim/lazy-lock.json`
   - formatting and LSP settings: `dot_config/nvim/stylua.toml`, `dot_config/nvim/.neoconf.json`
2. Add a focused file under `lua/plugins/` named after the concern (`colorscheme.lua`,
   `disabled.lua`), or extend the file that already owns it. Do not create one file per plugin
   mechanically, and do not add specs to `example.lua`.
3. Prefer minimal `opts` over a replacement `config` function when the plugin supports it, and
   preserve lazy loading with the appropriate `event`, `cmd`, `ft`, or `keys` trigger.
4. Place keymaps by ownership: global editor mappings in `lua/config/keymaps.lua`; plugin-owned,
   lazy-loading, and LSP mappings in the owning spec's `keys` field.
5. Change `lazy-lock.json` only through Lazy - `nvim --headless '+Lazy! sync' +qa` - never by hand.
6. After a sync, bring the regenerated lockfile back into the repository with
   `chezmoi add ~/.config/nvim/lazy-lock.json`. Lazy writes to the deployed copy, so the source
   otherwise stays behind and the next `chezmoi apply` silently reverts the plugin versions.
7. Verify: `stylua --check <changed-lua-files>` when available, then
   `nvim --headless -i NONE '+qa'` and resolve every startup error.

## Gotchas

- **Editing `~/.config/nvim` directly** - chezmoi owns that path from `dot_config/nvim/`. The edit
  survives until the next `chezmoi apply`, then disappears.
- **Forgetting `chezmoi add` after a Lazy sync** - the single most likely way to lose plugin updates
  in this setup. The lockfile is generated at the destination, not at the source.
- **Adding configuration to `example.lua`** - it is the starter's inert catalogue. Create a named file
  for a real spec.
- **Centralising every keymap** - moving a plugin-owned mapping out of `keys` breaks its lazy loading.
  Determine ownership first.
- **Running a formatter across all Lua files during a focused change** - unrelated files may already
  differ from current StyLua output; check only what you changed.
- **Installing tooling to verify** - use `stylua` if it is on `PATH` or under
  `~/.local/share/nvim/mason/bin/`; otherwise report that check as unavailable and continue.

## Constraints

- Neovim artifacts stay under `dot_config/nvim/` and follow their existing owner.
- Plugin changes preserve LazyVim loading behaviour and prefer `opts` over `config`.
- Keymaps go where their owner is: global in `config/keymaps.lua`, plugin-owned in the spec's `keys`.
- `lazy-lock.json` changes only through Lazy, and every change is followed by `chezmoi add`.
- Never edit the deployed `~/.config/nvim` as if it were a separate configuration.
- Do not install tooling implicitly; ask before adding a dependency.
