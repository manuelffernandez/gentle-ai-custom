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
- changes to overlay policy/runbooks/docs

### 2. Exact parity between paired automation scripts

This repo now has three script pairs:

- `apply-gentle-ai-custom.sh` / `apply-gentle-ai-custom.ps1` → canonical user-facing entrypoints
- `audit-gentle-ai-upstream.sh` / `audit-gentle-ai-upstream.ps1` → canonical maintainer-facing upstream audit entrypoints
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` / `.ps1` → internal Gentle AI depuration helpers

If one side changes, the paired script must be updated in the same commit.

**Canonical entrypoint parity items:**
- target parsing and usage/help behavior
- installation of custom skills/wrappers
- invocation of the Gentle AI overlay helper

**Maintainer audit parity items:**
- invocation of the shared upstream audit logic
- upstream prompt + metadata alignment checks
- profile-generation invariant checks
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
- `overlay/gentle-ai/README.md` — overlay/control-plane behavior
- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` — maintenance procedure
- affected `SKILL.md` files — source of truth for runtime agent behavior

### 4. Append to the overlay update log on every overlay change (MANDATORY)

Any change to overlay assets MUST add an entry to `overlay/gentle-ai/logs/update-log.md` in the same commit (or commit chain).

"Overlay assets" means any of:

- `overlay/gentle-ai/**` (policy, state, runbooks, scripts, snapshots)
- `apply-gentle-ai-custom.sh` / `.ps1`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- this file (`AGENTS.md`)
- `README.md` when it documents overlay behavior

Each log entry MUST include:

- date and short title
- WHAT changed (one bullet per affected file or coherent area)
- WHY the change was needed (discovery, bug, intent shift, upstream change)
- relevant verification performed (manual test, idempotency check, etc.)

Rule 3 = live state. Rule 4 = decision history. Both deliverables are required for changes that touch overlapping files (AGENTS.md, README.md, SKILL.md, runbook, overlay/gentle-ai/README.md).

- Rule 3 keeps the live documentation aligned with current behavior.
- Rule 4 preserves the decision history — why things became the way they are.

A change that updates docs (rule 3) without logging the decision (rule 4) is incomplete. Likewise, logging a decision (rule 4) without updating the docs that describe current behavior (rule 3) leaves the live docs lying about the system.

---

## Repository structure

```text
gentle-ai-custom/
├── .agents/
│   └── skills/
│       └── gentle-ai-overlay-maintainer/
│           └── SKILL.md                         # Skill: maintain the Gentle AI overlay against upstream changes
├── overlay/
│   └── gentle-ai/
│       ├── README.md                            # Human guide for the Gentle AI control-plane
│       ├── policy/
│       │   ├── gentle-ai-policy.json           # Machine-readable keep/prune + overrides + OpenCode paths
│       │   ├── maintenance-intent.md           # Semantic source of truth for what to preserve/depure and why
│       │   └── orchestrator-policy.md          # Sanitization intent for orchestrators
│       ├── state/
│       │   └── upstream-state.json             # Last maintained upstream version/tag/commit boundary
│       ├── runbooks/
│       │   └── maintain-upstream-overlay.md    # Human maintenance runbook
│       ├── logs/
│       │   └── update-log.md                   # Incremental decision history
│       ├── scripts/
│       │   ├── apply-gentle-ai-policy.sh       # Internal helper: depure Gentle AI runtime assets
│       │   ├── apply-gentle-ai-policy.ps1      # Internal helper: Windows equivalent
│       │   └── audit-gentle-ai-upstream.py     # Shared maintainer audit logic for base prompt + metadata + invariants
│       └── snapshots/
│           └── upstream/
│               └── opencode/
│                   └── orchestrators/          # Versioned baseline snapshot + metadata (gentle-orchestrator.last.md + .meta.yaml)
├── shared/
│   ├── skills/
│   │   ├── commit-planner/
│   │   │   └── SKILL.md
│   │   └── pr-finalizer/
│   │       └── SKILL.md
│   └── commands/
│       ├── commit-plan-body.md
│       ├── commit-apply-body.md
│       ├── commit-fast-body.md
│       ├── pr-create-body.md
│       └── pr-regenerate-body.md
├── apply-gentle-ai-custom.sh                   # Canonical Linux/macOS entrypoint (public)
├── apply-gentle-ai-custom.ps1                  # Canonical Windows entrypoint (public)
├── audit-gentle-ai-upstream.sh                 # Canonical Linux/macOS maintainer audit entrypoint (public)
├── audit-gentle-ai-upstream.ps1                # Canonical Windows maintainer audit entrypoint (public)
├── AGENTS.md
├── CLAUDE.md
└── README.md
```

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

This repo does **not** mirror the upstream codebase. Upstream lives at `/home/manuel/Documentos/gentle-ai` and is treated as input only.

The maintenance model is intentionally split into:

- `maintenance-intent.md` → semantic intent
- `gentle-ai-policy.json` → runtime policy
- `upstream-state.json` → last maintained upstream boundary
- `update-log.md` → historical record

---

## Skills catalog

### `commit-planner`

- **Source**: `shared/skills/commit-planner/SKILL.md`
- **Purpose**: plan and execute coherently grouped local commits after implementation.
- **Commands**: `/commit-plan`, `/commit-apply`, `/commit-fast`

### `pr-finalizer`

- **Source**: `shared/skills/pr-finalizer/SKILL.md`
- **Purpose**: generate or regenerate PR content from committed diff.
- **Commands**: `/pr-create`, `/pr-regenerate`

### `gentle-ai-overlay-maintainer`

- **Source**: `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- **Purpose**: maintain this repo's overlay against upstream Gentle AI updates with version-aware auditing and explicit human approval gates.
- **Use when**: auditing upstream changes, refreshing sanitization rules, deciding what to keep/depure after a new Gentle AI version, and updating intent/policy/state/log coherently.

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

SDD profile orchestrators (`sdd-orchestrator-<name>` and `sdd-<phase>-<name>`) are **NOT** baked into the versioned policy. They are reconciled from a per-machine local config — see `## SDD profile local config` below. The versioned policy keeps only portable baseline keys (`gentle-orchestrator`) so the repo never carries machine-specific model/variant choices.

Profile orchestrator snapshots also stay out of the versioned repo. The repo keeps the portable `gentle-orchestrator.last.md` baseline plus `gentle-orchestrator.last.meta.yaml`; the helper keeps operational snapshots per machine under `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.

The maintainer must not infer evolving user intent only from the JSON policy. Intent changes belong first in `maintenance-intent.md`, then in policy/runtime artifacts if the user approves them.

---

## SDD profile local config

Per-machine SDD profile assignments live OUTSIDE the repo at:

```
~/.config/gentle-ai-custom/opencode-sdd-profiles.json
```

### Behavior

- If the file does NOT exist → the helper does not touch any SDD profile entry in `opencode.json`.
- If the file exists → the helper validates it STRICTLY and fails closed before any write if validation fails.
- For each managed profile → the helper creates or updates `sdd-orchestrator-<name>` plus the 10 phase agents (`sdd-init-<name>`, …, `sdd-onboard-<name>`) with the configured `model` + `variant`.
- Profiles present in `opencode.json` but NOT named in the local config are LEFT UNTOUCHED and surfaced as `WARNING - unmanaged SDD profiles left untouched`. Nothing is deleted automatically.

### V1 schema (strict, no inheritance)

```jsonc
{
  "version": 1,
  "profiles": [
    {
      "name": "vertex",
      "orchestrator": { "model": "provider/model", "variant": "..." },
      "phases": {
        "sdd-init":     { "model": "provider/model", "variant": "..." },
        "sdd-explore":  { "model": "provider/model", "variant": "..." },
        "sdd-propose":  { "model": "provider/model", "variant": "..." },
        "sdd-spec":     { "model": "provider/model", "variant": "..." },
        "sdd-design":   { "model": "provider/model", "variant": "..." },
        "sdd-tasks":    { "model": "provider/model", "variant": "..." },
        "sdd-apply":    { "model": "provider/model", "variant": "..." },
        "sdd-verify":   { "model": "provider/model", "variant": "..." },
        "sdd-archive":  { "model": "provider/model", "variant": "..." },
        "sdd-onboard":  { "model": "provider/model", "variant": "..." }
      }
    }
  ]
}
```

Hard rules:

- Top-level config MUST contain exactly `version` and `profiles`. Extra top-level fields are rejected.
- `version` MUST equal `1`.
- `profiles` MUST be a non-empty array.
- Each profile MUST have exactly the fields `name`, `orchestrator`, `phases`. Extra fields are rejected.
- `name` MUST match `^[a-z0-9][a-z0-9._-]*$` (safe agent-key suffix) and MUST be unique across profiles in the file.
- `orchestrator` and each phase assignment MUST be exactly `{ "model": "...", "variant": "..." }`.
- `model` MUST be a non-empty string.
- `variant` MUST be a string. It may be `""` if the assignment has no variant, but the field is REQUIRED — there is no implicit default.
- `phases` MUST contain exactly the 10 SDD phase keys listed above. No defaults are inherited from anywhere.

### Why orchestrator prompts are NOT in this schema

The helper only manages `model`/`variant` on profile orchestrators. Orchestrator `prompt` content still comes from `gentle-ai sync` and is sanitized by the existing inline-orchestrator pass. If a profile is configured locally but the orchestrator agent does not yet exist in `opencode.json`, the helper creates a stub `{ model, variant }` agent without a prompt — running `gentle-ai sync` then materializes the prompt, and the next overlay run picks it up via the prefix-matched sanitization.

---

## Orchestrator rule

The OpenCode orchestrator is inline upstream by design. The maintainer audit/apply flow must therefore:

1. audit the upstream base asset with `bash audit-gentle-ai-upstream.sh` before maintainer sync/reinstall work
2. read the inline prompt from `opencode.json`
3. snapshot every orchestrator into `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`
4. additionally keep `gentle-orchestrator.last.md` plus `gentle-orchestrator.last.meta.yaml` versioned under `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`
5. sanitize PR/budget/chained-PR/review-workload flow
6. generate `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md`
7. repoint the orchestrator to that generated file
8. fail closed if the materialized `gentle-orchestrator` does not match the last audited baseline/metadata

Recovery rules:

- `gentle-orchestrator`: prefer the local operational snapshot; if missing, fall back to the repo versioned snapshot and mirror it back into the local directory.
- `sdd-orchestrator-<name>`: use only the local operational snapshot directory. If missing, fail closed and require `gentle-ai sync`.

Do **not** switch back to a static repo-owned prompt file as the operational source of truth.

If sanitization anchors are missing, fail closed and surface the warning.

---

## Update flow

Before maintainer sync/reinstall work, run:

```bash
bash audit-gentle-ai-upstream.sh
```

After any Gentle AI operation other than a bare `brew upgrade`, always run:

```bash
bash apply-gentle-ai-custom.sh all
```

**Why**: `gentle-ai sync` resets orchestrator prompts to upstream inline content and reinstalls all skills (including pruned ones). The audit script catches upstream base-prompt or profile-invariant drift before you sync. The apply script then re-applies the full overlay: skill pruning, model overrides, snapshot capture, sanitization, `{file:...}` rewrite, and automatic verification that `gentle-orchestrator` still matches the last audited baseline.

| Operation | Resets prompts | Restores pruned skills | Run script after |
|---|---|---|---|
| `brew upgrade` only | No | No | No |
| `gentle-ai sync` | **Yes** | **Yes** | **Always** |
| TUI reinstall | **Yes** (topology may change) | **Yes** | **Always** (audit first) |

When reinstalling, the overlay maintainer agent must audit before running the script in case agent topology changed.

During an upstream audit, the maintainer must make the adoption path explicit:

- If the upstream delta is overlay-relevant but preserves agent topology, recommend `gentle-ai sync` and then re-apply the overlay.
- If the upstream delta adds, removes, or renames agents, changes presets/topology, or changes how upstream materializes `opencode.json`, recommend a full reinstall before re-applying the overlay.
- If both kinds of changes exist, topology wins: recommend reinstall.

---

## Runtime caveat

If the scripts update `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new orchestrator prompt takes effect.

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. Suggested scope examples: `overlay`, `maintainer-skill`, `custom-layer`, `commit-planner`, `pr-finalizer`.
