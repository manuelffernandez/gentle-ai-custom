# Maintenance

Human guide to the Gentle AI overlay.

This file describes the current operating model: `apply-gentle-ai-custom` reinstalls repo-owned assets from `overlay/gentle-ai/assets/owned/...`; `audit-gentle-ai-upstream` and `sync-gentle-ai-upstream-assets` maintain the relationship with upstream and the approved baseline.

## What this file does not define

- agent/runtime behavior -> `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- repo ownership and policy -> `AGENTS.md`
- keep/prune intent and desired orchestrator behavior -> `policy/maintenance-intent.md`
- machine-readable policy -> `policy/gentle-ai-policy.json`
- maintained upstream boundary -> `state/upstream-state.json`
- closed-event ledger -> `logs/update-log.md`

## Quick path

1. Update the `gentle-ai` binary.
2. Run `git pull` in the resolved upstream `gentle-ai` clone.
3. From `gentle-ai-custom`, run `bash audit-gentle-ai-upstream.sh`.
4. Convert the audit into a concise decision summary before any mutation:
   - what is new upstream
   - recommend adopt
   - recommend discard
   - why
   - recommended runtime path: `gentle-ai sync` vs full reinstall
5. STOP for explicit approval before updating this repo, advancing `state/upstream-state.json`, running `bash sync-gentle-ai-upstream-assets.sh`, or refreshing runtime.
6. If a new upstream boundary was approved, run `bash sync-gentle-ai-upstream-assets.sh` to refresh `overlay/gentle-ai/assets/upstream/...`.
7. Execute the recommended upstream refresh path:
   - `gentle-ai sync` if topology did not change
   - full reinstall if topology changed or sync no longer materializes the right state
8. Re-apply the overlay with `bash apply-gentle-ai-custom.sh opencode`.
9. Read `Summary:`, verify the final on-disk state, run one fresh-context consistency review, and return a closing summary of what was actually adopted vs discarded and why.
10. If `~/.config/opencode/opencode.json` changed, restart OpenCode.

## Operating model

| Artifact | Role |
| --- | --- |
| `policy/maintenance-intent.md` | Human intent: what to keep, depure, and protect |
| `policy/gentle-ai-policy.json` | Runtime policy consumed by the Go CLI and wrappers |
| `policy/managed-assets.json` | Canonical map of approved upstream assets and installable owned assets |
| `assets/upstream/` | Approved upstream copies for review/diff |
| `assets/owned/` | Repo-owned assets that `apply` installs into runtime |
| `shared/skills/` | Portable repo-owned skills installed globally by `apply` |
| `shared/commands/` | Source bodies for custom wrappers rendered by `apply` |
| `state/upstream-state.json` | Last maintained upstream boundary |

## What each command does

### `audit-gentle-ai-upstream`

- uses `last_maintained_commit` from `state/upstream-state.json`
- discovers drift with `git diff --name-status --find-renames <last_maintained_commit>..HEAD`
- filters that drift through `policy/managed-assets.json`
- keeps verifying upstream structural invariants (`profiles.go`, `inject.go`, etc.)

### Decision handoff before mutation

Before the maintainer edits repo files or refreshes runtime, turn the audit into an approval gate:

- `What is new upstream` — concise change summary for the reviewed upstream range
- `Recommend adopt` — overlay-relevant behavior/assets worth carrying forward
- `Recommend discard` — upstream additions the overlay should keep pruning or reject
- `Why` — rationale for both lists
- `Recommended runtime path` — `gentle-ai sync` vs full reinstall, with the topology reason when relevant

No repo mutation happens before that handoff is approved.

### `sync-gentle-ai-upstream-assets`

- copies approved upstream assets into `assets/upstream/...`
- advances `state/upstream-state.json` when the new upstream was accepted
- does not touch local runtime under `~/.config/opencode/`

### `apply-gentle-ai-custom`

- installs repo-owned SDD/runtime assets from `assets/owned/...`
- rewrites `opencode.json` so the base and SDD profiles use those prompt files
- prunes rejected upstream skills only in the selected registered targets
- applies `agent_overrides`
- reconciles `default_profile` and `profiles`
- installs repo-owned skills from `shared/skills/`
- renders custom wrappers from `shared/commands/`

`apply` no longer depends on sanitization, inline prompt capture, local operational snapshots, or snapshot-based recovery.

## Update types and impact

| Update path | What changes | Overlay impact |
| --- | --- | --- |
| `brew upgrade gentle-ai` | Only the binary | Usually does not reset the overlay |
| `gentle-ai sync` | Prompts, skills, MCP configs, SDD assets | Restores upstream runtime state; re-applying the overlay is mandatory |
| TUI reinstall | Full installation, topology, presets, and config | Resets everything and may change agents/presets |

## Highest-signal indicators

| Signal | Meaning | Action |
| --- | --- | --- |
| `base prompt drift: yes` | Upstream `gentle-orchestrator` changed relative to the approved upstream asset | Read `Drift summary:` first |
| `profile ... mismatch` / `base asset injection invariant: mismatch` | Upstream SDD profile mechanics changed | Stop and audit before recommending `sync` |
| `topology: unknown orchestrator matched by prefix only` | A new upstream orchestrator appeared | Audit it and decide whether policy must include it |
| `topology: expected orchestrator missing from opencode.json` | A known orchestrator disappeared or was renamed | Audit upstream and update policy/intent if needed |
| `WARNING - unmanaged SDD profiles left untouched` | `opencode.json` contains profiles absent from the active `profiles` source | Decide whether to manage them in local config or remove them manually |
| `owned asset writes - ...` | `apply` installed or left repo-owned assets untouched | Review `--verbose` output if you need file-level detail |

## Post-state verification

After `apply`, confirm this:

- pruned skills no longer exist in each selected registered target
- each effective `agent_override` resolves to the expected `model` / `variant`
- `agent.gentle-orchestrator.prompt` points to `~/.config/opencode/prompts/sdd/orchestrators/gentle-orchestrator.overlay.md`
- each managed `sdd-orchestrator-<name>` points to that same owned prompt file
- each `sdd-<phase>` and each managed `sdd-<phase>-<name>` points to its owned prompt file under `~/.config/opencode/prompts/sdd/`
- runtime files/directories declared in `policy/managed-assets.json` exist on disk
- if `default_profile` exists, the base family keeps the correct `model` and `variant`
- if `profiles` exists, each declared profile keeps the correct `model` and `variant`
- one fresh-context reviewer/subagent pass checked the changed maintainer artifacts and final summary for workflow consistency

## Local overlay config

The canonical per-machine config lives outside the repo at `~/.config/gentle-ai-custom/opencode-local-config.json`.

Operational rules:

- `upstream_repo_path` takes precedence over `GENTLE_AI_CUSTOM_UPSTREAM_REPO`, and both take precedence over `../gentle-ai`
- `opencode_config_path` is optional; when omitted, the default remains `~/.config/opencode/opencode.json`
- `agent_overrides` manages only explicit built-in agent assignments such as `general` or `explore`
- `default_profile` manages only the base `gentle-orchestrator` family plus unsuffixed SDD phases
- `profiles` manages only named SDD families (`sdd-orchestrator-<name>` plus phases)
- existing undeclared profiles remain untouched and are reported as unmanaged

## Maintenance checklist

- [ ] `maintenance-intent.md` still reflects what to keep and depure
- [ ] `managed-assets.json` still aligns with `assets/upstream/` and `assets/owned/`
- [ ] `upstream-state.json` still points to the last upstream boundary that was actually maintained
- [ ] `apply-gentle-ai-custom` still reinstalls prompt refs from owned assets
- [ ] the public shell and PowerShell entrypoints remain equivalent
- [ ] SDD profile assignments remain local and did not leak back into versioned policy
- [ ] upstream resolution still respects: local config -> env -> fallback `../gentle-ai`

## References

- `README.md`
- `AGENTS.md`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- `policy/maintenance-intent.md`
- `policy/gentle-ai-policy.json`
- `policy/managed-assets.json`
- `state/upstream-state.json`
- `logs/update-log.md`
