package overlay

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUpstreamRepoPrecedence(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv(upstreamRepoEnvVar, filepath.Join(homeDir, "env-upstream"))

	repoRoot := filepath.Join(t.TempDir(), "gentle-ai-custom")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	localUpstream := filepath.Join(homeDir, "custom-upstream")
	envUpstream := filepath.Join(homeDir, "env-upstream")
	fallbackUpstream := filepath.Join(filepath.Dir(repoRoot), "gentle-ai")
	for _, dir := range []string{localUpstream, envUpstream, fallbackUpstream} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir upstream dir %s: %v", dir, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".config", "gentle-ai-custom"), 0o755); err != nil {
		t.Fatalf("mkdir local config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, ".config", "gentle-ai-custom", "opencode-local-config.json"), []byte("{\n  \"version\": 1,\n  \"upstream_repo_path\": \""+localUpstream+"\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	pathValue, source, err := resolveUpstreamRepo(repoRoot, testPolicy())
	if err != nil {
		t.Fatalf("resolve upstream repo: %v", err)
	}
	if pathValue != localUpstream {
		t.Fatalf("resolved upstream = %q, want %q", pathValue, localUpstream)
	}
	if source == "$"+upstreamRepoEnvVar {
		t.Fatalf("expected local config precedence, got env source %q", source)
	}
}

func TestResolveUpstreamRepoFallsBackToEnvAndSibling(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	repoRoot := filepath.Join(t.TempDir(), "gentle-ai-custom")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}

	envUpstream := filepath.Join(homeDir, "env-upstream")
	if err := os.MkdirAll(envUpstream, 0o755); err != nil {
		t.Fatalf("mkdir env upstream: %v", err)
	}
	t.Setenv(upstreamRepoEnvVar, envUpstream)

	pathValue, source, err := resolveUpstreamRepo(repoRoot, testPolicy())
	if err != nil {
		t.Fatalf("resolve upstream from env: %v", err)
	}
	if pathValue != envUpstream || source != "$"+upstreamRepoEnvVar {
		t.Fatalf("resolved upstream = %q from %q, want %q from env", pathValue, source, envUpstream)
	}

	t.Setenv(upstreamRepoEnvVar, "")
	fallbackUpstream := filepath.Join(filepath.Dir(repoRoot), "gentle-ai")
	if err := os.MkdirAll(fallbackUpstream, 0o755); err != nil {
		t.Fatalf("mkdir fallback upstream: %v", err)
	}

	pathValue, source, err = resolveUpstreamRepo(repoRoot, testPolicy())
	if err != nil {
		t.Fatalf("resolve upstream from fallback: %v", err)
	}
	if pathValue != fallbackUpstream {
		t.Fatalf("resolved upstream = %q, want %q", pathValue, fallbackUpstream)
	}
	if source == "$"+upstreamRepoEnvVar {
		t.Fatalf("expected sibling fallback source, got %q", source)
	}
}

func TestResolveLocalRuntimeConfigUsesLegacyProfilesWhenNewConfigOmitsThem(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	configDir := filepath.Join(homeDir, ".config", "gentle-ai-custom")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "opencode-local-config.json"), []byte("{\n  \"version\": 1,\n  \"opencode_config_path\": \"/tmp/custom-opencode.json\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "opencode-sdd-profiles.json"), []byte(legacyProfilesJSON()), 0o644); err != nil {
		t.Fatalf("write legacy profiles: %v", err)
	}

	resolved, err := resolveLocalRuntimeConfig(testPolicy(), testPhasesSet())
	if err != nil {
		t.Fatalf("resolve local runtime config: %v", err)
	}
	if resolved.ConfigPath != "/tmp/custom-opencode.json" {
		t.Fatalf("config path = %q, want custom override", resolved.ConfigPath)
	}
	if !resolved.UsedLegacyProfiles || !resolved.ProfilesDefined {
		t.Fatalf("expected legacy profiles to remain active when new config omits profiles")
	}
	if len(resolved.Profiles) != 1 || resolved.Profiles[0].Name != "cheap" {
		t.Fatalf("resolved profiles = %#v, want one legacy profile named cheap", resolved.Profiles)
	}
	if len(resolved.AgentOverrides) != 0 {
		t.Fatalf("expected no agent overrides when local config omits them, got %#v", resolved.AgentOverrides)
	}
}

func TestResolveLocalRuntimeConfigReadsDefaultProfileFromCanonicalConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	configDir := filepath.Join(homeDir, ".config", "gentle-ai-custom")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configJSON := `{
  "version": 1,
  "default_profile": {
    "orchestrator": { "model": "openai/gpt-5.4", "variant": "high" },
    "phases": {
      "sdd-init": { "model": "openai/gpt-5.4", "variant": "medium" },
      "sdd-apply": { "model": "openai/gpt-5.3-codex", "variant": "high" }
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "opencode-local-config.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	resolved, err := resolveLocalRuntimeConfig(testPolicy(), testPhasesSet())
	if err != nil {
		t.Fatalf("resolve local runtime config: %v", err)
	}
	if resolved.DefaultProfile == nil {
		t.Fatalf("expected default profile to be loaded")
	}
	if resolved.DefaultProfile.Orchestrator.Model != "openai/gpt-5.4" {
		t.Fatalf("unexpected default profile orchestrator: %#v", resolved.DefaultProfile.Orchestrator)
	}
	if resolved.DefaultProfile.Phases["sdd-init"].Variant != "medium" {
		t.Fatalf("unexpected default profile phases: %#v", resolved.DefaultProfile.Phases)
	}
}

func TestResolveLocalRuntimeConfigEmptyProfilesSuppressesLegacy(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	configDir := filepath.Join(homeDir, ".config", "gentle-ai-custom")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	// Canonical config explicitly declares profiles: [] (empty array).
	if err := os.WriteFile(filepath.Join(configDir, "opencode-local-config.json"), []byte("{\n  \"version\": 1,\n  \"profiles\": []\n}\n"), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	// Legacy file is present — must be suppressed because profiles was defined (as []).
	if err := os.WriteFile(filepath.Join(configDir, "opencode-sdd-profiles.json"), []byte(legacyProfilesJSON()), 0o644); err != nil {
		t.Fatalf("write legacy profiles: %v", err)
	}

	resolved, err := resolveLocalRuntimeConfig(testPolicy(), testPhasesSet())
	if err != nil {
		t.Fatalf("resolve local runtime config: %v", err)
	}
	if !resolved.ProfilesDefined {
		t.Fatalf("expected ProfilesDefined=true when profiles: [] is explicitly set")
	}
	if len(resolved.Profiles) != 0 {
		t.Fatalf("expected empty profiles slice, got %#v", resolved.Profiles)
	}
	if resolved.UsedLegacyProfiles {
		t.Fatalf("expected legacy fallback suppressed when canonical config defines profiles: []")
	}
}

func testPolicy() Policy {
	policy := Policy{}
	policy.Upstream.RepoName = "gentle-ai"
	policy.OpenCode.ConfigPath = "~/.config/opencode/opencode.json"
	policy.OpenCode.LocalConfigPath = "~/.config/gentle-ai-custom/opencode-local-config.json"
	policy.OpenCode.LegacyProfilesLocalConfigPath = "~/.config/gentle-ai-custom/opencode-sdd-profiles.json"
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}
	return policy
}

func testPhasesSet() map[string]bool {
	return map[string]bool{
		"sdd-init":  true,
		"sdd-apply": true,
	}
}

func legacyProfilesJSON() string {
	return `{
  "version": 1,
  "profiles": [
    {
      "name": "cheap",
      "orchestrator": { "model": "openai/gpt-5.4-mini", "variant": "low" },
      "phases": {
        "sdd-init": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-apply": { "model": "openai/gpt-5.4-mini", "variant": "low" }
      }
    }
  ]
}
`
}
