package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func RunAuditUpstream(repoRoot string) int {
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	upstreamRepo := filepath.Clean(expandUser(policy.Upstream.RepoPath))
	statePath := filepath.Join(repoRoot, policy.Maintenance.StateFile)
	snapshotPath := filepath.Join(repoRoot, policy.OpenCode.OrchestratorSnapshotDir, policy.OpenCode.BaseOrchestratorKey+".last.md")
	metaPath := filepath.Join(repoRoot, policy.OpenCode.OrchestratorSnapshotMetadata)
	upstreamPromptPath := filepath.Join(upstreamRepo, policy.Upstream.OrchestratorPromptPath)
	upstreamProfilesPath := filepath.Join(upstreamRepo, "internal", "components", "sdd", "profiles.go")
	upstreamInjectPath := filepath.Join(upstreamRepo, "internal", "components", "sdd", "inject.go")

	if info, err := os.Stat(upstreamRepo); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "ERROR: upstream repo not found: %s\n", upstreamRepo)
		return 1
	}

	var state UpstreamState
	if err := readJSONFile(statePath, &state); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: state file is not valid JSON at %s: %v\n", statePath, err)
		return 1
	}

	metadata, err := parseSimpleYAML(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", metaPath, err)
		return 1
	}
	snapshotText, err := readText(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", snapshotPath, err)
		return 1
	}
	upstreamPromptText, err := readText(upstreamPromptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", upstreamPromptPath, err)
		return 1
	}
	upstreamProfilesText, err := readText(upstreamProfilesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", upstreamProfilesPath, err)
		return 1
	}
	upstreamInjectText, err := readText(upstreamInjectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", upstreamInjectPath, err)
		return 1
	}

	baseKey := policy.OpenCode.BaseOrchestratorKey
	if baseKey == "" {
		baseKey = "gentle-orchestrator"
	}
	expectedPhaseCSV := strings.Join(policy.OpenCode.SDDPhases, ",")
	expectedMetadata := map[string]string{
		"schema_version":                    "1",
		"snapshot_file":                     filepath.Base(snapshotPath),
		"snapshot_source":                   "upstream-opencode-inline-asset",
		"state_file":                        policy.Maintenance.StateFile,
		"upstream_repo_name":                filepath.Base(upstreamRepo),
		"upstream_prompt_rel_path":          policy.Upstream.OrchestratorPromptPath,
		"upstream_inject_source_rel_path":   "internal/components/sdd/inject.go",
		"upstream_profiles_source_rel_path": "internal/components/sdd/profiles.go",
		"last_maintained_version":           state.LastMaintainedVersion,
		"last_maintained_tag":               state.LastMaintainedTag,
		"last_maintained_commit":            state.LastMaintainedCommit,
		"last_reviewed_at":                  state.LastReviewedAt,
		"base_orchestrator_key":             baseKey,
		"profile_orchestrator_prefix":       policy.OpenCode.ProfileOrchestratorPrefix,
		"profile_phase_order_csv":           expectedPhaseCSV,
		"profile_task_scope_rule":           "deny-all-then-allow-suffixed-phases-and-global-jd",
	}

	var failures []string
	var notes []string

	actualSnapshotHash := sha256Text(snapshotText)
	if metadata["snapshot_sha256"] != actualSnapshotHash {
		failures = append(failures, fmt.Sprintf("metadata snapshot_sha256 is %q, expected %q from %s", metadata["snapshot_sha256"], actualSnapshotHash, snapshotPath))
	}
	for field, expected := range expectedMetadata {
		actual := metadata[field]
		if actual != expected {
			failures = append(failures, fmt.Sprintf("metadata field %q is %q, expected %q", field, actual, expected))
		}
	}

	upstreamHead, gitErr := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if gitErr != nil {
		failures = append(failures, fmt.Sprintf("cannot inspect upstream git state: %v", gitErr))
	}
	upstreamDescribe, gitErr := runGit(upstreamRepo, false, "describe", "--tags", "--always")
	if gitErr != nil {
		failures = append(failures, fmt.Sprintf("cannot inspect upstream git state: %v", gitErr))
	}
	upstreamExactTag, gitErr := runGit(upstreamRepo, true, "describe", "--tags", "--exact-match")
	if gitErr != nil {
		failures = append(failures, fmt.Sprintf("cannot inspect upstream git state: %v", gitErr))
	}

	if upstreamHead != "" && state.LastMaintainedCommit != "" && upstreamHead != state.LastMaintainedCommit {
		notes = append(notes, fmt.Sprintf("upstream HEAD %s differs from last maintained commit %s; prompt/invariant drift checks below show whether the baseline still holds", upstreamHead, state.LastMaintainedCommit))
	}
	if upstreamExactTag != "" && state.LastMaintainedTag != "" && upstreamExactTag != state.LastMaintainedTag {
		notes = append(notes, fmt.Sprintf("upstream exact tag %s differs from last maintained tag %s; review state/log if you are closing a new upstream audit", upstreamExactTag, state.LastMaintainedTag))
	}

	promptMatches := normalizeLFTerminated(snapshotText) == normalizeLFTerminated(upstreamPromptText)
	if !promptMatches {
		failures = append(failures, fmt.Sprintf("base prompt drift detected: %s no longer matches %s; review/update the audited baseline before sync/apply", upstreamPromptPath, snapshotPath))
	}

	phaseOrder, phaseErr := extractProfilePhaseOrder(upstreamProfilesText)
	if phaseErr != nil {
		failures = append(failures, phaseErr.Error())
	}
	if len(phaseOrder) > 0 && !sameStrings(phaseOrder, policy.OpenCode.SDDPhases) {
		failures = append(failures, fmt.Sprintf("upstream profilePhaseOrder is %#v, expected %#v from policy/metadata", phaseOrder, policy.OpenCode.SDDPhases))
	}

	const profilePrefixSnippet = `const orchPrefix = "sdd-orchestrator-"`
	const profileKeyBuilderSnippet = `keys = append(keys, "sdd-orchestrator"+suffix)`
	if !strings.Contains(upstreamProfilesText, profilePrefixSnippet) {
		failures = append(failures, "upstream profiles.go no longer declares DetectProfiles prefix 'sdd-orchestrator-'")
	}
	if !strings.Contains(upstreamProfilesText, profileKeyBuilderSnippet) {
		failures = append(failures, "upstream ProfileAgentKeys no longer builds profile orchestrator keys from 'sdd-orchestrator'+suffix")
	}

	requiredProfilesSnippets := []string{
		"taskPerms := map[string]any{",
		`"*": "deny",`,
		"taskPerms[phase+suffix] = \"allow\"",
		"taskPerms[jd] = \"allow\"",
	}
	for _, snippet := range requiredProfilesSnippets {
		if !strings.Contains(upstreamProfilesText, snippet) {
			failures = append(failures, fmt.Sprintf("upstream profile task scoping snippet missing from profiles.go: %q", snippet))
		}
	}

	requiredInjectSnippets := []string{
		`orchestratorRaw, ok := agentsMap["gentle-orchestrator"]`,
		`orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))`,
	}
	for _, snippet := range requiredInjectSnippets {
		if !strings.Contains(upstreamInjectText, snippet) {
			failures = append(failures, fmt.Sprintf("upstream inject.go no longer contains expected base orchestrator asset binding snippet: %q", snippet))
		}
	}

	metadataOK := true
	for _, failure := range failures {
		if strings.HasPrefix(failure, "metadata field") || strings.HasPrefix(failure, "metadata snapshot_sha256") {
			metadataOK = false
			break
		}
	}
	profileNamingOK := strings.Contains(upstreamProfilesText, profilePrefixSnippet) && strings.Contains(upstreamProfilesText, profileKeyBuilderSnippet)
	taskScopingOK := containsAll(upstreamProfilesText, requiredProfilesSnippets)
	baseAssetInjectionOK := containsAll(upstreamInjectText, requiredInjectSnippets)
	phaseOrderOK := len(phaseOrder) > 0 && sameStrings(phaseOrder, policy.OpenCode.SDDPhases)

	fmt.Println("Auditing Gentle AI upstream baseline...")
	fmt.Printf("- Repo root: %s\n", repoRoot)
	fmt.Printf("- Upstream repo: %s\n", upstreamRepo)
	if upstreamDescribe != "" {
		fmt.Printf("- Upstream HEAD: %s (%s)\n", upstreamDescribe, upstreamHead)
	}
	fmt.Printf("- Base snapshot: %s\n", snapshotPath)
	fmt.Printf("- Base metadata: %s\n", metaPath)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  state/metadata alignment: %s\n", statusWord(metadataOK))
	fmt.Printf("  snapshot hash verification: %s\n", statusWord(metadata["snapshot_sha256"] == actualSnapshotHash))
	if promptMatches {
		fmt.Println("  base prompt drift: no")
	} else {
		fmt.Println("  base prompt drift: yes")
	}
	fmt.Printf("  profile phase order: %s\n", statusWord(phaseOrderOK))
	fmt.Printf("  profile orchestrator naming: %s\n", statusWord(profileNamingOK))
	fmt.Printf("  profile task scoping invariant: %s\n", statusWord(taskScopingOK))
	fmt.Printf("  base asset injection invariant: %s\n", statusWord(baseAssetInjectionOK))

	if len(notes) > 0 {
		fmt.Println()
		for _, note := range notes {
			fmt.Printf("NOTE: %s\n", note)
		}
	}

	if len(failures) > 0 {
		fmt.Println()
		for _, failure := range failures {
			fmt.Printf("FAIL: %s\n", failure)
		}
		fmt.Println()
		fmt.Println("Action:")
		fmt.Println("1. Review the upstream delta against the committed baseline.")
		fmt.Println("2. Update `gentle-orchestrator.last.md`, `.meta.yaml`, `upstream-state.json`, docs, and the update log if the new upstream state is accepted.")
		fmt.Println("3. Run `gentle-ai sync` or reinstall as appropriate, then re-run `bash apply-gentle-ai-custom.sh all` so runtime verification passes.")
		return 1
	}

	fmt.Println()
	fmt.Println("Done. The committed base snapshot and metadata still match the current upstream prompt/invariants.")
	return 0
}

func extractProfilePhaseOrder(profilesGo string) ([]string, error) {
	re := regexp.MustCompile(`(?s)var\s+profilePhaseOrder\s*=\s*\[\]string\s*\{(.*?)\n\}`)
	match := re.FindStringSubmatch(profilesGo)
	if match == nil {
		return nil, fmt.Errorf("could not locate profilePhaseOrder in upstream profiles.go")
	}
	valueRE := regexp.MustCompile(`"([^"]+)"`)
	values := valueRE.FindAllStringSubmatch(match[1], -1)
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value[1])
	}
	return result, nil
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsAll(text string, snippets []string) bool {
	for _, snippet := range snippets {
		if !strings.Contains(text, snippet) {
			return false
		}
	}
	return true
}

func statusWord(ok bool) string {
	if ok {
		return "ok"
	}
	return "mismatch"
}
