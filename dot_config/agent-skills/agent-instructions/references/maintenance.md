# Instruction Maintenance

## The map

| Layer | File | Role |
| --- | --- | --- |
| Canonical | `harness/AGENTS.md` | technical rules that apply to every project |
| Canonical | `harness/SOUL.md` | voice, register, typography, priorities |
| Canonical | `harness/USER.md` | user preferences and technical context |
| Projection | `dot_claude/{AGENTS,SOUL,USER}.md.tmpl` | one line each: `{{ include "harness/..." }}` |
| Adapter | `dot_claude/CLAUDE.md` | Claude entry point, imports the three projections |
| Encrypted | `dot_claude/encrypted_private_CONTEXT.md.age` | named contexts, `pro` profile only |
| Adapter | `dot_codex/AGENTS.md.tmpl` | Codex entry point, inlines the three sources |
| Projection | `dot_claude_pro/symlink_*` | the `pro` profile pointing at `~/.claude` |
| Projection | `dot_{claude,codex}/skills/symlink_<slug>` | one link per skill, per host |
| Repository | `AGENTS.md` at the root | rules specific to this repository |
| Adapter | `CLAUDE.md` at the root | one line pointing at `AGENTS.md` |
| Conditional | `dot_config/agent-skills/<slug>/SKILL.md` | procedures loaded on demand |

`harness/` is listed in `.chezmoiignore`: it is a source read by templates, never deployed as such.
The three canonical files are the ones an agent loads on every task, in every project - which is
why what goes in them is a decision, not a preference.

## Placement policy

Ask, in order:

1. **Does every task need it?** No, and it stops here: it is a skill. A rule about pull requests
   does not belong in a file loaded while editing a shell script.
2. **Is it about this repository only?** Then it belongs in the root `AGENTS.md`, not in `harness/`.
3. **Is it a voice or a person, rather than a rule?** Then `harness/SOUL.md` or `harness/USER.md`.
   Their French is deliberate and is not an inconsistency to fix.
4. **Does something already say it?** Then amend that place. Two statements of one rule are one
   rule and one future contradiction.

An always-loaded rule earns its place by changing behaviour on tasks that have nothing to do with
its subject. Everything else is a skill with a description that routes to it.

## Consumers

A canonical source is never alone. On any change, walk this list and update what applies in the
same commit:

- the projections: `dot_claude/*.md.tmpl`, and `dot_claude_pro/symlink_*` for the `pro` profile;
- the adapters: `dot_claude/CLAUDE.md`, the root `CLAUDE.md`;
- `.chezmoiignore`, when a new source must stay undeployed, or a fragment is `pro`-only;
- the indexes: `dot_config/agent-skills/README.md`, `docs/adr/README.md`;
- the barrier: `scripts/verify.sh`, which must cover the new shape as well as the old;
- the documentation: `README.md`, `docs/`, and the ADR that records the decision if one does.

## Verification

- `chezmoi diff` - the deployed effect. A template that renders is not a template that lands where
  you think.
- `chezmoi execute-template < <file>` - one template in isolation, when the diff is unclear.
- `bash scripts/verify.sh` - renders every template across the three profile combinations, and
  checks the skills, the subagents and the ADR index.
- Both values of `.profile` for anything conditional. The current machine proves one.

Report what you exercised and where. Green on the `perso` profile says nothing about `pro`.

## Recorded decisions

Read them in `docs/adr/` of the dotfiles checkout; a deployed skill has no path to them.

- `012-instructions-ia-versionnees.md` - why the instructions are versioned in the repository
  rather than kept in `~`.
- `016-depot-public-sensible-chiffre.md` - the repository is public: no name, host or client in an
  instruction file, in clear.

Never contradict an ADR in force silently. Follow it, or state the conflict and deliver the change
along with the ADR that would need superseding.
