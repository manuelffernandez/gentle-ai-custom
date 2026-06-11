package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
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

	sharedSkillsRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedSharedSkills.Root))
	commandBodiesRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedCommandBodies.Root))
	commandBodyByKey := mapCommandBodiesByKey(commandBodiesRoot, target.RepoOwnedCommandBodies.Entries)

	if err := validateCustomSources(sharedSkillsRoot, customSkillNames(target), target.RepoOwnedCommandBodies.Entries, commandBodyByKey); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		options.recorder.print()
		return 1
	}
	if err := validateOwnedAssetSources(repoRoot, target.OwnedOverlayAssets); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		options.recorder.print()
		return 1
	}

	for _, targetName := range targets {
		entry := agentRegistry[targetName]
		agent := entry.agent
		if code := agent.ApplyOverlay(repoRoot, applyPolicyOptions{
			verbose:      options.verbose,
			recorder:     options.recorder,
			skillTargets: registeredSkillTargets([]string{targetName}),
		}); code != 0 {
			options.recorder.print()
			return code
		}
		if err := installAgentAssets(agent, target, sharedSkillsRoot, commandBodyByKey, options.recorder); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
		}
		if err := reconcileOwnedRuntimeOutputs(agent, repoRoot, target, sharedSkillsRoot, options.recorder); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
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

func validateCustomSources(sharedSkillsRoot string, skillNames []string, commandEntries []RepoOwnedCommandEntry, commandBodyByKey map[string]string) error {
	for _, skillName := range skillNames {
		src := filepath.Join(sharedSkillsRoot, skillName, "SKILL.md")
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("Missing source: %s", src)
		}
	}
	for _, entry := range commandEntries {
		source := commandBodyByKey[entry.Key]
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("Missing source: %s", source)
		}
	}
	return nil
}

func validateOwnedAssetSources(repoRoot string, assets []OwnedOverlayAsset) error {
	for _, asset := range assets {
		if asset.RepoOwnedPath == "" {
			// Upstream-source-only asset: no owned file; nothing to validate here.
			continue
		}
		src := ownedAssetSourcePath(repoRoot, asset)
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("Missing source: %s", src)
		}
	}
	return nil
}

// installAgentAssets is the generic replacement for the former applyCustomTarget
// + buildCustomTarget pair. It installs skills and renders command files for the
// given agent using only the Agent interface — no switch on agent name.
func installAgentAssets(agent Agent, target ManagedAssetsTarget, sharedSkillsRoot string, commandBodyByKey map[string]string, recorder *verboseRecorder) error {
	basePath, err := agent.BasePath()
	if err != nil {
		return fmt.Errorf("cannot resolve agent base path: %w", err)
	}

	for _, skillName := range customSkillNames(target) {
		src := filepath.Join(sharedSkillsRoot, skillName, "SKILL.md")
		if err := installSkill(basePath, skillName, src, agent.Name(), recorder); err != nil {
			return err
		}
		assetsDir := filepath.Join(sharedSkillsRoot, skillName, "assets")
		if err := installSkillAssets(basePath, skillName, assetsDir, agent.Name(), recorder); err != nil {
			return err
		}
	}

	for _, def := range target.RepoOwnedCommandBodies.Entries {
		cmd := customCommand{
			fileRelPath: filepath.Join("commands", def.Key+".md"),
			skillName:   def.OwnerSkill,
			mode:        def.Mode,
			commandType: def.CommandType,
			description: def.Description,
			bodyPath:    commandBodyByKey[def.Key],
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

func customSkillNames(target ManagedAssetsTarget) []string {
	skills := append([]string(nil), target.RepoOwnedSharedSkills.Allowlist...)
	sort.Strings(skills)
	return skills
}

func mapCommandBodiesByKey(commandBodiesRoot string, entries []RepoOwnedCommandEntry) map[string]string {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		bodyPath := entry.BodyPath
		if rel, err := filepath.Rel("shared/commands", filepath.ToSlash(bodyPath)); err == nil && !strings.HasPrefix(rel, "..") {
			bodyPath = filepath.Join(commandBodiesRoot, rel)
		} else {
			bodyPath = filepath.Join(filepath.Dir(commandBodiesRoot), filepath.FromSlash(bodyPath))
		}
		result[entry.Key] = bodyPath
	}
	return result
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
