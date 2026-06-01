package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedTargets = []string{"opencode", "claude", "codex", "gemini", "antigravity"}

type customSourceFiles struct {
	commitSkill      string
	prSkill          string
	planBody         string
	applyBody        string
	fastBody         string
	prCreateBody     string
	prRegenerateBody string
}

type customCommand struct {
	fileRelPath string
	renderer    string
	skillName   string
	commandName string
	mode        string
	commandType string
	description string
	bodyPath    string
}

type customTarget struct {
	name     string
	basePath string
	message  string
	commands []customCommand
}

func RunApplyCustom(repoRoot string, args []string) int {
	targets, exitCode := normalizeTargets(args)
	if exitCode >= 0 {
		return exitCode
	}

	sources := customSourceFiles{
		commitSkill:      filepath.Join(repoRoot, "shared", "skills", "commit-planner", "SKILL.md"),
		prSkill:          filepath.Join(repoRoot, "shared", "skills", "pr-finalizer", "SKILL.md"),
		planBody:         filepath.Join(repoRoot, "shared", "commands", "commit-plan-body.md"),
		applyBody:        filepath.Join(repoRoot, "shared", "commands", "commit-apply-body.md"),
		fastBody:         filepath.Join(repoRoot, "shared", "commands", "commit-fast-body.md"),
		prCreateBody:     filepath.Join(repoRoot, "shared", "commands", "pr-create-body.md"),
		prRegenerateBody: filepath.Join(repoRoot, "shared", "commands", "pr-regenerate-body.md"),
	}

	if err := validateCustomSources(sources); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	for _, targetName := range targets {
		if err := applyCustomTarget(buildCustomTarget(targetName, sources)); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
	}

	if shouldApplyGentleOverlay(targets) {
		if code := RunApplyPolicy(repoRoot); code != 0 {
			return code
		}
	}

	fmt.Println("Reminder: run audit-gentle-ai-upstream before maintainer sync/reinstall work, and re-run this script after syncs, upgrades, or managed config refreshes.")
	return 0
}

func normalizeTargets(args []string) ([]string, int) {
	if len(args) == 0 {
		printApplyCustomUsage(os.Stderr)
		return nil, 1
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printApplyCustomUsage(os.Stderr)
		return nil, 0
	}
	if len(args) == 1 && args[0] == "all" {
		return append([]string(nil), supportedTargets...), -1
	}

	seen := map[string]bool{}
	var result []string
	for _, target := range args {
		switch target {
		case "-h", "--help":
			printApplyCustomUsage(os.Stderr)
			return nil, 0
		case "all":
			fmt.Fprintln(os.Stderr, "Use 'all' by itself, or pass explicit targets only.")
			return nil, 1
		}
		if !isSupportedTarget(target) {
			fmt.Fprintf(os.Stderr, "Unknown target: %s\n", target)
			return nil, 1
		}
		if !seen[target] {
			seen[target] = true
			result = append(result, target)
		}
	}
	return result, -1
}

func printApplyCustomUsage(out *os.File) {
	fmt.Fprintf(out, "Usage: %s all | [opencode|claude|codex|gemini|antigravity ...]\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintf(out, "  %s opencode\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "  %s claude codex\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "  %s gemini\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "  %s antigravity\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "  %s all\n", filepath.Base(os.Args[0]))
}

func isSupportedTarget(target string) bool {
	for _, supported := range supportedTargets {
		if target == supported {
			return true
		}
	}
	return false
}

func validateCustomSources(sources customSourceFiles) error {
	for _, source := range []string{
		sources.commitSkill,
		sources.prSkill,
		sources.planBody,
		sources.applyBody,
		sources.fastBody,
		sources.prCreateBody,
		sources.prRegenerateBody,
	} {
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("Missing source: %s", source)
		}
	}
	return nil
}

func buildCustomTarget(name string, sources customSourceFiles) customTarget {
	commandDefs := []struct {
		name        string
		relPath     string
		mode        string
		commandType string
		description string
		bodyPath    string
	}{
		{name: "commit-plan", mode: "plan", commandType: "read-only", description: "Propose a post-SDD commit plan without changing git state", bodyPath: sources.planBody},
		{name: "commit-apply", mode: "apply", commandType: "state-changing", description: "Execute an approved post-SDD commit plan, or generate one first if missing", bodyPath: sources.applyBody},
		{name: "commit-fast", mode: "auto", commandType: "state-changing", description: "Generate and execute a commit plan in one shot without approval pause", bodyPath: sources.fastBody},
		{name: "pr-create", mode: "create", commandType: "state-changing", description: "Draft a PR from committed changes and optionally create it after approval", bodyPath: sources.prCreateBody},
		{name: "pr-regenerate", mode: "regenerate", commandType: "state-changing", description: "Regenerate or update an existing PR from the current committed diff after approval", bodyPath: sources.prRegenerateBody},
	}

	var target customTarget
	switch name {
	case "opencode":
		target = customTarget{name: name, basePath: filepath.Join(expandUser("~/.config"), "opencode"), message: "Applied OpenCode overlays -> %s", commands: make([]customCommand, 0, len(commandDefs))}
		for _, def := range commandDefs {
			target.commands = append(target.commands, customCommand{fileRelPath: filepath.Join("commands", def.name+".md"), renderer: "opencode", skillName: skillNameForCommand(def.name), mode: def.mode, commandType: def.commandType, description: def.description, bodyPath: def.bodyPath})
		}
	case "claude":
		target = customTarget{name: name, basePath: expandUser("~/.claude"), message: "Applied Claude overlays -> %s", commands: make([]customCommand, 0, len(commandDefs))}
		for _, def := range commandDefs {
			target.commands = append(target.commands, customCommand{fileRelPath: filepath.Join("commands", def.name+".md"), renderer: "claude", skillName: skillNameForCommand(def.name), mode: def.mode, commandType: def.commandType, description: def.description, bodyPath: def.bodyPath})
		}
	case "codex":
		target = customTarget{name: name, basePath: expandUser("~/.codex"), message: "Applied Codex overlays -> %s", commands: make([]customCommand, 0, len(commandDefs))}
		for _, def := range commandDefs {
			target.commands = append(target.commands, customCommand{fileRelPath: filepath.Join("prompts", def.name+".md"), renderer: "codex", skillName: skillNameForCommand(def.name), mode: def.mode, commandType: def.commandType, description: def.description, bodyPath: def.bodyPath})
		}
	case "gemini":
		target = customTarget{name: name, basePath: expandUser("~/.gemini"), message: "Applied Gemini overlays -> %s", commands: make([]customCommand, 0, len(commandDefs))}
		for _, def := range commandDefs {
			target.commands = append(target.commands, customCommand{fileRelPath: filepath.Join("skills", def.name, "SKILL.md"), renderer: "gemini", skillName: skillNameForCommand(def.name), commandName: def.name, mode: def.mode, commandType: def.commandType, description: def.description, bodyPath: def.bodyPath})
		}
	case "antigravity":
		target = customTarget{name: name, basePath: filepath.Join(expandUser("~/.gemini"), "antigravity"), message: "Applied Antigravity overlays -> %s", commands: make([]customCommand, 0, len(commandDefs))}
		for _, def := range commandDefs {
			target.commands = append(target.commands, customCommand{fileRelPath: filepath.Join("skills", def.name, "SKILL.md"), renderer: "antigravity", skillName: skillNameForCommand(def.name), commandName: def.name, mode: def.mode, commandType: def.commandType, description: def.description, bodyPath: def.bodyPath})
		}
	}
	return target
}

func applyCustomTarget(target customTarget) error {
	sharedRoot := filepath.Dir(filepath.Dir(target.commands[0].bodyPath))
	skillSources := map[string]string{
		"commit-planner":      filepath.Join(sharedRoot, "skills", "commit-planner", "SKILL.md"),
		"pr-finalizer":        filepath.Join(sharedRoot, "skills", "pr-finalizer", "SKILL.md"),
		"code-modularization": filepath.Join(sharedRoot, "skills", "code-modularization", "SKILL.md"),
	}
	if err := installSkill(target.basePath, "commit-planner", skillSources["commit-planner"]); err != nil {
		return err
	}
	if err := installSkill(target.basePath, "pr-finalizer", skillSources["pr-finalizer"]); err != nil {
		return err
	}
	if err := installSkill(target.basePath, "code-modularization", skillSources["code-modularization"]); err != nil {
		return err
	}
	for _, command := range target.commands {
		if err := renderCustomCommand(target.basePath, command); err != nil {
			return err
		}
	}
	fmt.Printf(target.message+"\n", target.basePath)
	return nil
}

func installSkill(targetDir, skillName, skillSource string) error {
	return copyFile(skillSource, filepath.Join(targetDir, "skills", skillName, "SKILL.md"))
}

func skillNameForCommand(command string) string {
	if strings.HasPrefix(command, "commit-") {
		return "commit-planner"
	}
	return "pr-finalizer"
}

func renderCustomCommand(targetDir string, command customCommand) error {
	bodyRaw, err := os.ReadFile(command.bodyPath)
	if err != nil {
		return err
	}
	body := normalizeLF(string(bodyRaw))
	content := buildCustomCommandContent(command, body)
	return writeTextFile(filepath.Join(targetDir, command.fileRelPath), content)
}

func buildCustomCommandContent(command customCommand, body string) string {
	var lines []string
	switch command.renderer {
	case "opencode":
		lines = []string{
			"---",
			fmt.Sprintf("description: %s", command.description),
			"---",
			"",
			fmt.Sprintf("Read the skill file at `~/.config/opencode/skills/%s/SKILL.md` FIRST, then follow it exactly.", command.skillName),
			"",
			"CONTEXT:",
			"- Working directory: !`echo -n \"$(pwd)\"`",
			"- Current project: !`echo -n \"$(basename \"$(pwd)\")\"`",
			fmt.Sprintf("- Mode: %s", command.mode),
			fmt.Sprintf("- Command type: %s", command.commandType),
			"",
		}
	case "claude":
		lines = []string{
			"---",
			fmt.Sprintf("description: %s", command.description),
			"argument-hint: [optional-context]",
			"allowed-tools:",
			"  - Read",
			"  - Glob",
			"  - Bash(git:*)",
			"  - Bash(gh:*)",
			"  - Bash(pwd:*)",
			"  - Bash(basename:*)",
		}
		if command.mode == "apply" || command.mode == "auto" {
			lines = append(lines, "disable-model-invocation: true")
		}
		lines = append(lines,
			"---",
			"",
			fmt.Sprintf("Read the skill file at `~/.claude/skills/%s/SKILL.md` FIRST, then follow it exactly.", command.skillName),
			"",
			"CONTEXT:",
			"- Working directory: !`pwd`",
			"- Current project: !`basename \"$PWD\"`",
			fmt.Sprintf("- Mode: %s", command.mode),
			fmt.Sprintf("- Command type: %s", command.commandType),
			"",
		)
	case "codex":
		lines = []string{
			"---",
			fmt.Sprintf("description: %s", command.description),
			"argument-hint: [optional-context]",
			"allowed-tools:",
			"  - Read",
			"  - Glob",
			"  - Bash(git:*)",
			"  - Bash(gh:*)",
			"  - Bash(pwd:*)",
			"  - Bash(basename:*)",
			"---",
			"",
			fmt.Sprintf("Read the skill file at `~/.codex/skills/%s/SKILL.md` FIRST, then follow it exactly.", command.skillName),
			"",
			"CONTEXT:",
			"- Working directory: !`pwd`",
			"- Current project: !`basename \"$PWD\"`",
			fmt.Sprintf("- Mode: %s", command.mode),
			fmt.Sprintf("- Command type: %s", command.commandType),
			"",
		}
	case "gemini":
		lines = []string{
			"---",
			fmt.Sprintf("name: %s", command.commandName),
			fmt.Sprintf("description: %s", command.description),
			"---",
			"",
			fmt.Sprintf("Read the skill file at `~/.gemini/skills/%s/SKILL.md` FIRST, then follow it exactly.", command.skillName),
			"",
			"CONTEXT:",
			"- Working directory: !`pwd`",
			"- Current project: !`basename \"$PWD\"`",
			fmt.Sprintf("- Mode: %s", command.mode),
			fmt.Sprintf("- Command type: %s", command.commandType),
			"",
		}
	case "antigravity":
		lines = []string{
			"---",
			fmt.Sprintf("name: %s", command.commandName),
			fmt.Sprintf("description: %s", command.description),
			"---",
			"",
			fmt.Sprintf("Read the skill file at `~/.gemini/antigravity/skills/%s/SKILL.md` FIRST, then follow it exactly.", command.skillName),
			"",
			"CONTEXT:",
			"- Working directory: !`pwd`",
			"- Current project: !`basename \"$PWD\"`",
			fmt.Sprintf("- Mode: %s", command.mode),
			fmt.Sprintf("- Command type: %s", command.commandType),
			"",
		}
	}
	return strings.Join(lines, "\n") + body
}

func shouldApplyGentleOverlay(targets []string) bool {
	for _, target := range targets {
		if target == "opencode" || target == "claude" {
			return true
		}
	}
	return false
}
