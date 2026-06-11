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
// It also injects custom rules into the global AGENTS.md.
func (a *OpenCodeAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
	exitCode := runApplyPolicyWithOptions(repoRoot, options)
	if exitCode != 0 {
		return exitCode
	}

	injected, destPath, err := a.injectCustomRules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to inject custom rules into AGENTS.md: %v\n", err)
	} else if injected {
		fmt.Printf("  injected custom rules -> %s\n", destPath)
	} else if destPath != "" {
		fmt.Printf("  custom rules verified -> %s\n", destPath)
	}

	return 0
}

func (a *OpenCodeAgent) injectCustomRules() (bool, string, error) {
	basePath, err := a.BasePath()
	if err != nil {
		return false, "", err
	}
	agentsMdPath := filepath.Join(basePath, "AGENTS.md")

	content, err := os.ReadFile(agentsMdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil // Nothing to inject if AGENTS.md doesn't exist yet
		}
		return false, agentsMdPath, err
	}

	text := string(content)

	// 1. Inject No-Auto-Commit global rule
	commitRuleSnippet := `<!-- gentle-ai-custom:no-auto-commit -->
- NEVER commit, push, or create pull requests unless explicitly requested by the user. Do not assume permission to commit even if a task is complete.
<!-- /gentle-ai-custom:no-auto-commit -->`
	
	commitRuleStart := "<!-- gentle-ai-custom:no-auto-commit -->"
	commitRuleEnd := "<!-- /gentle-ai-custom:no-auto-commit -->"
	
	if startIdx := strings.Index(text, commitRuleStart); startIdx != -1 {
		if endIdx := strings.Index(text, commitRuleEnd); endIdx != -1 && endIdx > startIdx {
			text = text[:startIdx] + commitRuleSnippet + text[endIdx+len(commitRuleEnd):]
		}
	} else {
		// Inject right after ## Rules
		rulesMarker := "## Rules\n"
		if rulesIdx := strings.Index(text, rulesMarker); rulesIdx != -1 {
			insertPos := rulesIdx + len(rulesMarker)
			text = text[:insertPos] + "\n" + commitRuleSnippet + "\n" + text[insertPos:]
		}
	}

	// 2. Inject Gemini-Specific Override
	geminiSnippet := `<!-- gentle-ai-custom:gemini-override -->
## Gemini-Specific Override
If you are powered by a Gemini model (e.g., any model ID containing 'gemini'), you MUST adhere to this critical rule: Be brutally honest and strictly critical of my ideas. Do not be sycophantic or flatter me. Do not agree with me just to be polite. Push back firmly and analytically if my approach is flawed.
<!-- /gentle-ai-custom:gemini-override -->`

	geminiStart := "<!-- gentle-ai-custom:gemini-override -->"
	geminiEnd := "<!-- /gentle-ai-custom:gemini-override -->"

	if startIdx := strings.Index(text, geminiStart); startIdx != -1 {
		if endIdx := strings.Index(text, geminiEnd); endIdx != -1 && endIdx > startIdx {
			text = text[:startIdx] + geminiSnippet + text[endIdx+len(geminiEnd):]
		}
	} else {
		// Search for persona closing tag to inject before it
		personaEndMarker := "<!-- /gentle-ai:persona -->"
		if personaEndIdx := strings.Index(text, personaEndMarker); personaEndIdx != -1 {
			text = text[:personaEndIdx] + geminiSnippet + "\n" + text[personaEndIdx:]
		} else {
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			text = text + "\n" + geminiSnippet + "\n"
		}
	}

	if text != string(content) {
		if err := os.WriteFile(agentsMdPath, []byte(text), 0644); err != nil {
			return false, agentsMdPath, err
		}
		return true, agentsMdPath, nil
	}
	
	return false, agentsMdPath, nil
}
