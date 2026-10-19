package overlay

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	if err := runVersionPreflight(repoRoot, policy, os.Stdin, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}
	manifest, _, err := loadManagedAssetsManifest(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}
	if err := validateTargetParity(manifest, registeredAgentNames()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}
	for _, targetName := range targets {
		if targetName != "codex" {
			continue
		}
		if _, _, err := prepareCodexOverlay(repoRoot, applyPolicyOptions{verbose: options.verbose, recorder: options.recorder}); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
		}
	}
	chainedCleanup, err := preflightChainedReferenceCleanup(repoRoot, targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		options.recorder.print()
		return 1
	}

	for _, targetName := range targets {
		target, _, err := loadManagedAssetsTarget(repoRoot, targetName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			options.recorder.print()
			return 1
		}
		if err := validateRuntimeOwnershipBoundary(target); err != nil {
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
		if path := chainedCleanup[targetName]; path != "" {
			if err := requireSanitizedChainedSkill(repoRoot, targetName, target); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				options.recorder.print()
				return 1
			}
			if err := removeApprovedChainedReference(path, repoRoot, options.recorder); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				options.recorder.print()
				return 1
			}
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

const chainedLegacyReference = "chained-pr/references/chaining-details.md"
const chainedApprovedReference = "overlay/gentle-ai/assets/upstream/opencode/skills/chained-pr/references/chaining-details.md"

// This is a one-file migration for the two selected sanitized skill copies,
// not an owned-state claim or a generic upstream skill pruning rule.
func preflightChainedReferenceCleanup(repoRoot string, targets []string) (map[string]string, error) {
	result := map[string]string{}
	for _, name := range targets {
		if name != "opencode" && name != "codex" {
			continue
		}
		agent := agentRegistry[name].agent
		basePath, err := agent.BasePath()
		if err != nil {
			return nil, err
		}
		path := filepath.Join(basePath, "skills", filepath.FromSlash(chainedLegacyReference))
		if err := inspectChainedReferencePath(path); err != nil {
			return nil, err
		}
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing unsafe chained-pr reference %s: not a regular file", path)
		}
		installedSkill := filepath.Join(basePath, "skills", "chained-pr", "SKILL.md")
		installedInfo, err := os.Lstat(installedSkill)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil && !installedInfo.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing chained-pr reference cleanup: installed skill %s is not a regular file", installedSkill)
		}
		target, _, err := loadManagedAssetsTarget(repoRoot, name)
		if err != nil {
			return nil, err
		}
		if err := requireSanitizedChainedSkill(repoRoot, name, target); err != nil {
			return nil, err
		}
		if err := verifyApprovedChainedReference(repoRoot, path); err != nil {
			return nil, err
		}
		result[name] = path
	}
	return result, nil
}

func inspectChainedReferencePath(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := ensureNoSymlinkParents(home, path); err != nil {
		return fmt.Errorf("refusing unsafe chained-pr reference %s: %w", path, err)
	}
	return nil
}

func requireSanitizedChainedSkill(repoRoot, name string, target ManagedAssetsTarget) error {
	wantRoot := map[string]string{"opencode": "~/.config/opencode/skills", "codex": "~/.codex/skills"}[name]
	if target.RepoOwnedSharedSkills.Root != "shared/skills" || target.RepoOwnedSharedSkills.RuntimeRoot != wantRoot ||
		!slices.Contains(target.RepoOwnedSharedSkills.Allowlist, "chained-pr") {
		return fmt.Errorf("refusing chained-pr reference cleanup: %s does not install the sanitized skill", name)
	}
	source := filepath.Join(repoRoot, "shared", "skills", "chained-pr", "SKILL.md")
	asset := filepath.Join(repoRoot, "shared", "skills", "chained-pr", "assets", "chaining-details.md")
	for _, path := range []string{source, asset} {
		if err := ensureNoSymlinkParents(repoRoot, path); err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("sanitized chained-pr source %s is not a regular file", path)
		}
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if !bytes.Contains(content, []byte("[assets/chaining-details.md](assets/chaining-details.md)")) ||
		bytes.Contains(content, []byte("references/chaining-details.md")) {
		return fmt.Errorf("refusing chained-pr reference cleanup: sanitized source no longer uses only its own asset")
	}
	return nil
}

func verifyApprovedChainedReference(repoRoot, path string) error {
	approvedPath := filepath.Join(repoRoot, filepath.FromSlash(chainedApprovedReference))
	if err := ensureNoSymlinkParents(repoRoot, approvedPath); err != nil {
		return err
	}
	info, err := os.Lstat(approvedPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("approved upstream chained-pr snapshot %s is not a regular file", approvedPath)
	}
	approved, err := os.ReadFile(approvedPath)
	if err != nil {
		return err
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, approved) {
		return fmt.Errorf("refusing chained-pr reference cleanup: %s differs from approved upstream snapshot", path)
	}
	return nil
}

func removeApprovedChainedReference(path, repoRoot string, recorder *verboseRecorder) error {
	if err := inspectChainedReferencePath(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing unsafe chained-pr reference %s: not a regular file", path)
	}
	if err := verifyApprovedChainedReference(repoRoot, path); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	recorder.record(path, "removed exact obsolete chained-pr reference after sanitized skill install")
	return nil
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
		if asset.Class != "repo_owned_runtime" {
			if asset.RepoOwnedPath != "" {
				return fmt.Errorf("managed asset %q is class %q but declares repo_owned_path; upstream snapshots cannot be runtime install sources", asset.Key, asset.Class)
			}
			continue
		}
		if asset.RepoOwnedPath == "" {
			return fmt.Errorf("repo-owned runtime asset %q does not declare repo_owned_path", asset.Key)
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
