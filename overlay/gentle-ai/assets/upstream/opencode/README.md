## Upstream asset snapshots for OpenCode

This tree is reserved for the approved upstream-equivalent inputs and retained assets that feed the owned OpenCode materialization.

Intended contents:

- `persona-gentleman.md` — upstream persona source
- `engram-protocol.md` — mirrored copy of `internal/assets/claude/engram-protocol.md`, kept here because it feeds the owned OpenCode AGENTS build
- `prompts/orchestrators/` — upstream orchestrator baselines
- `sdd-overlay-single.json` / `sdd-overlay-multi.json` — upstream overlay JSONs
- `commands/` — retained upstream command snapshots used to materialize the OpenCode command surface
- `plugins/` — upstream plugins
- `skills/` — selected upstream skill snapshots

No stored `AGENTS.md` lives in this tree. This tree is the active maintainer snapshot target for `sync-gentle-ai-upstream-assets`; runtime installation still comes from `../owned/opencode/`.
