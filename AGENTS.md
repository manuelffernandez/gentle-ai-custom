# AGENTS.md — gentle-ai-custom

Operating instructions for AI agents (Claude, OpenCode, Codex, and equivalents) working in this repository.

---

## Operating rules

### 1. Self-update this document

Any change that affects the repository structure, command flow, overlay behavior, or skill layout **must be accompanied by the corresponding update to this `AGENTS.md`**.

This includes:
- new or moved skills (`shared/skills/...`, `.agents/skills/...`)
- new or renamed scripts
- changes to installation or maintenance flow
- changes to overlay policy/maintenance/docs

For files whose primary reader is an agent/LLM (for example `AGENTS.md`, `SKILL.md`, maintenance docs, intent/policy docs), optimize the writing for machine consumption:
- default to English unless a file has an explicit reason to be another language
- prefer explicit paths, headings, bullets, tables, and checklists over narrative prose
- avoid ASCII tree diagrams for repository maps; use short `Key paths` lists instead
- keep one source of truth per topic and link to it instead of duplicating detailed rules
- separate intent, policy, procedure, and runtime behavior into the correct artifacts
- state invariants and stop conditions directly; do not bury them inside examples
- keep examples minimal and only when they materially reduce ambiguity
- write for diffability: short sections, stable headings, and low-noise edits

### 2. Exact parity between paired automation scripts

This repo now has three script pairs:

- `apply-gentle-ai-custom.sh` / `apply-gentle-ai-custom.ps1` → canonical user-facing entrypoints
- `audit-gentle-ai-upstream.sh` / `audit-gentle-ai-upstream.ps1` → canonical maintainer-facing upstream audit entrypoints
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` / `.ps1` → internal Gentle AI depuration helpers

All three pairs are thin wrappers over the shared Go CLI in `cmd/gentle-ai-overlay` + `internal/overlay`.

If one side changes, the paired script must be updated in the same commit.

**Canonical entrypoint parity items:**
- target parsing and usage/help behavior
- `--verbose` flag support and file-level change reporting
- installation of custom skills/wrappers
- invocation of the Gentle AI overlay helper

**Maintainer audit parity items:**
- invocation of the shared upstream audit logic
- upstream prompt + metadata alignment checks
- profile-generation invariant checks
- brief human-readable drift summary output when drift is detected (especially for base prompt drift)
- fail/success criteria and actionable error output

**Overlay helper parity items:**
- keep/prune skill policy
- `agent_overrides` application
- orchestrator snapshot behavior
- orchestrator sanitization rules
- generated prompt output path and naming
- fail-closed behavior when sanitization anchors are missing

Never leave either pair in a divergent state.

### 3. Update documentation on functional changes

Any modification that affects operability must be reflected in documentation:

- `README.md` — primary human guide and entrypoint usage
- `AGENTS.md` — repository structure, policies, and agent workflow
- `overlay/gentle-ai/README.md` — local orientation for the overlay asset directory
- `overlay/gentle-ai/maintenance.md` — maintenance guide
- affected `SKILL.md` files — source of truth for runtime agent behavior

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
  - sanitizer behavior or required anchors changed
  - maintainer workflow requirements changed
  - the criteria for audit/apply/recovery/verification changed
- **Tooling/runtime changes with upstream-maintenance impact**
  - scripts, shared Go runtime, snapshots, metadata, or verification logic changed in a way that materially affects the ability to audit, apply, recover, or verify the overlay against upstream
- **Maintenance incidents and recoveries**
  - broken state, snapshot recovery, topology drift, sanitizer breakage, verification failures, or similar incidents were investigated and closed

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

Each entry MUST include:

- `Date`
- `Title`
- `Type` (`audit`, `adoption`, `policy-change`, `tooling-change`, or `incident`)
- `Upstream scope/range` when applicable
- `Decision`
- `Why it mattered`
- `Affected artifacts`
- `Verification`
- `Follow-up` (optional)

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

## Key paths

- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — project-only maintainer workflow for auditing and re-applying the overlay.
- `shared/skills/` — canonical repo-owned skills that can be installed globally.
- `shared/commands/` — reusable prompt bodies for commit/PR wrapper commands.
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — keep/prune policy, OpenCode paths, and runtime overrides.
- `overlay/gentle-ai/policy/maintenance-intent.md` — semantic source of truth for what to preserve, depure, and remove from orchestrator sanitization.
- `overlay/gentle-ai/state/upstream-state.json` — last maintained upstream boundary.
- `overlay/gentle-ai/maintenance.md` — centralized human maintenance guide.
- `overlay/gentle-ai/logs/update-log.md` — high-signal ledger of closed upstream-maintenance decisions and incidents.
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` — versioned upstream orchestrator baseline and metadata.
- `~/.config/gentle-ai-custom/opencode-local-config.json` — canonical per-machine OpenCode overlay config.
- `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` — legacy profile-only fallback, read only when the canonical config omits `profiles`.
- `cmd/gentle-ai-overlay/main.go` — shared Go CLI entrypoint for apply/audit commands.
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
   - capture inline orchestrators from OpenCode config
   - sanitize PR/budget workflow content
   - emit generated orchestrator prompt files under the OpenCode prompts tree

This repo does **not** mirror the upstream codebase. Upstream is treated as input only and is resolved in this order: `~/.config/gentle-ai-custom/opencode-local-config.json` (`upstream_repo_path`) -> `$GENTLE_AI_CUSTOM_UPSTREAM_REPO` -> `../gentle-ai` fallback relative to this repo.

The maintenance model is intentionally split into:

- `maintenance-intent.md` → semantic intent and orchestrator sanitization goals
- `gentle-ai-policy.json` → runtime policy
- `upstream-state.json` → last maintained upstream boundary
- `update-log.md` → closed maintenance-event record

---

## Repo-owned skills

Canonical source files live here; `SKILL.md` is always the source of truth for behavior details.

- `shared/skills/code-design/SKILL.md`
- `shared/skills/commit-planner/SKILL.md`
- `shared/skills/package-security/SKILL.md`
- `shared/skills/pr-finalizer/SKILL.md`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — project-only maintainer workflow

---

## Overlay policy baseline

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
- `work-unit-commits`

Built-in OpenCode agent overrides:
- `general` → `openai/gpt-5.4` / `high`
- `explore` → `google-vertex/gemini-3.1-pro-preview` / `high`

SDD profile orchestrators (`sdd-orchestrator-<name>` and `sdd-<phase>-<name>`) are **NOT** baked into the versioned policy. They are reconciled from a per-machine local config — see `## Local OpenCode overlay config` below. The versioned policy keeps only portable baseline keys (`gentle-orchestrator`) so the repo never carries machine-specific model/variant choices.

Profile orchestrator snapshots also stay out of the versioned repo. The repo keeps the portable `gentle-orchestrator.last.md` baseline plus `gentle-orchestrator.last.meta.yaml`; the helper keeps operational snapshots per machine under `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.

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
- If `profiles` is omitted, the helper falls back to the legacy `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` file when present.
- If `profiles` is present (including `[]`), the canonical config becomes the source of truth and the legacy file is ignored.
- `default_profile` has no legacy fallback; it exists only in the canonical local config.
- Managed profiles create or update `sdd-orchestrator-<name>` plus the 10 `sdd-<phase>-<name>` agents in `opencode.json`.
- Profiles present in `opencode.json` but not declared in the active config source are left untouched and surfaced as `WARNING - unmanaged SDD profiles left untouched`. Nothing is deleted automatically.
- This local config governs upstream path selection, optional OpenCode config path selection, agent model assignments, and named profile assignments; prompt materialization still comes from `gentle-ai sync`, and the overlay sanitizes that inline upstream content afterward.

Detailed schema, validation behavior, and recovery guidance belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Orchestrator invariants

- OpenCode orchestrators originate as inline upstream prompts and are rewritten to generated overlay prompt files.
- `gentle-orchestrator.last.md` plus `gentle-orchestrator.last.meta.yaml` remain the only versioned upstream orchestrator baseline in the repo.
- Profile orchestrator snapshots are local-only under `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.
- Generated prompt files under `~/.config/opencode/prompts/sdd/orchestrators/` are derived runtime outputs, not the source of truth.
- Do **not** switch back to a static repo-owned prompt file as the operational source of truth.
- If sanitization anchors are missing, or if the materialized `gentle-orchestrator` no longer matches the audited baseline/metadata contract, fail closed and surface the warning.

Detailed audit, recovery, and apply procedure belongs in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Update flow

This section stays intentionally short: it is the human maintainer's quick path, not the full operating procedure.

Canonical order:

1. update the `gentle-ai` binary
2. `git pull` in the resolved upstream `gentle-ai` repo (default: `../gentle-ai` relative to this repo)
3. open `gentle-ai-custom` and run the maintainer audit
4. read `Summary:` and, if present, `Drift summary:`; if the audit reveals overlay-relevant drift, update this repo first
5. run the correct upstream refresh path (`gentle-ai sync` or full reinstall)
6. re-apply the overlay with `apply-gentle-ai-custom.sh opencode` or `apply-gentle-ai-custom.sh all`

Canonical commands:

```bash
bash audit-gentle-ai-upstream.sh
bash apply-gentle-ai-custom.sh opencode
bash apply-gentle-ai-custom.sh all
```

Adoption rule:

- If the upstream delta is overlay-relevant but preserves agent topology, use `gentle-ai sync` and then re-apply the overlay.
- If the upstream delta adds, removes, or renames agents, changes presets/topology, or changes how upstream materializes `opencode.json`, recommend a full reinstall before re-applying the overlay.
- If both kinds of changes exist, topology wins: recommend reinstall.

Operational reminders:

- `gentle-ai sync` resets orchestrator prompts to upstream inline content and restores pruned skills, so re-apply is mandatory afterward.
- `apply-gentle-ai-custom.sh opencode` is the minimum OpenCode refresh; `all` is equivalent — `opencode` is the only registered agent.

Detailed triage, decision gates, drift interpretation, recovery, and post-state verification belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Runtime caveat

If the scripts update `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new orchestrator prompt takes effect.

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. Suggested scope examples: `overlay`, `maintainer-skill`, `custom-layer`, `commit-planner`, `pr-finalizer`.
