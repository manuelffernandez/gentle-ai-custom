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
	policy                  Policy
	managedTarget           ManagedAssetsTarget
	verbose                 bool
	recorder                *verboseRecorder
	configPath              string
	resolvedAgentOverrides  []AgentOverride
	resolvedDefaultProfile  *validatedProfile
	resolvedProfiles        []validatedProfile
	profileConfigSourcePath string
	profilesDefined         bool
	agents                  map[string]any
	configData              map[string]any
	configChanged           bool
	metrics                 applyMetrics
	managedProfiles         map[string]bool
	skillTargets            []string
	originalAgentKeys       map[string]bool
	sddPhasesSet            map[string]bool
	expectedPromptRefs      map[string]string
}

type applyPolicyOptions struct {
	verbose      bool
	recorder     *verboseRecorder
	skillTargets []string
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
	options.skillTargets = registeredSkillTargets(registeredAgentNames())
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
	managedTarget, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	state := &applyPolicyState{
		policy:               policy,
		managedTarget:        managedTarget,
		verbose:              options.verbose,
		recorder:             options.recorder,
		configPath:           expandUser(policy.OpenCode.ConfigPath),
		skillTargets:         append([]string(nil), options.skillTargets...),
		managedProfiles:      map[string]bool{},
		originalAgentKeys:    map[string]bool{},
		sddPhasesSet:         map[string]bool{},
		expectedPromptRefs:   map[string]string{},
	}
	// options.recorder must be initialized by the caller before invoking this
	// function. Both RunApplyPolicy and OpenCodeAgent.ApplyOverlay guarantee
	// this. A nil recorder here would silently accumulate verbose entries into
	// an unreachable object.
	if state.recorder == nil {
		fmt.Fprintf(os.Stderr, "internal error: recorder must be non-nil\n")
		os.Exit(1)
	}
	fail := func(err error) int {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if len(state.skillTargets) == 0 {
		return fail(fmt.Errorf("no registered skill targets resolved for prune scope"))
	}
	for _, phase := range policy.OpenCode.SDDPhases {
		state.sddPhasesSet[phase] = true
	}
	if err := state.loadLocalRuntimeConfig(); err != nil {
		return fail(err)
	}

	fmt.Println("Applying Gentle AI overlay policy...")
	if err := state.pruneSkills(); err != nil {
		return fail(err)
	}
	if err := state.installOwnedAssets(repoRoot); err != nil {
		return fail(err)
	}
	if err := state.verifyOwnedAssetTargets(); err != nil {
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

	state.applyAgentOverrides()
	if err := state.reconcileProfiles(); err != nil {
		return fail(err)
	}
	state.detectTopologyDrift()
	if err := state.rewriteManagedPromptReferences(); err != nil {
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
	// Prune scope is bound to the selected registered agent(s), not the legacy
	// policy target list. Unselected runtimes must remain untouched.
	for _, targetDirRaw := range s.skillTargets {
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
				s.metrics.prunedCount++
				s.recordVerbose(skillPath, fmt.Sprintf("removed pruned skill directory %q", skill))
				fmt.Printf("  removed %s\n", skill)
			} else {
				fmt.Printf("  already absent %s\n", skill)
			}
		}

		for _, keep := range s.policy.Skills.Keep {
			if !pathExists(filepath.Join(targetDir, keep)) {
				s.metrics.missingKeepSummary = append(s.metrics.missingKeepSummary, targetDir+" -> "+keep)
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

func (s *applyPolicyState) loadLocalRuntimeConfig() error {
	resolved, err := resolveLocalRuntimeConfig(s.policy, s.sddPhasesSet)
	if err != nil {
		return err
	}
	s.configPath = resolved.ConfigPath
	if policyDefault := expandUser(s.policy.OpenCode.ConfigPath); s.configPath != policyDefault {
		fmt.Printf("  opencode config path overridden by local config: %s\n", s.configPath)
	}
	s.resolvedAgentOverrides = resolved.AgentOverrides
	s.resolvedDefaultProfile = resolved.DefaultProfile
	s.resolvedProfiles = resolved.Profiles
	s.profileConfigSourcePath = resolved.ProfilesSourcePath
	s.profilesDefined = resolved.ProfilesDefined
	return nil
}

// --- Agent overrides ---

func (s *applyPolicyState) applyAgentOverrides() {
	for _, override := range s.resolvedAgentOverrides {
		current, ok := jsonObject(s.agents[override.Key])
		if !ok {
			current = map[string]any{}
			s.agents[override.Key] = current
			s.metrics.createdOverrides = append(s.metrics.createdOverrides, override.Key)
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
		s.metrics.topologyWarnings = append(s.metrics.topologyWarnings, message)
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
		s.metrics.topologyWarnings = append(s.metrics.topologyWarnings, message)
		fmt.Printf("  topology: %s\n", message)
	}

	sort.Strings(s.metrics.createdOverrides)
	for _, key := range s.metrics.createdOverrides {
		message := fmt.Sprintf("agent_override target was missing from upstream (created): %s", key)
		s.metrics.topologyWarnings = append(s.metrics.topologyWarnings, message)
		fmt.Printf("  topology: %s\n", message)
	}
}

func (s *applyPolicyState) isOrchestrator(key string) bool {
	for _, exact := range s.policy.OpenCode.OrchestratorAgentKeys {
		if key == exact {
			return true
		}
	}
	for _, prefix := range s.policy.OpenCode.OrchestratorAgentPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func (s *applyPolicyState) isProfileOrchestrator(key string) bool {
	prefix := s.policy.OpenCode.ProfileOrchestratorPrefix
	return prefix != "" && strings.HasPrefix(key, prefix)
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

	for _, override := range s.resolvedAgentOverrides {
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

	if s.resolvedDefaultProfile != nil {
		baseKey := s.policy.OpenCode.BaseOrchestratorKey
		if err := verifyProfileAssignment(verifyAgents, baseKey, s.resolvedDefaultProfile.Orchestrator); err != nil {
			return fmt.Errorf("post-write verification failed for default profile orchestrator %q: %w", baseKey, err)
		}
		for _, phase := range s.policy.OpenCode.SDDPhases {
			if err := verifyProfileAssignment(verifyAgents, phase, s.resolvedDefaultProfile.Phases[phase]); err != nil {
				return fmt.Errorf("post-write verification failed for default profile phase %q: %w", phase, err)
			}
		}
	}

	keys := make([]string, 0, len(s.expectedPromptRefs))
	for key := range s.expectedPromptRefs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		expectedRef := s.expectedPromptRefs[key]
		actual, _ := jsonObject(verifyAgents[key])
		if jsonString(actual["prompt"]) != expectedRef {
			return fmt.Errorf("post-write verification failed: agent %q prompt is %q after write, expected %q", key, jsonString(actual["prompt"]), expectedRef)
		}
		promptPath := strings.TrimSuffix(strings.TrimPrefix(expectedRef, "{file:"), "}")
		if !pathExists(promptPath) {
			return fmt.Errorf("post-write verification failed: prompt file missing for %q at %s", key, promptPath)
		}
	}

	var profiles []string
	for name := range s.managedProfiles {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	for _, name := range profiles {
		profile, ok := s.managedProfileAssignment(name)
		if !ok {
			return fmt.Errorf("post-write verification failed: managed profile %q missing from resolved profiles", name)
		}
		orchKey := s.policy.OpenCode.ProfileOrchestratorPrefix + name
		if err := verifyProfileAssignment(verifyAgents, orchKey, profile.Orchestrator); err != nil {
			return fmt.Errorf("post-write verification failed for profile %q orchestrator agent %q: %w", name, orchKey, err)
		}
		for _, phase := range s.policy.OpenCode.SDDPhases {
			phaseKey := phase + "-" + name
			if err := verifyProfileAssignment(verifyAgents, phaseKey, profile.Phases[phase]); err != nil {
				return fmt.Errorf("post-write verification failed for profile %q phase agent %q: %w", name, phaseKey, err)
			}
		}
	}

	return nil
}

func verifyProfileAssignment(agents map[string]any, key string, assignment profileAssignment) error {
	actual, ok := jsonObject(agents[key])
	if !ok {
		return fmt.Errorf("agent missing from OpenCode config after write")
	}
	if jsonString(actual["model"]) != assignment.Model {
		return fmt.Errorf("model is %q after write, expected %q", jsonString(actual["model"]), assignment.Model)
	}
	if assignment.Variant != "" {
		if jsonString(actual["variant"]) != assignment.Variant {
			return fmt.Errorf("variant is %q after write, expected %q", jsonString(actual["variant"]), assignment.Variant)
		}
	} else if jsonString(actual["variant"]) != "" {
		return fmt.Errorf("variant is %q after write, expected empty", jsonString(actual["variant"]))
	}
	return nil
}

func (s *applyPolicyState) managedProfileAssignment(name string) (validatedProfile, bool) {
	for _, profile := range s.resolvedProfiles {
		if profile.Name == name {
			return profile, true
		}
	}
	return validatedProfile{}, false
}
