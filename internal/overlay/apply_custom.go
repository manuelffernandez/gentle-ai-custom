package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedTargets = []string{"opencode", "claude", "codex", "gemini", "antigravity"}

// customSkills is the single source of truth for installable skills.
// To add a new skill: add its directory name here. If the skill has an
// assets/ subdirectory it will be installed automatically.
var customSkills = []string{
	"commit-planner",
	"pr-finalizer",
	"code-modularization",
	"package-security",
}

type customSourceFiles struct {
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

type applyCustomOptions struct {
	verbose  bool
	recorder *verboseRecorder
}

func RunApplyCustom(repoRoot string, args []string) int {
	options, targets, exitCode := normalizeTargets(args)
	if exitCode >= 0 {
		return exitCode
	}
	options.recorder = newVerboseRecorder(options.verbose)

	sharedRoot := filepath.Join(repoRoot, "shared")
	sources := customSourceFiles{
		planBody:         filepath.Join(sharedRoot, "commands", "commit-plan-body.md"),
		applyBody:        filepath.Join(sharedRoot, "commands", "commit-apply-body.md"),
		fastBody:         filepath.Join(sharedRoot, "commands", "commit-fast-body.md"),
		prCreateBody:     filepath.Join(sharedRoot, "commands", "pr-create-body.md"),
		prRegenerateBody: filepath.Join(sharedRoot, "commands", "pr-regenerate-body.md"),
	}

	if err := validateCustomSources(sharedRoot, sources); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	for _, targetName := range targets {
		if err := applyCustomTarget(buildCustomTarget(targetName, sources), sharedRoot, options.recorder); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
		}
	}

	if shouldApplyGentleOverlay(targets) {
		if code := runApplyPolicyWithOptions(repoRoot, applyPolicyOptions{verbose: options.verbose, recorder: options.recorder}); code != 0 {
			return code
		}
	} else {
		options.recorder.print()
	}

	fmt.Println("Reminder: run audit-gentle-ai-upstream before maintainer sync/reinstall work, and re-run this script after syncs, upgrades, or managed config refreshes.")
	return 0
}

func normalizeTargets(args []string) (applyCustomOptions, []string, int) {
	var options applyCustomOptions
	if len(args) == 0 {
		printApplyCustomUsage(os.Stderr)
		return options, nil, 1
	}

	var positional []string
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			printApplyCustomUsage(os.Stdout)
			return options, nil, 0
		case "--verbose":
			options.verbose = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown apply-custom flag: %s\n", arg)
				printApplyCustomUsage(os.Stderr)
				return options, nil, 1
			}
			positional = append(positional, arg)
		}
	}

	if len(positional) == 0 {
		printApplyCustomUsage(os.Stderr)
		return options, nil, 1
	}
	if len(positional) == 1 && positional[0] == "all" {
		return options, append([]string(nil), supportedTargets...), -1
	}

	seen := map[string]bool{}
	var result []string
	for _, target := range positional {
		switch target {
		case "all":
			fmt.Fprintln(os.Stderr, "Use 'all' by itself, or pass explicit targets only.")
			return options, nil, 1
		}
		if !isSupportedTarget(target) {
			fmt.Fprintf(os.Stderr, "Unknown target: %s\n", target)
			return options, nil, 1
		}
		if !seen[target] {
			seen[target] = true
			result = append(result, target)
		}
	}
	return options, result, -1
}

func printApplyCustomUsage(out *os.File) {
	prefix := usageCommandName("apply-custom")
	fmt.Fprintf(out, "Usage: %s [--verbose] all | [opencode|claude|codex|gemini|antigravity ...]\n", prefix)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintf(out, "  %s opencode\n", prefix)
	fmt.Fprintf(out, "  %s opencode --verbose\n", prefix)
	fmt.Fprintf(out, "  %s claude codex\n", prefix)
	fmt.Fprintf(out, "  %s gemini\n", prefix)
	fmt.Fprintf(out, "  %s antigravity\n", prefix)
	fmt.Fprintf(out, "  %s all\n", prefix)
}

func isSupportedTarget(target string) bool {
	for _, supported := range supportedTargets {
		if target == supported {
			return true
		}
	}
	return false
}

func validateCustomSources(sharedRoot string, sources customSourceFiles) error {
	for _, skillName := range customSkills {
		src := filepath.Join(sharedRoot, "skills", skillName, "SKILL.md")
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("Missing source: %s", src)
		}
	}
	for _, source := range []string{
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

func applyCustomTarget(target customTarget, sharedRoot string, recorder *verboseRecorder) error {
	if len(target.commands) == 0 {
		return fmt.Errorf("no commands defined for target %q", target.name)
	}
	for _, skillName := range customSkills {
		src := filepath.Join(sharedRoot, "skills", skillName, "SKILL.md")
		if err := installSkill(target.basePath, skillName, src, target.name, recorder); err != nil {
			return err
		}
		assetsDir := filepath.Join(sharedRoot, "skills", skillName, "assets")
		if err := installSkillAssets(target.basePath, skillName, assetsDir, target.name, recorder); err != nil {
			return err
		}
	}
	for _, command := range target.commands {
		if err := renderCustomCommand(target.basePath, command, target.name, recorder); err != nil {
			return err
		}
	}
	fmt.Printf(target.message+"\n", target.basePath)
	return nil
}

func installSkill(targetDir, skillName, skillSource, targetName string, recorder *verboseRecorder) error {
	destination := filepath.Join(targetDir, "skills", skillName, "SKILL.md")
	status, err := copyFileWithStatus(skillSource, destination)
	if err != nil {
		return err
	}
	if shouldRecordWriteStatus(status) {
		recorder.record(destination, fmt.Sprintf("installed %s skill for %s target (%s)", skillName, targetName, describeWriteStatus(status)))
	}
	return nil
}

func installSkillAssets(targetDir, skillName, assetsDir, targetName string, recorder *verboseRecorder) error {
	entries, err := os.ReadDir(assetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(assetsDir, entry.Name())
		dst := filepath.Join(targetDir, "skills", skillName, "assets", entry.Name())
		status, err := copyFileWithStatus(src, dst)
		if err != nil {
			return err
		}
		if shouldRecordWriteStatus(status) {
			recorder.record(dst, fmt.Sprintf("installed %s/%s asset for %s target (%s)", skillName, entry.Name(), targetName, describeWriteStatus(status)))
		}
	}
	return nil
}

func skillNameForCommand(command string) string {
	if strings.HasPrefix(command, "commit-") {
		return "commit-planner"
	}
	return "pr-finalizer"
}

func renderCustomCommand(targetDir string, command customCommand, targetName string, recorder *verboseRecorder) error {
	bodyRaw, err := os.ReadFile(command.bodyPath)
	if err != nil {
		return err
	}
	body := normalizeLF(string(bodyRaw))
	content := buildCustomCommandContent(command, body)
	destination := filepath.Join(targetDir, command.fileRelPath)
	status, err := writeTextFileWithStatus(destination, content)
	if err != nil {
		return err
	}
	if shouldRecordWriteStatus(status) {
		recorder.record(destination, fmt.Sprintf("rendered %s command for %s target (%s)", command.fileRelPath, targetName, describeWriteStatus(status)))
	}
	return nil
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
