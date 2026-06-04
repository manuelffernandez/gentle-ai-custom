package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// customSkills is the single source of truth for installable skills.
// To add a new skill: add its directory name here. If the skill has an
// assets/ subdirectory it will be installed automatically.
var customSkills = []string{
	"commit-planner",
	"pr-finalizer",
	"code-design",
	"package-security",
}

type customSourceFiles struct {
	planBody         string
	applyBody        string
	fastBody         string
	prCreateBody     string
	prRegenerateBody string
}

// customCommand holds the metadata for a single command file to be rendered.
// The fileRelPath, skillName, mode, commandType, and description fields are
// agent-agnostic; the agent decides how to format them into the final content.
type customCommand struct {
	fileRelPath string
	skillName   string
	mode        string
	commandType string
	description string
	bodyPath    string
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

	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}
	if err := runVersionPreflight(repoRoot, policy, os.Stdin, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}

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
		options.recorder.print()
		return 1
	}

	for _, targetName := range targets {
		agent := agentRegistry[targetName]
		if err := installAgentAssets(agent, sources, sharedRoot, options.recorder); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
		}
		if code := agent.ApplyOverlay(repoRoot, applyPolicyOptions{verbose: options.verbose, recorder: options.recorder}); code != 0 {
			options.recorder.print()
			return code
		}
	}

	options.recorder.print()
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
		return options, registeredAgentNames(), -1
	}

	seen := map[string]bool{}
	var result []string
	for _, target := range positional {
		if target == "all" {
			fmt.Fprintln(os.Stderr, "Use 'all' by itself, or pass explicit targets only.")
			return options, nil, 1
		}
		if _, ok := agentRegistry[target]; !ok {
			fmt.Fprintf(os.Stderr, "unsupported agent: %s\n", target)
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
	names := registeredAgentNames()
	targets := strings.Join(names, " | ")
	fmt.Fprintf(out, "Usage: %s [--verbose] all | %s\n", prefix, targets)
	fmt.Fprintln(out, "Examples:")
	for _, name := range names {
		fmt.Fprintf(out, "  %s %s\n", prefix, name)
		fmt.Fprintf(out, "  %s %s --verbose\n", prefix, name)
	}
	fmt.Fprintf(out, "  %s all\n", prefix)
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

// installAgentAssets is the generic replacement for the former applyCustomTarget
// + buildCustomTarget pair. It installs skills and renders command files for the
// given agent using only the Agent interface — no switch on agent name.
func installAgentAssets(agent Agent, sources customSourceFiles, sharedRoot string, recorder *verboseRecorder) error {
	// Command definitions are agent-agnostic; each agent formats them differently
	// via BuildCommandContent. fileRelPath is hardcoded to commands/{name}.md
	// (YAGNI — revisit when a second agent with a different layout is added).
	commandDefs := []struct {
		name        string
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

	basePath := agent.BasePath()

	for _, skillName := range customSkills {
		src := filepath.Join(sharedRoot, "skills", skillName, "SKILL.md")
		if err := installSkill(basePath, skillName, src, agent.Name(), recorder); err != nil {
			return err
		}
		assetsDir := filepath.Join(sharedRoot, "skills", skillName, "assets")
		if err := installSkillAssets(basePath, skillName, assetsDir, agent.Name(), recorder); err != nil {
			return err
		}
	}

	for _, def := range commandDefs {
		cmd := customCommand{
			fileRelPath: filepath.Join("commands", def.name+".md"),
			skillName:   skillNameForCommand(def.name),
			mode:        def.mode,
			commandType: def.commandType,
			description: def.description,
			bodyPath:    def.bodyPath,
		}
		if err := renderCustomCommand(basePath, agent, cmd, recorder); err != nil {
			return err
		}
	}

	fmt.Printf("Applied %s overlays -> %s\n", agent.Name(), basePath)
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

// renderCustomCommand reads the body file, delegates content rendering to the
// agent, and writes the result to the command file under the agent's basePath.
func renderCustomCommand(targetDir string, agent Agent, command customCommand, recorder *verboseRecorder) error {
	bodyRaw, err := os.ReadFile(command.bodyPath)
	if err != nil {
		return err
	}
	body := normalizeLF(string(bodyRaw))
	content := agent.BuildCommandContent(command, body)
	destination := filepath.Join(targetDir, command.fileRelPath)
	status, err := writeTextFileWithStatus(destination, content)
	if err != nil {
		return err
	}
	if shouldRecordWriteStatus(status) {
		recorder.record(destination, fmt.Sprintf("rendered %s command for %s target (%s)", command.fileRelPath, agent.Name(), describeWriteStatus(status)))
	}
	return nil
}
