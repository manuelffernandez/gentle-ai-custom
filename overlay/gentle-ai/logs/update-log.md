# Gentle AI Overlay Update Log

## 2026-05-29 — Removed compatibility alias layer

- Deleted `inject-skills.sh` and `inject-skills.ps1`.
- Moved the full installation + wrapper-rendering flow directly into `apply-gentle-ai-custom.sh` and `apply-gentle-ai-custom.ps1`.
- Kept `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh/.ps1` as internal helpers invoked only when targets include `opencode` or `claude`.
- Updated docs to reflect that the only public entrypoint pair is now `apply-gentle-ai-custom.sh/.ps1`.

## 2026-05-29 — Unified custom layer and dynamic orchestrator generation

- Added `apply-gentle-ai-custom.sh` and `.ps1` as canonical entrypoints.
- Converted `inject-skills.sh` and `.ps1` into compatibility aliases that still execute the full custom-layer workflow.
- Reframed `gentle-ai-custom` as a unified installation + depuration + maintenance layer for Gentle AI.
- Reworked the overlay helper so it no longer depends on a static repo-owned orchestrator prompt file.
- The helper now reads inline orchestrators from `opencode.json`, snapshots them per agent, sanitizes them, and generates `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md` files.
- Deleted the obsolete static derived prompt artifact and the obsolete single-snapshot placeholder.
- Added `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` as the runtime entrypoint for overlay maintenance.
- Added `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` as the human maintenance runbook.

## 2026-05-28 — Script fixes, parity alignment, and agent_overrides

- Fixed `apply-gentle-ai-policy.sh` Python guard: resolved `PYTHON_CMD="${PYTHON:-python3}"` at startup; guard and both inline Python calls now use `${PYTHON_CMD}` consistently.
- Fixed `apply-gentle-ai-policy.ps1` `Get-PromptContentForSnapshot`: added `-PathType Leaf` to `Test-Path` so directories are not mistakenly matched as prompt files.
- Fixed `apply-gentle-ai-policy.ps1` newline parity: replaced `[Environment]::NewLine` (CRLF on Windows) with `` "`n" `` in both snapshot-write and config-write paths; bash counterpart already writes LF.
- Fixed missing-keep warning separator parity: bash now uses `$(IFS=', '; echo "${missing_keep[*]}")` (comma-space), matching PS1's `$MissingKeep -join ', '`.
- Cleaned up `AGENTS.md` section 2: split the single shared parity bullet list into two explicit sub-lists, one per script pair (`inject-skills.*` vs `apply-gentle-ai-policy.*`), removing ambiguity about which items apply to which pair.
- Added `agent_overrides` to `gentle-ai-policy.json`: `general=openai/gpt-5.4/high`, `explore=google-vertex/gemini-3.1-pro-preview/high`.
- Updated both scripts to apply `agent_overrides` from policy atomically (single write) alongside the prompt redirect; bash uses `config_changed` flag, PS1 uses `$ConfigChanged` flag.
- Fixed `overlay/gentle-ai/README.md` Windows path example: changed relative `./overlay/...` to absolute `~\Documentos\gentle-ai-custom\...`.

## 2026-05-28 — Baseline overlay/control-plane bootstrap

- Created overlay structure under `overlay/gentle-ai/`.
- Persisted keep/prune baseline in `policy/gentle-ai-policy.json`.
- Added orchestrator derivation policy (`policy/orchestrator-policy.md`).
- Added upstream-audit prompt (`prompts/audit-gentle-ai-update.md`).
- Added derived OpenCode orchestrator prompt with PR/budget/chained-PR workflow removed.
- Added initial upstream snapshot placeholder; the first policy-apply run refreshes it with the effective upstream prompt content before redirect.
- Added paired apply scripts (`scripts/apply-gentle-ai-policy.sh` + `.ps1`) with parity commitments.

Baseline decisions:

- KEEP: `_shared`, SDD core phases, `skill-registry`, `skill-creator`, `skill-improver`, `cognitive-doc-design`, `comment-writer`, `judgment-day`, `go-testing`.
- PRUNE: `branch-pr`, `chained-pr`, `issue-creation`, `work-unit-commits`.
- Historical baseline: the first version of the overlay redirected `gentle-orchestrator` to a repo-owned derived prompt file. This was replaced on 2026-05-29 by dynamic per-orchestrator generation from inline prompts.
