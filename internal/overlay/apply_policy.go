package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// applyPolicyState holds all mutable state accumulated while running the apply
// pipeline. It is created once in RunApplyPolicy and threaded through each phase.
type applyPolicyState struct {
	policy               Policy
	verbose              bool
	recorder             *verboseRecorder
	configPath           string
	generatedDir         string
	repoSnapshotDir      string
	localSnapshotDir     string
	localProfilesPath    string
	repoSnapshotMetaFile string
	repoSnapshotBaseFile string
	repoSnapshotBaseline string
	agents               map[string]any
	configData           map[string]any
	configChanged        bool
	prunedCount          int
	missingKeepSummary   []string
	createdOverrides     []string
	generatedCount       int
	recoveredCount       int
	keptCount            int
	skippedCount         int
	repoSnapshots        snapshotCounters
	localSnapshots       snapshotCounters
	localSnapshotMigrate int
	repoSnapshotBackfill int
	topologyWarnings     []string
	writtenOrchestrators map[string]bool
	profilesManagedCount int
	profileAgentsCreated int
	profileAgentsUpdated int
	profileAgentsSame    int
	managedProfiles      map[string]bool
	unmanagedProfiles    []string
	baseRuntimePrompt    string
	baseGeneratedPath    string
	originalAgentKeys    map[string]bool
	state                UpstreamState
	sddPhasesSet         map[string]bool
}

type applyPolicyOptions struct {
	verbose  bool
	recorder *verboseRecorder
}

// RunApplyPolicy is the main entrypoint for the standalone `apply-policy`
// subcommand. It loads the policy, builds initial state, and runs each phase
// of the apply pipeline in order.
func RunApplyPolicy(repoRoot string, args []string) int {
	options, exitCode := normalizeApplyPolicyArgs(args)
	if exitCode >= 0 {
		return exitCode
	}
	options.recorder = newVerboseRecorder(options.verbose)
	code := runApplyPolicyWithOptions(repoRoot, options)
	options.recorder.print()
	return code
}

func runApplyPolicyWithOptions(repoRoot string, options applyPolicyOptions) int {
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	state := &applyPolicyState{
		policy:               policy,
		verbose:              options.verbose,
		recorder:             options.recorder,
		configPath:           expandUser(policy.OpenCode.ConfigPath),
		generatedDir:         expandUser(policy.OpenCode.GeneratedOrchestratorsDir),
		repoSnapshotDir:      filepath.Join(repoRoot, policy.OpenCode.OrchestratorSnapshotDir),
		localSnapshotDir:     expandUser(policy.OpenCode.LocalOrchestratorSnapshotDir),
		localProfilesPath:    expandUser(policy.OpenCode.SDDProfilesLocalConfigPath),
		repoSnapshotMetaFile: filepath.Join(repoRoot, policy.OpenCode.OrchestratorSnapshotMetadata),
		writtenOrchestrators: map[string]bool{},
		managedProfiles:      map[string]bool{},
		originalAgentKeys:    map[string]bool{},
		sddPhasesSet:         map[string]bool{},
	}
	// options.recorder must be initialized by the caller before invoking this
	// function. Both RunApplyPolicy and OpenCodeAgent.ApplyOverlay guarantee
	// this. A nil recorder here would silently accumulate verbose entries into
	// an unreachable object — use a panic to catch any future caller that
	// forgets to initialize it.
	if state.recorder == nil {
		panic("runApplyPolicyWithOptions: options.recorder must be non-nil; caller must initialize it")
	}
	fail := func(err error) int {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	state.repoSnapshotBaseFile = filepath.Join(state.repoSnapshotDir, policy.OpenCode.BaseOrchestratorKey+".last.md")
	for _, phase := range policy.OpenCode.SDDPhases {
		state.sddPhasesSet[phase] = true
	}

	fmt.Println("Applying Gentle AI overlay policy...")
	if err := state.pruneSkills(); err != nil {
		return fail(err)
	}

	if !pathExists(state.configPath) {
		fmt.Printf("- skip missing OpenCode config: %s\n", state.configPath)
		state.printMissingOpenCodeSummary()
		return 0
	}

	if err := state.loadOpenCodeConfig(); err != nil {
		return fail(err)
	}
	if err := state.loadAuditedBaseline(repoRoot); err != nil {
		return fail(err)
	}

	state.applyAgentOverrides()
	if err := state.reconcileProfiles(); err != nil {
		return fail(err)
	}
	state.detectTopologyDrift()
	if err := state.generateOverlays(); err != nil {
		return fail(err)
	}
	if err := state.persistConfig(); err != nil {
		return fail(err)
	}
	if err := state.verifyPersistedState(); err != nil {
		return fail(err)
	}
	state.printSummary()
	return 0
}

func normalizeApplyPolicyArgs(args []string) (applyPolicyOptions, int) {
	var options applyPolicyOptions
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			printApplyPolicyUsage(os.Stdout)
			return options, 0
		case "--verbose":
			options.verbose = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown apply-policy flag: %s\n", arg)
			} else {
				fmt.Fprintf(os.Stderr, "apply-policy does not accept positional argument: %s\n", arg)
			}
			printApplyPolicyUsage(os.Stderr)
			return options, 1
		}
	}
	return options, -1
}

func printApplyPolicyUsage(out *os.File) {
	fmt.Fprintf(out, "Usage: %s [--verbose]\n", usageCommandName("apply-policy"))
}

// --- Skills ---

func (s *applyPolicyState) pruneSkills() error {
	for _, targetDirRaw := range s.policy.Skills.Targets {
		targetDir := expandUser(targetDirRaw)
		info, err := os.Stat(targetDir)
		if err != nil || !info.IsDir() {
			fmt.Printf("- skip missing skills dir: %s\n", targetDir)
			continue
		}

		fmt.Printf("- pruning unwanted skills in %s\n", targetDir)
		for _, skill := range s.policy.Skills.Prune {
			skillPath := filepath.Join(targetDir, skill)
			if pathExists(skillPath) {
				if err := os.RemoveAll(skillPath); err != nil {
					return fmt.Errorf("failed to remove skill %s: %w", skillPath, err)
				}
				s.prunedCount++
				s.recordVerbose(skillPath, fmt.Sprintf("removed pruned skill directory %q", skill))
				fmt.Printf("  removed %s\n", skill)
			} else {
				fmt.Printf("  already absent %s\n", skill)
			}
		}

		for _, keep := range s.policy.Skills.Keep {
			if !pathExists(filepath.Join(targetDir, keep)) {
				s.missingKeepSummary = append(s.missingKeepSummary, targetDir+" -> "+keep)
			}
		}
	}
	return nil
}

// --- Config loading ---

func (s *applyPolicyState) loadOpenCodeConfig() error {
	raw, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("Cannot read OpenCode config at %s: %v", s.configPath, err)
	}
	if err := json.Unmarshal(raw, &s.configData); err != nil {
		home, _ := os.UserHomeDir()
		return fmt.Errorf("OpenCode config at %s is not valid JSON: %v. Restore it from a backup under %s/.gentle-ai/backups/ or re-run `gentle-ai sync` to regenerate it.", s.configPath, err, home)
	}
	agents, ok := jsonObject(s.configData["agent"])
	if !ok {
		return fmt.Errorf("OpenCode config does not contain an agent map")
	}
	s.agents = agents
	for key := range agents {
		s.originalAgentKeys[key] = true
	}
	return nil
}

func (s *applyPolicyState) loadAuditedBaseline(repoRoot string) error {
	if _, err := os.Stat(s.repoSnapshotBaseFile); err != nil {
		return fmt.Errorf("audited base snapshot missing for orchestrator %q at %s. Restore the committed baseline before re-running apply.", s.policy.OpenCode.BaseOrchestratorKey, s.repoSnapshotBaseFile)
	}
	baseline, err := readText(s.repoSnapshotBaseFile)
	if err != nil {
		return fmt.Errorf("Cannot read audited base snapshot at %s: %v", s.repoSnapshotBaseFile, err)
	}
	s.repoSnapshotBaseline = strings.TrimRight(baseline, "\r\n")

	statePath := filepath.Join(repoRoot, s.policy.Maintenance.StateFile)
	if err := readJSONFile(statePath, &s.state); err != nil {
		return fmt.Errorf("state file at %s is not valid JSON: %v", statePath, err)
	}
	metadata, err := parseSimpleYAML(s.repoSnapshotMetaFile)
	if err != nil {
		return fmt.Errorf("Cannot read audited snapshot metadata at %s: %v", s.repoSnapshotMetaFile, err)
	}
	expectedMetadata := map[string]string{
		"schema_version":                    "1",
		"snapshot_file":                     filepath.Base(s.repoSnapshotBaseFile),
		"snapshot_source":                   "upstream-opencode-inline-asset",
		"state_file":                        s.policy.Maintenance.StateFile,
		"upstream_repo_name":                filepath.Base(strings.TrimRight(expandUser(s.policy.Upstream.RepoPath), string(filepath.Separator))),
		"upstream_prompt_rel_path":          s.policy.Upstream.OrchestratorPromptPath,
		"upstream_inject_source_rel_path":   "internal/components/sdd/inject.go",
		"upstream_profiles_source_rel_path": "internal/components/sdd/profiles.go",
		"last_maintained_version":           s.state.LastMaintainedVersion,
		"last_maintained_tag":               s.state.LastMaintainedTag,
		"last_maintained_commit":            s.state.LastMaintainedCommit,
		"last_reviewed_at":                  s.state.LastReviewedAt,
		"base_orchestrator_key":             s.policy.OpenCode.BaseOrchestratorKey,
		"profile_orchestrator_prefix":       s.policy.OpenCode.ProfileOrchestratorPrefix,
		"profile_phase_order_csv":           strings.Join(s.policy.OpenCode.SDDPhases, ","),
		"profile_task_scope_rule":           "deny-all-then-allow-suffixed-phases-and-global-jd",
	}
	for field, expected := range expectedMetadata {
		if metadata[field] != expected {
			return fmt.Errorf("audited snapshot metadata mismatch: field %q in %s is %q, expected %q. Repair the committed baseline before re-running apply.", field, s.repoSnapshotMetaFile, metadata[field], expected)
		}
	}
	actualHash := sha256Text(s.repoSnapshotBaseline)
	if metadata["snapshot_sha256"] != actualHash {
		return fmt.Errorf("audited snapshot metadata mismatch: snapshot_sha256 in %s is %q, expected %q from %s. Repair the committed baseline before re-running apply.", s.repoSnapshotMetaFile, metadata["snapshot_sha256"], actualHash, s.repoSnapshotBaseFile)
	}
	return nil
}

// --- Agent overrides ---

func (s *applyPolicyState) applyAgentOverrides() {
	for _, override := range s.policy.AgentOverrides {
		current, ok := jsonObject(s.agents[override.Key])
		if !ok {
			current = map[string]any{}
			s.agents[override.Key] = current
			s.createdOverrides = append(s.createdOverrides, override.Key)
			s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s: created missing object before applying override", override.Key))
			fmt.Printf("  agent override %s reset to object before applying model\n", override.Key)
		}
		oldModel := jsonString(current["model"])
		if oldModel != override.Model {
			current["model"] = override.Model
			s.configChanged = true
			s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.model: %s -> %s", override.Key, quotedValue(oldModel), quotedValue(override.Model)))
		}
		if override.Variant != "" {
			oldVariant := jsonString(current["variant"])
			if oldVariant != override.Variant {
				current["variant"] = override.Variant
				s.configChanged = true
				s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.variant: %s -> %s", override.Key, quotedValue(oldVariant), quotedValue(override.Variant)))
			}
		} else if _, hasVariant := current["variant"]; hasVariant {
			oldVariant := jsonString(current["variant"])
			delete(current, "variant")
			s.configChanged = true
			s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.variant: removed %s", override.Key, quotedValue(oldVariant)))
		}
		suffix := ""
		if override.Variant != "" {
			suffix = fmt.Sprintf(" (%s)", override.Variant)
		}
		fmt.Printf("  agent override %s -> %s%s\n", override.Key, override.Model, suffix)
	}
}

// --- Topology drift detection ---

func (s *applyPolicyState) detectTopologyDrift() {
	orchestratorsInConfig := map[string]bool{}
	for key := range s.originalAgentKeys {
		if s.isOrchestrator(key) {
			orchestratorsInConfig[key] = true
		}
	}
	known := map[string]bool{}
	for _, key := range s.policy.OpenCode.OrchestratorAgentKeys {
		known[key] = true
	}

	var unknown []string
	for key := range orchestratorsInConfig {
		if !known[key] && !s.isProfileOrchestrator(key) {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	for _, key := range unknown {
		message := fmt.Sprintf("unknown orchestrator matched by prefix only: %s", key)
		s.topologyWarnings = append(s.topologyWarnings, message)
		fmt.Printf("  topology: %s\n", message)
	}

	var missing []string
	for _, key := range s.policy.OpenCode.OrchestratorAgentKeys {
		if !s.originalAgentKeys[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		message := fmt.Sprintf("expected orchestrator missing from opencode.json: %s", key)
		s.topologyWarnings = append(s.topologyWarnings, message)
		fmt.Printf("  topology: %s\n", message)
	}

	sort.Strings(s.createdOverrides)
	for _, key := range s.createdOverrides {
		message := fmt.Sprintf("agent_override target was missing from upstream (created): %s", key)
		s.topologyWarnings = append(s.topologyWarnings, message)
		fmt.Printf("  topology: %s\n", message)
	}
}

// --- Persistence and verification ---

func (s *applyPolicyState) persistConfig() error {
	if !s.configChanged {
		return nil
	}
	status, err := writeJSONIndentedWithStatus(s.configPath, s.configData)
	if err != nil {
		return err
	}
	if shouldRecordWriteStatus(status) {
		s.recordVerbose(s.configPath, fmt.Sprintf("saved OpenCode config (%s)", describeWriteStatus(status)))
	}
	return nil
}

func (s *applyPolicyState) verifyPersistedState() error {
	verifyData, err := readJSONAny(s.configPath)
	if err != nil {
		return err
	}
	verifyAgents, ok := jsonObject(verifyData["agent"])
	if !ok {
		return fmt.Errorf("post-write verification failed: OpenCode config does not contain an agent map after write")
	}

	for _, override := range s.policy.AgentOverrides {
		actual, _ := jsonObject(verifyAgents[override.Key])
		if jsonString(actual["model"]) != override.Model {
			return fmt.Errorf("post-write verification failed: agent %q model is %q after write, expected %q", override.Key, jsonString(actual["model"]), override.Model)
		}
		if override.Variant != "" {
			if jsonString(actual["variant"]) != override.Variant {
				return fmt.Errorf("post-write verification failed: agent %q variant is %q after write, expected %q", override.Key, jsonString(actual["variant"]), override.Variant)
			}
		} else if jsonString(actual["variant"]) != "" {
			return fmt.Errorf("post-write verification failed: agent %q variant is %q after write, expected empty", override.Key, jsonString(actual["variant"]))
		}
	}

	keys := make([]string, 0, len(s.writtenOrchestrators))
	for key := range s.writtenOrchestrators {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		expectedRef := "{file:" + filepath.Join(s.generatedDir, key+".overlay.md") + "}"
		actual, _ := jsonObject(verifyAgents[key])
		if jsonString(actual["prompt"]) != expectedRef {
			return fmt.Errorf("post-write verification failed: orchestrator %q prompt is %q after write, expected %q", key, jsonString(actual["prompt"]), expectedRef)
		}
		overlayPath := filepath.Join(s.generatedDir, key+".overlay.md")
		if !pathExists(overlayPath) {
			return fmt.Errorf("post-write verification failed: overlay file missing for %q at %s", key, overlayPath)
		}
	}

	var profiles []string
	for name := range s.managedProfiles {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	for _, name := range profiles {
		orchKey := s.policy.OpenCode.ProfileOrchestratorPrefix + name
		if _, ok := verifyAgents[orchKey]; !ok {
			return fmt.Errorf("post-write verification failed: profile %q orchestrator agent %q missing from %s after write", name, orchKey, s.configPath)
		}
		for _, phase := range s.policy.OpenCode.SDDPhases {
			phaseKey := phase + "-" + name
			if _, ok := verifyAgents[phaseKey]; !ok {
				return fmt.Errorf("post-write verification failed: profile %q phase agent %q missing from %s after write", name, phaseKey, s.configPath)
			}
		}
	}

	if s.baseRuntimePrompt == "" || s.baseGeneratedPath == "" {
		return fmt.Errorf("audited baseline verification failed: orchestrator %q was not materialized during apply. Run `gentle-ai sync` to restore the inline upstream prompt, then re-run this script.", s.policy.OpenCode.BaseOrchestratorKey)
	}
	if normalizeLFTerminated(s.baseRuntimePrompt) != normalizeLFTerminated(s.repoSnapshotBaseline) {
		return fmt.Errorf("audited baseline mismatch for orchestrator %q: runtime source prompt does not match %s. Run `bash audit-gentle-ai-upstream.sh` before adopting a new upstream baseline, then re-run `gentle-ai sync` and this script.", s.policy.OpenCode.BaseOrchestratorKey, s.repoSnapshotBaseFile)
	}
	expectedBaseOverlay, err := sanitizePrompt(s.repoSnapshotBaseline, s.policy)
	if err != nil {
		return err
	}
	actualBaseOverlay, err := readText(s.baseGeneratedPath)
	if err != nil {
		return fmt.Errorf("Cannot read generated overlay for audited base orchestrator at %s: %v", s.baseGeneratedPath, err)
	}
	if normalizeLFTerminated(actualBaseOverlay) != normalizeLFTerminated(expectedBaseOverlay) {
		return fmt.Errorf("audited baseline mismatch for orchestrator %q: generated overlay at %s does not match the sanitized audited snapshot. Re-run apply after restoring the audited baseline, or run `gentle-ai sync` if local runtime state is stale.", s.policy.OpenCode.BaseOrchestratorKey, s.baseGeneratedPath)
	}
	return nil
}
