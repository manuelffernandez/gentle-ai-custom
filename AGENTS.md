# AGENTS.md — gentle-ai-custom

Operating instructions for AI agents (Claude, OpenCode, Codex, and equivalents) working in this repository.

## Agent-readable Markdown standard

For Markdown whose primary audience is agents or LLMs (`AGENTS.md`, `SKILL.md`, maintainer docs, intent/policy docs, and similar runtime instruction files), follow the overlay-owned runtime `Agent-Readable Markdown Standard` defined in `overlay/gentle-ai/assets/owned/opencode/AGENTS.md` and installed to `~/.config/opencode/AGENTS.md`.

Repo-specific additions in this file stay local, preserve the runtime standard, and only add clarifying constraints needed for this repository.

---

## Operating rules

### 1. Self-update this document

Any change that affects the repository structure, command flow, overlay behavior, or skill layout **must be accompanied by the corresponding update to this `AGENTS.md`**.

This includes:
- new or moved skills (`shared/skills/...`, `.agents/skills/...`)
- new or renamed scripts
- changes to installation or maintenance flow
- changes to overlay policy/maintenance/docs

Repo-specific additions that stay local to this file:

- default to English unless a file has an explicit reason to be another language
- keep one source of truth per topic and link to it instead of duplicating detailed rules
- separate intent, policy, procedure, and runtime behavior into the correct artifacts
- write for diffability: short sections, stable headings, and low-noise edits

### 2. Exact parity between paired automation scripts

This repo now has four script pairs:

- `apply-gentle-ai-custom.sh` / `apply-gentle-ai-custom.ps1` → canonical user-facing entrypoints
- `audit-gentle-ai-upstream.sh` / `audit-gentle-ai-upstream.ps1` → canonical maintainer-facing upstream audit entrypoints
- `sync-gentle-ai-upstream-assets.sh` / `sync-gentle-ai-upstream-assets.ps1` → canonical maintainer-facing upstream snapshot sync entrypoints
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` / `.ps1` → internal Gentle AI depuration helpers

All three pairs are thin wrappers over the shared Go CLI in `cmd/gentle-ai-overlay` + `internal/overlay`.

If one side changes, the paired script must be updated in the same commit.

**Canonical entrypoint parity items:**
- target parsing and usage/help behavior
- `--verbose` flag support and file-level change reporting
- installation of custom skills/wrappers
- invocation of the Gentle AI overlay helper
- advisory version preflight against `overlay/gentle-ai/state/upstream-state.json` before any apply writes

**Maintainer audit parity items:**
- invocation of the shared upstream audit logic
- invocation of the shared upstream snapshot-sync logic
- upstream prompt + metadata alignment checks
- profile-generation invariant checks
- brief human-readable drift summary output when drift is detected (especially for base prompt drift)
- fail/success criteria and actionable error output

**Overlay helper parity items:**
- keep/prune skill policy
- skill prune scope derived from the selected registered CLI targets only
- `agent_overrides` application
- owned runtime asset installation rules
- generated prompt output path and naming
- fail-closed behavior when owned asset/runtime verification breaks

Never leave either pair in a divergent state.

### 3. Update documentation on functional changes

Any modification that affects operability must be reflected in documentation:

- `README.md` — primary human guide and entrypoint usage
- `AGENTS.md` — repository structure, policies, and agent workflow
- `overlay/gentle-ai/README.md` — local orientation for the overlay asset directory
- `overlay/gentle-ai/maintenance.md` — maintenance guide
- affected `SKILL.md` files — source of truth for runtime agent behavior

For maintainer audit reports, keep the wording simple and consistent:
- report rows must use `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, and `Follow-up`
- `Upstream change` must be a concise human summary of the upstream delta and why it matters; do not use a file-path list as the main content
- `Follow-up` is optional and should be empty when no further action is needed
- `Scope` is `Managed` or `Unmanaged`
- `Impact` is `Behavioral`, `Runtime`, or `Housekeeping`
- `Decision` is `Adquirir`, `Sanitizar`, or `Ignorar`
- `Runtime` includes maintained-target wiring, install, config, or materialization changes even when agent behavior stays the same
- `Housekeeping` covers irrelevant docs, unrelated agents, or internal fixes with no maintained-target effect
- do not use `descartar` as the main report label

**Language and Tone Rule for README.md:**
`README.md` MUST be written in Spanish. Use simple, clear, and accessible language to ensure it is easily understood, rather than reading like a highly specialized technical manual. Avoid overly complex words where possible. If technical details are necessary, include them but always aim to explain them in a simple and straightforward way.

### 4. Use the overlay update log only for closed upstream-maintenance events (MANDATORY)

`overlay/gentle-ai/logs/update-log.md` is a **single-file maintenance ledger**.

- Keep using that exact file.
- Do **not** create month/day/task subfolders or sibling log files unless the user explicitly changes this policy.
- Do **not** treat the log as a mirror of every repo change; Git history already carries implementation-level granularity.

Update `overlay/gentle-ai/logs/update-log.md` **ONLY** when a **closed maintenance event** materially affects how this repo aligns with, audits, adopts, applies, recovers, or verifies Gentle AI upstream.

Eligible events:

- **Upstream audit closure**
  - a reviewed upstream range was closed
  - drift was assessed and the result was recorded
  - this includes "no overlay changes required" outcomes when the audit itself was completed
- **Upstream adoption / rejection / postponement decisions**
  - the repo accepted a new upstream baseline
  - the repo intentionally rejected or deferred an upstream change
  - the adoption path (`gentle-ai sync` vs reinstall) changed or was explicitly decided
- **Maintenance-contract changes**
  - `maintenance-intent.md` changed semantically
  - keep/prune policy changed
  - repo-owned orchestrator/runtime behavior changed
  - maintainer workflow requirements changed
  - the criteria for audit/apply/recovery/verification changed
- **Tooling/runtime changes with upstream-maintenance impact**
  - scripts, shared Go runtime, snapshots, metadata, or verification logic changed in a way that materially affects the ability to audit, apply, recover, or verify the overlay against upstream
- **Maintenance incidents and recoveries**
  - broken state, topology drift, owned-asset/runtime breakage, verification failures, or similar incidents were investigated and closed

Forbidden events — **do NOT** update the log for:

- README/docs wording, structure, or pedagogy tweaks that do not change the maintenance contract
- repo-local refactors, features, or new skills with no upstream-maintenance impact
- cosmetic cleanup, copy edits, examples, formatting-only changes, or readability passes
- intermediate iterations on the same maintenance event
- changes whose useful traceability is already sufficiently covered by Git history because no maintenance decision, audit closure, or incident closure happened

Anti-noise rule:

- **One closed maintenance event = one consolidated log entry.**
- Do **not** add one entry per file, commit, or micro-iteration.
- If the same event spans multiple commits or sessions, wait until the event is closed and then write **one** consolidated entry.

Each entry MUST include these fields. `Follow-up` may be omitted or left empty when no extra action is needed:

- `Date`
- `Title`
- `Type` (`audit`, `adoption`, `policy-change`, `tooling-change`, or `incident`)
- `Upstream scope/range` when applicable
- `Decision`
- `Why it mattered`
- `Affected artifacts`
- `Verification`
- `Follow-up` (optional; dejalo vacío cuando no haga falta ninguna acción extra)

Rule 3 = live state. Rule 4 = maintenance decision ledger.

- Rule 3 keeps the live documentation aligned with current behavior.
- Rule 4 records only the closed upstream-maintenance events that Git history alone would not communicate clearly enough.

If Git history is sufficient and no eligible maintenance event was closed, the log must **not** be updated.
If an eligible maintenance event changes the live docs or maintainer contract, both rule 3 and rule 4 apply together.

### 5. Locate a skill before editing it (MANDATORY)

Before reading or modifying any skill file, map every location where that skill exists:

| Location | Path pattern | Scope |
|---|---|---|
| Project agents | `.agents/skills/<name>/SKILL.md` | This repo only |
| Project shared | `shared/skills/<name>/SKILL.md` | Canonical source for installed skills |
| Global runtime | `~/.config/opencode/skills/<name>/SKILL.md` (and equivalents for other agents) | Installed copy, runtime-only |

**Resolution rules:**

- **Found in one place only** → proceed with that file without asking.
- **Found in multiple places** → stop and ask the user which copies to update: one specific location, a subset, or all of them. Do not assume.

The most common pattern in this repo is `shared/skills/<name>/` + global runtime coexisting. `.agents/skills/` skills are project-exclusive and never appear in the global runtime.

### 6. Propagate canonical skill changes through the installer (MANDATORY)

`shared/skills/<name>/SKILL.md` is the canonical source for any skill that gets installed globally. The global runtime copy is a derived artifact — never the source of truth.

**Rules:**

- Modify `shared/skills/<name>/SKILL.md` (the canonical source), never the global copy directly.
- After modifying the canonical source, propagate by running:
  ```bash
  bash apply-gentle-ai-custom.sh opencode
  ```
  This ensures the global copy is written by the installer, which may apply transformations, wrappers, or path rewrites that a manual copy would miss — the exact source of drift if bypassed.
- If the user declines to run the installer in the same session, surface a reminder that the global copy is now stale.
- Skills under `.agents/skills/` are project-exclusive. They are never propagated and require no installer step.

---

## Two-layer agentic configuration

This repo contains two conceptually distinct layers of agentic configuration. Conflating them is the most common agent error. Resolve ambiguity by asking which layer is intended before acting.

### Layer 1 — Repo-local runtime (for working IN this repo)

Files that take effect when OpenCode is open in this project. Never installed globally. No apply script step required.

| Path | Purpose |
|---|---|
| `.agents/skills/<name>/SKILL.md` | Skills loaded automatically when OpenCode is in this project |
| `.opencode/commands/<name>.md` | Slash commands available only in this project |

Current repo-local assets:
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — maintainer workflow skill
- `.opencode/commands/maintain.md` — `/maintain` slash command

### Layer 2 — Installer output (what this repo installs globally)

Files that `apply-gentle-ai-custom.sh` installs into the global OpenCode runtime. The files in this repo are the canonical sources; the global copies are derived artifacts.

| Path | Purpose |
|---|---|
| `shared/skills/<name>/SKILL.md` | Canonical source for globally-installed skills |
| `shared/commands/<name>-body.md` | Command body rendered and installed to `~/.config/opencode/commands/` |

### Disambiguation rule

- User is modifying something to use **while working in this repo** → Layer 1 (`.agents/`, `.opencode/`)
- User is modifying something that gets **installed into other projects** → Layer 2 (`shared/`)
- When unclear: **stop and ask** — do not assume Layer 2 because `shared/` is larger or more prominent in the codebase

---

## Key paths

### Layer 1 — Repo-local

- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — project-only maintainer workflow skill (loaded automatically when OpenCode is in this project).
- `.opencode/commands/maintain.md` — `/maintain` slash command, available only in this project.

### Layer 2 — Installer output (canonical sources)

- `shared/skills/` — canonical repo-owned skills that the apply script installs globally.
- `shared/commands/` — reusable prompt bodies for commit/PR wrapper commands rendered by the apply script.
- `overlay/gentle-ai/assets/` — canonical owned-assets tree: approved upstream snapshots plus repo-owned SDD/runtime assets.
- `overlay/gentle-ai/assets/upstream/opencode/` — approved upstream-equivalent OpenCode snapshot set: `persona-gentleman.md`, `engram-protocol.md` (mirrors `internal/assets/claude/engram-protocol.md` for the owned AGENTS build), the orchestrator prompt, overlay JSONs, plugins, retained `commands/` snapshots, and selected skill snapshots.
- `overlay/gentle-ai/assets/owned/opencode/AGENTS.md` — repo-owned runtime copy installed to `~/.config/opencode/AGENTS.md` with local overlay semantics.
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — keep/prune policy, OpenCode paths, and runtime overrides.
- `overlay/gentle-ai/policy/managed-assets.json` — canonical manifest for managed upstream/owned assets and repo-owned skill install intent.
- `overlay/gentle-ai/policy/maintenance-intent.md` — semantic source of truth for what to preserve, depure, and enforce in repo-owned orchestrator behavior, including the scoped inline-exploration override for user-requested 4+ file reads.
- `overlay/gentle-ai/state/upstream-state.json` — last maintained upstream boundary.
- `overlay/gentle-ai/maintenance.md` — centralized human maintenance guide.
- `overlay/gentle-ai/owned-assets-refactor.md` — architecture reference for the repo-owned managed-assets runtime model.
- `overlay/gentle-ai/logs/update-log.md` — high-signal ledger of closed upstream-maintenance decisions and incidents.
- `~/.config/gentle-ai-custom/opencode-local-config.json` — canonical per-machine OpenCode overlay config.
- `cmd/gentle-ai-overlay/main.go` — shared Go CLI entrypoint for apply/audit commands.
- `sync-gentle-ai-upstream-assets.sh` / `.ps1` — canonical public upstream snapshot sync entrypoints.
- `internal/overlay/` — implementation of overlay apply, audit, policy, profiles, snapshots, and verification.
- `apply-gentle-ai-custom.sh` / `.ps1` — canonical public entrypoints.
- `audit-gentle-ai-upstream.sh` / `.ps1` — canonical public upstream audit entrypoints.

---

## Repo meaning

This repository is now a **unified custom layer** on top of Gentle AI.

It does two classes of work:

1. **Custom overlays owned by the user**
   - `commit-planner`
   - `pr-finalizer`
   - generated wrappers/commands per target

2. **Maintenance/depuration of upstream Gentle AI behavior**
   - audit the upstream `gentle-orchestrator` asset before sync/reinstall work
   - prune unwanted workflow skills
   - set runtime model overrides for built-in OpenCode agents
	- refresh approved upstream asset snapshots for diff/review
	- maintain a repo-owned SDD/orchestrator runtime layer

The repo now uses a **repo-owned managed-assets** model:

- `overlay/gentle-ai/assets/upstream/` will hold approved upstream behavior assets for diff/audit review, including the OpenCode overlay JSONs/plugins and only the approved `internal/assets/claude/engram-protocol.md` snapshot tracked by the managed-assets manifest
- `overlay/gentle-ai/assets/owned/` will hold repo-owned SDD/runtime behavior assets applied to the local runtime
- `shared/skills/` remains the canonical source for portable repo-owned skills and is NOT folded into the overlay asset tree

This repo does **not** mirror the upstream codebase. Upstream is treated as input only and is resolved in this order: `~/.config/gentle-ai-custom/opencode-local-config.json` (`upstream_repo_path`) -> `$GENTLE_AI_CUSTOM_UPSTREAM_REPO` -> `../gentle-ai` fallback relative to this repo.

The maintenance model is intentionally split into:

- `maintenance-intent.md` → semantic intent and repo-owned orchestrator behavior goals
- `gentle-ai-policy.json` → runtime policy
- `upstream-state.json` → last maintained upstream boundary
- `update-log.md` → closed maintenance-event record

---

## Repo-owned skills

`SKILL.md` is always the source of truth for behavior details.

**Layer 1 — Project-local (never installed globally):**
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — maintainer workflow, only active when OpenCode is in this project

**Layer 2 — Globally installed via apply script:**
- `shared/skills/code-design/SKILL.md`
- `shared/skills/commit-planner/SKILL.md`
- `shared/skills/judgment-retrospective/SKILL.md`
- `shared/skills/package-security/SKILL.md`
- `shared/skills/pr-finalizer/SKILL.md`

**Overlay-owned runtime hooks installed via apply script:**
- `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/SKILL.md` — runtime Judgment Day hook that auto-runs the retrospective skill

---

## Overlay policy baseline

The `Keep` list below is the retained upstream-skill baseline only. Repo-owned shared skills and overlay-owned runtime hooks are tracked separately in the sections above, so do not infer the install surface from this list alone.

Keep:
- `_shared`
- `cognitive-doc-design`
- `comment-writer`
- `go-testing`
- `judgment-day`
- `sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard`
- `skill-registry`, `skill-creator`, `skill-improver`

Prune:
- `branch-pr`
- `chained-pr`
- `issue-creation`
- `hermes-ephemeral-delegation`
- `work-unit-commits`

Built-in OpenCode agent overrides:
- `general` → `openai/gpt-5.4` / `high`
- `explore` → `google-vertex/gemini-3.1-pro-preview` / `high`

SDD profile orchestrators (`sdd-orchestrator-<name>` and `sdd-<phase>-<name>`) are **NOT** baked into the versioned policy. They are reconciled from a per-machine local config — see `## Local OpenCode overlay config` below. The versioned policy keeps only portable baseline keys (`gentle-orchestrator`) so the repo never carries machine-specific model/variant choices.

The maintainer must not infer evolving user intent only from the JSON policy. Intent changes belong first in `maintenance-intent.md`, then in policy/runtime artifacts if the user approves them.

---

## Local OpenCode overlay config

The canonical per-machine config lives OUTSIDE the repo at:

```
~/.config/gentle-ai-custom/opencode-local-config.json
```

Supported top-level fields:

- `version` — required, must be `1`
- `upstream_repo_path` — optional absolute or `~`-expanded path to the upstream `gentle-ai` clone
- `opencode_config_path` — optional override for the OpenCode config file; default remains `~/.config/opencode/opencode.json`
- `agent_overrides` — optional array of explicit agent-key `model` / `variant` assignments (for example `general`, `explore`)
- `default_profile` — optional base SDD family assignment for `gentle-orchestrator` plus the unsuffixed 10 `sdd-<phase>` agents
- `profiles` — optional array of named SDD profile families (`sdd-orchestrator-<name>` plus the 10 suffixed SDD phases)

Repo-level rules:

- This file is local-only and is never versioned in this repo.
- The versioned policy must not carry per-profile orchestrator or phase `model` / `variant` assignments.
- `agent_overrides` means ONLY explicit agent-key assignments; it does not manage SDD profile families.
- `default_profile` means the base `gentle-orchestrator` family only; it does not replace explicit built-in agent overrides.
- `profiles` means ONLY named grouped SDD profile families; it does not replace explicit built-in agent overrides.
- If `agent_overrides` is omitted, the helper applies no explicit built-in agent model overrides.
- If `default_profile` is omitted, the helper leaves the base `gentle-orchestrator` family untouched.
- If `profiles` is omitted, the helper applies no named SDD profiles.
- Managed profiles create or update `sdd-orchestrator-<name>` plus the 10 `sdd-<phase>-<name>` agents in `opencode.json`.
- Profiles present in `opencode.json` but not declared in the active config source are left untouched and surfaced as `WARNING - unmanaged SDD profiles left untouched`. Nothing is deleted automatically.
- This local config governs upstream path selection, optional OpenCode config path selection, agent model assignments, and named profile assignments. Runtime prompt materialization comes from repo-owned assets installed by `apply-gentle-ai-custom`.

Detailed schema, validation behavior, and recovery guidance belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Orchestrator invariants

- `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md` is the canonical runtime source for the base orchestrator behavior.
- `gentle-ai sync` or reinstall can overwrite runtime prompt refs, so re-applying the overlay remains mandatory after upstream runtime refreshes.
- `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md` is the approved upstream audit baseline.
- Generated or installed files under `~/.config/opencode/prompts/sdd/orchestrators/` are deployment targets, not the source of truth.

Detailed audit, recovery, and apply procedure belongs in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Update flow

This section stays intentionally short: it is the human maintainer's quick path, not the full operating procedure.

Canonical order:

1. update the `gentle-ai` binary
2. `git pull` in the resolved upstream `gentle-ai` repo (default: `../gentle-ai` relative to this repo)
3. open `gentle-ai-custom` and run the maintainer audit
4. read `Summary:` and, if present, `Drift summary:`; convert the audit into a concise decision summary that states what is new upstream, the `Scope` / `Impact` classification, the `Decision` (`Adquirir`, `Sanitizar`, or `Ignorar`), whether repo sync is required, why, and the runtime refresh recommendation
5. STOP for explicit user approval before any repo mutation, upstream-boundary advance, upstream-asset sync, or runtime refresh
6. if a new upstream boundary was accepted, run `bash sync-gentle-ai-upstream-assets.sh` as the repo sync step
7. run the correct upstream refresh path (`gentle-ai sync` or full reinstall) as a separate later runtime step
8. re-apply the overlay with `apply-gentle-ai-custom.sh opencode` or `apply-gentle-ai-custom.sh all`
9. run one fresh-context consistency review over the maintainer changes, then return a closing summary of what was `Adquirir`, `Sanitizar`, or `Ignorar`, and why

Current audit discovery model:

- `audit-gentle-ai-upstream` uses `last_maintained_commit` from `overlay/gentle-ai/state/upstream-state.json`
- it discovers changed upstream files via `git diff --name-status --find-renames <last_maintained_commit>..HEAD`
- it filters that drift through `overlay/gentle-ai/policy/managed-assets.json`
- it still runs structural invariant checks for upstream integration mechanics beyond markdown assets

Canonical commands:

```bash
bash audit-gentle-ai-upstream.sh
bash sync-gentle-ai-upstream-assets.sh
bash apply-gentle-ai-custom.sh opencode
bash apply-gentle-ai-custom.sh all
```

Adoption rule:

- If the upstream delta is overlay-relevant and the maintained runtime target stays effectively compatible, use `gentle-ai sync` and then re-apply the overlay.
- Do NOT recommend a full reinstall only because upstream added support for a new agent or platform that does not affect the runtime target/materialized state this repo actually maintains.
- Recommend a full reinstall when adopted upstream changes affect topology, presets, or materialization for the maintained runtime target, or when `gentle-ai sync` cannot materialize the required state.
- Keep rejecting upstream changes that reintroduce `chained-pr` governance or review-workload gates into the repo-owned orchestrator behavior unless the user explicitly changes maintenance intent.

Operational reminders:

- `gentle-ai sync` rewrites runtime prompt refs and restores pruned skills, so re-apply is mandatory afterward.
- `apply-gentle-ai-custom.sh opencode` is the minimum OpenCode refresh; `all` is equivalent — `opencode` is the only registered agent.
- The maintainer workflow is approval-gated: audit first, decision summary second, mutations only after explicit user approval.

Detailed triage, decision gates, drift interpretation, recovery, and post-state verification belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Runtime caveat

If the scripts update `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new orchestrator prompt takes effect.

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. Suggested scope examples: `overlay`, `maintainer-skill`, `custom-layer`, `commit-planner`, `pr-finalizer`.
