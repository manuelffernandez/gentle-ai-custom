# Codex overlay support

## Objective

Add the approved first Codex target to `gentle-ai-custom`, installing existing portable repo-owned skills and applying only compatible local instruction rules and exact surgical patches to Codex's upstream-materialized files.

## Why and scope

- Implement repository behavior only; do not change installed runtime state.
- Preserve upstream ownership of `~/.codex/AGENTS.md`, skills, Engram, `config.toml`, and model profiles. Do not replace full Codex instructions with OpenCode instructions, render OpenCode commands, or apply OpenCode JSON settings.
- Reuse existing `shared/skills/` sources without copying them into Codex-specific source trees.
- Keep target selection (`codex` or `all`) isolated: cleanup/pruning must touch only selected targets and preserve Codex's upstream SDD and unrelated files.
- Apply only local rules that are applicable and compatible with upstream Codex materialization. Any surgical patch must validate exact anchors before writes and preserve all unrelated bytes/modes; incompatible Judgment Day or work-unit anchors remain explicitly unsupported rather than forced.
- Keep the upstream remote-operation authorization rule out of overlay-maintained prompts. Preserve its audit-only upstream copy and do not weaken credential-file safety or runtime tool guards.
- Never read credential files or secret-bearing machine configuration. Never run sync, reinstall, apply against the real runtime, commit, branch, push, or create a PR. The upstream checkout is read-only input.
- Installed binary is verified as v3.7.0 by the completed canonical audit; that does not verify materialized runtime state. Correct stale v2.9.1 claims without asserting runtime parity.

## Route and execution policy

- Route: delegated direct, one writer. The implementation map spans four or more files and includes multiple non-trivial writes.
- TDD: unknown; no `openspec/config.yaml` exists and no session mode is established. Use behavior-first tests, but do not claim strict TDD was configured.
- Checks: `go test ./internal/overlay/...`, `go test ./...`, and `git diff --check`, unless an unambiguous project mode is discovered.
- Delivery strategy: ask-on-risk. Forecast is advisory at approximately 500 authored changed lines; do not code-golf to meet it.
- Commit policy: NO commits, overriding any ODD close guidance. No other delivery operation is authorized.

## Tasks

- [x] CX-1: Added Codex as an isolated apply target with shared portable skills, exact Codex instruction/skill patches, and a Codex-only patch path that avoids OpenCode policy/config behavior. Tests verify `codex`/`all` selection, shared source copies, no command rendering, preservation of Codex SDD/config and unselected OpenCode/upstream Codex skills, exact local deltas, mode preservation, idempotence, and write-free failure on incompatible or malformed anchors. A selected Codex preflight runs before any target writes, including when the caller orders OpenCode first. Checks: `go test ./internal/overlay/...` and `go test ./...` passed.
- [x] CX-2: Declared Codex install and patch ownership in `managed-assets.json`; Codex claims no upstream skill pruning or whole-file upstream asset ownership, and its installed upstream patch targets remain protected from owned-state cleanup. Tests verify the shared allowlist and ownership boundary. Check: included in `go test ./internal/overlay/...`, passed.
- [x] CX-3: Updated root `AGENTS.md`, Spanish `README.md`, overlay README, maintenance guide, semantic intent, and maintainer skill for Codex target behavior and the verified v3.7.0 binary versus unverified materialized runtime. No upstream snapshot, audit behavior, or OpenCode runtime behavior changed. Checks: `go test ./...` passed; `git diff --check` passed.
- [x] CX-4: Constrained runtime patch targets to the selected runtime root before reading them, including lexical traversal, absolute paths, and symlink-resolved escapes. Relative OpenCode target paths continue to resolve within their runtime root. Added focused rejection and valid-path tests; the initial Codex implementation's earlier checks are not evidence for these corrected bytes.

## Progress and evidence

- Completed before this task: `bash audit-gentle-ai-upstream.sh` passed for installed binary v3.7.0, exact tag `v3.7.0`, commit `6dee8f833aec9e46015759c5065a9035795d9af1`, with no drift. This proves binary/tag parity only, not installed runtime materialization.
- Initial repository tree was clean. No OpenSpec configuration or active TDD mode was found.
- CodeGraph MCP and CLI queries were attempted but failed (`no such table: unresolved_refs`); source mapping continued with targeted repository reads.
- Final checks: `go test ./internal/overlay/...` passed (`ok gentle-ai-custom/internal/overlay 2.155s`); `go test ./...` passed (`cmd/gentle-ai-overlay` and `internal/overlay`); `git diff --check` passed.
- The end-to-end `all` test used a temporary HOME and a fake v3.7.0 binary, verifying both targets and confirming Codex receives no OpenCode command wrappers. No real runtime apply was run.
- Correction verification after `gofmt` on the changed Go files: `go test ./internal/overlay/...` passed (`ok gentle-ai-custom/internal/overlay 2.198s`); `go test ./...` passed (`cmd/gentle-ai-overlay` and cached `internal/overlay`); `git diff --check` passed. This verification covers the corrected path validation and tests. No runtime apply, sync, or upstream mutation was run.
- Residual risk: a concurrently modified runtime directory could still change path components after preflight; this narrowly scoped fix protects against static traversal and symlink escapes, not hostile concurrent filesystem mutation.
- Commit: not authorized.

## Next step

Report repository-only implementation and verification evidence, unresolved risks, and skill resolution. Do not apply changes to a real runtime.
