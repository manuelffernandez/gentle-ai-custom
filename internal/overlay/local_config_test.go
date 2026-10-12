package overlay

import (
	"os"
	"path/filepath"
	"strings"
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

func TestResolveUpstreamRepoIgnoresRuntimeProfileSchemaDuringAuditResolution(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	repoRoot := filepath.Join(t.TempDir(), "gentle-ai-custom")
	upstreamRepo := filepath.Join(homeDir, "upstream")
	configDir := filepath.Join(homeDir, ".config", "gentle-ai-custom")
	for _, dir := range []string{repoRoot, upstreamRepo, configDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	config := `{
  "version": 1,
  "upstream_repo_path": "` + upstreamRepo + `",
  "default_profile": {
    "orchestrator": {"model": "openai/model"},
    "phases": {"sdd-init": {"model": "openai/model"}}
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "opencode-local-config.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	pathValue, _, err := resolveUpstreamRepo(repoRoot, testPolicy())
	if err != nil {
		t.Fatalf("resolve upstream repo with stale runtime profile schema: %v", err)
	}
	if pathValue != upstreamRepo {
		t.Fatalf("resolved upstream = %q, want %q", pathValue, upstreamRepo)
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

func TestLoadLocalConfigRejectsLegacySDDConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		legacy string
	}{
		{name: "default profile", legacy: `"default_profile": {}`},
		{name: "named profiles", legacy: `"profiles": []`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "local.json")
			mustWriteFile(t, path, []byte(`{"version":1,`+tt.legacy+`}`))
			if _, err := loadLocalConfigFile(path); err == nil || !strings.Contains(err.Error(), "legacy") || !strings.Contains(err.Error(), "manually") {
				t.Fatalf("loadLocalConfigFile() error = %v, want actionable legacy refusal", err)
			}
		})
	}
}

func TestLoadLocalConfigPreservesUnrelatedOverridesAndGlobalActors(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	configDir := filepath.Join(homeDir, ".config", "gentle-ai-custom")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configJSON := `{"version":1,"agent_overrides":[{"key":"general","model":"openai/general"}],"global_actors":{"jd-judge-a":{"model":"openai/judge"},"review-risk":{"inherit":true}}}`
	if err := os.WriteFile(filepath.Join(configDir, "opencode-local-config.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	resolved, err := resolveLocalRuntimeConfig(testPolicy())
	if err != nil {
		t.Fatalf("resolve local runtime config: %v", err)
	}
	if len(resolved.AgentOverrides) != 1 || resolved.AgentOverrides[0].Key != "general" {
		t.Fatalf("unrelated agent override not preserved: %#v", resolved.AgentOverrides)
	}
	if resolved.GlobalActors["jd-judge-a"].Model != "openai/judge" || !resolved.GlobalActors["review-risk"].Inherit {
		t.Fatalf("global actor assignments not preserved: %#v", resolved.GlobalActors)
	}
	if _, err := loadLocalConfigFile(filepath.Join(configDir, "opencode-local-config.json")); err != nil {
		t.Fatalf("load v4-compatible local config: %v", err)
	}
}

func TestValidateActorAssignment(t *testing.T) {
	tests := []struct {
		name         string
		value        any
		allowInherit bool
		want         actorAssignment
		wantError    string
	}{
		{name: "explicit model and variant", value: map[string]any{"model": "openai/model", "variant": "high"}, allowInherit: true, want: actorAssignment{Model: "openai/model", Variant: "high"}},
		{name: "explicit model without variant", value: map[string]any{"model": "openai/model"}, allowInherit: true, want: actorAssignment{Model: "openai/model"}},
		{name: "explicit inheritance", value: map[string]any{"inherit": true}, allowInherit: true, want: actorAssignment{Inherit: true}},
		{name: "false inheritance", value: map[string]any{"inherit": false}, allowInherit: true, wantError: "must be exactly true"},
		{name: "string inheritance", value: map[string]any{"inherit": "true"}, allowInherit: true, wantError: "must be exactly true"},
		{name: "inherit and model", value: map[string]any{"inherit": true, "model": "openai/model"}, allowInherit: true, wantError: "cannot be combined"},
		{name: "inherit and variant", value: map[string]any{"inherit": true, "variant": "high"}, allowInherit: true, wantError: "cannot be combined"},
		{name: "variant only", value: map[string]any{"variant": "high"}, allowInherit: true, wantError: "model"},
		{name: "empty object", value: map[string]any{}, allowInherit: true, wantError: "model"},
		{name: "unknown field", value: map[string]any{"model": "openai/model", "other": true}, allowInherit: true, wantError: "unexpected fields"},
		{name: "inheritance rejected when disabled", value: map[string]any{"inherit": true}, allowInherit: false, wantError: "explicit inheritance is not allowed"},
		{name: "malformed value", value: "openai/model", allowInherit: true, wantError: "must be an object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateActorAssignment(tt.value, "actor", tt.allowInherit)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("validateActorAssignment() error = %v, want containing %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateActorAssignment() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("validateActorAssignment() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestValidateAgentOverridesRejectsLegacySDDRoles(t *testing.T) {
	if _, err := validateAgentOverrides([]any{map[string]any{"key": "sdd-apply", "model": "openai/model"}}, "agent_overrides"); err == nil || !strings.Contains(err.Error(), "choose an appropriate ODD role") {
		t.Fatalf("validateAgentOverrides() error = %v, want actionable SDD role refusal", err)
	}
}

func testPolicy() Policy {
	policy := Policy{}
	policy.Upstream.RepoName = "gentle-ai"
	policy.OpenCode.ConfigPath = "~/.config/opencode/opencode.json"
	policy.OpenCode.LocalConfigPath = "~/.config/gentle-ai-custom/opencode-local-config.json"
	return policy
}
