# Owned-assets runtime model

The owned-assets cutover is complete.

`apply-gentle-ai-custom` installs only true repo-owned assets from `overlay/gentle-ai/assets/owned/...`. Upstream Gentle AI materializes SDD prompts, contracts, commands, skills, and plugins; audit/sync maintain source snapshots without turning them into runtime artifacts.

## Sources of truth

| Concern | Source of truth |
| --- | --- |
| Upstream maintenance boundary | `overlay/gentle-ai/state/upstream-state.json` |
| Managed upstream/owned mapping | `overlay/gentle-ai/policy/managed-assets.json` |
| Approved upstream asset snapshots | `overlay/gentle-ai/assets/upstream/` |
| Repo-owned runtime assets | `overlay/gentle-ai/assets/owned/` |
| Surgical runtime patches | `overlay/gentle-ai/patches/` |
| Repo-owned portable skills | `shared/skills/` |
| Repo-owned custom command bodies | `shared/commands/` |

## Runtime rules

1. Upstream Gentle AI owns and materializes the native SDD/review runtime surface.
2. `overlay/gentle-ai/assets/owned/...` is canonical only for true local assets such as the extended `AGENTS.md`.
3. `shared/skills/` remains canonical for portable repo-owned skills installed outside the managed SDD/runtime asset tree.
4. `shared/commands/` remains canonical for commit/PR wrapper bodies rendered by agent-specific wrappers.

## Command responsibilities

### `audit-gentle-ai-upstream`

- defaults to a clean upstream checkout at the exact release tag for installed `gentle-ai --version`; `--release v<V>` can select another exact clean tag without changing the installed binary or runtime
- reads `last_maintained_commit` from `upstream-state.json`
- diffs upstream with `git diff --name-status --find-renames <last_maintained_commit>..<verified-release-commit>`
- filters changes through `managed-assets.json`
- reports managed drift plus structural invariant drift
- validates the audited `gentle-orchestrator` baseline/metadata against upstream and state

### `sync-gentle-ai-upstream-assets`

- repeats the exact release check before copy or write; supports the same optional `--release v<V>` selector while retaining installed-binary parity by default
- copies approved upstream assets into `overlay/gentle-ai/assets/upstream/...`
- updates the audited `gentle-orchestrator` baseline + metadata
- advances `upstream-state.json` with the verified version, exact tag, and release commit when the new upstream state is accepted
- does not touch runtime files under `~/.config/opencode/`

### `apply-gentle-ai-custom`

- validates the installed upstream runtime bundle before any write
- preserves upstream-rendered orchestrator/phase prompts and review plugins
- applies declared surgical patches only after exact-anchor compatibility checks
- tracks every custom-installed output and protects upstream-returned or patch-target paths during stale-state migration without claiming upstream files as custom-owned
- prunes rejected upstream skills in the selected registered targets
- applies built-in `agent_overrides`
- reconciles `default_profile` plus named `profiles`, including local JD/review assignment wiring
- installs repo-owned portable skills from `shared/skills/`
- renders repo-owned custom wrappers from `shared/commands/`

## Key paths

- `assets/upstream/opencode/` — approved source snapshots for audit and sync
- `assets/owned/opencode/AGENTS.md` — true repo-owned runtime instruction file
- `patches/opencode/judgment-retrospective.json` — exact local Judgment Day delta

## Notes

- Required SDD phase assets, `_shared/`, and strict-TDD references remain in the managed audit set, but are installed by upstream.
- The repo-owned OpenCode `AGENTS.md` encodes the local depuration of PR/review governance directly in the owned prompt file; apply no longer derives that behavior dynamically from upstream runtime state.
- Managed named profiles clone installed base agents and then change only local model/variant assignments while reusing the available base prompt values required by the profile family.
- Optional named Judgment Day actors clone their unsuffixed upstream structural actor. Native review actors remain global; explicit inheritance removes only assignment fields and never fabricates upstream structure.
