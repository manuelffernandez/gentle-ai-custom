# Tasks: agent-strategy-refactor

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~270 (additions + deletions) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Import Circular Resolution

`overlay` → `agents` → `overlay` is circular. Resolution: **keep both `agent.go` and `opencode_agent.go` in package `overlay`** (no subdirectory). `OpenCodeAgent` calls `runApplyPolicyWithOptions` and uses `applyPolicyOptions` freely as same-package symbols. The `agents/` subdirectory is deferred until a second agent requires it. YAGNI.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full refactor (all phases) | PR 1 | Single PR, Low budget risk |

---

## Phase 1: Package Structure + Interface

- [x] 1.1 Create `internal/overlay/agent.go` — `Agent` interface (4 methods: `Name()`, `BasePath()`, `BuildCommandContent(cmd customCommand, body string) string`, `ApplyOverlay(repoRoot string, options applyPolicyOptions) int`), empty `var agentRegistry = map[string]Agent{}`, and `registeredAgentNames() []string` (sorted keys).
- [x] 1.2 Verify: `go build ./...` passes with new empty-registry file.

## Phase 2: OpenCodeAgent Implementation

- [x] 2.1 Create `internal/overlay/opencode_agent.go` (package `overlay`) — `OpenCodeAgent` struct with all 4 methods: `Name()` → `"opencode"`, `BasePath()` → `filepath.Join(home, ".config", "opencode")`, `BuildCommandContent()` → exact current `case "opencode"` block from `buildCustomCommandContent`, `ApplyOverlay()` → `return runApplyPolicyWithOptions(repoRoot, options)`.
- [x] 2.2 Populate `agentRegistry` in `agent.go`: `"opencode": &OpenCodeAgent{}`.
- [x] 2.3 Verify: `go build ./...` and `go vet ./...` clean.

## Phase 3: Refactor apply_custom.go

- [x] 3.1 Refactor `normalizeTargets` — validate positional args against `agentRegistry` keys (error: `"unsupported agent: %s"`); `all` expands to `registeredAgentNames()`. Remove `isSupportedTarget` call (used only here).
- [x] 3.2 Add `installAgentAssets(agent Agent, sources customSourceFiles, sharedRoot string, recorder *verboseRecorder) error` — generic replacement for `applyCustomTarget`. Uses `agent.BasePath()` and `agent.Name()` in place of `target.basePath`/`target.name`; command `fileRelPath` hardcoded as `"commands/{name}.md"` (YAGNI). Message: `fmt.Printf("Applied %s overlays -> %s\n", agent.Name(), agent.BasePath())`.
- [x] 3.3 Update `renderCustomCommand` to accept `agent Agent` and call `agent.BuildCommandContent(command, body)` instead of `buildCustomCommandContent(command, body)`. Remove `renderer` field from `customCommand` struct (no longer used by dispatch).
- [x] 3.4 Refactor `RunApplyCustom` dispatch loop — resolve `agent` from `agentRegistry[targetName]`, call `installAgentAssets(agent, ...)`, then `agent.ApplyOverlay(repoRoot, applyPolicyOptions{...})`. Remove `shouldApplyGentleOverlay` call and its `else` recorder branch (overlay always runs via `ApplyOverlay`).
- [x] 3.5 Update `printApplyCustomUsage` — remove all legacy target names (`claude`, `codex`, `gemini`, `antigravity`) from usage and examples. List only `opencode` and `all`.
- [x] 3.6 Verify: `go build ./...` passes; behavior of `opencode` target unchanged.

## Phase 4: Dead Code Elimination

- [x] 4.1 Delete from `apply_custom.go`: `var supportedTargets`, `buildCustomTarget`, `buildCustomCommandContent`, `shouldApplyGentleOverlay`, `isSupportedTarget`, `customTarget` struct.
- [x] 4.2 Verify: `grep -n "claude\|codex\|gemini\|antigravity" internal/overlay/apply_custom.go` → zero matches. `go build ./...` passes.

## Phase 5: CLI Usage Update

- [x] 5.1 Update `cmd/gentle-ai-overlay/main.go` `printMainUsage` — remove `claude|codex|gemini|antigravity` from the `apply-custom` usage line.
- [x] 5.2 Verify: `go run ./cmd/gentle-ai-overlay apply-custom --help` lists only `opencode` and `all`.

## Phase 6: Final Verification

- [x] 6.1 `go build ./...` clean, `go vet ./...` clean.
- [x] 6.2 `bash apply-gentle-ai-custom.sh opencode --verbose` produces output identical to pre-refactor.
- [x] 6.3 `bash apply-gentle-ai-custom.sh all --verbose` runs exactly the `opencode` pipeline.
- [x] 6.4 `bash apply-gentle-ai-custom.sh gemini` → stderr contains `"unsupported agent: gemini"`, exits 1.
- [x] 6.5 `bash apply-gentle-ai-custom.sh --help` does not mention `claude`, `codex`, `gemini`, or `antigravity`.
