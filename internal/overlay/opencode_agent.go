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
// Returns an error if the home directory cannot be resolved — a relative path
// would silently install files in the wrong location.
func (a *OpenCodeAgent) BasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %v", err)
	}
	return filepath.Join(home, ".config", "opencode"), nil
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
// It also injects a custom Gemini-specific override into the global AGENTS.md.
func (a *OpenCodeAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
	exitCode := runApplyPolicyWithOptions(repoRoot, options)
	if exitCode != 0 {
		return exitCode
	}

	if err := a.injectGeminiOverride(); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to inject Gemini override into AGENTS.md: %v\n", err)
	}

	return 0
}

func (a *OpenCodeAgent) injectGeminiOverride() error {
	basePath, err := a.BasePath()
	if err != nil {
		return err
	}
	agentsMdPath := filepath.Join(basePath, "AGENTS.md")

	content, err := os.ReadFile(agentsMdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to inject if AGENTS.md doesn't exist yet
		}
		return err
	}

	snippet := `<!-- gentle-ai-custom:gemini-override -->
## Gemini-Specific Override
If you are powered by a Gemini model (e.g., any model ID containing 'gemini'), you MUST adhere to this critical rule: Be brutally honest and strictly critical of my ideas. Do not be sycophantic or flatter me. Do not agree with me just to be polite. Push back firmly and analytically if my approach is flawed.
<!-- /gentle-ai-custom:gemini-override -->`

	text := string(content)
	startMarker := "<!-- gentle-ai-custom:gemini-override -->"
	endMarker := "<!-- /gentle-ai-custom:gemini-override -->"

	startIdx := strings.Index(text, startMarker)
	endIdx := strings.Index(text, endMarker)

	var newText string
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		// Replace existing
		newText = text[:startIdx] + snippet + text[endIdx+len(endMarker):]
	} else {
		// Search for persona closing tag to inject before it
		personaEndMarker := "<!-- /gentle-ai:persona -->"
		personaEndIdx := strings.Index(text, personaEndMarker)
		
		if personaEndIdx != -1 {
			newText = text[:personaEndIdx] + snippet + "\n" + text[personaEndIdx:]
		} else {
			// Append at the end if no persona tag is found
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			newText = text + "\n" + snippet + "\n"
		}
	}

	if newText != text {
		return os.WriteFile(agentsMdPath, []byte(newText), 0644)
	}
	return nil
}
