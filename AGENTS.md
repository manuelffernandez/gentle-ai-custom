# AGENTS.md — gentle-ai-custom

Operating instructions for AI agents (Claude, OpenCode, Codex, and equivalents) working in this repository.

---

## Operating rules

### 1. Self-update this document

Any change that affects the repository structure — creating folders, adding or removing files, relocating a resource, or altering the purpose of a component — **must be accompanied by the corresponding update to this `AGENTS.md`**.

This applies to:
- new skills (`shared/skills/<name>/`)
- new command bodies (`shared/commands/<name>-body.md`)
- changes to installation targets
- new root files (`AGENTS.md`, `CLAUDE.md`, installers, etc.)

This document must always reflect the actual state of the repository, not the state at the time it was written.

### 2. Exact parity between paired automation scripts

`inject-skills.sh` (bash) and `inject-skills.ps1` (PowerShell 5.1+) must maintain **exact behavioral parity**.

`overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` (bash) and `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1` (PowerShell 5.1+) must also maintain **exact behavioral parity**.

If one is modified, the other must be updated in the same commit to preserve identical behavior on Linux/macOS and Windows.

**`inject-skills.*` pair — parity items:**
- path variables
- source validations (`validate_sources` / `Assert-Sources`)
- command rendering per target (`apply_opencode`, `apply_claude`, `apply_codex`, `apply_gemini`, `apply_antigravity`)
- mode-specific conditions (e.g. `disable-model-invocation`)

**`apply-gentle-ai-policy.*` pair — parity items:**
- skill pruning logic
- upstream snapshot capture
- prompt redirection
- `agent_overrides` application

Never leave either pair in a divergent state.

### 3. Update documentation on functional changes

Any modification that affects the operability of a skill, command, or workflow must be reflected in the documentation:

- **`README.md`**: update the commands section when a command is added, removed, or its behavior changes; update the structure section when files or folders are added.
- **`AGENTS.md`**: update the repository structure and the skills/commands catalog.
- **`SKILL.md` of the affected skill**: keep it as the source of truth for behavior.

Adding a new command without documenting it in `README.md`, or changing the behavior of an existing one without updating its description, is not acceptable.

---

## Repository structure

```
gentle-ai-custom/
├── overlay/
│   └── gentle-ai/
│       ├── README.md                              # Human guide for the Gentle AI overlay/control-plane
│       ├── policy/
│       │   ├── gentle-ai-policy.json             # Machine-readable keep/prune baseline + path config
│       │   └── orchestrator-policy.md            # Rules for deriving a clean SDD orchestrator prompt
│       ├── prompts/
│       │   └── audit-gentle-ai-update.md         # Reusable audit prompt for future upstream updates
│       ├── derived/
│       │   └── opencode/
│       │       └── gentle-orchestrator.md        # Local standalone orchestrator prompt without PR/budget flow
│       ├── snapshots/
│       │   └── upstream/
│       │       └── opencode/
│       │           └── gentle-orchestrator.last.md # Last upstream prompt seen before redirecting to derived prompt
│       ├── logs/
│       │   └── update-log.md                     # Incremental overlay decisions and upstream audit history
│       └── scripts/
│           ├── apply-gentle-ai-policy.sh         # Linux/macOS policy applicator
│           └── apply-gentle-ai-policy.ps1        # Windows PowerShell policy applicator
├── shared/
│   ├── skills/
│   │   ├── commit-planner/
│   │   │   └── SKILL.md          # Skill: commit planning and execution
│   │   └── pr-finalizer/
│   │       └── SKILL.md          # Skill: PR creation and regeneration
│   └── commands/
│       ├── commit-plan-body.md   # Command body for /commit-plan (plan mode)
│       ├── commit-apply-body.md  # Command body for /commit-apply (apply mode)
│       ├── commit-fast-body.md   # Command body for /commit-fast (auto mode)
│       ├── pr-create-body.md     # Command body for /pr-create (create mode)
│       └── pr-regenerate-body.md # Command body for /pr-regenerate (regenerate mode)
├── .atl/
│   └── skill-registry.md         # Skill registry for ATL orchestrator resolution
├── inject-skills.sh               # Bash installer (Linux/macOS)
├── inject-skills.ps1              # PowerShell installer (Windows, 5.1+)
├── AGENTS.md                      # This file — instructions for AI agents
├── CLAUDE.md                      # Delegates to AGENTS.md for Claude Code
└── README.md                      # Usage documentation for humans (Spanish)
```

Agent-specific wrappers (OpenCode commands, Claude commands, Codex prompts, Gemini command skills, Antigravity command skills) are **not versioned** in this repo. They are generated at install time from the sources in `shared/`.

The `overlay/gentle-ai/` tree is different: it is a **persistent control-plane** for adapting upstream Gentle AI to local preferences. It does not mirror the upstream repo; it stores local policy, local prompts, local audit instructions, and the paired scripts that reapply those decisions after every sync/update.

---

## Skills catalog

### `commit-planner`

- **Source**: `shared/skills/commit-planner/SKILL.md`
- **Purpose**: plan and execute coherently grouped local commits after finishing implementation.
- **Modes**: `plan` (read-only), `apply` (requires approval), `auto` (executes without approval pause).
- **Exposed commands**: `/commit-plan`, `/commit-apply`, `/commit-fast`

> Nota OpenCode: los wrappers generados para `/commit-plan`, `/commit-apply` y `/commit-fast` no fijan `agent:` en el frontmatter; dependen del agente por defecto del entorno.

### `pr-finalizer`

- **Source**: `shared/skills/pr-finalizer/SKILL.md`
- **Purpose**: generate or regenerate PR content from the committed diff; refresh remote refs automatically, validate remote head state read-only, and create or update the PR on GitHub immediately after the content approval step, without a second confirmation.
- **Modes**: `create`, `regenerate`
- **Exposed commands**: `/pr-create`, `/pr-regenerate`

---

## Gentle AI overlay control-plane

### Purpose

The `overlay/gentle-ai/` subtree exists to maintain a **local, durable adaptation layer** on top of the upstream Gentle AI repo at `/home/manuel/Documentos/gentle-ai`.

Its job is to:

- preserve the approved keep/prune skill baseline
- remove unwanted repo-management conventions from the OpenCode SDD orchestrator
- provide a reusable audit prompt for future upstream updates
- keep a local snapshot/log trail so each future review is incremental rather than starting from zero

### Source-of-truth split

- **Upstream input**: `/home/manuel/Documentos/gentle-ai`
- **Local decisions**: `overlay/gentle-ai/policy/`, `overlay/gentle-ai/derived/`, `overlay/gentle-ai/logs/`
- **Mechanical reapplication**: `overlay/gentle-ai/scripts/`

### Audit workflow

When auditing a new Gentle AI update:

1. Start from this repository (`gentle-ai-custom`), not from the upstream repo.
2. Read the upstream repo as an external source: `/home/manuel/Documentos/gentle-ai`.
3. Use `overlay/gentle-ai/prompts/audit-gentle-ai-update.md` as the reusable audit prompt.
4. Update these artifacts together when decisions change:
   - `policy/gentle-ai-policy.json`
   - `policy/orchestrator-policy.md`
   - `derived/opencode/gentle-orchestrator.md`
   - `snapshots/upstream/opencode/gentle-orchestrator.last.md`
   - `logs/update-log.md`
   - `scripts/apply-gentle-ai-policy.sh`
   - `scripts/apply-gentle-ai-policy.ps1`
   - `README.md`
   - `AGENTS.md`

### Current approved baseline

**Keep**:

- `_shared`
- `cognitive-doc-design`
- `comment-writer`
- `go-testing`
- `judgment-day`
- `sdd-apply`
- `sdd-archive`
- `sdd-design`
- `sdd-explore`
- `sdd-init`
- `sdd-onboard`
- `sdd-propose`
- `sdd-spec`
- `sdd-tasks`
- `sdd-verify`
- `skill-creator`
- `skill-improver`
- `skill-registry`

**Prune**:

- `branch-pr`
- `chained-pr`
- `issue-creation`
- `work-unit-commits`

### Orchestrator rule

The derived OpenCode `gentle-orchestrator` prompt must remove **all PR/budget/chained-PR/review-workload flow** while preserving the useful SDD orchestration behavior. Do not reintroduce those concepts unless the user explicitly changes this policy.

### Runtime caveat

If `apply-gentle-ai-policy.*` updates `~/.config/opencode/opencode.json`, OpenCode must be restarted before the new `gentle-orchestrator` prompt takes effect. Agents working from this repo should surface that reminder explicitly.

---

## Installation targets

| Target | Skills destination | Commands/prompts destination |
|--------|--------------------|------------------------------|
| `opencode` | `~/.config/opencode/skills/` | `~/.config/opencode/commands/` |
| `claude` | `~/.claude/skills/` | `~/.claude/commands/` |
| `codex` | `~/.codex/skills/` | `~/.codex/prompts/` |
| `gemini` | `~/.gemini/skills/` | `~/.gemini/skills/<command-name>/` (skill entries) |
| `antigravity` | `~/.gemini/antigravity/skills/` | `~/.gemini/antigravity/skills/<command-name>/` (skill entries) |

---

## Commit convention

Use **Conventional Commits**: `feat`, `fix`, `refactor`, `docs`, `chore`. The scope is the affected component (e.g. `commit-planner`, `pr-finalizer`, `overlay`).
