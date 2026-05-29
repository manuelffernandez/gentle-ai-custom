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

This repo now has two script pairs:

- `apply-gentle-ai-custom.sh` / `apply-gentle-ai-custom.ps1` → canonical user-facing entrypoints
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` / `.ps1` → internal Gentle AI depuration helpers

If one side changes, the paired script must be updated in the same commit.

**Canonical entrypoint parity items:**
- target parsing and usage/help behavior
- installation of custom skills/wrappers
- invocation of the Gentle AI overlay helper

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
│       │   └── apply-gentle-ai-policy.ps1      # Internal helper: Windows equivalent
│       └── snapshots/
│           └── upstream/
│               └── opencode/
│                   └── orchestrators/          # Per-orchestrator snapshots written at runtime
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

The maintainer must not infer evolving user intent only from the JSON policy. Intent changes belong first in `maintenance-intent.md`, then in policy/runtime artifacts if the user approves them.

---

## Orchestrator rule

The OpenCode orchestrator is inline upstream by design. The helper scripts must therefore:

1. read the inline prompt from `opencode.json`
2. snapshot it per orchestrator
3. sanitize PR/budget/chained-PR/review-workload flow
4. generate `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md`
5. repoint the orchestrator to that generated file

Do **not** switch back to a static repo-owned prompt file as the operational source of truth.

If sanitization anchors are missing, fail closed and surface the warning.

---

## Update flow

After any Gentle AI operation other than a bare `brew upgrade`, always run:

```bash
bash apply-gentle-ai-custom.sh all
```

**Why**: `gentle-ai sync` resets orchestrator prompts to upstream inline content and reinstalls all skills (including pruned ones). The script re-applies the full overlay: skill pruning, model overrides, snapshot capture, sanitization, and `{file:...}` rewrite.

| Operation | Resets prompts | Restores pruned skills | Run script after |
|---|---|---|---|
| `brew upgrade` only | No | No | No |
| `gentle-ai sync` | **Yes** | **Yes** | **Always** |
| TUI reinstall | **Yes** (topology may change) | **Yes** | **Always** (audit first) |

When reinstalling, the overlay maintainer agent must audit before running the script in case agent topology changed.

---

## Runtime caveat

If the scripts update `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new orchestrator prompt takes effect.

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. Suggested scope examples: `overlay`, `maintainer-skill`, `custom-layer`, `commit-planner`, `pr-finalizer`.
