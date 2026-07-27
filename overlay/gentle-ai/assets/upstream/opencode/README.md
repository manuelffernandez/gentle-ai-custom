## Upstream asset snapshots for OpenCode

This tree is reserved for approved upstream-equivalent inputs and retained assets used for audit and sync review.

Intended contents:

- `persona-gentleman.md` — upstream persona source
- `engram-protocol.md` — mirrored copy of `internal/assets/engram/protocol.md`, kept here because it feeds the owned OpenCode AGENTS build
- `prompts/orchestrators/` — upstream orchestrator baselines
- `sdd-overlay-single.json` / `sdd-overlay-multi.json` — upstream overlay JSONs
- `commands/` — retained upstream command snapshots
- `plugins/` — upstream plugins
- `skills/` — selected upstream skill snapshots

No stored `AGENTS.md` lives in this tree. This is the active snapshot target for `sync-gentle-ai-upstream-assets`; upstream Gentle AI, not custom apply, materializes these native runtime assets.
