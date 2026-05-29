# Gentle AI Overlay Update Log

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
- OpenCode `agent.gentle-orchestrator.prompt` must point to the local derived prompt file via `{file:/abs/path/...}`.
