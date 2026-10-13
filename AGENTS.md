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

All four pairs are thin wrappers over the shared Go CLI in `cmd/gentle-ai-overlay` + `internal/overlay`.

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
- synchronous release-parity activation gate before broad repository exploration or repository-reading/drift-analysis delegation
- upstream prompt + metadata alignment checks
- current v4 orchestrator/routing injection, remote-authorization projection, and actor-topology checks
- brief human-readable drift summary output when drift is detected (especially for base prompt drift)
- fail/success criteria and actionable error output

**Overlay helper parity items:**
- keep/prune skill policy
- skill prune scope derived from the selected registered CLI targets only
- OpenCode `permission.skill` denies derived from pruned names without a repo-owned same-name replacement, applied globally and to every configured agent without touching externally owned skill copies
- `agent_overrides` application
- global Judgment Day/review model assignments and explicit inheritance reconciliation
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
- one report row represents one atomic decision, and `Decision` applies exactly to its `Upstream change`
- distinct local remediation belongs in `Follow-up` or a separate row; runtime sanitization always uses a separate row
- every required summary adjustment must map to an explicit row
- never merge rows when `Scope`, `Impact`, or `Decision` differ
- reconcile the summary, rows, and final recommendation before presenting them
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
  - upstream/runtime ownership or local patch behavior changed
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
- Never claim the global copy is updated until the installer has run successfully. If the user declines to run it in the same session, say the global copy remains stale.
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
- `overlay/gentle-ai/assets/` — approved upstream audit snapshots plus true repo-owned runtime assets.
- `overlay/gentle-ai/assets/upstream/opencode/` — approved upstream-equivalent OpenCode snapshots for audit and diff review. The maintained repository boundary is v4.0.0 at `ff77164d4f56f1665b22fb6fac51c2ccbb769400`; the authorized sync reported 11 new, 13 changed, 31 unchanged, and 35 deleted snapshots across 19 directories, and the post-sync audit passed with no drift and all structural flags `ok`. This verifies repository snapshots only, not either materialized runtime. RDD is maintainer-approved as active and functional, while effective mode remains machine-owned.
- All five OpenCode `plugins-v2` source files are audit/sync-only and are not custom apply sources. OpenCode 1.18.32 uses the V1 plugin path; supporting V2 materialization requires additional custom apply schema work, so do not claim V2 runtime support.
- `overlay/gentle-ai/assets/owned/opencode/AGENTS.md` — full-file source of truth for repo-owned global agent behavior installed to `~/.config/opencode/AGENTS.md`. Its inherited persona and OpenCode full Engram sections must remain semantically equal to the approved `persona-gentleman.md` and `engram-protocol.md` snapshots; its six overlay-only sections cover implementation scope, plain-language communication, Agent-Readable Markdown Standard, Credential-File Safety, Gemini override, and Notion Project Knowledge.
- `overlay/gentle-ai/patches/opencode/judgment-retrospective.json` — exact local delta applied to a compatible upstream-installed Judgment Day skill.
- `overlay/gentle-ai/patches/codex/` — exact Codex-only deltas for the local instructions block and compatible upstream Judgment Day/work-unit skill anchors. Codex preserves upstream `AGENTS.md`, skills, Engram, `config.toml`, and model profiles outside those deltas; the audit does not verify an installed Codex runtime.
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — OpenCode keep/prune policy and Codex upstream-preservation/command-surface constraints.
- `overlay/gentle-ai/policy/managed-assets.json` — canonical manifest for managed upstream/owned assets and repo-owned skill install intent.
- `overlay/gentle-ai/policy/maintenance-intent.md` — semantic source of truth for the upstream-first baseline, the small local customization layer to preserve, and the old local patches that must not be carried forward by default.
- `overlay/gentle-ai/patches/codex/` — Codex-only local instruction and exact skill deltas. Codex's upstream-materialized `AGENTS.md` and other assets are preserved outside declared deltas.
- The v3.7.0 explicit partial-archive acceptance is historical policy evidence only; v4's maintained target is ODD without SDD.
- `overlay/gentle-ai/state/upstream-state.json` — last maintained upstream boundary.
- `overlay/gentle-ai/maintenance.md` — centralized human maintenance guide.
- `overlay/gentle-ai/owned-assets-refactor.md` — architecture reference for the repo-owned managed-assets runtime model.
- `overlay/gentle-ai/logs/update-log.md` — high-signal ledger of closed upstream-maintenance decisions and incidents.
- `~/.config/gentle-ai-custom/opencode-local-config.json` — canonical per-machine OpenCode overlay config.
- `cmd/gentle-ai-overlay/main.go` — shared Go CLI entrypoint for apply/audit commands.
- `sync-gentle-ai-upstream-assets.sh` / `.ps1` — canonical public upstream snapshot sync entrypoints.
- `internal/overlay/` — implementation of overlay apply, audit, policy, global actor assignments, snapshots, and verification.
- `apply-gentle-ai-custom.sh` / `.ps1` — canonical public entrypoints.
- `audit-gentle-ai-upstream.sh` / `.ps1` — canonical public upstream audit entrypoints.

---

## Repo meaning

This repository is now a **unified custom layer** on top of Gentle AI.

It does two classes of work:

1. **Custom overlays owned by the user**
   - `commit-planner` as the explicit user-invoked operational layer for local commits
   - sanitized repo-owned `branch-pr` as the preferred final PR workflow; explicit `/pr-create` and `/pr-regenerate` invocation makes generated title/body audit-visible rather than a content-approval gate, while push authorization and unresolved required metadata, identity, or chain context remain blocking
   - sanitized repo-owned `chained-pr` runtime copy for explicit GitHub repo scoping and plan-only automatic activation until the user approves branch/PR mutation
   - `pr-finalizer` as temporary deprecated compatibility during the migration
   - generated wrappers/commands per target

2. **Maintenance/depuration of upstream Gentle AI behavior**
   - audit the upstream `gentle-orchestrator` asset before sync/reinstall work
   - prune only explicitly rejected upstream skills that stay outside the maintained runtime target
   - keep approved upstream `work-unit-commits`, `chained-pr`, `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` snapshots for audit/sync when they do not conflict with the local layer; the last three are not repo-owned install sources or forced runtime selections
   - deny pruned skills without a repo-owned same-name replacement through OpenCode permissions so externally discovered copies stay unavailable by default without changing other runtimes' copies
   - keep the upstream `branch-pr` copy pruned because its mandatory issue-first rule conflicts with the local PR workflow, then reinstall the repo-owned sanitized `branch-pr`
   - set runtime model overrides for built-in OpenCode agents
	- refresh approved upstream asset snapshots for diff/review
	- preserve the upstream-materialized ODD/review runtime layer and apply only explicit compatible local deltas

The repo now uses a **repo-owned managed-assets** model:

- `overlay/gentle-ai/assets/upstream/` will hold approved upstream behavior assets for diff/audit review, including the OpenCode overlay JSONs/plugins, the conditionally composed `background-subagents.md` prompt source, and only the approved `internal/assets/engram/protocol.md` snapshot tracked by the managed-assets manifest
- `overlay/gentle-ai/assets/owned/` holds only true repo-owned runtime assets; raw upstream snapshots never become install sources
- focused repository tests compare the inherited persona and OpenCode full Engram sections in the owned `AGENTS.md` with their approved snapshots, so snapshot refreshes cannot silently leave the install source stale
- `overlay/gentle-ai/patches/` holds exact surgical deltas for compatible installed upstream assets
- `shared/skills/` remains the canonical source for portable repo-owned skills and is NOT folded into the overlay asset tree

This repo does **not** mirror the upstream codebase. Upstream is treated as input only and is resolved in this order: `~/.config/gentle-ai-custom/opencode-local-config.json` (`upstream_repo_path`) -> `$GENTLE_AI_CUSTOM_UPSTREAM_REPO` -> `../gentle-ai` fallback relative to this repo.

The maintenance model is intentionally split into:

- `maintenance-intent.md` → semantic intent and upstream/runtime ownership boundaries
- `gentle-ai-policy.json` → runtime policy
- `upstream-state.json` → last maintained upstream boundary
- `update-log.md` → closed maintenance-event record

Approved upstream runtime assets use replacement-first adoption: upstream Gentle AI materializes the finished artifact, custom apply preserves it, and only explicit compatible deltas from `maintenance-intent.md` are applied.

---

## Repo-owned skills

`SKILL.md` is always the source of truth for behavior details.

**Layer 1 — Project-local (never installed globally):**
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — maintainer workflow, only active when OpenCode is in this project

**Layer 2 — Globally installed via apply script:**
- `shared/skills/branch-pr/SKILL.md`
- `shared/skills/chained-pr/SKILL.md`
- `shared/skills/code-design/SKILL.md`
- `shared/skills/commit-planner/SKILL.md`
- `shared/skills/judgment-retrospective/SKILL.md`
- `shared/skills/notion-project-knowledge/SKILL.md`
- `shared/skills/package-security/SKILL.md`
- `shared/skills/pr-finalizer/SKILL.md`

The same eight `shared/skills/` sources are installed under `~/.codex/skills/` when `codex` or `all` is selected. Codex apply does not prune upstream skills or render command wrappers.

`package-security` is the canonical source for package-manager guidance. It prohibits credential-file inspection and distinguishes npm's advisory-by-default reviewed script policy from pnpm's version-gated build approvals; use its local templates and verify exact CLI feature support rather than assuming a major version guarantees it.

`notion-project-knowledge` follows current Notion guides for retrieval and parent-agent preservation of project knowledge and personal activity evidence. Its activation is a separate marker-bounded block in the owned OpenCode `AGENTS.md` and the Codex local-instructions patch. Codex replacement zero remains unchanged for legacy-block migration; a second sequential replacement adds the activation after the local-instructions closing marker. Runtime installation and Notion operations remain separately scoped; repository proof does not verify installed copies.

**Overlay-owned runtime hook:**
- `overlay/gentle-ai/patches/opencode/judgment-retrospective.json` — exact-anchor patch that adds the terminal retrospective handoff to a compatible upstream-installed Judgment Day skill

---

## Overlay policy baseline

The `Keep` list below is the retained upstream-skill baseline only. Repo-owned shared skills and overlay-owned runtime hooks are tracked separately in the sections above, so do not infer the install surface from this list alone.

`chained-pr` is dual-tracked on purpose: its approved upstream snapshot stays retained for audit/sync, while `shared/skills/` provides the sanitized runtime copy that apply installs. `issue-creation` and legacy SDD skill names stay pruned from the default OpenCode runtime surface. No apply has run as part of the v4 snapshot adoption; runtime files and user-owned SDD settings were not changed.

Apply reports only skills actually removed during pruning; already-absent names are silent. SDD pruning and applicable OpenCode skill permission denies remain persistent protective policy, including when no SDD skills are currently installed.

Keep:
- `_shared`
- `chained-pr`
- `cognitive-doc-design`
- `comment-writer`
- `gentle-ai-bench`
- `go-testing`
- `judgment-day`
- `rdd-defect-workflow`
- `skill-registry`, `skill-creator`, `skill-improver`
- `systemic-issue-triage`
- `work-unit-commits`

Prune:
- `branch-pr`
- `issue-creation`
- `hermes-ephemeral-delegation`
- `sdd-init`, `sdd-explore`, `sdd-research`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard`

Built-in OpenCode agent overrides are machine-local assignments from
`~/.config/gentle-ai-custom/opencode-local-config.json`; their current values
are intentionally not part of versioned policy.

Custom apply does not generate SDD agents or profiles. Existing local SDD profile settings and SDD agent keys are migration blockers: apply fails before writes and preserves them for manual review. Global Judgment Day/review assignments remain supported independently.

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
- `global_actors` — optional model/variant or explicit-inheritance assignments for global Judgment Day and review actors

Repo-level rules:

- This file is local-only and is never versioned in this repo.
- `agent_overrides` means ONLY explicit non-SDD agent-key assignments.
- Any `default_profile` or `profiles` field is rejected with migration guidance; remove it manually and reassign SDD model choices yourself. Apply never deletes or maps SDD settings to ODD roles.
- SDD-prefixed `agent_overrides` and SDD-prefixed keys already present in `opencode.json` are also rejected before writes and left untouched.
- `global_actors` accepts only the global Judgment Day actors (`jd-judge-a`, `jd-judge-b`, `jd-fix-agent`) and review actors, including `review-validator`.
- `global_actors` assignments accept a model and optional variant or `{"inherit": true}`. Omission preserves the runtime assignment; inheritance removes both `model` and `variant`.
- If `agent_overrides` is omitted, the helper applies no explicit built-in agent model overrides.
- This local config governs upstream path selection, optional OpenCode config path selection, explicit unrelated agent overrides, and global actor assignments. Upstream Gentle AI owns runtime prompt materialization.

Detailed schema, validation behavior, and recovery guidance belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Orchestrator invariants

- The installed `agent.gentle-orchestrator.prompt` is upstream-owned. Custom apply does not clone it into custom profiles or certify its review semantics.
- Upstream-owned plugins and contracts may be absent or opaque to custom apply unless a concrete local patch consumes them. They must never become custom install sources.
- Codex owns its materialized `~/.codex/AGENTS.md`, skills, Engram files, `config.toml`, and model profiles. Codex apply installs only the allowlisted repo-owned skills, leaving unrelated upstream skills untouched; it inserts the local instruction block after the upstream Engram marker and the separate Notion activation immediately after that local block, changes only the `odd-explorer` and `odd-worker` table rows from Luna/high to Sol/high in that same exact-anchor `AGENTS.md` patch, and applies declared exact skill deltas. Other ODD, JD, and review rows, unrelated bytes, and file modes are preserved. It does not require an SDD marker; account model availability is unverified.
- Codex apply must not render OpenCode command wrappers, prune Codex skills, replace the full Codex `AGENTS.md`, or modify Codex config, Engram, SDD prompts, or model profiles. Its declared patch anchors are preflighted before any selected-target write; upstream Codex runtime materialization is not verified by the current audit.
- Codex patch preflight removes only the exact stale second local-instructions block in the orchestrator-to-CodeGraph gap, including its retired no-commit line. It accepts an absent legacy block and an already-applied or unapplied Engram-adjacent current block; duplicate, partial, changed, or relocated blocks fail before writes, including for `all`. This repository-only migration does not run apply or verify the installed Codex runtime.
- Owned-state reconciliation tracks every custom-installed output, including single-file assets without a sync mode. Stale paths returned to upstream or used as surgical patch targets are protected but never recorded as wholly custom-owned.
- `gentle-ai sync` or reinstall restores the upstream runtime surface; re-applying the overlay remains mandatory for local skills, pruning, assignments, and declared surgical patches.
- `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md` is the approved upstream audit baseline.
- `internal/assets/opencode/background-subagents.md` is an upstream-owned prompt source composed conditionally into OpenCode orchestrators. The manifest tracks it as audit/snapshot input, never as a custom runtime install source.
- `internal/assets/generic/remote-authorization-contract.md` remains an upstream-owned audit/snapshot input. Structural audit checks its canonical helper and all maintained OpenCode projection carriers. Apply preflights maintained inline prompts, referenced files, and any existing legacy shared SDD prompt files still declared among owned overlay assets, then removes only the exact recognized upstream or previously relaxed authorization section; malformed, duplicate, or unknown sections and unsafe paths fail before writes. Missing optional legacy files are skipped. This removal does not alter the upstream snapshot or its structural projection audit.
- The repo-owned global `AGENTS.md` intentionally does not carry a remote-operation authorization section. This does not change credential-file safety or unrelated runtime/tool permission guards.
- The v2.9 orchestrator audit follows the bound `composeOpenCodeOrchestratorPrompt` -> `composeOrchestratorPrompt` -> `sddOrchestratorAsset` -> `assets.MustRead` chain and rejects unbound or decoy assignments.
- Historical v3.7.0 `sdd-research` source skill and OpenCode command were output-only with orchestrator-owned persistence; the former local precedence patch is retired.
- Codex reuses the eight canonical shared skill sources, preserves upstream instructions except for its declared local instruction block, separate Notion activation, and two exact ODD table-row deltas, and leaves unrelated skills alone. The upstream audit currently checks OpenCode managed assets, not installed Codex runtime materialization.
- Custom apply must not redirect base or named agents to raw `.overlay.md` prompt copies.
- Historical v2.9.1 repository sync advanced `upstream-state.json` from `be49554794917ae92a6dc9dbfa2eb3db5cf70084` (`v2.9.0`) to `9ec0cf443622f20fe511815ff5a83088c7467bff` (`v2.9.1`) and reported 0 managed file changes, 74 unchanged files, 0 deletions, and 18 synced directories. Its post-sync audit passed. Compatible OpenCode review consent and preserved-prompt migration were acquired upstream-first; Claude Code transcript-derived dispatch authority remained outside the maintained OpenCode target.
- After that historical v2.9.1 repository-only closure, the user reported manually running both `gentle-ai sync` and `bash apply-gentle-ai-custom.sh opencode`. Their original command summaries and exit output were not captured, so no apply counts are asserted. This evidence predates both the historical v3.7.0 boundary and the current v4.0.0 boundary; it does not verify either runtime.
- Independent read-only post-state verification passed 602 selected assertions over runtime assignments, named profiles, prompt parity, task scoping, skill permissions, absence of suffixed review actors, and obsolete prompt references. The installed orchestrator has the `gentle-ai.review-integration.consent/v3` compatibility semantics, including `multiple: false`, `custom: false`, and the rule that missing custom consent alone is not a compatibility block.
- All 7 repo-owned skill sources/assets, all 5 wrapper bodies, and the owned `AGENTS.md` matched their installed targets. `issue-creation` and `hermes-ephemeral-delegation` were absent, `branch-pr` was the repo-owned replacement, and all 3 installed runtime patch targets matched approved upstream snapshots after reversing only their declared local patches.
- Historical v2.9.0 repository evidence remains valid: its authorized sync advanced the maintained boundary from v2.5.0 and refreshed the approved snapshots for the v2.9 OpenCode composition model.
- Historical installed-runtime evidence remains v2.5.0: the human-run full SDD component installation and separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully. Apply updated config, pruned 2 skills, changed and verified 1 owned asset, reported 0 topology warnings, managed 4 profiles, and reported agents created 3, updated 4, unchanged 49, with 0 unmanaged profiles. That v2.5.0 runtime verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. The sanitized repo-owned `branch-pr` was present; conflicting upstream `branch-pr`, `issue-creation`, and Hermes were absent. All 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, overrides, and 4 managed profiles matched canonical or local config. Every profile had the ordered 11-phase topology and the 3 named `sdd-research` actors; review actors remained global with no suffixed copies, and Judgment Day differed only by the retrospective patch.
- Historical v2.5.0 machine-state evidence: managed background intent and effective policy were `on`, with launcher `~/.gentle-ai/bin/opencode`; existing Desktop and plain OpenCode processes remained foreground. At that verification point, `~/.gentle-ai/state.json` still reported `installed_binary_version` as `2.4.0`; no root cause was asserted and external state was not edited.
- Receipt-Driven Development is maintainer-approved as active and functional based on the owner's operational validation in other repositories. This policy decision does not set machine state and does not imply upstream remote control: verify the effective user/machine/clone setting with `gentle-ai review mode status --cwd .`. On this machine, the observed state is global `on` with clone-local unset; the user explicitly excluded this v2.9.1 maintenance candidate from RDD. The correction-required canary in `overlay/gentle-ai/maintenance.md` is an optional regression diagnostic, not an activation blocker; candidate-bound receipts, correction bounds, the user-owned kill switch, and ordinary delivery policy remain mandatory.
- The adopted upstream ODD work-unit commit requirement applies at task close. Repo-owned OpenCode and Codex opt-out patches, plus the OpenCode no-auto-commit instruction, have been removed from source; runtime refresh/apply remains separately authorized and unverified. Task-close commits do not authorize pushes or PRs, which retain their explicit user-authorization gates.
- Historical v3.7.0 `sdd-research` source and command enforced output-only collection, so its local precedence patch was retired. The Judgment Day retrospective patch remains.
- The v3.7.0 snapshot adoption is closed historical evidence: sync advanced the boundary to `6dee8f833aec9e46015759c5065a9035795d9af1` with 36 managed changes (5 new, 31 changed, 43 unchanged, 0 deleted) across 19 directories; its post-sync audit passed. The current v4.0.0 repository boundary is recorded separately above and in `upstream-state.json`; runtime refresh and overlay apply remain separately unauthorized.
- v4 structural audit follows `internal/components/agentguidance/{orchestrator.go,inject.go,remote_authorization.go}` and `internal/opencode/models.go`: it checks the OpenCode orchestrator asset binding, always-loaded routing injection, remote-authorization projection into that routing carrier, and current ODD/JD/review agent sets. It no longer requires deleted SDD profile-generation paths. The 11 new OpenCode orchestrator/agent assets are upstream snapshots only; they do not create new custom runtime sources.
- The separately approved v4 repo snapshot sync deleted 35 snapshots under `assets/upstream/opencode/`: 18 exact SDD files declared in `retired_upstream_snapshots` and 17 stale files pruned inside still-declared synced directories (11 commands, 2 plugins, and 4 `_shared` files). Manifest retirement rejects traversal, overlap with active paths, symlink parents, and directory targets; directory pruning is bounded to declared synced assets. The sync validates release parity, mapped drift, and every structural gate first, and did not mutate runtime files or local user settings.

Detailed audit, recovery, and apply procedure belongs in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Update flow

This section stays intentionally short: it is the human maintainer's quick path, not the full operating procedure.

Canonical order:

1. update the `gentle-ai` binary when appropriate; keeping it unchanged is valid when live sessions depend on it
2. update tags in the resolved upstream `gentle-ai` repo and manually check out the exact release tag to inspect
3. open `gentle-ai-custom`; immediately after loading the maintainer skill/minimal context, run the canonical audit synchronously before broad repository exploration or repository-reading/drift-analysis delegation
4. read `Summary:` and, if present, `Drift summary:`; convert the audit into a concise decision summary that states what is new upstream, the `Scope` / `Impact` classification, the `Decision` (`Adquirir`, `Sanitizar`, or `Ignorar`), whether repo sync is required, why, and the runtime refresh recommendation
5. STOP for explicit user approval before any repo mutation, upstream-boundary advance, or upstream-asset sync
6. if a new upstream boundary was accepted and repo sync was separately approved, run `bash sync-gentle-ai-upstream-assets.sh` as the repo sync step
7. recommend and print exact instructions for `gentle-ai sync` or a full reinstall; do not execute either unless the human explicitly requests that specific runtime action
8. recommend overlay apply when needed; run `apply-gentle-ai-custom.sh opencode`, `codex`, or `all` only with separate explicit authorization
9. run one fresh-context consistency review over the maintainer changes, then return a closing summary of what was `Adquirir`, `Sanitizar`, or `Ignorar`, and why

Current audit discovery model:

- audit and sync default to the installed binary's exact `v<V>` tag. An explicit `--release v<V>` may select a different exact release for both commands without changing the binary or runtime.
- every selected release still requires a fully clean upstream checkout whose `HEAD` equals the peeled commit of that exact `refs/tags/v<V>`
- audit/sync upstream resolution reads only `upstream_repo_path` from the local config, so a new upstream topology can be audited before runtime config changes; runtime config validation still happens before apply writes
- the audit is an activation hard gate: if selected-release verification fails, stop immediately and delegate nothing; `last_maintained_commit` is only the later audit-range boundary
- missing/stale tags, any staged/unstaged/untracked/submodule dirt, or a branch ahead of the release block before analysis/copy/write; the tools never fetch or check out automatically
- the user's description of what was updated is diagnostic only; an explicit release selector must be used to inspect a release other than the installed binary's, and exact tag/HEAD identity is always verified, never inferred from `git describe`, nearest tags, version ordering, or ancestry
- `audit-gentle-ai-upstream` uses `last_maintained_commit` from `overlay/gentle-ai/state/upstream-state.json`
- it discovers changed upstream files via `git diff --name-status --find-renames <last_maintained_commit>..<verified-release-commit>`
- it filters that drift through `overlay/gentle-ai/policy/managed-assets.json`
- it still runs structural invariant checks for upstream integration mechanics beyond markdown assets
- v4 checks the current `agentguidance` orchestrator/routing injection, remote-authorization projection, and ODD/JD/review agent sets structurally
- deleted SDD profile-generator paths and 11-phase profile invariants are not current v4 audit requirements; any unsynced snapshot drift remains separately reported

Canonical commands:

```bash
bash audit-gentle-ai-upstream.sh
bash sync-gentle-ai-upstream-assets.sh
bash apply-gentle-ai-custom.sh opencode
bash apply-gentle-ai-custom.sh codex
bash apply-gentle-ai-custom.sh all
```

Adoption rule:

- If the upstream delta is overlay-relevant and the maintained runtime target stays effectively compatible, recommend `gentle-ai sync` and a later separately authorized overlay apply.
- Do NOT recommend a full reinstall only because upstream added support for a new agent or platform that does not affect the runtime target/materialized state this repo actually maintains.
- Recommend a full reinstall when adopted upstream changes affect topology, presets, or materialization for the maintained runtime target, or when `gentle-ai sync` cannot materialize the required state.
- Compare upstream commit/PR workflow changes against the explicit local customizations in `maintenance-intent.md`; sanitize only when they conflict with that smaller local layer.

Operational reminders:

- successful audit metadata reports the verified binary version, exact tag, and release commit; successful sync stores that exact identity in `upstream-state.json`
- apply keeps its separate advisory installed-binary versus maintained-baseline check and does not require an upstream clone
- `gentle-ai sync` rewrites runtime prompt refs and restores pruned skills, so re-apply is mandatory afterward.
- approval of audit findings, adoption, repo snapshot sync, or overlay edits never authorizes runtime refresh; only an explicit request for that specific runtime action does
- an audit using `--release` is preview evidence only; it does not mean that release was adopted or approve repo snapshot sync. Repo sync still requires separate explicit authorization and must use the selected exact release only after approval.
- runtime-refresh authorization never authorizes overlay apply; apply requires its own explicit approval
- `apply-gentle-ai-custom.sh opencode` and `codex` select one target; `all` selects both. Codex apply is independent of OpenCode permissions, configuration, command wrappers, and pruning.
- The maintainer workflow is approval-gated: audit first, decision summary second, mutations only after explicit user approval.

Detailed triage, decision gates, drift interpretation, recovery, and post-state verification belong in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` and `overlay/gentle-ai/maintenance.md`.

---

## Runtime caveat

If the scripts update `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new orchestrator prompt takes effect.

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. Suggested scope examples: `overlay`, `maintainer-skill`, `custom-layer`, `commit-planner`, `pr-finalizer`.
