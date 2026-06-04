package overlay

import "sort"

// Agent defines the behavior contract for a target installation agent.
// Each registered agent handles skill installation, command rendering, and
// overlay application for its specific AI tool.
type Agent interface {
	// Name returns the canonical identifier used in CLI targets and the registry.
	Name() string
	// BasePath returns the absolute path to the agent's configuration directory.
	BasePath() string
	// BuildCommandContent renders a command file's content from the given
	// customCommand metadata and body text. The returned string is written
	// verbatim to the command file on disk.
	BuildCommandContent(cmd customCommand, body string) string
	// ApplyOverlay runs the overlay pipeline for this agent and returns an exit
	// code (0 = success). The agent is responsible for printing its own
	// summary. recorder.print() is called by RunApplyCustom after the agent
	// loop completes — agents must not call it themselves.
	ApplyOverlay(repoRoot string, options applyPolicyOptions) int
}

// agentRegistry maps target names to their Agent implementation.
// Register new agents here once a second agent is required (YAGNI).
var agentRegistry = map[string]Agent{}

// registeredAgentNames returns the sorted list of registered agent names.
// Used by normalizeTargets when the caller requests the "all" target.
// CONTRACT: "all" processes agents in alphabetical order. No agent may depend
// on another agent having run first — agents must be fully independent.
func registeredAgentNames() []string {
	names := make([]string, 0, len(agentRegistry))
	for k := range agentRegistry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
