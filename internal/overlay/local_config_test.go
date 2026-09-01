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
		{name: "named inheritance rejected", value: map[string]any{"inherit": true}, allowInherit: false, wantError: "require an explicit model"},
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

func TestLoadLocalConfigValidatesActorFamilies(t *testing.T) {
	phases := []string{"sdd-init", "sdd-apply"}
	phaseSet := testPhasesSet()
	validBase := func() map[string]any {
		return map[string]any{
			"version": 1,
			"default_profile": map[string]any{
				"orchestrator": map[string]any{"model": "openai/orchestrator"},
				"phases": map[string]any{
					"sdd-init":  map[string]any{"model": "openai/init"},
					"sdd-apply": map[string]any{"model": "openai/apply"},
				},
				"judgment_day": map[string]any{"jd-judge-a": map[string]any{"model": "openai/judge"}},
				"review": map[string]any{
					"review-risk":      map[string]any{"inherit": true},
					"review-validator": map[string]any{"model": "openai/validator"},
				},
			},
			"profiles": []any{map[string]any{
				"name": "named", "orchestrator": map[string]any{"model": "openai/orchestrator"},
				"phases": map[string]any{
					"sdd-init":  map[string]any{"model": "openai/init"},
					"sdd-apply": map[string]any{"model": "openai/apply"},
				},
				"judgment_day": map[string]any{"jd-fix-agent": map[string]any{"model": "openai/fix"}},
			}},
		}
	}
	tests := []struct {
		name      string
		mutate    func(map[string]any)
		wantError string
	}{
		{name: "valid explicit and inherit"},
		{name: "unknown Judgment Day actor", mutate: func(config map[string]any) {
			config["default_profile"].(map[string]any)["judgment_day"] = map[string]any{"jd-other": map[string]any{"model": "openai/model"}}
		}, wantError: "unknown actors"},
		{name: "unknown review actor", mutate: func(config map[string]any) {
			config["default_profile"].(map[string]any)["review"] = map[string]any{"review-other": map[string]any{"inherit": true}}
		}, wantError: "unknown actors"},
		{name: "named review rejected", mutate: func(config map[string]any) {
			config["profiles"].([]any)[0].(map[string]any)["review"] = map[string]any{"review-risk": map[string]any{"inherit": true}}
		}, wantError: "unexpected fields"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validBase()
			if tt.mutate != nil {
				tt.mutate(config)
			}
			path := filepath.Join(t.TempDir(), "local.json")
			writeJSONFixture(t, path, config)
			got, err := loadLocalConfigFile(path, phases, phaseSet)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("loadLocalConfigFile() error = %v, want containing %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadLocalConfigFile() error = %v", err)
			}
			if !got.DefaultProfile.Review["review-risk"].Inherit || got.Profiles[0].JudgmentDay["jd-fix-agent"].Model != "openai/fix" {
				t.Fatalf("actor families were not parsed: %#v %#v", got.DefaultProfile.Review, got.Profiles[0].JudgmentDay)
			}
			if got.DefaultProfile.Review["review-validator"].Model != "openai/validator" {
				t.Fatalf("review-validator assignment was not parsed: %#v", got.DefaultProfile.Review)
			}
		})
	}
}

func testPolicy() Policy {
	policy := Policy{}
	policy.Upstream.RepoName = "gentle-ai"
	policy.OpenCode.ConfigPath = "~/.config/opencode/opencode.json"
	policy.OpenCode.LocalConfigPath = "~/.config/gentle-ai-custom/opencode-local-config.json"
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}
	return policy
}

func testPhasesSet() map[string]bool {
	return map[string]bool{
		"sdd-init":  true,
		"sdd-apply": true,
	}
}
