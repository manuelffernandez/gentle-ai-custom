# Proposal: agent-strategy-refactor

## Intent

The `apply-gentle-ai-custom.sh` CLI exposes a uniform multi-agent interface but delivers inconsistent behavior per target:
- **OpenCode**: full treatment (custom skills + overlay policy)
- **Claude**: partial (custom skills + skill pruning, can touch opencode.json)
- **Codex / Gemini / Antigravity**: custom skills only, no policy

The owner uses exclusively OpenCode. The inconsistency is a maintenance liability and the multi-agent surface creates confusion without delivering value.

This change eliminates all non-OpenCode agent support and refactors the dispatch to a Strategy/Interface pattern (`Agent` interface), so the command's public contract matches its actual behavior and future agents can be added with complete, consistent treatment.

## Scope

**Changes:**
- Define an `Agent` interface in `internal/overlay/`
- Implement `OpenCodeAgent` (extracts logic from existing switches)
- Remove the two hardcoded switches in `apply_custom.go`:
  - `buildCustomTarget` — switch for paths/config per agent
  - `buildCustomCommandContent` — switch for YAML frontmatter per agent
- Remove `shouldApplyGentleOverlay()` conditional function
- Refactor `RunApplyCustom` to iterate over a `map[string]Agent` registry instead of raw strings
- Move `runApplyPolicyWithOptions` invocation inside `OpenCodeAgent.ApplyOverlay()`
- Update shell script help/usage to reflect the single supported target

**Does NOT change:**
- Functional behavior of OpenCode (same output, same files written, same policy applied)
- Internal logic of `apply_policy.go` beyond the integration point
- `audit-gentle-ai-upstream.sh` and its internals
- Any overlay policy rules or profile configuration

## Approach

### 1. Define the `Agent` interface

```go
// internal/overlay/agent.go
type Agent interface {
    Name() string
    BasePath() string
    BuildCommandContent(cmd customCommand, body string) string
    ApplyOverlay(repoRoot string, options applyPolicyOptions) int
}
```

### 2. Implement `OpenCodeAgent`

```go
// internal/overlay/opencode_agent.go
type OpenCodeAgent struct{}

func (a *OpenCodeAgent) Name() string { return "opencode" }
func (a *OpenCodeAgent) BasePath() string { return filepath.Join(os.UserHomeDir(), ".config/opencode") }
func (a *OpenCodeAgent) BuildCommandContent(cmd customCommand, body string) string {
    // logic extracted from the current switch case "opencode" in buildCustomCommandContent
}
func (a *OpenCodeAgent) ApplyOverlay(repoRoot string, opts applyPolicyOptions) int {
    // calls runApplyPolicyWithOptions directly
}
```

### 3. Refactor `RunApplyCustom`

Replace the `normalizeTargets` + switch dispatch with an agent registry:

```go
var agentRegistry = map[string]Agent{
    "opencode": &OpenCodeAgent{},
}

func resolveAgents(targets []string) ([]Agent, error) {
    if contains(targets, "all") {
        return allAgents(), nil
    }
    // resolve by name, error on unknown
}
```

Each agent runs its full pipeline: install skills → apply overlay. No conditional branching.

### 4. Update `all` target

`all` becomes equivalent to listing all registered agents — today that means only `opencode`. The behavior is identical to `apply-gentle-ai-custom.sh opencode`.

### 5. Clean up shell scripts

Update help/usage in `apply-gentle-ai-custom.sh` and `.ps1` to document the current supported targets accurately.

## Out of Scope

- Writing tests (no test infrastructure exists in the project)
- Refactoring internal logic of `apply_policy.go` beyond the call site
- Adding Claude or any other agent support (future work, implement `Agent` interface when needed)
- Changes to overlay policy rules, profile configuration, or snapshot behavior

## Risks

**Low overall risk.**

- **Output ordering**: Moving the policy summary log from global to per-agent level may change the order of output lines when multiple agents are registered in the future. This is acceptable and actually desirable (each agent owns its output). No impact with a single agent.
- **`all` semantics change**: Today `all` installs skills for 5 agents. After the refactor, `all` = only OpenCode. If anyone (unlikely, this is a personal tool) relied on `all` to install skills in Claude/Codex/Gemini directories, that stops working. Acceptable given the stated goal.
- **No regression risk on OpenCode behavior**: The policy pipeline (`apply_policy.go`) is untouched internally. The `OpenCodeAgent` is a thin wrapper around the existing call.

## Success Criteria

1. `bash apply-gentle-ai-custom.sh opencode` produces identical output to today
2. `bash apply-gentle-ai-custom.sh all` applies only OpenCode (single registered agent)
3. `apply_custom.go` contains no references to claude, codex, gemini, or antigravity
4. Adding a new agent in the future requires only: implement `Agent` interface + register in `agentRegistry`
5. No regression in overlay policy behavior (same files written, same OpenCode config mutations)
