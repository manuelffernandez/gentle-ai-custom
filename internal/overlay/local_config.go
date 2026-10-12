package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const upstreamRepoEnvVar = "GENTLE_AI_CUSTOM_UPSTREAM_REPO"

type localConfigFile struct {
	UpstreamRepoPath      string
	OpenCodeConfigPath    string
	AgentOverrides        []AgentOverride
	AgentOverridesDefined bool
	GlobalActors          map[string]actorAssignment
}

type resolvedLocalRuntimeConfig struct {
	ConfigPath     string
	AgentOverrides []AgentOverride
	GlobalActors   map[string]actorAssignment
}

func loadLocalConfigFile(path string) (localConfigFile, error) {
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
	for _, key := range []string{"default_profile", "profiles"} {
		if _, exists := top[key]; exists {
			return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s still contains legacy %q SDD configuration; remove that key manually and reassign any SDD model choices yourself. Global Judgment Day/review assignments can be configured under \"global_actors\"", path, key)
		}
	}
	for key := range top {
		if key != "version" && key != "upstream_repo_path" && key != "opencode_config_path" && key != "agent_overrides" && key != "global_actors" {
			return localConfigFile{}, fmt.Errorf("local OpenCode overlay config at %s has unexpected top-level fields [%s]; only \"version\", \"upstream_repo_path\", \"opencode_config_path\", \"agent_overrides\", and \"global_actors\" are allowed", path, strings.Join(sortedStringKeys(top, "version", "upstream_repo_path", "opencode_config_path", "agent_overrides", "global_actors"), ", "))
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
	if value, ok := top["global_actors"]; ok {
		actors, err := validateGlobalActors(value, path+".global_actors")
		if err != nil {
			return localConfigFile{}, err
		}
		config.GlobalActors = actors
	}

	return config, nil
}

func resolveLocalRuntimeConfig(policy Policy) (resolvedLocalRuntimeConfig, error) {
	resolved := resolvedLocalRuntimeConfig{
		ConfigPath: expandUser(policy.OpenCode.ConfigPath),
	}
	localConfigPath := expandUser(policy.OpenCode.LocalConfigPath)

	if pathExists(localConfigPath) {
		config, err := loadLocalConfigFile(localConfigPath)
		if err != nil {
			return resolvedLocalRuntimeConfig{}, err
		}
		if config.OpenCodeConfigPath != "" {
			resolved.ConfigPath = expandUser(config.OpenCodeConfigPath)
		}
		if config.AgentOverridesDefined {
			resolved.AgentOverrides = config.AgentOverrides
		}
		resolved.GlobalActors = config.GlobalActors
	}

	return resolved, nil
}

func resolveUpstreamRepo(repoRoot string, policy Policy) (string, string, error) {
	localConfigPath := expandUser(policy.OpenCode.LocalConfigPath)
	if pathExists(localConfigPath) {
		upstreamRepoPath, err := loadLocalUpstreamRepoPath(localConfigPath)
		if err != nil {
			return "", "", err
		}
		if upstreamRepoPath != "" {
			pathValue, err := ensureExistingDir(expandUser(upstreamRepoPath), fmt.Sprintf("local config %s (upstream_repo_path)", localConfigPath))
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

func loadLocalUpstreamRepoPath(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read local OpenCode overlay config at %s: %v", path, err)
	}
	var config struct {
		Version          int `json:"version"`
		UpstreamRepoPath any `json:"upstream_repo_path"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return "", fmt.Errorf("local OpenCode overlay config at %s is not valid JSON: %v. Fix or remove the file before re-running this script.", path, err)
	}
	if config.Version != 1 {
		return "", fmt.Errorf("local OpenCode overlay config at %s has unsupported \"version\" %v; expected 1", path, config.Version)
	}
	if config.UpstreamRepoPath == nil {
		return "", nil
	}
	upstreamRepoPath, ok := config.UpstreamRepoPath.(string)
	if !ok {
		return "", fmt.Errorf("local OpenCode overlay config at %s: field \"upstream_repo_path\" must be a string", path)
	}
	return strings.TrimSpace(upstreamRepoPath), nil
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
		if isLegacySDDAgentKey(key) {
			return nil, fmt.Errorf("%s[%d] targets legacy SDD agent %q; remove that override and choose an appropriate ODD role manually", label, idx, key)
		}
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

func validateGlobalActors(value any, label string) (map[string]actorAssignment, error) {
	actors, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: must be an object keyed by global Judgment Day or review actor name", label)
	}
	validated := make(map[string]actorAssignment, len(actors))
	for key, value := range actors {
		if !containsString(judgmentDayActorKeys, key) && !containsString(reviewActorKeys, key) {
			return nil, fmt.Errorf("%s: unsupported global actor %q", label, key)
		}
		assignment, err := validateActorAssignment(value, label+"."+key, true)
		if err != nil {
			return nil, err
		}
		validated[key] = assignment
	}
	return validated, nil
}

func isLegacySDDAgentKey(key string) bool {
	return strings.HasPrefix(key, "sdd-")
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
