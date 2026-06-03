# Design: agent-strategy-refactor

## Technical Approach

Replace `buildCustomTarget` (switch → struct) and `buildCustomCommandContent` (switch → string) with an `Agent` interface + registry map. `OpenCodeAgent` encapsulates basePath, command rendering, and overlay invocation. `RunApplyCustom` becomes a generic dispatch loop over registry-resolved agents. Dead agent names (claude, codex, gemini, antigravity) and `shouldApplyGentleOverlay` are eliminated.

## Architecture Decisions

| Decision | Choice | Alternative | Rationale |
|----------|--------|-------------|-----------|
| Dispatch contract | 4-method interface | Struct embedding, function map | Compile-time contract; clean extension; idiomatic Go |
| Agent resolution | `map[string]Agent` package var | Slice, factory, init() | O(1) lookup; natural dedup; `all` = sorted keys |
| `ApplyOverlay` return | `int` (exit code) | `error` | Matches `runApplyPolicyWithOptions` signature; avoids double error printing since policy already writes to stderr |
| Command paths | Hardcoded `commands/{name}.md` in dispatch helper | Per-agent `CommandPath()` method | Only OpenCode exists; YAGNI. Future agents add a method then |
| Recorder printing | Delegated to `ApplyOverlay` (via `runApplyPolicyWithOptions` → `printSummary`) | Caller-owned in `RunApplyCustom` | `printSummary()` already calls `recorder.print()` at line 74 of summary.go; keeps single responsibility |

## Data Flow

```
CLI args ──→ normalizeTargets ──→ resolve from agentRegistry
                                       │
                           ┌───────────┤
                           ▼           ▼
                    installSkills   renderCommands
                    (generic)      (agent.BuildCommandContent)
                           │           │
                           └─────┬─────┘
                                 ▼
                    agent.ApplyOverlay()
                           │
                    [OpenCode: runApplyPolicyWithOptions]
                           │
                    printSummary() + recorder.print()
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/overlay/agent.go` | Create | `Agent` interface, `agentRegistry` var, `registeredAgentNames()` sorted helper |
| `internal/overlay/opencode_agent.go` | Create | `OpenCodeAgent` struct; `Name()`, `BasePath()`, `BuildCommandContent()`, `ApplyOverlay()`. Extracts logic from both switch cases |
| `internal/overlay/apply_custom.go` | Modify | Delete: `buildCustomTarget`, `buildCustomCommandContent`, `shouldApplyGentleOverlay`, `supportedTargets`, `isSupportedTarget`. Refactor: `RunApplyCustom` dispatch loop, `normalizeTargets` validates against registry, `printApplyCustomUsage` lists only `opencode`/`all`. Keep: `customSkills`, shared types, `installSkill*`, `renderCustomCommand` (refactored to use Agent), `skillNameForCommand`, `validateCustomSources` |
| `cmd/gentle-ai-overlay/main.go` | Modify | Update `printMainUsage` to remove legacy target names |

Shell scripts (`apply-gentle-ai-custom.sh`/`.ps1`) are thin wrappers passing `$@` to the Go CLI — no code changes needed. Help text comes from Go.

## Interfaces / Contracts

```go
// internal/overlay/agent.go
type Agent interface {
    Name() string
    BasePath() string
    BuildCommandContent(cmd customCommand, body string) string
    ApplyOverlay(repoRoot string, options applyPolicyOptions) int
}

var agentRegistry = map[string]Agent{
    "opencode": &OpenCodeAgent{},
}

func registeredAgentNames() []string {
    names := make([]string, 0, len(agentRegistry))
    for k := range agentRegistry { names = append(names, k) }
    sort.Strings(names)
    return names
}
```

```go
// internal/overlay/opencode_agent.go
type OpenCodeAgent struct{}

func (a *OpenCodeAgent) Name() string { return "opencode" }

func (a *OpenCodeAgent) BasePath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "opencode")
}

func (a *OpenCodeAgent) BuildCommandContent(cmd customCommand, body string) string {
    // Exact current "case opencode" logic from buildCustomCommandContent
    lines := []string{
        "---",
        fmt.Sprintf("description: %s", cmd.description),
        "---", "",
        fmt.Sprintf("Read the skill file at `~/.config/opencode/skills/%s/SKILL.md` FIRST, then follow it exactly.", cmd.skillName),
        "", "CONTEXT:",
        "- Working directory: !`echo -n \"$(pwd)\"`",
        "- Current project: !`echo -n \"$(basename \"$(pwd)\")\"`",
        fmt.Sprintf("- Mode: %s", cmd.mode),
        fmt.Sprintf("- Command type: %s", cmd.commandType), "",
    }
    return strings.Join(lines, "\n") + body
}

func (a *OpenCodeAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
    return runApplyPolicyWithOptions(repoRoot, options)
}
```

## Refactored RunApplyCustom (key changes)

```go
func RunApplyCustom(repoRoot string, args []string) int {
    options, targets, exitCode := normalizeTargets(args)
    // ... recorder, sources, validate (unchanged) ...

    for _, name := range targets {
        agent := agentRegistry[name]
        if err := installAgentAssets(agent, sharedRoot, sources, options.recorder); err != nil {
            // error handling...
        }
        if code := agent.ApplyOverlay(repoRoot, applyPolicyOptions{
            verbose: options.verbose, recorder: options.recorder,
        }); code != 0 {
            return code
        }
    }
    fmt.Println("Reminder: ...")
    return 0
}

// installAgentAssets replaces applyCustomTarget — generic over Agent interface
func installAgentAssets(agent Agent, sharedRoot string, sources customSourceFiles, recorder *verboseRecorder) error {
    // 1. Install skills (uses agent.BasePath(), agent.Name()) — same loop as current
    // 2. Render commands: iterate commandDefs, build customCommand with
    //    fileRelPath = "commands/{name}.md", call agent.BuildCommandContent()
    // 3. Print "Applied {name} overlays -> {basePath}"
}
```

`normalizeTargets` changes: validates against `agentRegistry` keys instead of `supportedTargets` slice. Error: `"unsupported agent: %s"`. `all` expands to `registeredAgentNames()`.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `BuildCommandContent` parity | Golden strings: capture current switch output, compare byte-for-byte |
| Unit | `normalizeTargets` registry validation | Table-driven: `opencode`→ok, `gemini`→error, `all`→`["opencode"]` |
| Unit | `registeredAgentNames` ordering | Verify sorted output |

## Migration / Rollout

No migration required. `all` intentionally narrows to OpenCode only — accepted per proposal.

## Open Questions

- [ ] None
