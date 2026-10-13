package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV4OpenCodeOrchestratorAndAgentAssetsAreAuditOnly(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"gentle-orchestrator",
		"opencode-agent-gentle-ai-explore",
		"opencode-agent-gentle-ai-verify",
		"opencode-agent-gentle-ai-worker",
		"opencode-agent-jd-fix-agent",
		"opencode-agent-jd-judge-a",
		"opencode-agent-jd-judge-b",
		"opencode-agent-review-readability",
		"opencode-agent-review-reliability",
		"opencode-agent-review-resilience",
		"opencode-agent-review-risk",
	}
	if len(want) != 11 {
		t.Fatal("test expectation must cover the 11 new upstream orchestrator/agent assets")
	}
	for _, key := range want {
		asset, found := findOwnedOverlayAsset(target.OwnedOverlayAssets, key)
		if !found {
			t.Errorf("managed-assets is missing v4 asset %q", key)
			continue
		}
		if asset.RepoOwnedPath != "" || len(asset.RuntimeTargets) != 0 {
			t.Errorf("upstream asset %q became a custom runtime source: repo_owned_path=%q runtime_targets=%v", key, asset.RepoOwnedPath, asset.RuntimeTargets)
		}
		if asset.UpstreamPath == "" || asset.RepoUpstreamPath == "" {
			t.Errorf("upstream asset %q is missing its source or snapshot path", key)
		}
	}
}

func TestV4ManifestDropsObsoleteSDDInputsButProtectsLegacyRuntimeFiles(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range target.StructuralInvariantSources {
		if strings.Contains(source, "/sdd/") || strings.HasPrefix(source, "internal/components/sdd/") {
			t.Errorf("v4 structural sources still reference removed SDD code: %q", source)
		}
	}
	for _, asset := range target.OwnedOverlayAssets {
		if strings.HasPrefix(asset.UpstreamPath, "internal/assets/skills/sdd-") || strings.HasPrefix(asset.UpstreamPath, "internal/assets/opencode/sdd-overlay-") {
			t.Errorf("manifest retains an obsolete SDD upstream input: %q", asset.UpstreamPath)
		}
	}
	if len(target.ProtectedRuntimeTargets) == 0 {
		t.Fatal("legacy SDD runtime paths must remain protected from owned-state cleanup")
	}
	if len(target.RetiredUpstreamSnapshots) == 0 {
		t.Fatal("retired snapshots must be explicitly listed for a separately approved sync")
	}
	for _, broadRoot := range []string{"internal/assets/opencode", "internal/assets/skills"} {
		for _, root := range target.WatchRoots {
			if root == broadRoot {
				t.Errorf("watch root %q would treat retired SDD deletions as current unmapped drift", root)
			}
		}
	}
}

func TestV4StructuralInvariantsFollowCurrentAgentGuidanceInjection(t *testing.T) {
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "assigned", "connected")
	mustInitGitRepo(t, upstream)
	target := v4StructuralFixtureTarget()
	result, failures := evaluateStructuralInvariants(upstream, target, Policy{})
	if len(failures) != 0 {
		t.Fatalf("evaluateStructuralInvariants() failures = %v", failures)
	}
	if !result.OrchestratorAssetOK || !result.RoutingInjectionOK || !result.RemoteAuthorizationOK || !result.RoleTopologyOK || !result.ODDTopologyOK || !result.ODDRegistrationOK {
		t.Fatalf("v4 structural result = %#v, want every current invariant true", result)
	}
}

func TestV4OrchestratorBindingRejectsDisconnectedRenderer(t *testing.T) {
	base := v411OrchestratorFixture()
	for _, tt := range []struct {
		name, old, replacement string
	}{
		{"public entrypoint bypasses renderer", "return renderOrchestrator(agent, source, capability, false)", "return \"unrelated\", nil"},
		{"renderer reads decoy path", "content, err := assets.Read(path)", "content, err := assets.Read(\"generic/orchestrator.md\")"},
		{"selected path overwritten", "content, err := assets.Read(path)", "path = \"generic/orchestrator.md\"\n content, err := assets.Read(path)"},
		{"renderer read unreachable", "content, err := assets.Read(path)", "return \"unrelated\", nil\n content, err := assets.Read(path)"},
		{"renderer discards selected content", "return content, nil", "return \"unrelated\", nil"},
		{"renderer returns before selected content", "if err != nil { return \"\", err }", "return \"unrelated\", nil\n if err != nil { return \"\", err }"},
		{"renderer empties selected content", "return content, nil", "content = content[:0]\n return content, nil"},
		{"renderer conditionally empties selected content", "return content, nil", "if agent == AgentOpenCode { content = content[:0] }\n return content, nil"},
		{"renderer switch empties selected content", "return content, nil", "switch agent { case AgentOpenCode: content = content[:0] }\n return content, nil"},
		{"renderer nested switch empties selected content", "return content, nil", "if agent == AgentOpenCode { switch agent { case AgentOpenCode: content = content[:0] } }\n return content, nil"},
		{"renderer returns empty selected content", "return content, nil", "return content[:0], nil"},
		{"wrong OpenCode asset", "return \"opencode/orchestrator.md\"", "return \"generic/orchestrator.md\""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := strings.Replace(base, tt.old, tt.replacement, 1)
			if changed == base {
				t.Fatal("test did not alter renderer fixture")
			}
			if orchestratorAssetBindingInvariantOK(changed) {
				t.Fatal("unbound OpenCode orchestrator passed")
			}
		})
	}
}

func TestV4OrchestratorBindingAcceptsKnownContentTransform(t *testing.T) {
	base := v411OrchestratorFixture()
	transformed := strings.Replace(base, "return content, nil", `content, err = expandSharedOrchestratorSections(content, true, false)
 if err != nil { return "", err }
 content = strings.ReplaceAll(content, "old", "new")
 return strings.TrimSpace(content) + "\n", nil`, 1)
	if transformed == base || !orchestratorAssetBindingInvariantOK(transformed) {
		t.Fatal("known v4.1.1 renderer transforms must preserve selected content")
	}
}

func TestV4OrchestratorBindingAcceptsDirectRenderer(t *testing.T) {
	base := v411OrchestratorFixture()
	direct := strings.Replace(base, `return renderOrchestrator(agent, source, capability, false)`, `path := orchestratorAsset(agent)
 content, err := assets.Read(path)
 if err != nil { return "", err }
 return content, nil`, 1)
	if direct == base || !orchestratorAssetBindingInvariantOK(direct) {
		t.Fatal("direct v4 renderer must preserve the selected OpenCode asset binding")
	}
}

func TestV4ODDRegistrationRejectsDisconnectedPathAwareMerge(t *testing.T) {
	base := v411RunFixture("connected")
	for _, tt := range []struct {
		name, old, replacement string
	}{
		{"wrong settings path", "MergeJSONObjectsForPath(settingsPath, raw, overlay)", "MergeJSONObjectsForPath(otherPath, raw, overlay)"},
		{"wrong source", "MergeJSONObjectsForPath(settingsPath, raw, overlay)", "MergeJSONObjectsForPath(settingsPath, otherRaw, overlay)"},
		{"wrong overlay", "MergeJSONObjectsForPath(settingsPath, raw, overlay)", "MergeJSONObjectsForPath(settingsPath, raw, otherOverlay)"},
		{"merge result discarded", "WriteFileAtomic(settingsPath, merged, 0644)", "WriteFileAtomic(settingsPath, otherMerged, 0644)"},
		{"merge hidden in unreachable branch", "merged, err := filemerge.MergeJSONObjectsForPath(settingsPath, raw, overlay)", "merged := []byte{}; var err error; if false { merged, err = filemerge.MergeJSONObjectsForPath(settingsPath, raw, overlay) }"},
		{"merged bytes replaced before write", "_, err = filemerge.WriteFileAtomic(settingsPath, merged, 0644)", "merged = []byte(`{}`)\n _, err = filemerge.WriteFileAtomic(settingsPath, merged, 0644)"},
		{"merged bytes conditionally replaced before write", "_, err = filemerge.WriteFileAtomic(settingsPath, merged, 0644)", "if err == nil { merged = []byte(`{}`) }\n _, err = filemerge.WriteFileAtomic(settingsPath, merged, 0644)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := strings.Replace(base, tt.old, tt.replacement, 1)
			if changed == base {
				t.Fatal("test did not alter installer fixture")
			}
			if openCodeParityInstallerWritesSpecs(changed) {
				t.Fatal("disconnected ODD parity settings write passed")
			}
		})
	}
}

func TestV4StructuralInvariantsFailClosedWhenRemoteProjectionIsMissing(t *testing.T) {
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "missing", "connected")
	mustInitGitRepo(t, upstream)
	result, failures := evaluateStructuralInvariants(upstream, v4StructuralFixtureTarget(), Policy{})
	if result.RemoteAuthorizationOK {
		t.Fatal("remote authorization invariant passed without projection into the OpenCode routing carrier")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "remote authorization contract") {
		t.Fatalf("failures = %v, want a remote authorization projection failure", failures)
	}
}

func TestV4StructuralInvariantsRejectDiscardedRemoteAuthorizationProjection(t *testing.T) {
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "discarded", "connected")
	mustInitGitRepo(t, upstream)
	result, failures := evaluateStructuralInvariants(upstream, v4StructuralFixtureTarget(), Policy{})
	if result.RoutingInjectionOK || result.RemoteAuthorizationOK {
		t.Fatalf("discarded projection passed: routing=%v remote=%v", result.RoutingInjectionOK, result.RemoteAuthorizationOK)
	}
	if !strings.Contains(strings.Join(failures, "\n"), "always-loaded carrier") {
		t.Fatalf("failures = %v, want routing-carrier failure", failures)
	}
}

func TestV4RoutingInjectionRejectsProjectionOverwrittenBeforeWrite(t *testing.T) {
	for _, mode := range []string{"overwritten", "conditional"} {
		t.Run(mode, func(t *testing.T) {
			upstream := t.TempDir()
			writeV4StructuralFixtureWithModes(t, upstream, mode, "connected")
			mustInitGitRepo(t, upstream)
			result, failures := evaluateStructuralInvariants(upstream, v4StructuralFixtureTarget(), Policy{})
			if result.RoutingInjectionOK || result.RemoteAuthorizationOK {
				t.Fatalf("%s projection passed: routing=%v remote=%v", mode, result.RoutingInjectionOK, result.RemoteAuthorizationOK)
			}
			if !strings.Contains(strings.Join(failures, "\n"), "always-loaded carrier") {
				t.Fatalf("failures = %v, want routing-carrier failure", failures)
			}
		})
	}
}

func TestV4ODDRegistrationFailsWhenSyncDisconnectsMaterialization(t *testing.T) {
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "assigned", "disconnected")
	mustInitGitRepo(t, upstream)
	result, failures := evaluateStructuralInvariants(upstream, v4StructuralFixtureTarget(), Policy{})
	if !result.ODDTopologyOK {
		t.Fatal("ODD name list should remain valid in disconnected-registration fixture")
	}
	if result.ODDRegistrationOK {
		t.Fatal("ODD registration passed when the scheduled routing step no longer invokes materialization")
	}
	syncText, err := os.ReadFile(filepath.Join(upstream, "internal", "cli", "sync.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !syncSchedulesOpenCodeRouting(string(syncText)) {
		t.Fatal("fixture must retain the sync routing stage and isolate its disconnected materialization call")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "settings materialization") {
		t.Fatalf("failures = %v, want registration failure", failures)
	}
}

func TestV4ODDRegistrationRejectsUnreachableInstallerCall(t *testing.T) {
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "assigned", "conditional")
	mustInitGitRepo(t, upstream)
	result, failures := evaluateStructuralInvariants(upstream, v4StructuralFixtureTarget(), Policy{})
	if !result.ODDTopologyOK {
		t.Fatal("ODD name list should remain valid in unreachable-installer fixture")
	}
	if result.ODDRegistrationOK {
		t.Fatal("ODD registration passed with the installer hidden under if false")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "ODD names") {
		t.Fatalf("failures = %v, want ODD registration failure", failures)
	}
}

func v4StructuralFixtureTarget() ManagedAssetsTarget {
	return ManagedAssetsTarget{StructuralInvariantSources: []string{
		"internal/components/agentguidance/orchestrator.go",
		"internal/components/agentguidance/inject.go",
		"internal/components/agentguidance/remote_authorization.go",
		"internal/opencode/models.go",
		"internal/components/opencodeagents/agents.go",
		"internal/cli/run.go",
		"internal/cli/sync.go",
	}}
}

func writeV4StructuralFixture(t *testing.T, root, remoteProjection string, registrationConnected bool) {
	registrationMode := "connected"
	if !registrationConnected {
		registrationMode = "disconnected"
	}
	writeV4StructuralFixtureWithModes(t, root, remoteProjection, registrationMode)
}

func writeV4StructuralFixtureWithModes(t *testing.T, root, remoteProjection, registrationMode string) {
	t.Helper()
	write := func(path, body string) {
		mustWriteFile(t, filepath.Join(root, filepath.FromSlash(path)), []byte(body))
	}
	write("internal/components/agentguidance/orchestrator.go", v411OrchestratorFixture())
	remoteCall := "rendered = InjectRemoteAuthorization(rendered)"
	if remoteProjection == "missing" {
		remoteCall = "rendered = strings.TrimSpace(rendered)"
	} else if remoteProjection == "discarded" {
		remoteCall = "InjectRemoteAuthorization(rendered)"
	} else if remoteProjection == "overwritten" {
		remoteCall += "\n\trendered = RenderRouting()"
	} else if remoteProjection == "conditional" {
		remoteCall = "if false { rendered = InjectRemoteAuthorization(rendered) }"
	}
	write("internal/components/agentguidance/inject.go", `package agentguidance
import "github.com/gentleman-programming/gentle-ai/v4/internal/components/filemerge"
func InjectRoutingWithOptions() {
	rendered := RenderRouting()
	`+remoteCall+`
	orchestrator, err := RenderOrchestratorWithSource()
	if err != nil { return }
	merge := func(existing string) string {
		existing = injectOrchestratorSection(existing, orchestrator)
		return filemerge.InjectMarkdownSection(existing, "agent-routing", rendered)
	}
	injectOrchestratorPrompt(delivery, agent, merge)
	_ = merge
}
`)
	write("internal/components/agentguidance/remote_authorization.go", `package agentguidance
import (
	"strings"
	"github.com/gentleman-programming/gentle-ai/v4/internal/assets"
	"github.com/gentleman-programming/gentle-ai/v4/internal/components/filemerge"
)
func InjectRemoteAuthorization(prompt string) string {
	contract := assets.MustRead("generic/remote-authorization-contract.md")
	return filemerge.InjectMarkdownSection(prompt, "remote-authorization", strings.TrimSpace(contract))
}
`)
	write("internal/opencode/models.go", `package opencode
func JDPhases() []string { return []string{"jd-judge-a", "jd-judge-b", "jd-fix-agent"} }
func ReviewLensPhases() []string { return []string{"review-risk", "review-readability", "review-reliability", "review-resilience"} }
const ReviewRefuterAgent = "review-refuter"
const ReviewValidatorAgent = "review-validator"
func GentleAIODDPhases() []string { return []string{"gentle-ai-explore", "gentle-ai-verify", "gentle-ai-worker"} }
`)
	write("internal/components/opencodeagents/agents.go", `package opencodeagents
import (
	"github.com/gentleman-programming/gentle-ai/v4/internal/assets"
	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
)
type Spec struct { Name string }
var parity = []Spec{{"gentle-ai-explore"}, {"gentle-ai-verify"}, {"gentle-ai-worker"}}
func Parity(agent model.AgentID) []Spec {
	if agent == model.AgentOpenCode { return append([]Spec(nil), parity...) }
	return nil
}
func Entry(spec Spec) (map[string]any, error) {
	prompt, err := assets.Read("opencode/agents/" + spec.Name + ".md")
	return map[string]any{"prompt": prompt}, err
}
`)
	write("internal/cli/run.go", v411RunFixture(registrationMode))
	scheduledRoute := `
	apply = append(apply, agentRoutingGuidanceStep{})
`
	write("internal/cli/sync.go", `package cli
type syncRuntime struct{}
type StagePlan struct { Apply []any }
func (r *syncRuntime) stagePlan() StagePlan {
	apply := []any{}
	`+scheduledRoute+`
	return StagePlan{Apply: apply}
}
`)
}

func v411OrchestratorFixture() string {
	return `package agentguidance
import "github.com/gentleman-programming/gentle-ai/v4/internal/assets"
func orchestratorAsset(agent AgentID) string {
	switch agent { case AgentOpenCode: return "opencode/orchestrator.md" }
	return "generic/orchestrator.md"
}
func RenderOrchestratorWithSource(agent AgentID, source ReviewContractSource, capability string) (string, error) {
	return renderOrchestrator(agent, source, capability, false)
}
func renderOrchestrator(agent AgentID, source ReviewContractSource, capability string, annotate bool) (string, error) {
	path := orchestratorAsset(agent)
	content, err := assets.Read(path)
	if err != nil { return "", err }
	return content, nil
}
	`
}

func v411RunFixture(registrationMode string) string {
	installCall := `
	parityChanged, err := installOpenCodeODDParityAgents(options.SettingsPath)
	_ = parityChanged
	_ = err
`
	if registrationMode == "disconnected" {
		installCall = ""
	} else if registrationMode == "conditional" {
		installCall = "\n\t\tif false {" + installCall + "\t\t}\n"
	}
	return `package cli
import (
	"encoding/json"
	"github.com/gentleman-programming/gentle-ai/v4/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/v4/internal/components/opencodeagents"
	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
)
type agentRoutingGuidanceStep struct { agent model.AgentID }
type syncOptions struct { SettingsPath string }
func (s agentRoutingGuidanceStep) Run(options syncOptions) {
	if s.agent == model.AgentOpenCode {
	` + installCall + `
	}
}
func installOpenCodeODDParityAgents(settingsPath string) (bool, error) {
	return installOpenCodeFamilyParityAgents(settingsPath, model.AgentOpenCode)
}
func installOpenCodeFamilyParityAgents(settingsPath string, agent model.AgentID) (bool, error) {
	raw := []byte{}
	specs := opencodeagents.Parity(agent)
	agentsOverlay := make(map[string]any, len(specs))
	for _, spec := range specs {
		entry, err := opencodeagents.Entry(spec)
		if err != nil { return false, err }
		agentsOverlay[spec.Name] = entry
	}
	overlay, err := json.Marshal(map[string]any{"agent": agentsOverlay})
	if err != nil { return false, err }
	merged, err := filemerge.MergeJSONObjectsForPath(settingsPath, raw, overlay)
	if err != nil { return false, err }
	_, err = filemerge.WriteFileAtomic(settingsPath, merged, 0644)
	return true, err
}
`
}

func TestRemoveRetiredUpstreamSnapshotsIsExactAndPreservesOtherFiles(t *testing.T) {
	repo := t.TempDir()
	root := filepath.Join(repo, filepath.FromSlash(upstreamSnapshotRoot))
	retired := filepath.Join(root, "skills", "sdd-apply", "SKILL.md")
	neighbor := filepath.Join(root, "skills", "sdd-apply", "notes.md")
	mustWriteFile(t, retired, []byte("retired upstream snapshot"))
	mustWriteFile(t, neighbor, []byte("unlisted file"))
	target := ManagedAssetsTarget{RetiredUpstreamSnapshots: []string{"skills/sdd-apply/SKILL.md"}}
	stats := &syncUpstreamAssetsStats{}
	if err := removeRetiredUpstreamSnapshots(repo, target, newVerboseRecorder(false), stats); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(retired); !os.IsNotExist(err) {
		t.Fatalf("retired snapshot still exists, lstat error = %v", err)
	}
	if got, err := os.ReadFile(neighbor); err != nil || string(got) != "unlisted file" {
		t.Fatalf("unlisted neighboring file = %q, error = %v", got, err)
	}
	if stats.FilesDeleted != 1 {
		t.Fatalf("FilesDeleted = %d, want 1", stats.FilesDeleted)
	}
}

func TestRemoveRetiredUpstreamSnapshotsRefusesUnsafeOrActivePaths(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		active string
	}{
		{name: "path traversal", path: "../../outside.txt"},
		{name: "normalized traversal", path: "safe/../outside.txt"},
		{name: "active exact path", path: "agents/current.md", active: "overlay/gentle-ai/assets/upstream/opencode/agents/current.md"},
		{name: "active directory ancestor", path: "agents", active: "overlay/gentle-ai/assets/upstream/opencode/agents/current.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			target := ManagedAssetsTarget{RetiredUpstreamSnapshots: []string{tt.path}}
			if tt.active != "" {
				target.OwnedOverlayAssets = []OwnedOverlayAsset{{RepoUpstreamPath: tt.active}}
			}
			if err := removeRetiredUpstreamSnapshots(repo, target, newVerboseRecorder(false), &syncUpstreamAssetsStats{}); err == nil {
				t.Fatal("removeRetiredUpstreamSnapshots() error = nil, want refusal")
			}
		})
	}
}

func TestRemoveRetiredUpstreamSnapshotsRefusesSymlinkParent(t *testing.T) {
	repo := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "keep.md")
	mustWriteFile(t, outsideFile, []byte("outside user file"))
	snapshotRoot := filepath.Join(repo, filepath.FromSlash(upstreamSnapshotRoot))
	if err := os.MkdirAll(snapshotRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(snapshotRoot, "linked")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	target := ManagedAssetsTarget{RetiredUpstreamSnapshots: []string{"linked/keep.md"}}
	if err := removeRetiredUpstreamSnapshots(repo, target, newVerboseRecorder(false), &syncUpstreamAssetsStats{}); err == nil {
		t.Fatal("removeRetiredUpstreamSnapshots() error = nil, want symlink-parent refusal")
	}
	if got, err := os.ReadFile(outsideFile); err != nil || string(got) != "outside user file" {
		t.Fatalf("outside file after refused cleanup = %q, error = %v", got, err)
	}
}

func TestRemoveRetiredUpstreamSnapshotsPreflightsAllPathsBeforeDeleting(t *testing.T) {
	repo := t.TempDir()
	root := filepath.Join(repo, filepath.FromSlash(upstreamSnapshotRoot))
	first := filepath.Join(root, "retired", "first.md")
	directory := filepath.Join(root, "retired", "directory")
	mustWriteFile(t, first, []byte("keep until every path validates"))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	target := ManagedAssetsTarget{RetiredUpstreamSnapshots: []string{"retired/first.md", "retired/directory"}}
	if err := removeRetiredUpstreamSnapshots(repo, target, newVerboseRecorder(false), &syncUpstreamAssetsStats{}); err == nil {
		t.Fatal("removeRetiredUpstreamSnapshots() error = nil, want directory refusal")
	}
	if got, err := os.ReadFile(first); err != nil || string(got) != "keep until every path validates" {
		t.Fatalf("first retired file after failed preflight = %q, error = %v", got, err)
	}
}
