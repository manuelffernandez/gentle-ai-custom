package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const upstreamRepoEnvVar = "GENTLE_AI_CUSTOM_UPSTREAM_REPO"

type localConfigFile struct {
	UpstreamRepoPath      string
	OpenCodeConfigPath    string
	AgentOverrides        []AgentOverride
	AgentOverridesDefined bool
	DefaultProfile        *validatedProfile
	DefaultProfileDefined bool
	Profiles              []validatedProfile
	ProfilesDefined       bool
}

type resolvedLocalRuntimeConfig struct {
	ConfigPath         string
	AgentOverrides     []AgentOverride
	DefaultProfile     *validatedProfile
	Profiles           []validatedProfile
	ProfilesSourcePath string
	ProfilesDefined    bool
}

func loadLocalConfigFile(path string, sddPhases []string, sddPhasesSet map[string]bool) (localConfigFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return localConfigFile{}, fmt.Errorf("cannot read local OpenCode overlay config at %s: %v", path, err)
	}
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s is not valid JSON: %v. Fix or remove the file before re-running this script.", path, err)
	}
	top, ok := data.(map[string]any)
	if !ok {
		return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s must be a JSON object at the top level", path)
	}
	version, ok := top["version"].(float64)
	if !ok || version != 1 {
		return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s has unsupported \"version\" %v; expected 1", path, top["version"])
	}
	for key := range top {
		if key != "version" && key != "upstream_repo_path" && key != "opencode_config_path" && key != "agent_overrides" && key != "default_profile" && key != "profiles" {
			return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s has unexpected top-level fields [%s]; only \"version\", \"upstream_repo_path\", \"opencode_config_path\", \"agent_overrides\", \"default_profile\", and \"profiles\" are allowed", path, strings.Join(sortedStringKeys(top, "version", "upstream_repo_path", "opencode_config_path", "agent_overrides", "default_profile", "profiles"), ", "))
		}
	}

	config := localConfigFile{}
	if value, ok := top["upstream_repo_path"]; ok {
		upstreamRepoPath, ok := value.(string)
		if !ok {
			return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s: field \"upstream_repo_path\" must be a string", path)
		}
		config.UpstreamRepoPath = strings.TrimSpace(upstreamRepoPath)
	}
	if value, ok := top["opencode_config_path"]; ok {
		opencodeConfigPath, ok := value.(string)
		if !ok {
			return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s: field \"opencode_config_path\" must be a string", path)
		}
		config.OpenCodeConfigPath = strings.TrimSpace(opencodeConfigPath)
	}
	if value, ok := top["agent_overrides"]; ok {
		agentOverrides, err := validateAgentOverrides(value, path+".agent_overrides")
		if err != nil {
			return localConfigFile{}, err
		}
		config.AgentOverrides = agentOverrides
		config.AgentOverridesDefined = true
	}
	if value, ok := top["default_profile"]; ok {
		defaultProfile, err := validateDefaultProfileValue(value, path+".default_profile", sddPhases, sddPhasesSet)
		if err != nil {
			return localConfigFile{}, err
		}
		config.DefaultProfile = defaultProfile
		config.DefaultProfileDefined = true
	}
	if value, ok := top["profiles"]; ok {
		profiles, err := validateProfilesValue(value, path+".profiles", sddPhases, sddPhasesSet)
		if err != nil {
			return localConfigFile{}, err
		}
		config.Profiles = profiles
		config.ProfilesDefined = true
	}

	return config, nil
}

func resolveLocalRuntimeConfig(policy Policy, sddPhasesSet map[string]bool) (resolvedLocalRuntimeConfig, error) {
	resolved := resolvedLocalRuntimeConfig{
		ConfigPath: expandUser(policy.OpenCode.ConfigPath),
	}
	localConfigPath := expandUser(policy.OpenCode.LocalConfigPath)

	if pathExists(localConfigPath) {
		config, err := loadLocalConfigFile(localConfigPath, policy.OpenCode.SDDPhases, sddPhasesSet)
		if err != nil {
			return resolvedLocalRuntimeConfig{}, err
		}
		if config.OpenCodeConfigPath != "" {
			resolved.ConfigPath = expandUser(config.OpenCodeConfigPath)
		}
		if config.AgentOverridesDefined {
			resolved.AgentOverrides = config.AgentOverrides
		}
		if config.DefaultProfileDefined {
			resolved.DefaultProfile = config.DefaultProfile
		}
		if config.ProfilesDefined {
			resolved.Profiles = config.Profiles
			resolved.ProfilesDefined = true
			resolved.ProfilesSourcePath = localConfigPath
		}
	}

	return resolved, nil
}

func resolveUpstreamRepo(repoRoot string, policy Policy) (string, string, error) {
	sddPhasesSet := make(map[string]bool, len(policy.OpenCode.SDDPhases))
	for _, phase := range policy.OpenCode.SDDPhases {
		sddPhasesSet[phase] = true
	}

	localConfigPath := expandUser(policy.OpenCode.LocalConfigPath)
	if pathExists(localConfigPath) {
		config, err := loadLocalConfigFile(localConfigPath, policy.OpenCode.SDDPhases, sddPhasesSet)
		if err != nil {
			return "", "", err
		}
		if config.UpstreamRepoPath != "" {
			pathValue, err := ensureExistingDir(expandUser(config.UpstreamRepoPath), fmt.Sprintf("local config %s (upstream_repo_path)", localConfigPath))
			if err != nil {
				return "", "", err
			}
			return pathValue, fmt.Sprintf("local config (%s)", localConfigPath), nil
		}
	}

	if value := strings.TrimSpace(os.Getenv(upstreamRepoEnvVar)); value != "" {
		pathValue, err := ensureExistingDir(expandUser(value), "$"+upstreamRepoEnvVar)
		if err != nil {
			return "", "", err
		}
		return pathValue, "$" + upstreamRepoEnvVar, nil
	}

	fallback := filepath.Clean(filepath.Join(repoRoot, "..", policy.Upstream.RepoName))
	if info, err := os.Stat(fallback); err == nil && info.IsDir() {
		return fallback, fmt.Sprintf("repo-relative fallback (%s)", filepath.Join("..", policy.Upstream.RepoName)), nil
	}

	return "", "", fmt.Errorf("cannot resolve upstream repo. Checked local config %s, $%s, and repo-relative fallback %s. Set \"upstream_repo_path\" in %s, export %s, or place the upstream repo at %s", localConfigPath, upstreamRepoEnvVar, fallback, localConfigPath, upstreamRepoEnvVar, fallback)
}

func ensureExistingDir(pathValue, source string) (string, error) {
	info, err := os.Stat(pathValue)
	if err != nil {
		return "", fmt.Errorf("%s points to a missing upstream repo: %s", source, pathValue)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s must point to a directory: %s", source, pathValue)
	}
	return filepath.Clean(pathValue), nil
}

func validateAgentOverrides(value any, label string) ([]AgentOverride, error) {
	rawOverrides, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s: must be an array of objects with \"key\", \"model\", and optional \"variant\"", label)
	}
	seen := map[string]bool{}
	validated := make([]AgentOverride, 0, len(rawOverrides))
	for idx, rawOverride := range rawOverrides {
		prefix := fmt.Sprintf("%s[%d]", label, idx)
		override, ok := rawOverride.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: must be an object", prefix)
		}
		for key := range override {
			if key != "key" && key != "model" && key != "variant" {
				return nil, fmt.Errorf("%s: unexpected fields [%s]; only \"key\", \"model\", and \"variant\" are allowed", prefix, strings.Join(sortedStringKeys(override, "key", "model", "variant"), ", "))
			}
		}
		keyRaw, ok := override["key"].(string)
		if !ok || strings.TrimSpace(keyRaw) == "" {
			return nil, fmt.Errorf("%s: field \"key\" must be a non-empty string", prefix)
		}
		key := strings.TrimSpace(keyRaw)
		modelRaw, ok := override["model"].(string)
		if !ok || strings.TrimSpace(modelRaw) == "" {
			return nil, fmt.Errorf("%s: field \"model\" must be a non-empty string", prefix)
		}
		model := strings.TrimSpace(modelRaw)
		var variant string
		if rawVariant, exists := override["variant"]; exists {
			v, ok := rawVariant.(string)
			if !ok {
				return nil, fmt.Errorf("%s: field \"variant\" must be a string (use \"\" for no variant)", prefix)
			}
			variant = v
		}
		if seen[key] {
			return nil, fmt.Errorf("%s: duplicate agent override key %q", prefix, key)
		}
		seen[key] = true
		validated = append(validated, AgentOverride{Key: key, Model: model, Variant: variant})
	}
	return validated, nil
}

func validateDefaultProfileValue(value any, label string, sddPhases []string, sddPhasesSet map[string]bool) (*validatedProfile, error) {
	profile, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: must be an object with \"orchestrator\" and \"phases\"", label)
	}
	for key := range profile {
		if key != "orchestrator" && key != "phases" {
			return nil, fmt.Errorf("%s: unexpected fields [%s]; only \"orchestrator\" and \"phases\" are allowed", label, strings.Join(sortedStringKeys(profile, "orchestrator", "phases"), ", "))
		}
	}
	orchestrator, err := validateAssignment(profile["orchestrator"], label+".orchestrator")
	if err != nil {
		return nil, err
	}
	phasesObj, ok := profile["phases"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s.phases: must be an object keyed by SDD phase name", label)
	}
	phaseKeys := map[string]bool{}
	for phaseName := range phasesObj {
		phaseKeys[phaseName] = true
	}
	var missing []string
	for _, phaseName := range sddPhases {
		if !phaseKeys[phaseName] {
			missing = append(missing, phaseName)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%s.phases: missing required phases %v (no defaults are inherited)", label, missing)
	}
	var unknown []string
	for phaseName := range phaseKeys {
		if !sddPhasesSet[phaseName] {
			unknown = append(unknown, phaseName)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		return nil, fmt.Errorf("%s.phases: unknown phases %v; allowed: %v", label, unknown, sddPhases)
	}
	validatedPhases := map[string]profileAssignment{}
	for _, phaseName := range sddPhases {
		assignment, err := validateAssignment(phasesObj[phaseName], label+".phases."+phaseName)
		if err != nil {
			return nil, err
		}
		validatedPhases[phaseName] = assignment
	}
	profileValue := &validatedProfile{Name: "default", Orchestrator: orchestrator, Phases: validatedPhases}
	return profileValue, nil
}
