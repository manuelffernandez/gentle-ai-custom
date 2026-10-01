package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var openCodeV2PluginNames = []string{
	"opencode-review-transport.ts",
	"skill-registry.ts",
	"sdd-task-result-artifacts.ts",
	"telemetry-runtime.ts",
	"model-variants.ts",
}

func writeOpenCodeV2PluginFixtures(t *testing.T, upstreamRepo string) {
	t.Helper()
	for _, name := range openCodeV2PluginNames {
		path := filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins-v2", name)
		mustWriteFile(t, path, []byte("// fixture: "+name+"\n"))
	}
}

func TestCategorizeGitDiffTreatsOnlyOpenCodeAndEngramSourcesAsManaged(t *testing.T) {
	target := testOpenCodeAgentsTarget()
	managed, unmanaged := categorizeGitDiff([]GitDiffEntry{
		{Status: "M", Path: "internal/assets/opencode/persona-gentleman.md"},
		{Status: "A", Path: "internal/assets/opencode/background-subagents.md"},
		{Status: "M", Path: "internal/assets/engram/protocol.md"},
		{Status: "A", Path: "internal/assets/generic/remote-authorization-contract.md"},
		{Status: "M", Path: "internal/assets/opencode/sdd-overlay-single.json"},
		{Status: "M", Path: "internal/assets/opencode/sdd-overlay-multi.json"},
		{Status: "M", Path: "internal/assets/opencode/plugins/model-variants.ts"},
		{Status: "M", Path: "internal/assets/opencode/plugins-v2/opencode-review-transport.ts"},
		{Status: "M", Path: "internal/assets/opencode/plugins-v2/skill-registry.ts"},
		{Status: "M", Path: "internal/assets/opencode/plugins-v2/sdd-task-result-artifacts.ts"},
		{Status: "M", Path: "internal/assets/opencode/plugins-v2/telemetry-runtime.ts"},
		{Status: "M", Path: "internal/assets/opencode/plugins-v2/model-variants.ts"},
		{Status: "M", Path: "internal/assets/claude/agents/jd-fix-agent.md"},
		{Status: "M", Path: "internal/assets/claude/commands/sdd-apply.md"},
		{Status: "A", Path: "internal/assets/claude/output-style-neutral.md"},
	}, buildManagedAssetCatalog(target))

	if len(managed) != 12 {
		t.Fatalf("managed entries = %d, want 12", len(managed))
	}
	if len(unmanaged) != 0 {
		t.Fatalf("unmanaged entries = %d, want 0", len(unmanaged))
	}
	keys := map[string]bool{}
	for _, entry := range managed {
		keys[entry.Record.Key] = true
	}
	for _, key := range []string{"opencode-persona-source", "opencode-background-subagents-source", "opencode-engram-source", "opencode-remote-authorization-source", "opencode-overlay-single", "opencode-overlay-multi", "opencode-plugins", "opencode-plugins-v2"} {
		if !keys[key] {
			t.Fatalf("managed record keys missing %q", key)
		}
	}
	for _, key := range []string{"claude-agents", "claude-commands", "claude-output-style-neutral"} {
		if keys[key] {
			t.Fatalf("managed record keys unexpectedly include %q", key)
		}
	}
}

func TestSyncManagedTargetCopiesOpenCodeAgentsSourceFilesFromUpstream(t *testing.T) {
	upstreamRepo := t.TempDir()
	personaContent := []byte("## Rules\n\n- Upstream persona.\n")
	backgroundPolicyContent := []byte("<!-- gentle-ai:opencode-background-subagents -->\nbackground policy\n<!-- /gentle-ai:opencode-background-subagents -->\n")
	engramContent := []byte("## Engram Persistent Memory — Protocol\n\n- mem_save\n")
	remoteAuthorizationContent := []byte("## Remote operation authorization\n\nRemain local without explicit authorization.\n")
	overlaySingleContent := []byte("{\n  \"name\": \"single\"\n}\n")
	overlayMultiContent := []byte("{\n  \"name\": \"multi\"\n}\n")
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), personaContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "background-subagents.md"), backgroundPolicyContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "engram", "protocol.md"), engramContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "generic", "remote-authorization-contract.md"), remoteAuthorizationContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-single.json"), overlaySingleContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-multi.json"), overlayMultiContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "model-variants.ts"), []byte("export const modelVariants = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "skill-registry.ts"), []byte("export const skillRegistry = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "background-agents.ts"), []byte("export const backgroundAgents = [];\n"))
	writeOpenCodeV2PluginFixtures(t, upstreamRepo)

	repoRoot := t.TempDir()
	stats := &syncUpstreamAssetsStats{}
	if err := syncManagedTarget(repoRoot, upstreamRepo, testOpenCodeAgentsTarget(), newVerboseRecorder(false), stats); err != nil {
		t.Fatalf("syncManagedTarget() error = %v", err)
	}

	// Fix #4: verify stats
	if stats.FilesNew != 14 {
		t.Fatalf("stats.FilesNew = %d, want 14", stats.FilesNew)
	}
	if stats.DirsSynced != 2 {
		t.Fatalf("stats.DirsSynced = %d, want 2", stats.DirsSynced)
	}

	// Verify persona-gentleman.md was copied verbatim
	gotPersona, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "persona-gentleman.md"))
	if err != nil {
		t.Fatalf("read persona mirror: %v", err)
	}
	if string(gotPersona) != string(personaContent) {
		t.Fatalf("persona mirror content mismatch\nwant: %q\ngot:  %q", personaContent, gotPersona)
	}
	gotBackgroundPolicy, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "background-subagents.md"))
	if err != nil {
		t.Fatalf("read background policy mirror: %v", err)
	}
	if string(gotBackgroundPolicy) != string(backgroundPolicyContent) {
		t.Fatalf("background policy mirror content mismatch\nwant: %q\ngot:  %q", backgroundPolicyContent, gotBackgroundPolicy)
	}

	// Verify engram-protocol.md was copied verbatim
	gotEngram, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "engram-protocol.md"))
	if err != nil {
		t.Fatalf("read engram mirror: %v", err)
	}
	if string(gotEngram) != string(engramContent) {
		t.Fatalf("engram mirror content mismatch\nwant: %q\ngot:  %q", engramContent, gotEngram)
	}
	gotRemoteAuthorization, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "remote-authorization-contract.md"))
	if err != nil {
		t.Fatalf("read remote authorization mirror: %v", err)
	}
	if string(gotRemoteAuthorization) != string(remoteAuthorizationContent) {
		t.Fatalf("remote authorization mirror content mismatch\nwant: %q\ngot:  %q", remoteAuthorizationContent, gotRemoteAuthorization)
	}

	gotSingle, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "sdd-overlay-single.json"))
	if err != nil {
		t.Fatalf("read single overlay mirror: %v", err)
	}
	if string(gotSingle) != string(overlaySingleContent) {
		t.Fatalf("single overlay mirror content mismatch\nwant: %q\ngot:  %q", overlaySingleContent, gotSingle)
	}

	gotMulti, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "sdd-overlay-multi.json"))
	if err != nil {
		t.Fatalf("read multi overlay mirror: %v", err)
	}
	if string(gotMulti) != string(overlayMultiContent) {
		t.Fatalf("multi overlay mirror content mismatch\nwant: %q\ngot:  %q", overlayMultiContent, gotMulti)
	}

	for _, name := range []string{"model-variants.ts", "skill-registry.ts", "background-agents.ts"} {
		if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "plugins", name)); err != nil {
			t.Fatalf("read plugin mirror %s: %v", name, err)
		}
	}
	for _, name := range openCodeV2PluginNames {
		if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "plugins-v2", name)); err != nil {
			t.Fatalf("read v2 plugin mirror %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "claude")); !os.IsNotExist(err) {
		t.Fatalf("sync wrote unsupported claude snapshot tree")
	}

	// Verify no AGENTS.md was written (no materialization in sync)
	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "AGENTS.md")); err == nil {
		t.Fatalf("sync wrote AGENTS.md but should not materialize during sync")
	}
}

func TestFormatUpstreamAssetSyncNoteRecordsPreviousAndNewCommit(t *testing.T) {
	got := formatUpstreamAssetSyncNote(
		"b6737129a295980962ec5e8daf41c971fac700b9",
		"b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5",
	)
	want := "Updated approved upstream asset snapshots via sync-upstream-assets from b6737129a295980962ec5e8daf41c971fac700b9 to b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5."
	if got != want {
		t.Fatalf("formatUpstreamAssetSyncNote() = %q, want %q", got, want)
	}
}

func TestFormatUpstreamAssetSyncNoteHandlesUnknownPreviousCommit(t *testing.T) {
	got := formatUpstreamAssetSyncNote("", "b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5")
	want := "Updated approved upstream asset snapshots via sync-upstream-assets to b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5."
	if got != want {
		t.Fatalf("formatUpstreamAssetSyncNote() = %q, want %q", got, want)
	}
}

func TestRunSyncUpstreamAssetsRecordsPreviousMaintainedCommitInStateNote(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), []byte("## Rules\n\n- Persona v1.\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "background-subagents.md"), []byte("background policy\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "engram", "protocol.md"), []byte("## Engram Persistent Memory — Protocol\n\n- v1\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "generic", "remote-authorization-contract.md"), []byte("## Remote operation authorization\n\n- v1\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-single.json"), []byte("{\n  \"name\": \"single\"\n}\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-multi.json"), []byte("{\n  \"name\": \"multi\"\n}\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "model-variants.ts"), []byte("export const modelVariants = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "skill-registry.ts"), []byte("export const skillRegistry = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "background-agents.ts"), []byte("export const backgroundAgents = [];\n"))
	writeOpenCodeV2PluginFixtures(t, upstreamRepo)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "components", "sdd", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles(key string) (string, bool) {
	profileName, managed := managedProfileAgentName(key)
	return profileName, managed
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	taskPerms := map[string]any{
		"*":       "deny",
		"general": "allow",
	}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "components", "sdd", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(agent)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	writeStructuralRoleFixture(t, upstreamRepo)
	mustInitGitRepo(t, upstreamRepo)

	previousCommit, err := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse previous HEAD: %v", err)
	}

	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), []byte("## Rules\n\n- Persona v2.\n"))
	if _, err := runGit(upstreamRepo, false, "add", "."); err != nil {
		t.Fatalf("git add updated upstream fixture: %v", err)
	}
	if _, err := runGit(upstreamRepo, false, "commit", "-m", "update persona snapshot"); err != nil {
		t.Fatalf("git commit updated upstream fixture: %v", err)
	}

	upstreamHead, err := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse upstream HEAD: %v", err)
	}
	if _, err := runGit(upstreamRepo, false, "tag", "v2.0.0"); err != nil {
		t.Fatalf("git tag upstream release: %v", err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.0.0'\n")

	repoRoot := t.TempDir()
	t.Setenv(upstreamRepoEnvVar, upstreamRepo)

	policy := Policy{}
	policy.Version = 1
	policy.Name = "test-overlay"
	policy.Description = "test policy"
	policy.Maintenance.IntentFile = "overlay/gentle-ai/policy/maintenance-intent.md"
	policy.Maintenance.StateFile = "overlay/gentle-ai/state/upstream-state.json"
	policy.Maintenance.LogFile = "overlay/gentle-ai/logs/update-log.md"
	policy.Upstream.RepoName = "gentle-ai"
	policy.Upstream.OrchestratorPromptPath = "internal/assets/opencode/sdd-orchestrator.md"
	policy.OpenCode.ConfigPath = "~/.config/opencode/opencode.json"
	policy.OpenCode.LocalConfigPath = "~/.config/gentle-ai-custom/opencode-local-config.json"
	policy.OpenCode.BaseOrchestratorKey = "gentle-orchestrator"
	policy.OpenCode.OrchestratorAgentKeys = []string{"gentle-orchestrator"}
	policy.OpenCode.OrchestratorAgentPrefixes = []string{"sdd-orchestrator"}
	policy.OpenCode.ProfileOrchestratorPrefix = "sdd-orchestrator-"
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "gentle-ai-policy.json"), policy)

	manifest := ManagedAssetsManifest{
		Version:     1,
		Status:      "active_runtime_canonical",
		Description: "test manifest",
		Targets: map[string]ManagedAssetsTarget{
			"opencode": {
				WatchRoots:                 testOpenCodeAgentsTarget().WatchRoots,
				StructuralInvariantSources: []string{"internal/components/sdd/profiles.go", "internal/components/sdd/inject.go", "internal/models.go"},
				OwnedOverlayAssets:         testOpenCodeAgentsTarget().OwnedOverlayAssets,
			},
		},
	}
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "managed-assets.json"), manifest)

	statePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "state", "upstream-state.json")
	writeJSONFixture(t, statePath, UpstreamState{
		LastMaintainedVersion: "1.0.0",
		LastMaintainedTag:     "v1.0.0",
		LastMaintainedCommit:  previousCommit,
		Notes:                 "before sync",
	})

	if exitCode := RunSyncUpstreamAssets(repoRoot, nil); exitCode != 0 {
		t.Fatalf("RunSyncUpstreamAssets() exit code = %d, want 0", exitCode)
	}

	var updatedState UpstreamState
	if err := readJSONFile(statePath, &updatedState); err != nil {
		t.Fatalf("read updated state: %v", err)
	}

	if updatedState.LastMaintainedCommit != upstreamHead {
		t.Fatalf("updated last_maintained_commit = %q, want %q", updatedState.LastMaintainedCommit, upstreamHead)
	}
	if updatedState.LastMaintainedVersion != "2.0.0" || updatedState.LastMaintainedTag != "v2.0.0" {
		t.Fatalf("updated release identity = version %q tag %q, want 2.0.0 and v2.0.0", updatedState.LastMaintainedVersion, updatedState.LastMaintainedTag)
	}

	wantNote := formatUpstreamAssetSyncNote(previousCommit, upstreamHead)
	if updatedState.Notes != wantNote {
		t.Fatalf("updated notes = %q, want %q", updatedState.Notes, wantNote)
	}

	regressionNote := formatUpstreamAssetSyncNote(upstreamHead, upstreamHead)
	if updatedState.Notes == regressionNote {
		t.Fatalf("updated notes regressed to head->head note %q", regressionNote)
	}
}

func TestOpenCodeAgentsOwnedSourceMatchesApprovedInheritedSections(t *testing.T) {
	if testing.Short() {
		t.Skip("integration: reads committed repo files to verify owned AGENTS.md invariants")
	}
	repoRoot := repoRootForTest(t)
	ownedSnapshot := mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md"))

	if err := validateOwnedAgentsComposition(ownedSnapshot); err != nil {
		t.Fatalf("owned AGENTS.md composition: %v", err)
	}
	plainLanguage := mustMarkdownBetweenUniqueMarkers(t, ownedSnapshot, "<!-- gentle-ai-custom:plain-language-communication -->", "<!-- /gentle-ai-custom:plain-language-communication -->")
	if !strings.Contains(plainLanguage, "When responding to the user, use simple, natural language without assuming familiarity with the implementation's logic.") {
		t.Fatal("owned AGENTS.md is missing the approved plain-language communication rule")
	}

	if strings.Contains(ownedSnapshot, "Agent Skills Index") {
		t.Fatalf("owned AGENTS.md still contains old root index content")
	}

	if !strings.HasPrefix(strings.TrimLeft(ownedSnapshot, " \t\r\n"), "<!-- gentle-ai:persona -->") {
		t.Fatalf("owned AGENTS.md does not start with the upstream persona section marker")
	}

	personaSnapshot := mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "persona-gentleman.md"))
	ownedPersona := mustMarkdownBetweenUniqueMarkers(t, ownedSnapshot, "<!-- gentle-ai:persona -->", "<!-- /gentle-ai:persona -->")
	if normalizeMarkdownSection(ownedPersona) != normalizeMarkdownSection(personaSnapshot) {
		t.Fatal("owned AGENTS.md persona section is stale relative to the approved upstream snapshot")
	}

	engramSnapshot := mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "engram-protocol.md"))
	upstreamFullEngram := mustMarkdownBetweenUniqueMarkers(t, engramSnapshot, "<!-- section:full -->", "<!-- /section:full -->")
	ownedEngram := mustMarkdownBetweenUniqueMarkers(t, ownedSnapshot, "<!-- gentle-ai:engram-protocol -->", "<!-- /gentle-ai:engram-protocol -->")
	if normalizeMarkdownSection(ownedEngram) != normalizeMarkdownSection(upstreamFullEngram) {
		t.Fatal("owned AGENTS.md Engram section is stale relative to the approved OpenCode full projection")
	}

	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "claude")); !os.IsNotExist(err) {
		t.Fatalf("owned AGENTS baseline should not materialize the claude snapshot tree")
	}
}

func TestValidateOwnedAgentsCompositionRejectsAmbiguousOrExtraBlocks(t *testing.T) {
	blocks := ownedAgentsBlocks()
	valid := renderMarkdownBlocks(blocks)
	if err := validateOwnedAgentsComposition(strings.ReplaceAll(valid, "\n", "\r\n")); err != nil {
		t.Fatalf("validateOwnedAgentsComposition() rejected valid CRLF composition: %v", err)
	}
	persona := renderMarkdownBlocks(blocks[:1])
	personaClosingBeforeOpening := blocks[0].closing + "\ncontent\n" + blocks[0].opening + "\n"

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "duplicate persona",
			content: valid + renderMarkdownBlocks(blocks[:1]),
		},
		{
			name:    "duplicate Engram",
			content: valid + renderMarkdownBlocks(blocks[3:4]),
		},
		{
			name:    "extra overlay top-level section",
			content: valid + "\n## Unexpected Overlay Policy\n",
		},
		{
			name:    "inherited blocks out of order",
			content: renderMarkdownBlocks([]markdownBlock{blocks[3], blocks[0], blocks[1], blocks[2], blocks[4], blocks[5], blocks[6], blocks[7], blocks[8]}),
		},
		{
			name:    "closing marker before opening marker",
			content: strings.Replace(valid, persona, personaClosingBeforeOpening, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOwnedAgentsComposition(tt.content); err == nil {
				t.Fatal("validateOwnedAgentsComposition() error = nil, want invalid composition rejection")
			}
		})
	}
}

type markdownBlock struct {
	name             string
	opening          string
	closing          string
	overlayH2Heading string
}

func ownedAgentsBlocks() []markdownBlock {
	return []markdownBlock{
		{name: "persona", opening: "<!-- gentle-ai:persona -->", closing: "<!-- /gentle-ai:persona -->"},
		{name: "Agent-Readable Markdown Standard", opening: "<!-- gentle-ai-custom:agent-readable-markdown -->", closing: "<!-- /gentle-ai-custom:agent-readable-markdown -->", overlayH2Heading: "Agent-Readable Markdown Standard"},
		{name: "Credential-File Safety", opening: "<!-- gentle-ai-custom:credential-file-safety -->", closing: "<!-- /gentle-ai-custom:credential-file-safety -->", overlayH2Heading: "Credential-File Safety"},
		{name: "Remote Operation Authorization", opening: "<!-- gentle-ai-custom:remote-authorization -->", closing: "<!-- /gentle-ai-custom:remote-authorization -->", overlayH2Heading: "Remote Operation Authorization"},
		{name: "OpenCode full Engram", opening: "<!-- gentle-ai:engram-protocol -->", closing: "<!-- /gentle-ai:engram-protocol -->"},
		{name: "no-auto-commit", opening: "<!-- gentle-ai-custom:no-auto-commit -->", closing: "<!-- /gentle-ai-custom:no-auto-commit -->"},
		{name: "Implementation Scope", opening: "<!-- gentle-ai-custom:implementation-scope -->", closing: "<!-- /gentle-ai-custom:implementation-scope -->", overlayH2Heading: "Implementation Scope"},
		{name: "Plain-Language Communication", opening: "<!-- gentle-ai-custom:plain-language-communication -->", closing: "<!-- /gentle-ai-custom:plain-language-communication -->", overlayH2Heading: "Plain-Language Communication"},
		{name: "Gemini Anti-Sycophancy Override", opening: "<!-- gentle-ai-custom:gemini-override -->", closing: "<!-- /gentle-ai-custom:gemini-override -->", overlayH2Heading: "Gemini Anti-Sycophancy Override"},
	}
}

func validateOwnedAgentsComposition(content string) error {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	cursor := 0
	for _, block := range ownedAgentsBlocks() {
		openingStart, contentStart, contentEnd, closingEnd, err := uniqueMarkdownMarkerBounds(content, block.opening, block.closing)
		if err != nil {
			return fmt.Errorf("%s block: %w", block.name, err)
		}
		if openingStart < cursor {
			return fmt.Errorf("%s block is out of order", block.name)
		}
		if unexpected := strings.TrimSpace(content[cursor:openingStart]); unexpected != "" {
			return fmt.Errorf("unexpected content before %s block: %q", block.name, unexpected)
		}
		if block.name != "persona" && block.name != "OpenCode full Engram" {
			headings := markdownH2Headings(content[contentStart:contentEnd])
			if block.overlayH2Heading == "" && len(headings) != 0 {
				return fmt.Errorf("%s block contains unexpected top-level sections %q", block.name, headings)
			}
			if block.overlayH2Heading != "" && (len(headings) != 1 || headings[0] != block.overlayH2Heading) {
				return fmt.Errorf("%s block top-level sections = %q, want [%q]", block.name, headings, block.overlayH2Heading)
			}
		}
		cursor = closingEnd
	}
	if unexpected := strings.TrimSpace(content[cursor:]); unexpected != "" {
		return fmt.Errorf("unexpected content after declared blocks: %q", unexpected)
	}
	return nil
}

func markdownH2Headings(content string) []string {
	var headings []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	return headings
}

func mustMarkdownBetweenUniqueMarkers(t *testing.T, content, opening, closing string) string {
	t.Helper()
	_, contentStart, contentEnd, _, err := uniqueMarkdownMarkerBounds(content, opening, closing)
	if err != nil {
		t.Fatal(err)
	}
	return content[contentStart:contentEnd]
}

func uniqueMarkdownMarkerBounds(content, opening, closing string) (openingStart, contentStart, contentEnd, closingEnd int, err error) {
	if count := strings.Count(content, opening); count != 1 {
		return 0, 0, 0, 0, fmt.Errorf("opening marker %q count = %d, want 1", opening, count)
	}
	if count := strings.Count(content, closing); count != 1 {
		return 0, 0, 0, 0, fmt.Errorf("closing marker %q count = %d, want 1", closing, count)
	}
	openingStart = strings.Index(content, opening)
	contentStart = openingStart + len(opening)
	contentEnd = strings.Index(content, closing)
	if contentEnd < contentStart {
		return 0, 0, 0, 0, fmt.Errorf("closing marker %q precedes opening marker %q", closing, opening)
	}
	closingEnd = contentEnd + len(closing)
	return openingStart, contentStart, contentEnd, closingEnd, nil
}

func renderMarkdownBlocks(blocks []markdownBlock) string {
	var rendered strings.Builder
	for _, block := range blocks {
		fmt.Fprintf(&rendered, "%s\n", block.opening)
		if block.overlayH2Heading != "" {
			fmt.Fprintf(&rendered, "## %s\n", block.overlayH2Heading)
		}
		fmt.Fprintf(&rendered, "content\n%s\n", block.closing)
	}
	return rendered.String()
}

func normalizeMarkdownSection(content string) string {
	return strings.TrimSpace(strings.ReplaceAll(content, "\r\n", "\n"))
}

func TestUpstreamRuntimeSnapshotsAreNeverOwnedInstallSources(t *testing.T) {
	repoRoot := repoRootForTest(t)
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatalf("loadManagedAssetsTarget() error = %v", err)
	}
	for key, wantPath := range map[string]string{
		"gentle-ai-bench":       "overlay/gentle-ai/assets/upstream/opencode/skills/gentle-ai-bench",
		"rdd-defect-workflow":   "overlay/gentle-ai/assets/upstream/opencode/skills/rdd-defect-workflow",
		"systemic-issue-triage": "overlay/gentle-ai/assets/upstream/opencode/skills/systemic-issue-triage",
	} {
		found := false
		for _, skill := range target.RetainedUpstreamSkills {
			if skill.Key == key {
				found = true
				if skill.RepoUpstreamPath != wantPath {
					t.Fatalf("retained skill %q snapshot path = %q, want %q", key, skill.RepoUpstreamPath, wantPath)
				}
			}
		}
		if !found {
			t.Fatalf("retained upstream skills missing %q", key)
		}
		for _, ownedSkill := range target.RepoOwnedSharedSkills.Allowlist {
			if ownedSkill == key {
				t.Fatalf("retained upstream skill %q must not be a repo-owned install source", key)
			}
		}
	}
	foundSnapshots := 0
	for _, asset := range target.OwnedOverlayAssets {
		if asset.RepoOwnedPath != "" && asset.Class != "repo_owned_runtime" {
			t.Fatalf("non-owned asset %q declares repo_owned_path", asset.Key)
		}
		if asset.Class != "upstream_runtime_snapshot" {
			continue
		}
		foundSnapshots++
		if asset.RepoOwnedPath != "" {
			t.Fatalf("upstream runtime snapshot %q must not declare repo_owned_path", asset.Key)
		}
		if asset.RepoUpstreamPath == "" {
			t.Fatalf("upstream runtime snapshot %q must retain an audit snapshot", asset.Key)
		}
	}
	if foundSnapshots == 0 {
		t.Fatal("managed-assets does not declare upstream runtime snapshots")
	}
	if len(target.RuntimePatches) != 3 || target.RuntimePatches[0].Key != "judgment-retrospective-handoff" || target.RuntimePatches[1].Key != "work-unit-commits-authorization" || target.RuntimePatches[2].Key != "remote-authorization-scope" {
		t.Fatalf("runtime patches = %#v, want Judgment Day, work-unit, and remote-authorization patches", target.RuntimePatches)
	}
	ownedRoot := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode")
	var ownedFiles []string
	if err := filepath.Walk(ownedRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ownedFiles = append(ownedFiles, filepath.Base(path))
		}
		return nil
	}); err != nil {
		t.Fatalf("walk owned runtime tree: %v", err)
	}
	if len(ownedFiles) != 1 || ownedFiles[0] != "AGENTS.md" {
		t.Fatalf("owned OpenCode runtime files = %v, want only AGENTS.md", ownedFiles)
	}
}

func TestOwnedAgentInstructionsOverrideODDCommitMandate(t *testing.T) {
	path := filepath.Join(repoRootForTest(t), "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md")
	content := mustReadFile(t, path)
	for _, required := range []string{
		"This rule overrides any upstream Gentle AI ODD routing or retained `work-unit-commits` instruction",
		"An implementation request, substantial ODD work, task completion, feature branch, or review workflow does not authorize `git commit`",
		"Leave the work unit uncommitted and report that delivery remains pending unless the user explicitly requests a commit",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("owned AGENTS.md is missing the ODD commit-authorization safeguard %q", required)
		}
	}
}

func TestValidateRuntimeOwnershipBoundaryRejectsUpstreamInstallSource(t *testing.T) {
	target := ManagedAssetsTarget{OwnedOverlayAssets: []OwnedOverlayAsset{
		{Key: "audit-only", Class: "upstream_runtime_snapshot", RepoOwnedPath: "overlay/gentle-ai/assets/owned/opencode/forbidden.md"},
	}}
	if err := validateRuntimeOwnershipBoundary(target); err == nil {
		t.Fatal("validateRuntimeOwnershipBoundary() error = nil, want upstream/custom ownership overlap rejection")
	}
}

func TestValidateRuntimeOwnershipBoundaryAcceptsCurrentManifest(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRuntimeOwnershipBoundary(target); err != nil {
		t.Fatalf("validateRuntimeOwnershipBoundary() error = %v", err)
	}
}

func TestBackgroundSubagentsPromptIsAuditOnlyUpstreamSource(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	asset, found := findOwnedOverlayAsset(target.OwnedOverlayAssets, "opencode-background-subagents-source")
	if !found {
		t.Fatal("managed-assets is missing the background-subagents prompt source")
	}
	if asset.Class != "upstream_source" || asset.Kind != "upstream_source_file" {
		t.Fatalf("background-subagents classification = %s/%s, want upstream_source/upstream_source_file", asset.Class, asset.Kind)
	}
	if asset.UpstreamPath != "internal/assets/opencode/background-subagents.md" || asset.RepoUpstreamPath != "overlay/gentle-ai/assets/upstream/opencode/background-subagents.md" {
		t.Fatalf("background-subagents paths = %q -> %q", asset.UpstreamPath, asset.RepoUpstreamPath)
	}
	if asset.RepoOwnedPath != "" || len(asset.RuntimeTargets) != 0 {
		t.Fatalf("background-subagents must remain audit-only, got repo_owned_path=%q runtime_targets=%v", asset.RepoOwnedPath, asset.RuntimeTargets)
	}
}

func TestRemoteAuthorizationContractIsAuditOnlyUpstreamSource(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	asset, found := findOwnedOverlayAsset(target.OwnedOverlayAssets, "opencode-remote-authorization-source")
	if !found {
		t.Fatal("managed-assets is missing the remote authorization contract")
	}
	if asset.Class != "upstream_source" || asset.Kind != "upstream_source_file" {
		t.Fatalf("remote authorization classification = %s/%s, want upstream_source/upstream_source_file", asset.Class, asset.Kind)
	}
	if asset.RepoOwnedPath != "" || len(asset.RuntimeTargets) != 0 {
		t.Fatalf("remote authorization must remain audit-only, got repo_owned_path=%q runtime_targets=%v", asset.RepoOwnedPath, asset.RuntimeTargets)
	}
	for _, suffix := range []string{"components/agentguidance/remote_authorization.go", "components/agentguidance/inject.go", "components/sdd/orchestrator.go", "components/sdd/prompts.go"} {
		if !declaresStructuralSource(target.StructuralInvariantSources, suffix) {
			t.Errorf("managed-assets structural sources missing %q", suffix)
		}
	}
}

func TestOpenCodeV2PluginsAreAuditOnlyUpstreamSource(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	asset, found := findOwnedOverlayAsset(target.OwnedOverlayAssets, "opencode-plugins-v2")
	if !found {
		t.Fatal("managed-assets is missing the OpenCode V2 plugin source directory")
	}
	if asset.Class != "upstream_source" || asset.Kind != "upstream_source_directory" {
		t.Fatalf("V2 plugin classification = %s/%s, want upstream_source/upstream_source_directory", asset.Class, asset.Kind)
	}
	if asset.UpstreamPath != "internal/assets/opencode/plugins-v2" || asset.RepoUpstreamPath != "overlay/gentle-ai/assets/upstream/opencode/plugins-v2" {
		t.Fatalf("V2 plugin paths = %q -> %q", asset.UpstreamPath, asset.RepoUpstreamPath)
	}
	if asset.RepoOwnedPath != "" || len(asset.RuntimeTargets) != 0 {
		t.Fatalf("V2 plugins must remain audit/sync-only, got repo_owned_path=%q runtime_targets=%v", asset.RepoOwnedPath, asset.RuntimeTargets)
	}
}

func TestAuditDriftSummaryDoesNotInferNoMaterializationDriftFromStructuralChecks(t *testing.T) {
	entries := []managedDiffEntry{{Record: managedAssetRecord{Key: "opencode-plugins-v2"}}}
	if !hasManagedDriftOutsideBasePrompt(entries) {
		t.Fatal("managed V2 plugin drift was not classified outside the base prompt")
	}

	summary := buildAuditDriftSummary(true, true, false, true, true, true, true, true, true)
	if len(summary) != 1 || !strings.Contains(summary[0], "do not rule out runtime materialization drift") {
		t.Fatalf("summary = %#v, want explicit materialization uncertainty", summary)
	}

	cleanSummary := buildAuditDriftSummary(true, false, false, true, true, true, true, true, true)
	if len(cleanSummary) != 1 || !strings.Contains(cleanSummary[0], "rather than topology or materialization drift") {
		t.Fatalf("clean summary = %#v, want the prompt-only summary", cleanSummary)
	}
}

func testOpenCodeAgentsTarget() ManagedAssetsTarget {
	return ManagedAssetsTarget{
		WatchRoots: []string{
			"internal/assets/opencode",
			"internal/assets/engram/protocol.md",
			"internal/assets/skills",
			"internal/assets/generic/remote-authorization-contract.md",
		},
		OwnedOverlayAssets: []OwnedOverlayAsset{
			{
				Key:              "opencode-persona-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/persona-gentleman.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/persona-gentleman.md",
			},
			{
				Key:              "opencode-background-subagents-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/background-subagents.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/background-subagents.md",
			},
			{
				Key:              "opencode-engram-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/engram/protocol.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/engram-protocol.md",
			},
			{
				Key:              "opencode-remote-authorization-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/generic/remote-authorization-contract.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/remote-authorization-contract.md",
			},
			{
				Key:              "opencode-overlay-single",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/sdd-overlay-single.json",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-single.json",
			},
			{
				Key:              "opencode-overlay-multi",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/sdd-overlay-multi.json",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-multi.json",
			},
			{
				Key:              "opencode-plugins",
				Class:            "upstream_source",
				Kind:             "upstream_source_directory",
				UpstreamPath:     "internal/assets/opencode/plugins",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/plugins",
			},
			{
				Key:              "opencode-plugins-v2",
				Class:            "upstream_source",
				Kind:             "upstream_source_directory",
				UpstreamPath:     "internal/assets/opencode/plugins-v2",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/plugins-v2",
			},
			{
				Key:            "opencode-agents",
				Class:          "repo_owned_runtime",
				Kind:           "agent_instruction_file",
				RepoOwnedPath:  "overlay/gentle-ai/assets/owned/opencode/AGENTS.md",
				RuntimeTargets: []string{"~/.config/opencode/AGENTS.md"},
			},
		},
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot determine working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot locate repo root: go.mod not found in any parent directory")
		}
		dir = parent
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal json fixture %s: %v", path, err)
	}
	mustWriteFile(t, path, append(raw, '\n'))
}

func TestProfileNamingInvariantOKAcceptsManagedProfileDetection(t *testing.T) {
	text := `package opencode

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles() {
	for key := range agentMap {
		profileName, managed := managedProfileAgentName(key)
		_, _ = profileName, managed
	}
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}
`
	if !profileNamingInvariantOK(text) {
		t.Fatal("profileNamingInvariantOK() = false, want true for managed-profile detection")
	}
}

func TestProfileNamingInvariantOKRejectsLiteralDecoys(t *testing.T) {
	text := `package opencode

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles() {
	message := "profileName, managed := managedProfileAgentName(key)"
	_ = message
}

func ProfileAgentKeys(suffix string) []string {
	message := "keys = append(keys, sdd-orchestrator+suffix)"
	_ = message
	return nil
}
`
	if profileNamingInvariantOK(text) {
		t.Fatal("profileNamingInvariantOK() = true, want false for string decoys")
	}
}

func TestProfileTaskScopingInvariantOKAcceptsEquivalentDenyFormatting(t *testing.T) {
	text := `
profileMap := map[string]any{}
taskPerms := map[string]any{
	"*":       "deny",
	"general": "allow",
	"explore": "allow",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	taskPerms[jd] = "allow"
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
profileMap["permission"] = map[string]any{
	"bash": true,
	"task": taskPerms,
}
`
	if !profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = false, want true for equivalent deny-by-default map formatting")
	}
}

func TestProfileTaskScopingInvariantOKRejectsUnwiredDenyMap(t *testing.T) {
	text := `
taskPerms := map[string]any{
	"*": "deny",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	taskPerms[jd] = "allow"
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
otherMap["permission"] = map[string]any{
	"task": randomPerms,
}
`
	if profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = true, want false when deny map is not wired to named profile permissions")
	}
}

func TestProfileTaskScopingInvariantOKRejectsCommentAndStringDecoys(t *testing.T) {
	text := `
// taskPerms := map[string]any{"*": "deny"}
message := "taskPerms[phase+suffix] = \"allow\"; taskPerms[jd] = \"allow\"; taskPerms[reviewAgent] = \"allow\"; profileMap[\"permission\"] = map[string]any{\"task\": taskPerms}"
_ = message
`
	if profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = true, want false for comment/string decoys")
	}
}

func TestProfileTaskScopingInvariantOKAcceptsV2ReplacementWiring(t *testing.T) {
	text := `
taskPerms := map[string]any{
	"*": "deny",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	if hasProfileAssignment(profile, jd) {
		taskPerms[jd+suffix] = "allow"
	} else {
		taskPerms[jd] = "allow"
	}
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
orchEntry := map[string]any{
	"permission": map[string]any{
		"task": map[string]any{
			"__replace__": taskPerms,
		},
	},
}
_ = orchEntry
`
	if !profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = false, want true for v2 __replace__ permission wiring")
	}
}

func TestEvaluateStructuralInvariantsRejectsProfileTaskScopingCommentDecoy(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles(key string) (string, bool) {
	profileName, managed := managedProfileAgentName(key)
	return profileName, managed
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring() {
	profileMap := map[string]any{
		"tools": map[string]any{"task": true},
	}
	_ = profileMap
	// taskPerms := map[string]any{"*": "deny"}
	message := "taskPerms[phase+suffix] = \"allow\"; taskPerms[jd] = \"allow\"; taskPerms[reviewAgent] = \"allow\"; profileMap[\"permission\"] = map[string]any{\"task\": taskPerms}"
	_ = message
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	writeStructuralRoleFixture(t, upstreamRepo)
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go", "internal/models.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if result.TaskScopingOK {
		t.Fatal("evaluateStructuralInvariants() TaskScopingOK = true, want false for profile task-scoping comment/string decoy")
	}
	if len(failures) == 0 {
		t.Fatal("evaluateStructuralInvariants() failures = none, want task-scoping invariant failure")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "deny-by-default task permission map") {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want task-scoping failure", failures)
	}
}

func TestBaseAssetInjectionInvariantOKValidatesCanonicalRenderArguments(t *testing.T) {
	tests := []struct {
		name      string
		composer  string
		arguments string
		want      bool
	}{
		{name: "real centralized composer shape", composer: "composeOrchestratorPrompt", arguments: "agent, renderOptions", want: true},
		{name: "centralized composer missing resolved options", composer: "composeOrchestratorPrompt", arguments: "agent", want: false},
		{name: "centralized composer with arbitrary agent", composer: "composeOrchestratorPrompt", arguments: "candidate, renderOptions", want: false},
		{name: "centralized composer with arbitrary options", composer: "composeOrchestratorPrompt", arguments: "agent, candidateOptions", want: false},
		{name: "current renderer with resolved options", composer: "renderSDDOrchestratorAsset", arguments: "agent, renderOptions", want: true},
		{name: "historical renderer with adapter-local agent", composer: "renderSDDOrchestratorAsset", arguments: "agent", want: true},
		{name: "historical renderer with OpenCode selector", composer: "renderSDDOrchestratorAsset", arguments: "model.AgentOpenCode", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = ` + tt.composer + `(` + tt.arguments + `)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
			if got := baseAssetInjectionInvariantOK(text); got != tt.want {
				t.Fatalf("baseAssetInjectionInvariantOK() = %v, want %v for %s(%s)", got, tt.want, tt.composer, tt.arguments)
			}
		})
	}
}

func TestBaseAssetInjectionInvariantOKAcceptsLegacyAssetRead(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if !baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = false, want true for legacy asset-read binding")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsDecoyRenderCall(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
_ = renderSDDOrchestratorAsset(model.AgentOpenCode)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for decoy render call without prompt assignment")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsWrongAssignmentTarget(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
profileMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false when render path is assigned to the wrong map")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsArbitraryRenderArgument(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = renderSDDOrchestratorAsset(candidate)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for arbitrary render argument")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsArbitraryRenderOptions(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = composeOrchestratorPrompt(agent, candidateOptions)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for unverified composition options")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsMissingPromptAssignment(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false when prompt assignment is missing")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsCommentAndStringDecoys(t *testing.T) {
	text := `
// orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
message := "orchestratorMap[\"prompt\"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))"
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
taskRaw, ok := permissionRaw["task"].(map[string]any)
_ = message
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for comment/string decoys")
	}
}

func TestBaseAssetInjectionInvariantOKAcceptsV290CompositionChain(t *testing.T) {
	injectSource := `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	orchestratorMap["prompt"] = orchestratorPrompt
}`
	if !baseAssetInjectionInvariantOK(injectSource, upstreamV290OrchestratorCompositionFixture) {
		t.Fatal("baseAssetInjectionInvariantOK() = false, want true for bound v2.9.0 composition chain")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsUnboundV290PromptAssignment(t *testing.T) {
	injectSource := `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt := arbitraryPrompt
	_, _ = composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	orchestratorMap["prompt"] = orchestratorPrompt
}`
	if baseAssetInjectionInvariantOK(injectSource, upstreamV290OrchestratorCompositionFixture) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for unbound composition decoy")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsNestedBindingDecoy(t *testing.T) {
	injectSource := `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt := arbitraryPrompt
	{
		orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
		_, _ = orchestratorPrompt, err
	}
	orchestratorMap["prompt"] = orchestratorPrompt
}`
	if baseAssetInjectionInvariantOK(injectSource, upstreamV290OrchestratorCompositionFixture) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for a canonical call bound only in a nested decoy scope")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsBrokenV290AssetChain(t *testing.T) {
	injectSource := `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	orchestratorMap["prompt"] = orchestratorPrompt
}`
	broken := strings.Replace(upstreamV290OrchestratorCompositionFixture, `assets.MustRead(path)`, `assets.MustRead("opencode/decoy.md")`, 1)
	if baseAssetInjectionInvariantOK(injectSource, broken) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false when the composer bypasses the selected canonical asset")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsReassignedV290Prompt(t *testing.T) {
	tests := []struct {
		name   string
		inject string
	}{
		{
			name: "direct overwrite",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	orchestratorPrompt = arbitraryPrompt
	orchestratorMap["prompt"] = orchestratorPrompt
	}`,
		},
		{
			name: "conditional overwrite",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	if replacePrompt { orchestratorPrompt = arbitraryPrompt }
	orchestratorMap["prompt"] = orchestratorPrompt
	}`,
		},
		{
			name: "self-consuming overwrite",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	orchestratorPrompt = arbitraryTransform(orchestratorPrompt)
	orchestratorMap["prompt"] = orchestratorPrompt
	}`,
		},
		{
			name: "conditional sink overwrite",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
	_ = err
	orchestratorMap["prompt"] = orchestratorPrompt
	if replacePrompt { orchestratorMap["prompt"] = arbitraryPrompt }
	}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if baseAssetInjectionInvariantOK(tt.inject, upstreamV290OrchestratorCompositionFixture) {
				t.Fatal("baseAssetInjectionInvariantOK() = true, want false when the composed prompt is overwritten before assignment")
			}
		})
	}
}

func TestBaseAssetInjectionInvariantOKRejectsV290DecoyControlFlow(t *testing.T) {
	tests := []struct {
		name   string
		inject string
	}{
		{
			name: "nested function literal",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	decoy := func() {
		orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
		_ = err
		orchestratorMap["prompt"] = orchestratorPrompt
	}
	_ = decoy
}`,
		},
		{
			name: "dead false branch",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	if false {
		orchestratorPrompt, err := composeOpenCodeOrchestratorPrompt(agent, renderOptions)
		_ = err
		orchestratorMap["prompt"] = orchestratorPrompt
	}
}`,
		},
		{
			name: "direct bypass of validated entrypoint",
			inject: `package sdd
func inlineOpenCodeSDDPrompts() {
	orchestratorMap["prompt"] = composeOrchestratorPrompt(agent, renderOptions)
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if baseAssetInjectionInvariantOK(tt.inject, upstreamV290OrchestratorCompositionFixture) {
				t.Fatal("baseAssetInjectionInvariantOK() = true, want false for decoy or bypassed v2.9 control flow")
			}
		})
	}
}

func TestOrchestratorCompositionInvariantOKRejectsDeadAndNestedDecoys(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{
			name: "nested compose entrypoint return",
			mutate: func(source string) string {
				return strings.Replace(source,
					`return projectSDDSessionPreflight(composeOrchestratorPrompt(agent, options...), "anchor")`,
					`decoy := func() string { return projectSDDSessionPreflight(composeOrchestratorPrompt(agent, options...), "anchor") }; _ = decoy; return "wrong"`, 1)
			},
		},
		{
			name: "dead selected asset read",
			mutate: func(source string) string {
				return strings.Replace(source, `content := assets.MustRead(path)`, `if false { content := assets.MustRead(path); _ = content }; content := "wrong"`, 1)
			},
		},
		{
			name: "selected asset chain after return",
			mutate: func(source string) string {
				return strings.Replace(source, `	path := sddOrchestratorAsset(agent)`, `	return "wrong"
	path := sddOrchestratorAsset(agent)`, 1)
			},
		},
		{
			name: "nested OpenCode case",
			mutate: func(source string) string {
				return strings.Replace(source,
					`switch agent {
	case model.AgentOpenCode, model.AgentKilocode:`,
					`decoy := func() string { switch agent { case model.AgentOpenCode, model.AgentKilocode: return "opencode/sdd-orchestrator.md" }; return "" }; _ = decoy
	switch agent {
	case model.AgentKilocode:`, 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if orchestratorCompositionInvariantOK(tt.mutate(upstreamV290OrchestratorCompositionFixture)) {
				t.Fatal("orchestratorCompositionInvariantOK() = true, want false for dead or nested decoy")
			}
		})
	}
}

func TestOrchestratorCompositionInvariantOKRejectsReassignedAssetChain(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{
			name:   "selected path overwritten before read",
			before: "\tcontent := assets.MustRead(path)",
			after:  "\tpath = arbitraryPath\n\tcontent := assets.MustRead(path)",
		},
		{
			name:   "selected content overwritten before return",
			before: "\treturn bindRuntimeAgentIdentity(content, agent)",
			after:  "\tcontent = arbitraryPrompt\n\treturn bindRuntimeAgentIdentity(content, agent)",
		},
		{
			name:   "selected content passed through untrusted transform",
			before: "\treturn bindRuntimeAgentIdentity(content, agent)",
			after:  "\tcontent = arbitraryTransform(content)\n\treturn bindRuntimeAgentIdentity(content, agent)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broken := strings.Replace(upstreamV290OrchestratorCompositionFixture, tt.before, tt.after, 1)
			if orchestratorCompositionInvariantOK(broken) {
				t.Fatal("orchestratorCompositionInvariantOK() = true, want false after canonical asset-chain reassignment")
			}
		})
	}
}

const upstreamV290OrchestratorCompositionFixture = `package sdd
func composeOpenCodeOrchestratorPrompt(agent model.AgentID, options ...OrchestratorRenderOptions) (string, error) {
	return projectSDDSessionPreflight(composeOrchestratorPrompt(agent, options...), "anchor")
}
func composeOrchestratorPrompt(agent model.AgentID, options ...OrchestratorRenderOptions) string {
	path := sddOrchestratorAsset(agent)
	content := assets.MustRead(path)
	if usesFallbackSessionPreflight(agent) {
		content = composeFallbackSessionPreflight(content, agent)
	}
	content = substituteSharedOrchestratorSections(content)
	content = renderBoundedReviewAssetBodyFromContent(agent, path, content)
	return bindRuntimeAgentIdentity(content, agent)
}
func sddOrchestratorAsset(agent model.AgentID) string {
	switch agent {
	case model.AgentOpenCode, model.AgentKilocode:
		return "opencode/sdd-orchestrator.md"
	default:
		return "generic/sdd-orchestrator.md"
	}
}`

func TestRemoteAuthorizationProjectionInvariantOK(t *testing.T) {
	sources := validRemoteAuthorizationProjectionSources()
	if !remoteAuthorizationProjectionInvariantOK(sources) {
		t.Fatal("remoteAuthorizationProjectionInvariantOK() = false, want true for complete OpenCode projection")
	}
	delete(sources, "components/sdd/profiles.go")
	if remoteAuthorizationProjectionInvariantOK(sources) {
		t.Fatal("remoteAuthorizationProjectionInvariantOK() = true, want false when named-profile projection is missing")
	}
}

func TestRemoteAuthorizationProjectionInvariantOKRejectsDecoys(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
	}{
		{
			name: "helper calls are unused",
			path: "components/agentguidance/remote_authorization.go",
			source: `package agentguidance
func InjectRemoteAuthorization(prompt string) string {
	contract := assets.MustRead("generic/remote-authorization-contract.md")
	_ = filemerge.InjectMarkdownSection(prompt, "remote-authorization", strings.TrimSpace(contract))
	return prompt
}`,
		},
		{
			name: "routing projection is nested in function literal",
			path: "components/agentguidance/inject.go",
			source: `package agentguidance
func InjectRoutingWithOptions() { decoy := func() { rendered = InjectRemoteAuthorization(rendered) }; _ = decoy }`,
		},
		{
			name: "shared prompt projection is dead",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { if false { content = agentguidance.InjectRemoteAuthorization(content) } }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt) }`,
		},
		{
			name: "subagent projection call is unused",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content) }
func injectRemoteAuthorizationIntoSubagentPrompts() { _ = agentguidance.InjectRemoteAuthorization(prompt) }`,
		},
		{
			name: "profile projection is nested in function literal",
			path: "components/sdd/profiles.go",
			source: `package sdd
func GenerateProfileOverlay() { decoy := func() { injectRemoteAuthorizationIntoSubagentPrompts(agentMap) }; _ = decoy }`,
		},
		{
			name: "phase projection is unused",
			path: "components/sdd/inject.go",
			source: `package sdd
func Inject() { _ = agentguidance.InjectRemoteAuthorization(contentStr) }
func inlineOpenCodeSDDPrompts() { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap) }`,
		},
		{
			name: "base projection is dead",
			path: "components/sdd/inject.go",
			source: `package sdd
func Inject() { contentStr = agentguidance.InjectRemoteAuthorization(contentStr) }
func inlineOpenCodeSDDPrompts() { if false { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap) } }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sources := validRemoteAuthorizationProjectionSources()
			sources[tt.path] = tt.source
			if remoteAuthorizationProjectionInvariantOK(sources) {
				t.Fatal("remoteAuthorizationProjectionInvariantOK() = true, want false for dead, nested, or unused projection")
			}
		})
	}
}

func TestRemoteAuthorizationProjectionInvariantOKRejectsReassignmentBeforeSink(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		source string
	}{
		{
			name: "contract overwritten before return",
			path: "components/agentguidance/remote_authorization.go",
			source: `package agentguidance
func InjectRemoteAuthorization(prompt string) string {
	contract := assets.MustRead("generic/remote-authorization-contract.md")
	contract = arbitraryContract
	return filemerge.InjectMarkdownSection(prompt, "remote-authorization", strings.TrimSpace(contract))
}`,
		},
		{
			name: "routing prompt overwritten before return",
			path: "components/agentguidance/inject.go",
			source: `package agentguidance
func InjectRoutingWithOptions() {
	rendered = InjectRemoteAuthorization(rendered)
	rendered = arbitraryPrompt
	switch delivery.kind { case deliveryPrompt: return injectPrompt(delivery, rendered); default: return injectOther(delivery, rendered) }
}`,
		},
		{
			name: "shared prompt overwritten before write",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content); content = arbitraryPrompt; _, _ = filemerge.WriteFileAtomic(path, []byte(content), 0) }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt) }`,
		},
		{
			name: "shared prompt conditionally overwritten before write",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content); if replacePrompt { content = arbitraryPrompt }; _, _ = filemerge.WriteFileAtomic(path, []byte(content), 0) }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt) }`,
		},
		{
			name: "shared prompt passed through untrusted transform before write",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content); content = arbitraryTransform(content); _, _ = filemerge.WriteFileAtomic(path, []byte(content), 0) }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt) }`,
		},
		{
			name: "inline subagent prompt overwritten after projection",
			path: "components/sdd/prompts.go",
			source: `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content); _, _ = filemerge.WriteFileAtomic(path, []byte(content), 0) }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt); agent["prompt"] = arbitraryPrompt }`,
		},
		{
			name: "profile agent map overwritten before marshal",
			path: "components/sdd/profiles.go",
			source: `package sdd
func GenerateProfileOverlay() { injectRemoteAuthorizationIntoSubagentPrompts(agentMap); agentMap = arbitraryMap; result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		},
		{
			name: "profile output carrier overwritten before marshal",
			path: "components/sdd/profiles.go",
			source: `package sdd
func GenerateProfileOverlay() { injectRemoteAuthorizationIntoSubagentPrompts(agentMap); overlay := map[string]any{"agent": agentMap}; overlay = arbitraryOverlay; result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		},
		{
			name: "agent file content overwritten before write",
			path: "components/sdd/inject.go",
			source: `package sdd
func Inject() { contentStr = agentguidance.InjectRemoteAuthorization(contentStr); contentStr = arbitraryPrompt; _, _ = filemerge.WriteFileAtomic(path, []byte(contentStr), 0) }
func inlineOpenCodeSDDPrompts() { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap); result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		},
		{
			name: "inline agent map overwritten before marshal",
			path: "components/sdd/inject.go",
			source: `package sdd
func Inject() { contentStr = agentguidance.InjectRemoteAuthorization(contentStr); _, _ = filemerge.WriteFileAtomic(path, []byte(contentStr), 0) }
func inlineOpenCodeSDDPrompts() { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap); agentsMap = arbitraryMap; result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		},
		{
			name: "inline output carrier overwritten before marshal",
			path: "components/sdd/inject.go",
			source: `package sdd
func Inject() { contentStr = agentguidance.InjectRemoteAuthorization(contentStr); _, _ = filemerge.WriteFileAtomic(path, []byte(contentStr), 0) }
func inlineOpenCodeSDDPrompts() { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap); overlay = arbitraryOverlay; result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sources := validRemoteAuthorizationProjectionSources()
			sources[tt.path] = tt.source
			if remoteAuthorizationProjectionInvariantOK(sources) {
				t.Fatal("remoteAuthorizationProjectionInvariantOK() = true, want false after projection reassignment before sink")
			}
		})
	}
}

func validRemoteAuthorizationProjectionSources() map[string]string {
	return map[string]string{
		"components/agentguidance/remote_authorization.go": `package agentguidance
func InjectRemoteAuthorization(prompt string) string {
	contract := assets.MustRead("generic/remote-authorization-contract.md")
	return filemerge.InjectMarkdownSection(prompt, "remote-authorization", strings.TrimSpace(contract))
}`,
		"components/agentguidance/inject.go": `package agentguidance
func InjectRoutingWithOptions() {
	rendered = InjectRemoteAuthorization(rendered)
	switch delivery.kind { case deliveryPrompt: return injectPrompt(delivery, rendered); default: return injectOther(delivery, rendered) }
}`,
		"components/sdd/prompts.go": `package sdd
func WriteSharedPromptFiles() { content = agentguidance.InjectRemoteAuthorization(content); _, _ = filemerge.WriteFileAtomic(path, []byte(content), 0) }
func injectRemoteAuthorizationIntoSubagentPrompts() { agent["prompt"] = agentguidance.InjectRemoteAuthorization(prompt) }`,
		"components/sdd/profiles.go": `package sdd
func GenerateProfileOverlay() { injectRemoteAuthorizationIntoSubagentPrompts(agentMap); overlay := map[string]any{"agent": agentMap}; result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
		"components/sdd/inject.go": `package sdd
func Inject() { if isMarkdown { contentStr = agentguidance.InjectRemoteAuthorization(contentStr) }; _, _ = filemerge.WriteFileAtomic(path, []byte(contentStr), 0) }
func inlineOpenCodeSDDPrompts() { injectRemoteAuthorizationIntoSubagentPrompts(agentsMap); result, _ := json.MarshalIndent(overlay, "", "  "); return result }`,
	}
}

var upstreamV250ProfilePhaseOrder = []string{
	"sdd-init",
	"sdd-explore",
	"sdd-research",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"sdd-onboard",
}

func writeUpstreamV250StructuralFixture(t *testing.T, upstreamRepo string, phases []string) {
	t.Helper()
	var phaseOrder strings.Builder
	for _, phase := range phases {
		phaseOrder.WriteString("\t\"")
		phaseOrder.WriteString(phase)
		phaseOrder.WriteString("\",\n")
	}
	profilesSource := `package opencode

var profilePhaseOrder = []string{
` + phaseOrder.String() + `}

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles(key string) (string, bool) {
	profileName, managed := managedProfileAgentName(key)
	return profileName, managed
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	taskPerms := map[string]any{
		"*":       "deny",
		"general": "allow",
	}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(profilesSource))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(agent)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	writeStructuralRoleFixture(t, upstreamRepo)
	mustInitGitRepo(t, upstreamRepo)
}

func TestEvaluateStructuralInvariantsAcceptsUpstreamV250Shape(t *testing.T) {
	upstreamRepo := t.TempDir()
	writeUpstreamV250StructuralFixture(t, upstreamRepo, upstreamV250ProfilePhaseOrder)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go", "internal/models.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = append([]string(nil), upstreamV250ProfilePhaseOrder...)

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if len(failures) != 0 {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want none for upstream v2.5.0 shape", failures)
	}
	if !result.TaskScopingOK {
		t.Fatal("evaluateStructuralInvariants() TaskScopingOK = false, want true for equivalent wiring")
	}
	if !result.BaseAssetInjectionOK {
		t.Fatal("evaluateStructuralInvariants() BaseAssetInjectionOK = false, want true")
	}
	if !result.PhaseOrderOK {
		t.Fatal("evaluateStructuralInvariants() PhaseOrderOK = false, want true for the full v2.5.0 phase order")
	}
}

func TestEvaluateStructuralInvariantsRejectsV250PhaseOrderDrift(t *testing.T) {
	tests := []struct {
		name   string
		phases []string
	}{
		{
			name: "missing sdd-research",
			phases: []string{
				"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks",
				"sdd-apply", "sdd-verify", "sdd-archive", "sdd-onboard",
			},
		},
		{
			name: "sdd-research after sdd-propose",
			phases: []string{
				"sdd-init", "sdd-explore", "sdd-propose", "sdd-research", "sdd-spec", "sdd-design",
				"sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive", "sdd-onboard",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstreamRepo := t.TempDir()
			writeUpstreamV250StructuralFixture(t, upstreamRepo, tt.phases)

			target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go", "internal/models.go"}}
			policy := Policy{}
			policy.OpenCode.SDDPhases = append([]string(nil), upstreamV250ProfilePhaseOrder...)

			result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
			if result.PhaseOrderOK {
				t.Fatal("evaluateStructuralInvariants() PhaseOrderOK = true, want false for incompatible v2.5.0 phase order")
			}
			if !strings.Contains(strings.Join(failures, "\n"), "upstream profilePhaseOrder") {
				t.Fatalf("evaluateStructuralInvariants() failures = %v, want profile phase-order failure", failures)
			}
		})
	}
}

func TestEvaluateStructuralInvariantsRejectsBaseAssetCommentDecoy(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func managedProfileAgentPrefixes() []string {
	return []string{"sdd-orchestrator-"}
}

func DetectProfiles(key string) (string, bool) {
	profileName, managed := managedProfileAgentName(key)
	return profileName, managed
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	orchestratorMap := map[string]any{
		"task": true,
	}
	_ = orchestratorMap
	taskPerms := map[string]any{"*": "deny"}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	// orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
	message := "orchestratorMap[\"prompt\"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))"
	_ = message
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	writeStructuralRoleFixture(t, upstreamRepo)
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go", "internal/models.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if result.BaseAssetInjectionOK {
		t.Fatal("evaluateStructuralInvariants() BaseAssetInjectionOK = true, want false for comment/string decoy")
	}
	if len(failures) == 0 {
		t.Fatal("evaluateStructuralInvariants() failures = none, want invariant failure")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "canonical SDD asset render path") {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want base asset injection failure", failures)
	}
}

func mustInitGitRepo(t *testing.T, repo string) {
	t.Helper()
	if _, err := runGit(repo, false, "init"); err != nil {
		t.Fatalf("git init %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "add", "."); err != nil {
		t.Fatalf("git add %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "commit", "-m", "test fixture"); err != nil {
		t.Fatalf("git commit %s: %v", repo, err)
	}
}

func writeStructuralRoleFixture(t *testing.T, repo string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repo, "internal", "models.go"), []byte(`package opencode

func JDPhases() []string {
	return []string{"jd-judge-a", "jd-judge-b", "jd-fix-agent"}
}

const (
	ReviewRefuterAgent = "review-refuter"
	ReviewValidatorAgent = "review-validator"
)

func ReviewLensPhases() []string {
	return []string{"review-risk", "review-readability", "review-reliability", "review-resilience"}
}
`))
}

func TestExtractStringConstantAcceptsGroupedDeclarations(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "file declaration", text: `package opencode
const (
	ReviewRefuterAgent = "review-refuter"
	ReviewValidatorAgent = "review-validator"
)`},
		{name: "standalone snippet", text: `const (
	ReviewRefuterAgent = "review-refuter"
	ReviewValidatorAgent = "review-validator"
)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractStringConstant(tt.text, "ReviewValidatorAgent")
			if err != nil {
				t.Fatal(err)
			}
			if got != "review-validator" {
				t.Fatalf("extractStringConstant() = %q, want review-validator", got)
			}
		})
	}
}

func TestStructuralRoleTopologyRejectsChangedActorSet(t *testing.T) {
	models := `package opencode

func JDPhases() []string {
	return []string{"jd-judge-a", "jd-judge-b", "jd-fix-agent", "jd-extra"}
}

const (
	ReviewRefuterAgent = "review-refuter"
	ReviewValidatorAgent = "review-validator"
)

func ReviewLensPhases() []string {
	return []string{"review-risk", "review-readability", "review-reliability", "review-resilience"}
}
`
	jdActors, err := extractReturnedStringSlice(models, "JDPhases")
	if err != nil {
		t.Fatal(err)
	}
	reviewActors, err := extractReturnedStringSlice(models, "ReviewLensPhases")
	if err != nil {
		t.Fatal(err)
	}
	refuter, err := extractStringConstant(models, "ReviewRefuterAgent")
	if err != nil {
		t.Fatal(err)
	}
	reviewActors = append(reviewActors, refuter)
	validator, err := extractStringConstant(models, "ReviewValidatorAgent")
	if err != nil {
		t.Fatal(err)
	}
	reviewActors = append(reviewActors, validator)
	if sameStrings(jdActors, judgmentDayActorKeys) || !sameStrings(reviewActors, reviewActorKeys) {
		t.Fatalf("unexpected topology comparison: JD=%v review=%v", jdActors, reviewActors)
	}
}
