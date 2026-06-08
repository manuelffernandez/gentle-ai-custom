package overlay

import "sort"

// Agent defines the behavior contract for a target installation agent.
// Each registered agent handles skill installation, command rendering, and
// overlay application for its specific AI tool.
type Agent interface {
	// Name returns the canonical identifier used in CLI targets and the registry.
	Name() string
	// BasePath returns the absolute path to the agent's configuration directory.
	BasePath() (string, error)
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

// registeredAgent binds a CLI target to its implementation and the runtime
// skill directories it owns. Prune scope is derived from this registration,
// not from the global policy target list.
type registeredAgent struct {
	agent        Agent
	skillTargets []string
}

// agentRegistry maps target names to their registeredAgent metadata.
// Register new agents here once a second agent is required (YAGNI).
var agentRegistry = map[string]registeredAgent{}

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

// registeredSkillTargets resolves the runtime skill directories for the given
// registered agent names, preserving the selected agent order and deduplicating
// overlapping directories.
func registeredSkillTargets(agentNames []string) []string {
	seen := map[string]bool{}
	targets := make([]string, 0)
	for _, name := range agentNames {
		entry, ok := agentRegistry[name]
		if !ok {
			continue
		}
		for _, target := range entry.skillTargets {
			if target == "" || seen[target] {
				continue
			}
			seen[target] = true
			targets = append(targets, target)
		}
	}
	return targets
}
