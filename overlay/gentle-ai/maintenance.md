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
   - what changed upstream and why it matters
   - `Scope`: `Managed` / `Unmanaged`
   - `Impact`: `Behavioral` / `Runtime` / `Housekeeping`
   - `Decision`: `Adquirir` / `Sanitizar` / `Ignorar`
   - `Why`
   - `Follow-up` (optional)
   - repo sync requirement: whether approval of a new boundary requires `bash sync-gentle-ai-upstream-assets.sh`
   - recommended runtime path: `gentle-ai sync` vs full reinstall
5. STOP for explicit approval before updating this repo, advancing `state/upstream-state.json`, running `bash sync-gentle-ai-upstream-assets.sh`, or refreshing runtime.
6. If a new upstream boundary was approved, run `bash sync-gentle-ai-upstream-assets.sh` as the repo sync step to refresh `overlay/gentle-ai/assets/upstream/...`.
7. Execute the recommended upstream refresh path:
   - `gentle-ai sync` if the maintained runtime target stays effectively compatible
   - full reinstall if adopted changes affect topology, presets, or materialization for the maintained runtime target, or if sync no longer materializes the right state
8. Re-apply the overlay with `bash apply-gentle-ai-custom.sh opencode`.
9. Read `Summary:`, verify the final on-disk state, run one fresh-context consistency review, and return a closing summary of what was actually `Adquirir`, `Sanitizar`, or `Ignorar`, and why.
10. If `~/.config/opencode/opencode.json` changed, restart OpenCode.

## Operating model

| Artifact | Role |
| --- | --- |
| `policy/maintenance-intent.md` | Human intent: what to keep, depure, and protect |
| `policy/gentle-ai-policy.json` | Runtime policy consumed by the Go CLI and wrappers |
| `policy/managed-assets.json` | Canonical map of approved upstream assets and installable owned assets |
| `assets/upstream/` | Approved upstream copies for review/diff, including the OpenCode snapshot inputs and retained assets: `persona-gentleman.md`, `engram-protocol.md` (mirrors `internal/assets/claude/engram-protocol.md`), the orchestrator prompt, overlay JSONs/plugins, retained `commands/` snapshots, and selected skill snapshots |
| `assets/owned/` | Repo-owned assets that `apply` installs into runtime, including `~/.config/opencode/AGENTS.md` with local overlay semantics |
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

Before the maintainer edits repo files or refreshes runtime, turn the audit into an approval gate. Use these report columns:

| Upstream change | Files | Scope | Impact | Decision | Why | Follow-up |
| --- | --- | --- | --- | --- | --- | --- |

- `Upstream change`: concise human summary of the upstream delta and why it matters; do not use a file-path list as the main content.
- `Scope`: `Managed` if the change is inside the overlay/runtime surface this repo maintains; `Unmanaged` if it is outside that surface or not relevant.
- `Impact`: `Behavioral` for agent/orchestrator behavior; `Runtime` for maintained-target wiring, install, config, or materialization; `Housekeeping` for docs, unrelated agents, or internal fixes with no maintained-target effect.
- `Decision`: `Adquirir`, `Sanitizar`, or `Ignorar`.
- `Follow-up`: optional; leave empty when nothing else is needed.
- Do not use `descartar` as the main report label.

No repo mutation happens before that handoff is approved.

### Runtime refresh recommendation

- Prefer `gentle-ai sync` when the adopted change matters to the runtime target this repo actually maintains, but that target's topology, presets, and materialized state remain effectively compatible.
- Do not recommend reinstall only because upstream added support for some other agent or platform outside the maintained runtime target/materialized state.
- Recommend a full reinstall when the adopted change affects topology, presets, or materialization for the maintained runtime target, or when `gentle-ai sync` cannot materialize the required state.
- Keep rejecting upstream attempts to reintroduce `chained-pr`, review-budget, or review-workload governance into the repo-owned orchestrator behavior unless human intent changes first.

`Runtime` includes changes to the maintained target's wiring, install path, config, or materialized files even when the agent behavior itself does not change.

`Housekeeping` covers changes that do not matter for the maintained target, such as irrelevant docs, unrelated agents, or internal fixes with no maintained-target effect.

### `sync-gentle-ai-upstream-assets`

- copies approved upstream assets into `assets/upstream/...`
- advances `state/upstream-state.json` when the new upstream was accepted
- does not touch local runtime under `~/.config/opencode/`
- always runs after approval of a newly accepted upstream boundary; it is the repo sync step, not the runtime refresh step

### `apply-gentle-ai-custom`

- installs repo-owned SDD/runtime assets from `assets/owned/...`
- rewrites `opencode.json` so the base and SDD profiles use those prompt files
- prunes rejected upstream skills only in the selected registered targets
- applies `agent_overrides`
- reconciles `default_profile` and `profiles`
- installs repo-owned skills from `shared/skills/`
- renders custom wrappers from `shared/commands/`

`apply` no longer depends on sanitization, inline prompt capture, local operational snapshots, snapshot-based recovery, or post-apply `AGENTS.md` mutation.

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
- `~/.config/opencode/AGENTS.md` matches the repo-owned asset from `assets/owned/opencode/AGENTS.md`
- the upstream OpenCode snapshot inputs under `assets/upstream/opencode/` include `persona-gentleman.md`, `engram-protocol.md`, the orchestrator prompt, overlay JSONs, plugins, and selected skill snapshots used to materialize that runtime file
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
