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

### 2. Exact parity between installers

`inject-skills.sh` (bash) and `inject-skills.ps1` (PowerShell 5.1+) must maintain **exact behavioral parity**.

If one is modified, the other must be updated in the same commit to preserve identical behavior on Linux/macOS and Windows. This includes:
- path variables
- source validations (`validate_sources` / `Assert-Sources`)
- command rendering per target (`apply_opencode`, `apply_claude`, `apply_codex`, `apply_gemini`, `apply_antigravity`)
- mode-specific conditions (e.g. `disable-model-invocation`)

Never leave the two scripts in a divergent state.

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

---

## Skills catalog

### `commit-planner`

- **Source**: `shared/skills/commit-planner/SKILL.md`
- **Purpose**: plan and execute coherently grouped local commits after finishing implementation.
- **Modes**: `plan` (read-only), `apply` (requires approval), `auto` (executes without approval pause).
- **Exposed commands**: `/commit-plan`, `/commit-apply`, `/commit-fast`

### `pr-finalizer`

- **Source**: `shared/skills/pr-finalizer/SKILL.md`
- **Purpose**: generate or regenerate PR content from the committed diff; create or update the PR on GitHub after explicit approval.
- **Modes**: `create`, `regenerate`
- **Exposed commands**: `/pr-create`, `/pr-regenerate`

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
