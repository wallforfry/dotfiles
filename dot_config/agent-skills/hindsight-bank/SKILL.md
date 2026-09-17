---
name: hindsight-bank
description: >
  Create a Hindsight bank or manage a local directory-to-bank association. Use when the user asks to
  create a memory bank, attach a folder to a bank, detach a folder, or change its bank. Make sure to
  use it whenever Hindsight bank routing changes, even if the user only mentions a local directory.
compatibility: Requires the dotfiles CLI, hindsight for remote-bank creation, chezmoi, and the encrypted native ~/.hindsight/coding-agent.json configuration.
metadata:
  category: ops
---

# Hindsight Bank Management

## Overview

This skill manages explicit Hindsight banks and the local directory-to-bank mapping. The native
`mapPathToBank` object in the encrypted `~/.hindsight/coding-agent.json` is the sole mapping
source; Cursor is reconciled from it. Creating or deleting a remote bank is separate from
attaching or detaching a local directory.

## Usage

Run `dotfiles hindsight bank create <bank-id>` only after the user explicitly asks for a new remote
bank. Run `dotfiles hindsight bank add <directory> <bank-id>` to create or replace one local mapping,
or `dotfiles hindsight bank remove <directory>` to detach it without deleting remote memory.

## Steps

1. Confirm the requested operation and exact bank identifier. For `create`, confirm the user wants
   a new remote bank, not merely a local mapping.
2. For an existing bank, use `hindsight bank list` or `hindsight bank stats <bank-id>` before adding
   a mapping when the bank identifier is uncertain.
3. Run the dotfiles command from this skill. It records or removes the requested canonical path
   directly in `mapPathToBank`, re-encrypts the chezmoi source, and applies the configuration.
4. Verify the mapping with `hindsight_diagnose` in a new agent session or by checking that the
   intended client exposes the expected bank. Do not display the Hindsight configuration file.

## Gotchas

- **Treating `remove` as remote deletion** - it only disables local routing; the bank and its data
  remain available. Use `hindsight bank delete` only after an explicit, separate request.
- **Expecting a worktree to inherit its parent mapping** - routing is path-specific; associate each
  worktree that needs its own local routing.
- **Creating a bank to fix a typo** - a new remote bank is persistent and empty; list banks first
  when the intended identifier is not certain.
- **Editing the deployed JSON manually** - chezmoi would later overwrite the change; use the CLI
  so it re-encrypts the source and triggers reconciliation.
- **Interrupted chezmoi application** - a client configuration can be only partly reconciled; do not
  revert the encrypted source manually, then rerun `chezmoi apply --force`.

## Constraints

- Never create or delete a remote bank without an explicit user request naming the operation.
- Never delete a bank as part of removing a local directory association.
- Never print the JSON configuration, API URL, or API token.
- Never map a relative or missing directory.
