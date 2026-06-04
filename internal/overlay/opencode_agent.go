package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	agentRegistry["opencode"] = registeredAgent{
		agent:        &OpenCodeAgent{},
		skillTargets: []string{"~/.config/opencode/skills"},
	}
}

// OpenCodeAgent handles installation of custom overlays for OpenCode.
// It encapsulates the basePath resolution, command YAML rendering, and
// policy overlay invocation specific to the OpenCode tool.
type OpenCodeAgent struct{}

// Name returns the canonical registry key for this agent.
func (a *OpenCodeAgent) Name() string { return "opencode" }

// BasePath returns the OpenCode configuration directory (~/.config/opencode).
// Fails fast if the home directory cannot be resolved — a relative path would
// silently install files in the wrong location.
func (a *OpenCodeAgent) BasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot resolve home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "opencode")
}

// BuildCommandContent renders a command file with OpenCode's YAML frontmatter
// format. The output is byte-for-byte identical to the former "opencode" case
// in buildCustomCommandContent.
func (a *OpenCodeAgent) BuildCommandContent(cmd customCommand, body string) string {
	lines := []string{
		"---",
		fmt.Sprintf("description: %s", cmd.description),
		"---",
		"",
		fmt.Sprintf("Read the skill file at `~/.config/opencode/skills/%s/SKILL.md` FIRST, then follow it exactly.", cmd.skillName),
		"",
		"CONTEXT:",
		"- Working directory: !`echo -n \"$(pwd)\"`",
		"- Current project: !`echo -n \"$(basename \"$(pwd)\")\"`",
		fmt.Sprintf("- Mode: %s", cmd.mode),
		fmt.Sprintf("- Command type: %s", cmd.commandType),
		"",
	}
	return strings.Join(lines, "\n") + body
}

// ApplyOverlay runs the Gentle AI policy pipeline for OpenCode and returns
// its exit code. The pipeline handles orchestrator generation, opencode.json
// mutations, and prints its own summary via runApplyPolicyWithOptions.
func (a *OpenCodeAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
	return runApplyPolicyWithOptions(repoRoot, options)
}
