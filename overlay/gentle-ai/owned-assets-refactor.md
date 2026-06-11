# Owned-assets runtime model

The owned-assets cutover is complete.

`apply-gentle-ai-custom` installs repo-owned managed runtime assets directly from `overlay/gentle-ai/assets/owned/...`, while `audit-gentle-ai-upstream` and `sync-gentle-ai-upstream-assets` maintain the approved upstream side of the control plane.

## Sources of truth

| Concern | Source of truth |
| --- | --- |
| Upstream maintenance boundary | `overlay/gentle-ai/state/upstream-state.json` |
| Managed upstream/owned mapping | `overlay/gentle-ai/policy/managed-assets.json` |
| Approved upstream asset snapshots | `overlay/gentle-ai/assets/upstream/` |
| Repo-owned SDD/runtime assets | `overlay/gentle-ai/assets/owned/` |
| Repo-owned portable skills | `shared/skills/` |
| Repo-owned custom command bodies | `shared/commands/` |

## Runtime rules

1. `overlay/gentle-ai/assets/owned/...` is the canonical runtime source for the OpenCode managed runtime assets, including the locally extended `AGENTS.md` surface.
2. Runtime directories under `~/.config/opencode/` are deployment targets only.
3. `shared/skills/` remains canonical for portable repo-owned skills installed outside the managed SDD/runtime asset tree.
4. `shared/commands/` remains canonical for commit/PR wrapper bodies rendered by agent-specific wrappers.

## Command responsibilities

### `audit-gentle-ai-upstream`

- reads `last_maintained_commit` from `upstream-state.json`
- diffs upstream with `git diff --name-status --find-renames <last_maintained_commit>..HEAD`
- filters changes through `managed-assets.json`
- reports managed drift plus structural invariant drift
- validates the audited `gentle-orchestrator` baseline/metadata against upstream and state

### `sync-gentle-ai-upstream-assets`

- copies approved upstream assets into `overlay/gentle-ai/assets/upstream/...`
- updates the audited `gentle-orchestrator` baseline + metadata
- advances `upstream-state.json` when the new upstream state is accepted
- does not touch runtime files under `~/.config/opencode/`

### `apply-gentle-ai-custom`

- copies repo-owned SDD/runtime assets from `overlay/gentle-ai/assets/owned/...` into runtime prompt/skill/command targets
- rewrites `opencode.json` prompt refs to those runtime files
- prunes rejected upstream skills in the selected registered targets
- applies built-in `agent_overrides`
- reconciles `default_profile` plus named `profiles`
- installs repo-owned portable skills from `shared/skills/`
- renders repo-owned custom wrappers from `shared/commands/`

## Directory map

```text
overlay/gentle-ai/
  assets/
    upstream/opencode/
      AGENTS.md
      prompts/orchestrators/
      skills/
      commands/
    owned/opencode/
      AGENTS.md
      prompts/orchestrators/
      skills/
      commands/
```

## Notes

- Required SDD phase assets live under `skills/` because upstream ships them from `internal/assets/skills/<phase>/SKILL.md` and runtime reuses the same content both as a skill and as a prompt file.
- `_shared/`, `strict-tdd.md`, and `strict-tdd-verify.md` are behavior-defining SDD support files and belong in the managed asset set.
- The repo-owned OpenCode `AGENTS.md` encodes the local depuration of PR/review governance directly in the owned prompt file; apply no longer derives that behavior dynamically from upstream runtime state.
- Directory-style owned assets can choose between `prune` and `merge` sync modes. `_shared` is fully owned and pruned; `sdd-commands` uses merge mode so OpenCode keeps repo-owned custom wrappers and any unrelated local commands outside the SDD command set.
