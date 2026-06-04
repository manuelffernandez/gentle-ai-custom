package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Policy struct {
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Maintenance struct {
		IntentFile string `json:"intent_file"`
		StateFile  string `json:"state_file"`
		LogFile    string `json:"log_file"`
	} `json:"maintenance"`
	Upstream struct {
		RepoName               string `json:"repo_name"`
		OrchestratorPromptPath string `json:"orchestrator_prompt_path"`
	} `json:"upstream"`
	Skills struct {
		Keep    []string `json:"keep"`
		Prune   []string `json:"prune"`
		Targets []string `json:"targets"`
	} `json:"skills"`
	OpenCode struct {
		ConfigPath          string   `json:"config_path"`
		LocalConfigPath     string   `json:"local_config_path"`
		BaseOrchestratorKey string   `json:"base_orchestrator_key"`
		GeneratedOrchestratorsDir     string   `json:"generated_orchestrators_dir"`
		OrchestratorSnapshotDir       string   `json:"orchestrator_snapshot_dir"`
		OrchestratorSnapshotMetadata  string   `json:"orchestrator_snapshot_metadata_file"`
		LocalOrchestratorSnapshotDir  string   `json:"local_orchestrator_snapshot_dir"`
		OrchestratorAgentKeys         []string `json:"orchestrator_agent_keys"`
		OrchestratorAgentPrefixes     []string `json:"orchestrator_agent_prefixes"`
		ProfileOrchestratorPrefix     string   `json:"profile_orchestrator_prefix"`
		SDDPhases                     []string `json:"sdd_phases"`
	} `json:"opencode"`
	Sanitizer struct {
		RequiredMarkers  []string `json:"required_markers"`
		ForbiddenMarkers []string `json:"forbidden_markers"`
	} `json:"sanitizer"`
}

type AgentOverride struct {
	Key     string `json:"key"`
	Model   string `json:"model"`
	Variant string `json:"variant"`
}

type UpstreamState struct {
	LastMaintainedVersion string `json:"last_maintained_version"`
	LastMaintainedTag     string `json:"last_maintained_tag"`
	LastMaintainedCommit  string `json:"last_maintained_commit"`
	LastReviewedAt        string `json:"last_reviewed_at"`
	Notes                 string `json:"notes"`
}

func loadPolicy(repoRoot string) (Policy, string, error) {
	policyPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "gentle-ai-policy.json")
	raw, err := os.ReadFile(policyPath)
	if err != nil {
		return Policy{}, policyPath, fmt.Errorf("cannot read policy file at %s: %w", policyPath, err)
	}
	var policy Policy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return Policy{}, policyPath, fmt.Errorf("policy file is not valid JSON at %s: %w", policyPath, err)
	}
	return policy, policyPath, nil
}
