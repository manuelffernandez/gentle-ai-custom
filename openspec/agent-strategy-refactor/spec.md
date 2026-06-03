# Spec: agent-strategy-refactor

## Purpose

Define the behavioral contract for the Agent interface pattern that replaces the hardcoded multi-agent switches in `apply_custom.go`. This spec describes the structural and behavioral requirements for `agent.go`, `opencode_agent.go`, and the refactored `RunApplyCustom` dispatch.

---

## Requirements

### Requirement: Agent Interface Contract

The package MUST expose an `Agent` interface in `internal/overlay/` with exactly these methods: `Name() string`, `BasePath() string`, `BuildCommandContent(cmd customCommand, body string) string`, `ApplyOverlay(repoRoot string, options applyPolicyOptions) int`.

No additional methods SHALL be required by the dispatch loop.

#### Scenario: Interface is the only dispatch contract

- GIVEN the `RunApplyCustom` function executes an agent pipeline
- WHEN it calls any agent method
- THEN it MUST use only the four methods defined by the `Agent` interface, with no type assertions or direct struct access

---

### Requirement: OpenCodeAgent Preserves Existing Behavior

`OpenCodeAgent` MUST implement `Agent` and produce output identical to the current `case "opencode"` branches in `buildCustomTarget` and `buildCustomCommandContent`.

| Method | Expected behavior |
|--------|-------------------|
| `Name()` | Returns `"opencode"` |
| `BasePath()` | Returns `~/.config/opencode` (via `os.UserHomeDir`) |
| `BuildCommandContent()` | Emits the same YAML frontmatter as the current `case "opencode"` in `buildCustomCommandContent` |
| `ApplyOverlay()` | Calls `runApplyPolicyWithOptions` with the same args; returns its exit code |

#### Scenario: Command content is identical to current output

- GIVEN an `OpenCodeAgent` and a `customCommand` with renderer `"opencode"`
- WHEN `BuildCommandContent` is called
- THEN the returned string MUST be byte-for-byte identical to what the current switch produces for the same input

#### Scenario: Overlay pipeline produces same files

- GIVEN `OpenCodeAgent.ApplyOverlay()` is invoked with valid `repoRoot` and options
- WHEN the policy pipeline runs
- THEN orchestrator files and `opencode.json` mutations MUST be identical to the pre-refactor output

---

### Requirement: Registry-Based Agent Resolution

`RunApplyCustom` MUST resolve agents from a `map[string]Agent` registry. The initial registry MUST contain exactly one entry: `"opencode"`.

#### Scenario: Known target resolves to agent

- GIVEN the registry contains `"opencode": &OpenCodeAgent{}`
- WHEN the CLI is invoked with target `opencode`
- THEN `RunApplyCustom` runs the full pipeline for `OpenCodeAgent` and returns exit code 0

#### Scenario: Unknown target returns descriptive error

- GIVEN the registry has only `"opencode"` registered
- WHEN the CLI is invoked with an unknown target (e.g., `gemini`)
- THEN `RunApplyCustom` MUST print an error containing `"unsupported agent: gemini"` to stderr and return exit code 1

---

### Requirement: `all` Expands to Registry

When the target is `all`, the CLI MUST expand it to all agents currently registered, in registry-iteration order.

#### Scenario: `all` with single registered agent

- GIVEN the registry contains only `"opencode"`
- WHEN the CLI is invoked with `all`
- THEN exactly one agent runs: `OpenCodeAgent`; output is identical to `opencode` target

---

### Requirement: Dead Code Eliminated

`apply_custom.go` MUST NOT contain references to `"claude"`, `"codex"`, `"gemini"`, or `"antigravity"` after the refactor. The functions `buildCustomTarget`, `buildCustomCommandContent`, and `shouldApplyGentleOverlay` MUST NOT exist.

#### Scenario: No legacy renderer cases remain

- GIVEN the refactored `apply_custom.go`
- WHEN searched for strings `"claude"`, `"codex"`, `"gemini"`, `"antigravity"`
- THEN no matches MUST be found in that file

---

### Requirement: Shell Script Usage Reflects Single Target

`apply-gentle-ai-custom.sh` and `.ps1` help output MUST list only `opencode` and `all` as valid targets. Legacy target names MUST NOT appear in usage or examples.

#### Scenario: Help output is accurate

- GIVEN the refactored shell script
- WHEN `--help` is passed
- THEN usage text lists `opencode` and `all`; does not mention `claude`, `codex`, `gemini`, or `antigravity`
