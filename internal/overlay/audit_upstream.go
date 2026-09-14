package overlay

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type auditUpstreamOptions struct{}

func RunAuditUpstream(repoRoot string, args []string) int {
	if exitCode := normalizeAuditUpstreamArgs(args); exitCode >= 0 {
		return exitCode
	}
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	manifest, manifestPath, err := loadManagedAssetsManifest(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	upstreamRepo, upstreamRepoSource, err := resolveUpstreamRepo(repoRoot, policy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	release, err := verifyUpstreamReleaseParity(upstreamRepo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: upstream release parity preflight failed: %v\n", err)
		return 1
	}
	target, ok := manifest.Targets["opencode"]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: managed-assets manifest at %s does not define target %q\n", manifestPath, "opencode")
		return 1
	}
	var state UpstreamState
	statePath := filepath.Join(repoRoot, policy.Maintenance.StateFile)
	if err := readJSONFile(statePath, &state); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: state file is not valid JSON at %s: %v\n", statePath, err)
		return 1
	}

	var failures []string
	var notes []string
	var driftSummary []string

	if strings.TrimSpace(state.LastMaintainedCommit) == "" {
		fmt.Fprintf(os.Stderr, "ERROR: last_maintained_commit is empty in %s\n", statePath)
		return 1
	}
	upstreamHead := release.Commit

	if upstreamHead != "" && state.LastMaintainedCommit != "" && upstreamHead != state.LastMaintainedCommit {
		notes = append(notes, fmt.Sprintf("upstream HEAD %s differs from last maintained commit %s; prompt/invariant drift checks below show whether the baseline still holds", upstreamHead, state.LastMaintainedCommit))
	}
	if state.LastMaintainedTag != "" && release.Tag != state.LastMaintainedTag {
		notes = append(notes, fmt.Sprintf("verified release tag %s differs from last maintained tag %s; review state/log if you are closing a new upstream audit", release.Tag, state.LastMaintainedTag))
	}
	diffEntries, err := runGitDiff(upstreamRepo, state.LastMaintainedCommit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot diff upstream git state from %s: %v\n", state.LastMaintainedCommit, err)
		return 1
	}
	catalog := buildManagedAssetCatalog(target)
	managedEntries, unmanagedEntries := categorizeGitDiff(diffEntries, catalog)
	basePromptDrift := false
	if entry, ok := findManagedAssetDiff(managedEntries, "gentle-orchestrator"); ok {
		basePromptDrift = true
		promptSummary, err := summarizeManagedPromptDrift(upstreamRepo, state.LastMaintainedCommit, entry)
		if err != nil {
			failures = append(failures, err.Error())
		} else {
			driftSummary = append(driftSummary, promptSummary...)
		}
	}
	driftSummary = append(driftSummary, summarizeManagedDiffEntries(managedEntries)...)
	driftSummary = append(driftSummary, summarizeUnmanagedDiffEntries(unmanagedEntries)...)

	if len(managedEntries) > 0 {
		failures = append(failures, fmt.Sprintf("managed asset drift detected in %d upstream file(s) since %s", len(managedEntries), state.LastMaintainedCommit))
	}
	if len(unmanagedEntries) > 0 {
		failures = append(failures, fmt.Sprintf("unmapped watched upstream drift detected in %d file(s); review managed-assets.json before adopting this upstream state", len(unmanagedEntries)))
	}

	structuralResult, structuralFailures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	failures = append(failures, structuralFailures...)
	profileNamingOK := structuralResult.ProfileNamingOK
	taskScopingOK := structuralResult.TaskScopingOK
	baseAssetInjectionOK := structuralResult.BaseAssetInjectionOK
	remoteAuthorizationOK := structuralResult.RemoteAuthorizationOK
	roleTopologyOK := structuralResult.RoleTopologyOK
	phaseOrderOK := structuralResult.PhaseOrderOK
	driftSummary = append(driftSummary, buildAuditDriftSummary(basePromptDrift, phaseOrderOK, profileNamingOK, taskScopingOK, baseAssetInjectionOK, remoteAuthorizationOK, roleTopologyOK)...)

	fmt.Println("Auditing Gentle AI upstream baseline...")
	fmt.Printf("- Repo root: %s\n", repoRoot)
	fmt.Printf("- Upstream repo: %s\n", upstreamRepo)
	fmt.Printf("- Upstream source: %s\n", upstreamRepoSource)
	fmt.Printf("- Verified binary version: %s\n", release.Version)
	fmt.Printf("- Verified exact tag: %s\n", release.Tag)
	fmt.Printf("- Verified release commit: %s\n", release.Commit)
	fmt.Printf("- Upstream baseline limit: %s\n", state.LastMaintainedCommit)
	fmt.Printf("- Managed assets manifest: %s\n", manifestPath)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  managed assets drift: %s\n", managedDriftStatus(len(managedEntries)))
	fmt.Printf("  watched but unmanaged drift: %s\n", unmanagedDriftStatus(len(unmanagedEntries)))
	if !basePromptDrift {
		fmt.Println("  base prompt drift: no")
	} else {
		fmt.Println("  base prompt drift: yes")
	}
	fmt.Printf("  profile phase order: %s\n", statusWord(phaseOrderOK))
	fmt.Printf("  profile orchestrator naming: %s\n", statusWord(profileNamingOK))
	fmt.Printf("  profile task scoping invariant: %s\n", statusWord(taskScopingOK))
	fmt.Printf("  base asset injection invariant: %s\n", statusWord(baseAssetInjectionOK))
	fmt.Printf("  remote authorization projection invariant: %s\n", statusWord(remoteAuthorizationOK))
	fmt.Printf("  JD/review actor topology invariant: %s\n", statusWord(roleTopologyOK))
	if len(driftSummary) > 0 {
		fmt.Println()
		fmt.Println("Drift summary:")
		for _, item := range driftSummary {
			fmt.Printf("  - %s\n", item)
		}
	}

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
		fmt.Println("1. Review the managed drift against the last maintained commit and decide what is relevant to the overlay.")
		fmt.Println("2. Update this repo (owned assets, approved upstream snapshots, docs, and/or state) if the new upstream state is accepted.")
		fmt.Println("3. Run `gentle-ai sync` or reinstall as appropriate, then re-run `bash apply-gentle-ai-custom.sh opencode` (or `all` if you also want the multi-target refresh) so runtime verification passes.")
		return 1
	}

	fmt.Println()
	fmt.Println("Done. No managed upstream drift or structural invariant mismatch was detected for the current upstream range.")
	return 0
}

func normalizeAuditUpstreamArgs(args []string) int {
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			printAuditUpstreamUsage(os.Stdout)
			return 0
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown audit-upstream flag: %s\n", arg)
			} else {
				fmt.Fprintf(os.Stderr, "audit-upstream does not accept positional argument: %s\n", arg)
			}
			printAuditUpstreamUsage(os.Stderr)
			return 1
		}
	}
	return -1
}

func printAuditUpstreamUsage(out *os.File) {
	fmt.Fprintf(out, "Usage: %s\n", usageCommandName("audit-upstream"))
}

type structuralInvariantResult struct {
	PhaseOrderOK          bool
	ProfileNamingOK       bool
	TaskScopingOK         bool
	BaseAssetInjectionOK  bool
	RemoteAuthorizationOK bool
	RoleTopologyOK        bool
}

type managedAssetRecord struct {
	Key           string
	Class         string
	Kind          string
	UpstreamPath  string
	UpstreamPaths []string
	Directory     bool
}

type managedAssetCatalog struct {
	WatchRoots []string
	Records    []managedAssetRecord
}

type managedDiffEntry struct {
	Diff   GitDiffEntry
	Record managedAssetRecord
	Path   string
}

func buildManagedAssetCatalog(target ManagedAssetsTarget) managedAssetCatalog {
	var records []managedAssetRecord
	for _, asset := range target.OwnedOverlayAssets {
		paths := ownedAssetUpstreamPaths(asset)
		records = append(records, managedAssetRecord{
			Key:           asset.Key,
			Class:         asset.Class,
			Kind:          asset.Kind,
			UpstreamPath:  asset.UpstreamPath,
			UpstreamPaths: paths,
			Directory:     strings.HasSuffix(asset.Kind, "_directory"),
		})
	}
	for _, asset := range target.RetainedUpstreamSkills {
		records = append(records, managedAssetRecord{
			Key:          asset.Key,
			Class:        "retained_optional",
			Kind:         "retained_upstream_skill",
			UpstreamPath: asset.UpstreamPath,
			Directory:    !strings.HasSuffix(asset.UpstreamPath, ".md"),
		})
	}
	for _, asset := range target.PrunedUpstreamSkills {
		records = append(records, managedAssetRecord{
			Key:          asset.Key,
			Class:        "pruned_optional",
			Kind:         "pruned_upstream_skill",
			UpstreamPath: asset.UpstreamPath,
			Directory:    !strings.HasSuffix(asset.UpstreamPath, ".md"),
		})
	}
	return managedAssetCatalog{WatchRoots: target.WatchRoots, Records: records}
}

func categorizeGitDiff(entries []GitDiffEntry, catalog managedAssetCatalog) ([]managedDiffEntry, []GitDiffEntry) {
	managed := make([]managedDiffEntry, 0, len(entries))
	unmanaged := make([]GitDiffEntry, 0)
	for _, entry := range entries {
		match, path, watched := matchManagedDiffEntry(entry, catalog)
		if match != nil {
			managed = append(managed, managedDiffEntry{Diff: entry, Record: *match, Path: path})
			if hasUnmappedWatchedSide(entry, catalog) {
				unmanaged = append(unmanaged, entry)
			}
			continue
		}
		if watched {
			unmanaged = append(unmanaged, entry)
		}
	}
	sort.Slice(managed, func(i, j int) bool {
		if managed[i].Record.Key == managed[j].Record.Key {
			return managed[i].Path < managed[j].Path
		}
		return managed[i].Record.Key < managed[j].Record.Key
	})
	sort.Slice(unmanaged, func(i, j int) bool {
		return diffPrimaryPath(unmanaged[i]) < diffPrimaryPath(unmanaged[j])
	})
	return managed, unmanaged
}

func matchManagedDiffEntry(entry GitDiffEntry, catalog managedAssetCatalog) (*managedAssetRecord, string, bool) {
	watched := false
	paths := []string{entry.Path}
	if entry.OldPath != "" {
		paths = append(paths, entry.OldPath)
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		if pathMatchesWatchRoots(path, catalog.WatchRoots) {
			watched = true
		}
		for _, record := range catalog.Records {
			if recordMatchesPath(record, path) {
				match := record
				return &match, path, true
			}
		}
	}
	return nil, "", watched
}

func recordMatchesPath(record managedAssetRecord, path string) bool {
	for _, recordPath := range record.UpstreamPaths {
		if recordPath == path {
			return true
		}
	}
	if record.Directory {
		return path == record.UpstreamPath || strings.HasPrefix(path, record.UpstreamPath+"/")
	}
	return path == record.UpstreamPath
}

func pathMatchesWatchRoots(path string, watchRoots []string) bool {
	for _, root := range watchRoots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

func hasUnmappedWatchedSide(entry GitDiffEntry, catalog managedAssetCatalog) bool {
	paths := []string{entry.Path}
	if entry.OldPath != "" {
		paths = append(paths, entry.OldPath)
	}
	for _, path := range paths {
		if path == "" || !pathMatchesWatchRoots(path, catalog.WatchRoots) {
			continue
		}
		matched := false
		for _, record := range catalog.Records {
			if recordMatchesPath(record, path) {
				matched = true
				break
			}
		}
		if !matched {
			return true
		}
	}
	return false
}

func findManagedAssetDiff(entries []managedDiffEntry, key string) (managedDiffEntry, bool) {
	for _, entry := range entries {
		if entry.Record.Key == key {
			return entry, true
		}
	}
	return managedDiffEntry{}, false
}

func summarizeManagedPromptDrift(upstreamRepo, baseCommit string, entry managedDiffEntry) ([]string, error) {
	if entry.Diff.Status == "A" {
		return []string{"The base orchestrator asset was added after the last maintained commit. Review the new prompt before adopting this upstream state."}, nil
	}
	if entry.Diff.Status == "D" {
		return []string{"The base orchestrator asset was removed upstream. Review integration mechanics before adopting this upstream state."}, nil
	}
	oldPath := entry.Path
	if entry.Diff.OldPath != "" {
		oldPath = entry.Diff.OldPath
	}
	previousText, err := runGitShow(upstreamRepo, baseCommit, oldPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read base orchestrator asset from %s at %s: %v", baseCommit, oldPath, err)
	}
	currentText, err := runGitShow(upstreamRepo, "HEAD", entry.Diff.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot read current orchestrator asset at %s: %v", entry.Diff.Path, err)
	}
	return summarizePromptDrift(previousText, currentText), nil
}

func summarizeManagedDiffEntries(entries []managedDiffEntry) []string {
	if len(entries) == 0 {
		return nil
	}
	summary := make([]string, 0, len(entries))
	for _, entry := range entries {
		summary = append(summary, formatManagedDiffEntry(entry))
	}
	return summary
}

func summarizeUnmanagedDiffEntries(entries []GitDiffEntry) []string {
	if len(entries) == 0 {
		return nil
	}
	summary := make([]string, 0, len(entries))
	for _, entry := range entries {
		summary = append(summary, formatUnmanagedDiffEntry(entry))
	}
	return summary
}

func formatManagedDiffEntry(entry managedDiffEntry) string {
	scope := entry.Record.Class + "/" + entry.Record.Kind
	if entry.Diff.Status == "R" {
		return fmt.Sprintf("Managed %s %s -> %s maps to %s (%s).", entry.Diff.Status, entry.Diff.OldPath, entry.Diff.Path, entry.Record.Key, scope)
	}
	return fmt.Sprintf("Managed %s %s maps to %s (%s).", entry.Diff.Status, entry.Path, entry.Record.Key, scope)
}

func formatUnmanagedDiffEntry(entry GitDiffEntry) string {
	if entry.Status == "R" {
		return fmt.Sprintf("Unmapped watched %s %s -> %s. Add or classify it in managed-assets.json if it matters to the overlay.", entry.Status, entry.OldPath, entry.Path)
	}
	return fmt.Sprintf("Unmapped watched %s %s. Add or classify it in managed-assets.json if it matters to the overlay.", entry.Status, diffPrimaryPath(entry))
}

func diffPrimaryPath(entry GitDiffEntry) string {
	if entry.Path != "" {
		return entry.Path
	}
	return entry.OldPath
}

func readStructuralSource(upstreamRepo string, paths []string, suffix string) (string, error) {
	for _, path := range paths {
		if strings.HasSuffix(path, suffix) {
			return runGitShow(upstreamRepo, "HEAD", path)
		}
	}
	return "", fmt.Errorf("structural invariant source matching %q is not declared in managed-assets.json", suffix)
}

func declaresStructuralSource(paths []string, suffix string) bool {
	for _, path := range paths {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func managedDriftStatus(count int) string {
	if count == 0 {
		return "ok"
	}
	return fmt.Sprintf("%d tracked file(s) changed", count)
}

func evaluateStructuralInvariants(upstreamRepo string, target ManagedAssetsTarget, policy Policy) (structuralInvariantResult, []string) {
	var failures []string
	result := structuralInvariantResult{}
	upstreamProfilesText, err := readStructuralSource(upstreamRepo, target.StructuralInvariantSources, "profiles.go")
	if err != nil {
		failures = append(failures, err.Error())
	}
	upstreamInjectText, err := readStructuralSource(upstreamRepo, target.StructuralInvariantSources, "inject.go")
	if err != nil {
		failures = append(failures, err.Error())
	}
	upstreamModelsText, err := readStructuralSource(upstreamRepo, target.StructuralInvariantSources, "models.go")
	if err != nil {
		failures = append(failures, err.Error())
	}
	phaseOrder, phaseErr := extractProfilePhaseOrder(upstreamProfilesText)
	if phaseErr != nil {
		failures = append(failures, phaseErr.Error())
	}
	if len(phaseOrder) > 0 && !sameStrings(phaseOrder, policy.OpenCode.SDDPhases) {
		failures = append(failures, fmt.Sprintf("upstream profilePhaseOrder is %#v, expected %#v from policy", phaseOrder, policy.OpenCode.SDDPhases))
	}
	profileNamingOK := profileNamingInvariantOK(upstreamProfilesText)
	if !profileNamingOK {
		failures = append(failures, "upstream profiles.go no longer detects and builds named profiles through the canonical managed-profile mechanism")
	}
	if !profileTaskScopingInvariantOK(upstreamProfilesText) {
		failures = append(failures, "upstream profiles.go no longer preserves a deny-by-default task permission map for named profile orchestrators")
	}
	requiredInjectSnippets := []string{
		`orchestratorRaw, ok := agentsMap["gentle-orchestrator"]`,
		`taskRaw, ok := permissionRaw["task"].(map[string]any)`,
	}
	for _, snippet := range requiredInjectSnippets {
		if !strings.Contains(upstreamInjectText, snippet) {
			failures = append(failures, fmt.Sprintf("upstream inject.go no longer contains expected base orchestrator asset binding snippet: %q", snippet))
		}
	}
	upstreamOrchestratorText := ""
	if declaresStructuralSource(target.StructuralInvariantSources, "components/sdd/orchestrator.go") {
		upstreamOrchestratorText, err = readStructuralSource(upstreamRepo, target.StructuralInvariantSources, "components/sdd/orchestrator.go")
		if err != nil {
			failures = append(failures, err.Error())
		}
	}
	if !baseAssetInjectionInvariantOK(upstreamInjectText, upstreamOrchestratorText) {
		failures = append(failures, "upstream inject.go no longer binds the OpenCode orchestrator prompt from the canonical SDD asset render path")
	}
	remoteAuthorizationOK := true
	if declaresStructuralSource(target.StructuralInvariantSources, "components/agentguidance/remote_authorization.go") {
		remoteSources := make(map[string]string)
		for _, suffix := range []string{
			"components/agentguidance/remote_authorization.go",
			"components/agentguidance/inject.go",
			"components/sdd/prompts.go",
			"components/sdd/profiles.go",
			"components/sdd/inject.go",
		} {
			remoteSources[suffix], err = readStructuralSource(upstreamRepo, target.StructuralInvariantSources, suffix)
			if err != nil {
				failures = append(failures, err.Error())
				remoteAuthorizationOK = false
			}
		}
		if remoteAuthorizationOK {
			remoteAuthorizationOK = remoteAuthorizationProjectionInvariantOK(remoteSources)
		}
		if !remoteAuthorizationOK {
			failures = append(failures, "upstream remote authorization contract no longer reaches all maintained OpenCode prompt carriers")
		}
	}
	jdActors, jdErr := extractReturnedStringSlice(upstreamModelsText, "JDPhases")
	if jdErr != nil {
		failures = append(failures, jdErr.Error())
	}
	reviewActors, reviewErr := extractReturnedStringSlice(upstreamModelsText, "ReviewLensPhases")
	if reviewErr != nil {
		failures = append(failures, reviewErr.Error())
	}
	refuter, refuterErr := extractStringConstant(upstreamModelsText, "ReviewRefuterAgent")
	if refuterErr != nil {
		failures = append(failures, refuterErr.Error())
	} else {
		reviewActors = append(reviewActors, refuter)
	}
	validator, validatorErr := extractStringConstant(upstreamModelsText, "ReviewValidatorAgent")
	if validatorErr != nil {
		failures = append(failures, validatorErr.Error())
	} else {
		reviewActors = append(reviewActors, validator)
	}
	roleTopologyOK := sameStrings(jdActors, judgmentDayActorKeys) && sameStrings(reviewActors, reviewActorKeys)
	if !roleTopologyOK && jdErr == nil && reviewErr == nil && refuterErr == nil && validatorErr == nil {
		failures = append(failures, fmt.Sprintf("upstream JD/review actor topology is JD=%#v review=%#v, expected JD=%#v review=%#v", jdActors, reviewActors, judgmentDayActorKeys, reviewActorKeys))
	}
	result.ProfileNamingOK = profileNamingOK
	result.TaskScopingOK = profileTaskScopingInvariantOK(upstreamProfilesText)
	result.BaseAssetInjectionOK = containsAll(upstreamInjectText, requiredInjectSnippets) && baseAssetInjectionInvariantOK(upstreamInjectText, upstreamOrchestratorText)
	result.RemoteAuthorizationOK = remoteAuthorizationOK
	result.PhaseOrderOK = len(phaseOrder) > 0 && sameStrings(phaseOrder, policy.OpenCode.SDDPhases)
	result.RoleTopologyOK = roleTopologyOK
	return result, failures
}

func extractReturnedStringSlice(text, functionName string) ([]string, error) {
	pattern := fmt.Sprintf(`(?s)func\s+%s\(\)\s+\[\]string\s*\{\s*return\s+\[\]string\s*\{(.*?)\}\s*\}`, regexp.QuoteMeta(functionName))
	match := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if match == nil {
		return nil, fmt.Errorf("could not locate upstream %s string slice", functionName)
	}
	valueRE := regexp.MustCompile(`"([^"]+)"`)
	values := valueRE.FindAllStringSubmatch(match[1], -1)
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value[1])
	}
	return result, nil
}

func extractStringConstant(text, constantName string) (string, error) {
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return "", fmt.Errorf("could not parse upstream source for %s string constant: %w", constantName, err)
	}
	var value string
	var valueErr error
	ast.Inspect(file, func(node ast.Node) bool {
		genDecl, ok := node.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			return true
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, name := range valueSpec.Names {
				if name.Name != constantName || index >= len(valueSpec.Values) {
					continue
				}
				literal, ok := stringLiteralValue(valueSpec.Values[index])
				if !ok {
					valueErr = fmt.Errorf("upstream %s constant is not a string literal", constantName)
					return false
				}
				value = literal
				return false
			}
		}
		return true
	})
	if valueErr != nil {
		return "", valueErr
	}
	if value != "" {
		return value, nil
	}
	return "", fmt.Errorf("could not locate upstream %s string constant", constantName)
}

func profileNamingInvariantOK(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	state := struct {
		managedPrefix   bool
		managedDetect   bool
		orchestratorKey bool
	}{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		switch fn.Name.Name {
		case "managedProfileAgentPrefixes":
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if lit, ok := node.(*ast.BasicLit); ok {
					value, valid := stringLiteralValue(lit)
					state.managedPrefix = state.managedPrefix || valid && value == "sdd-orchestrator-"
				}
				return true
			})
		case "DetectProfiles":
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				assign, ok := node.(*ast.AssignStmt)
				if !ok || len(assign.Rhs) != 1 || len(assign.Lhs) < 2 {
					return true
				}
				call, ok := assign.Rhs[0].(*ast.CallExpr)
				if !ok || len(call.Args) != 1 || !isIdent(call.Fun, "managedProfileAgentName") || !isIdent(call.Args[0], "key") {
					return true
				}
				state.managedDetect = isIdent(assign.Lhs[0], "profileName") && isIdent(assign.Lhs[1], "managed")
				return true
			})
		case "ProfileAgentKeys":
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || !isIdent(call.Fun, "append") || len(call.Args) < 2 {
					return true
				}
				binary, ok := call.Args[1].(*ast.BinaryExpr)
				state.orchestratorKey = state.orchestratorKey || ok && binary.Op == token.ADD && isStringLiteral(binary.X, "sdd-orchestrator") && isIdent(binary.Y, "suffix")
				return true
			})
		}
	}
	return state.managedPrefix && state.managedDetect && state.orchestratorKey
}

func profileTaskScopingInvariantOK(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	state := profileTaskScopingState{}
	ast.Inspect(file, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		inspectProfileTaskScopingAssignment(assign, &state)
		return true
	})
	return state.denyDefault && state.phaseAllow && state.jdAllow && state.reviewAllow && state.wiredToProfilePermission
}

type profileTaskScopingState struct {
	denyDefault              bool
	phaseAllow               bool
	jdAllow                  bool
	reviewAllow              bool
	wiredToProfilePermission bool
}

func inspectProfileTaskScopingAssignment(assign *ast.AssignStmt, state *profileTaskScopingState) {
	for i, lhs := range assign.Lhs {
		if i >= len(assign.Rhs) {
			continue
		}
		rhs := assign.Rhs[i]
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "taskPerms" {
			if lit, ok := rhs.(*ast.CompositeLit); ok && compositeLitHasStringEntry(lit, "*", "deny") {
				state.denyDefault = true
			}
		}
		if index, ok := lhs.(*ast.IndexExpr); ok && isIdent(index.X, "taskPerms") && isStringLiteral(rhs, "allow") {
			if isBinaryIdentPlusIdent(index.Index, "phase", "suffix") {
				state.phaseAllow = true
			}
			if isIdent(index.Index, "jd") || isBinaryIdentPlusIdent(index.Index, "jd", "suffix") {
				state.jdAllow = true
			}
			if isIdent(index.Index, "reviewAgent") {
				state.reviewAllow = true
			}
		}
		if profilePermissionAssignmentUsesTaskPerms(lhs, rhs) {
			state.wiredToProfilePermission = true
		}
	}
}

func profilePermissionAssignmentUsesTaskPerms(lhs ast.Expr, rhs ast.Expr) bool {
	if index, ok := lhs.(*ast.IndexExpr); ok && isProfileMapIdent(index.X) {
		if key, ok := stringLiteralValue(index.Index); ok && key == "permission" {
			return permissionMapUsesTaskPerms(rhs)
		}
	}
	ident, ok := lhs.(*ast.Ident)
	if !ok || !isProfileMapName(ident.Name) {
		return false
	}
	lit, ok := rhs.(*ast.CompositeLit)
	if !ok {
		return false
	}
	return compositeLitHasKeyMatching(lit, "permission", permissionMapUsesTaskPerms)
}

func isProfileMapIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && isProfileMapName(ident.Name)
}

func isProfileMapName(name string) bool {
	switch name {
	case "orchEntry", "profileMap", "orchestratorMap":
		return true
	default:
		return false
	}
}

func permissionMapUsesTaskPerms(expr ast.Expr) bool {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	return compositeLitHasKeyMatching(lit, "task", taskPermissionValueUsesTaskPerms)
}

func taskPermissionValueUsesTaskPerms(expr ast.Expr) bool {
	if isIdent(expr, "taskPerms") {
		return true
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	return compositeLitHasKeyMatching(lit, "__replace__", func(value ast.Expr) bool {
		return isIdent(value, "taskPerms")
	})
}

func baseAssetInjectionInvariantOK(text string, compositionSources ...string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	compositionRequired := false
	compositionOK := false
	for _, source := range compositionSources {
		if strings.TrimSpace(source) == "" {
			continue
		}
		compositionRequired = true
		if orchestratorCompositionInvariantOK(source) {
			compositionOK = true
		}
	}
	if compositionRequired && !compositionOK {
		return false
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if compositionRequired && fn.Name.Name != "inlineOpenCodeSDDPrompts" {
			continue
		}
		if blockBindsOrchestratorPrompt(fn.Body, compositionOK, !compositionRequired) {
			return true
		}
	}
	return false
}

func blockBindsOrchestratorPrompt(block *ast.BlockStmt, compositionOK, legacyAllowed bool) bool {
	bound := map[*ast.Object]bool{}
	promptAssignmentValid := false
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 && len(assign.Lhs) > 0 {
			for i := range assign.Lhs {
				if i >= len(assign.Rhs) || !isOrchestratorPromptAssignment(assign.Lhs[i]) {
					continue
				}
				promptAssignmentValid = false
				if call, ok := assign.Rhs[i].(*ast.CallExpr); ok {
					if compositionOK && isComposeOpenCodeOrchestratorCall(call) {
						promptAssignmentValid = true
					}
					if legacyAllowed && isCanonicalOrchestratorPromptSource(call) {
						promptAssignmentValid = true
					}
				}
				if ident, ok := assign.Rhs[i].(*ast.Ident); ok && ident.Obj != nil && bound[ident.Obj] && compositionOK {
					promptAssignmentValid = true
				}
			}
			for _, lhsExpr := range assign.Lhs {
				ident, ok := lhsExpr.(*ast.Ident)
				if !ok || ident.Obj == nil {
					continue
				}
				delete(bound, ident.Obj)
			}
			if call, ok := assign.Rhs[0].(*ast.CallExpr); ok && isComposeOpenCodeOrchestratorCall(call) {
				if ident, ok := assign.Lhs[0].(*ast.Ident); ok && ident.Obj != nil {
					bound[ident.Obj] = true
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsOrchestratorPrompt(child) {
				promptAssignmentValid = blockBindsOrchestratorPrompt(child, compositionOK, legacyAllowed)
			}
			for object := range bound {
				if blockAssignsObject(child, object) {
					delete(bound, object)
				}
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return promptAssignmentValid
}

func isComposeOpenCodeOrchestratorCall(call *ast.CallExpr) bool {
	return isIdent(call.Fun, "composeOpenCodeOrchestratorPrompt") && len(call.Args) == 2 &&
		isIdent(call.Args[0], "agent") && isIdent(call.Args[1], "renderOptions")
}

func orchestratorCompositionInvariantOK(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	composeOpenCode := findFunction(file, "composeOpenCodeOrchestratorPrompt")
	compose := findFunction(file, "composeOrchestratorPrompt")
	selector := findFunction(file, "sddOrchestratorAsset")
	return composeOpenCodeHasCanonicalReturn(composeOpenCode) &&
		composeOrchestratorReadsSelectedAsset(compose) &&
		sddOrchestratorAssetSelectsOpenCode(selector)
}

func composeOpenCodeHasCanonicalReturn(fn *ast.FuncDecl) bool {
	ret := finalTopLevelReturn(fn)
	if ret == nil || len(ret.Results) != 1 {
		return false
	}
	outer, ok := ret.Results[0].(*ast.CallExpr)
	if !ok || !isIdent(outer.Fun, "projectSDDSessionPreflight") || len(outer.Args) < 1 {
		return false
	}
	inner, ok := outer.Args[0].(*ast.CallExpr)
	return ok && isIdent(inner.Fun, "composeOrchestratorPrompt") && len(inner.Args) >= 1 && isIdent(inner.Args[0], "agent")
}

func composeOrchestratorReadsSelectedAsset(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	var pathObject, contentObject *ast.Object
	pathTrusted := false
	contentTrusted := false
	assetReadIndex := -1
	for index, stmt := range fn.Body.List {
		if statementTerminatesBlock(stmt) && index != len(fn.Body.List)-1 {
			return false
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) == 0 || len(assign.Rhs) != 1 {
			continue
		}
		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Obj == nil {
			continue
		}
		call, callOK := assign.Rhs[0].(*ast.CallExpr)
		if pathObject != nil && lhs.Obj == pathObject {
			pathTrusted = false
		}
		if contentObject != nil && lhs.Obj == contentObject {
			contentTrusted = contentTrusted && isCanonicalOrchestratorContentDerivation(assign.Rhs[0], contentObject)
		}
		if isIdent(lhs, "path") && callOK && isIdent(call.Fun, "sddOrchestratorAsset") && len(call.Args) == 1 && isIdent(call.Args[0], "agent") {
			pathObject = lhs.Obj
			pathTrusted = true
			continue
		}
		if isIdent(lhs, "content") && pathTrusted && callOK && isSelector(call.Fun, "assets", "MustRead") && len(call.Args) == 1 && identUsesObject(call.Args[0], pathObject) {
			contentObject = lhs.Obj
			contentTrusted = true
			assetReadIndex = index
		}
		for _, child := range reachableChildBlocks(stmt) {
			if pathTrusted && blockAssignsObject(child, pathObject) {
				pathTrusted = false
			}
			if contentTrusted && !blockPreservesOrchestratorContent(child, contentObject) {
				contentTrusted = false
			}
		}
	}
	ret := finalTopLevelReturn(fn)
	if contentObject == nil || !contentTrusted || ret == nil || assetReadIndex >= len(fn.Body.List)-1 || len(ret.Results) != 1 {
		return false
	}
	call, ok := ret.Results[0].(*ast.CallExpr)
	return ok && isIdent(call.Fun, "bindRuntimeAgentIdentity") && len(call.Args) == 2 && identUsesObject(call.Args[0], contentObject) && isIdent(call.Args[1], "agent")
}

func sddOrchestratorAssetSelectsOpenCode(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	for _, stmt := range fn.Body.List {
		if statementTerminatesBlock(stmt) {
			break
		}
		switchStmt, ok := stmt.(*ast.SwitchStmt)
		if !ok || !isIdent(switchStmt.Tag, "agent") || switchStmt.Body == nil {
			continue
		}
		for _, caseStmt := range switchStmt.Body.List {
			clause, ok := caseStmt.(*ast.CaseClause)
			if !ok || !expressionsContainSelector(clause.List, "model", "AgentOpenCode") {
				continue
			}
			for _, bodyStmt := range clause.Body {
				if ret, ok := bodyStmt.(*ast.ReturnStmt); ok && len(ret.Results) == 1 && isStringLiteral(ret.Results[0], "opencode/sdd-orchestrator.md") {
					return true
				}
			}
		}
	}
	return false
}

func finalTopLevelReturn(fn *ast.FuncDecl) *ast.ReturnStmt {
	if fn == nil || fn.Body == nil || len(fn.Body.List) == 0 {
		return nil
	}
	ret, _ := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	return ret
}

func remoteAuthorizationProjectionInvariantOK(sources map[string]string) bool {
	remoteFile := parseInvariantSource(sources["components/agentguidance/remote_authorization.go"])
	guidanceFile := parseInvariantSource(sources["components/agentguidance/inject.go"])
	promptsFile := parseInvariantSource(sources["components/sdd/prompts.go"])
	profilesFile := parseInvariantSource(sources["components/sdd/profiles.go"])
	injectFile := parseInvariantSource(sources["components/sdd/inject.go"])
	if remoteFile == nil || guidanceFile == nil || promptsFile == nil || profilesFile == nil || injectFile == nil {
		return false
	}
	return remoteAuthorizationHelperIsCanonical(findFunction(remoteFile, "InjectRemoteAuthorization")) &&
		routingProjectionReachesReturn(findFunction(guidanceFile, "InjectRoutingWithOptions")) &&
		functionProjectsIdentToWrite(findFunction(promptsFile, "WriteSharedPromptFiles"), "content", "agentguidance", "InjectRemoteAuthorization") &&
		functionAssignsPromptProjection(findFunction(promptsFile, "injectRemoteAuthorizationIntoSubagentPrompts")) &&
		functionProjectionSurvivesUntilMarshal(findFunction(profilesFile, "GenerateProfileOverlay"), "injectRemoteAuthorizationIntoSubagentPrompts", "agentMap") &&
		functionProjectsIdentToWrite(findFunction(injectFile, "Inject"), "contentStr", "agentguidance", "InjectRemoteAuthorization") &&
		functionProjectionSurvivesUntilMarshal(findFunction(injectFile, "inlineOpenCodeSDDPrompts"), "injectRemoteAuthorizationIntoSubagentPrompts", "agentsMap")
}

func remoteAuthorizationHelperIsCanonical(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	var contractObject *ast.Object
	contractTrusted := false
	for index, stmt := range fn.Body.List {
		if statementTerminatesBlock(stmt) && index != len(fn.Body.List)-1 {
			return false
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			continue
		}
		contract, ok := assign.Lhs[0].(*ast.Ident)
		call, callOK := assign.Rhs[0].(*ast.CallExpr)
		if ok && contractObject != nil && contract.Obj == contractObject {
			contractTrusted = false
		}
		if ok && contract.Obj != nil && callOK && isSelector(call.Fun, "assets", "MustRead") && len(call.Args) == 1 && isStringLiteral(call.Args[0], "generic/remote-authorization-contract.md") {
			contractObject = contract.Obj
			contractTrusted = true
		}
		for _, child := range reachableChildBlocks(stmt) {
			if contractTrusted && blockAssignsObject(child, contractObject) {
				contractTrusted = false
			}
		}
	}
	ret := finalTopLevelReturn(fn)
	if contractObject == nil || !contractTrusted || ret == nil || len(ret.Results) != 1 {
		return false
	}
	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok || !isSelector(call.Fun, "filemerge", "InjectMarkdownSection") || len(call.Args) != 3 || !isIdent(call.Args[0], "prompt") || !isStringLiteral(call.Args[1], "remote-authorization") {
		return false
	}
	trim, ok := call.Args[2].(*ast.CallExpr)
	return ok && isSelector(trim.Fun, "strings", "TrimSpace") && len(trim.Args) == 1 && identUsesObject(trim.Args[0], contractObject)
}

func routingProjectionReachesReturn(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	var renderedObject *ast.Object
	trusted := false
	for _, stmt := range fn.Body.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
			lhs, lhsOK := assign.Lhs[0].(*ast.Ident)
			if lhsOK && lhs.Name == "rendered" {
				if renderedObject == nil {
					renderedObject = lhs.Obj
				}
				trusted = false
				if call, ok := assign.Rhs[0].(*ast.CallExpr); ok && isIdent(call.Fun, "InjectRemoteAuthorization") && len(call.Args) == 1 && sameIdentifierBinding(lhs, call.Args[0]) {
					trusted = true
				}
			}
		}
		if trusted {
			if ret, ok := stmt.(*ast.ReturnStmt); ok && returnUsesTrackedIdentifier(ret, "rendered", renderedObject) {
				return true
			}
			if switchStmt, ok := stmt.(*ast.SwitchStmt); ok && switchReturnsUseTrackedIdentifier(switchStmt, "rendered", renderedObject) {
				return true
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if trusted && blockAssignsTrackedIdentifier(child, "rendered", renderedObject) {
				trusted = false
			}
		}
	}
	return false
}

func functionProjectsIdentToWrite(fn *ast.FuncDecl, identName, packageName, functionName string) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	return blockProjectsIdentToWrite(fn.Body, identName, packageName, functionName)
}

func blockProjectsIdentToWrite(block *ast.BlockStmt, identName, packageName, functionName string) bool {
	var object *ast.Object
	trusted := false
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for index, lhsExpr := range assign.Lhs {
				lhs, ok := lhsExpr.(*ast.Ident)
				if !ok || lhs.Name != identName {
					continue
				}
				if object == nil {
					object = lhs.Obj
				}
				rhs := assignmentRHS(assign, index)
				trusted = false
				if call, ok := rhs.(*ast.CallExpr); ok && len(call.Args) == 1 && sameIdentifierBinding(lhs, call.Args[0]) && callMatches(call, packageName, functionName) {
					trusted = true
				}
			}
		}
		if trusted && statementCallsWithTrackedIdentifier(stmt, "filemerge", "WriteFileAtomic", identName, object) {
			return true
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockProjectsIdentToWrite(child, identName, packageName, functionName) {
				return true
			}
			if blockAssignsTrackedIdentifier(child, identName, object) {
				object, trusted = projectionStateAfterBlock(child, identName, packageName, functionName, object, trusted)
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func projectionStateAfterBlock(block *ast.BlockStmt, identName, packageName, functionName string, object *ast.Object, trusted bool) (*ast.Object, bool) {
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for index, lhsExpr := range assign.Lhs {
				lhs, ok := lhsExpr.(*ast.Ident)
				if !ok || lhs.Name != identName {
					continue
				}
				if object == nil {
					object = lhs.Obj
				}
				rhs := assignmentRHS(assign, index)
				trusted = false
				if call, ok := rhs.(*ast.CallExpr); ok && len(call.Args) == 1 && sameIdentifierBinding(lhs, call.Args[0]) && callMatches(call, packageName, functionName) {
					trusted = true
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			object, trusted = projectionStateAfterBlock(child, identName, packageName, functionName, object, trusted)
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return object, trusted
}

func functionAssignsPromptProjection(fn *ast.FuncDecl) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	return blockEndsWithPromptProjection(fn.Body)
}

func blockEndsWithPromptProjection(block *ast.BlockStmt) bool {
	touched := false
	valid := false
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 && isPromptMapAssignment(assign.Lhs[0]) {
			touched = true
			call, callOK := assign.Rhs[0].(*ast.CallExpr)
			valid = callOK && isSelector(call.Fun, "agentguidance", "InjectRemoteAuthorization") && len(call.Args) == 1 && isIdent(call.Args[0], "prompt")
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsPromptMap(child) {
				touched = true
				valid = blockEndsWithPromptProjection(child)
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return touched && valid
}

func functionProjectionSurvivesUntilMarshal(fn *ast.FuncDecl, functionName, argumentName string) bool {
	if fn == nil || fn.Body == nil {
		return false
	}
	projected := false
	var argumentObject *ast.Object
	var containerObject *ast.Object
	containerTrusted := fn.Name.Name == "inlineOpenCodeSDDPrompts"
	for _, stmt := range fn.Body.List {
		if expr, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := expr.X.(*ast.CallExpr); ok && isIdent(call.Fun, functionName) && len(call.Args) == 1 && isIdent(call.Args[0], argumentName) {
				projected = true
				argumentObject = identifierObject(call.Args[0])
			}
		}
		if projected && statementAssignsTrackedIdentifier(stmt, argumentName, argumentObject) {
			return false
		}
		if projected {
			if assign, ok := stmt.(*ast.AssignStmt); ok {
				for index, lhsExpr := range assign.Lhs {
					lhs, ok := lhsExpr.(*ast.Ident)
					if !ok || lhs.Name != "overlay" {
						continue
					}
					containerObject = lhs.Obj
					containerTrusted = expressionUsesTrackedIdentifier(assignmentRHS(assign, index), argumentName, argumentObject)
				}
			}
		}
		if projected && containerTrusted && statementCallsWithTrackedIdentifier(stmt, "json", "MarshalIndent", "overlay", containerObject) {
			return true
		}
		for _, child := range reachableChildBlocks(stmt) {
			if projected && blockAssignsTrackedIdentifier(child, argumentName, argumentObject) {
				return false
			}
			if projected && blockAssignsTrackedIdentifier(child, "overlay", containerObject) {
				return false
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func reachableChildBlocks(stmt ast.Stmt) []*ast.BlockStmt {
	switch current := stmt.(type) {
	case *ast.BlockStmt:
		return []*ast.BlockStmt{current}
	case *ast.IfStmt:
		if isBoolLiteral(current.Cond, false) {
			return elseBlocks(current.Else)
		}
		if isBoolLiteral(current.Cond, true) {
			return []*ast.BlockStmt{current.Body}
		}
		return append([]*ast.BlockStmt{current.Body}, elseBlocks(current.Else)...)
	case *ast.ForStmt:
		if isBoolLiteral(current.Cond, false) {
			return nil
		}
		return []*ast.BlockStmt{current.Body}
	case *ast.RangeStmt:
		return []*ast.BlockStmt{current.Body}
	}
	return nil
}

func elseBlocks(stmt ast.Stmt) []*ast.BlockStmt {
	switch value := stmt.(type) {
	case *ast.BlockStmt:
		return []*ast.BlockStmt{value}
	case *ast.IfStmt:
		return reachableChildBlocks(value)
	default:
		return nil
	}
}

func statementTerminatesBlock(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	default:
		return false
	}
}

func isBoolLiteral(expr ast.Expr, value bool) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == strconv.FormatBool(value)
}

func parseInvariantSource(text string) *ast.File {
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return nil
	}
	return file
}

func findFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func isSelector(expr ast.Expr, packageName, selectorName string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	return ok && selector.Sel != nil && selector.Sel.Name == selectorName && isIdent(selector.X, packageName)
}

func identUsesObject(expr ast.Expr, object *ast.Object) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Obj != nil && ident.Obj == object
}

func sameIdentifierBinding(left *ast.Ident, right ast.Expr) bool {
	rightIdent, ok := right.(*ast.Ident)
	if !ok || left.Name != rightIdent.Name {
		return false
	}
	if left.Obj == nil || rightIdent.Obj == nil {
		return true
	}
	return left.Obj == rightIdent.Obj
}

func expressionUsesObject(expr ast.Expr, object *ast.Object) bool {
	if expr == nil || object == nil {
		return false
	}
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		ident, ok := node.(*ast.Ident)
		if ok && ident.Obj == object {
			found = true
			return false
		}
		return !found
	})
	return found
}

func isCanonicalOrchestratorContentDerivation(expr ast.Expr, object *ast.Object) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || !expressionUsesObject(call, object) {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	switch ident.Name {
	case "composeFallbackSessionPreflight",
		"substituteSharedOrchestratorSections",
		"appendOpenCodeBackgroundPolicy",
		"replacePiClosedSingleSelectRoute",
		"replaceOpenCodeConsentV3QuestionRoute",
		"renderBoundedReviewAssetBodyFromContent":
		return true
	default:
		return false
	}
}

func expressionUsesTrackedIdentifier(expr ast.Expr, name string, object *ast.Object) bool {
	if expr == nil {
		return false
	}
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		ident, ok := node.(*ast.Ident)
		if !ok || ident.Name != name {
			return !found
		}
		if object == nil || ident.Obj == nil || ident.Obj == object {
			found = true
			return false
		}
		return true
	})
	return found
}

func blockAssignsObject(block *ast.BlockStmt, object *ast.Object) bool {
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Obj == object {
					return true
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsObject(child, object) {
				return true
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func blockPreservesOrchestratorContent(block *ast.BlockStmt, object *ast.Object) bool {
	if object == nil {
		return false
	}
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for index, lhs := range assign.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Obj != object {
					continue
				}
				if !isCanonicalOrchestratorContentDerivation(assignmentRHS(assign, index), object) {
					return false
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if !blockPreservesOrchestratorContent(child, object) {
				return false
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return true
}

func blockAssignsTrackedIdentifier(block *ast.BlockStmt, name string, object *ast.Object) bool {
	for _, stmt := range block.List {
		if statementAssignsTrackedIdentifier(stmt, name, object) {
			return true
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsTrackedIdentifier(child, name, object) {
				return true
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func blockAssignsPromptMap(block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if isPromptMapAssignment(lhs) {
					return true
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsPromptMap(child) {
				return true
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func blockAssignsOrchestratorPrompt(block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if isOrchestratorPromptAssignment(lhs) {
					return true
				}
			}
		}
		for _, child := range reachableChildBlocks(stmt) {
			if blockAssignsOrchestratorPrompt(child) {
				return true
			}
		}
		if statementTerminatesBlock(stmt) {
			break
		}
	}
	return false
}

func assignmentRHS(assign *ast.AssignStmt, index int) ast.Expr {
	if index < len(assign.Rhs) {
		return assign.Rhs[index]
	}
	if len(assign.Rhs) == 1 {
		return assign.Rhs[0]
	}
	return nil
}

func callMatches(call *ast.CallExpr, packageName, functionName string) bool {
	if packageName == "" {
		return isIdent(call.Fun, functionName)
	}
	return isSelector(call.Fun, packageName, functionName)
}

func returnUsesTrackedIdentifier(ret *ast.ReturnStmt, name string, object *ast.Object) bool {
	for _, result := range ret.Results {
		if expressionUsesTrackedIdentifier(result, name, object) {
			return true
		}
	}
	return false
}

func switchReturnsUseTrackedIdentifier(stmt *ast.SwitchStmt, name string, object *ast.Object) bool {
	if stmt.Body == nil || len(stmt.Body.List) == 0 {
		return false
	}
	for _, caseStmt := range stmt.Body.List {
		clause, ok := caseStmt.(*ast.CaseClause)
		if !ok {
			return false
		}
		caseReturns := false
		for _, bodyStmt := range clause.Body {
			ret, ok := bodyStmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			if !returnUsesTrackedIdentifier(ret, name, object) {
				return false
			}
			caseReturns = true
		}
		if !caseReturns {
			return false
		}
	}
	return true
}

func statementCallsWithTrackedIdentifier(stmt ast.Stmt, packageName, functionName, identName string, object *ast.Object) bool {
	found := false
	ast.Inspect(stmt, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok || !callMatches(call, packageName, functionName) {
			return !found
		}
		for _, arg := range call.Args {
			if expressionUsesTrackedIdentifier(arg, identName, object) {
				found = true
				return false
			}
		}
		return !found
	})
	return found
}

func isPromptMapAssignment(expr ast.Expr) bool {
	index, ok := expr.(*ast.IndexExpr)
	return ok && isIdent(index.X, "agent") && isStringLiteral(index.Index, "prompt")
}

func identifierObject(expr ast.Expr) *ast.Object {
	ident, _ := expr.(*ast.Ident)
	if ident == nil {
		return nil
	}
	return ident.Obj
}

func statementAssignsTrackedIdentifier(stmt ast.Stmt, name string, object *ast.Object) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, lhs := range assign.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok || ident.Name != name {
			continue
		}
		if object == nil || ident.Obj == nil || ident.Obj == object {
			return true
		}
	}
	return false
}

func expressionsContainSelector(expressions []ast.Expr, packageName, selectorName string) bool {
	for _, expr := range expressions {
		if isSelector(expr, packageName, selectorName) {
			return true
		}
	}
	return false
}

func parseInvariantGoSource(text string) (*ast.File, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("empty source")
	}
	source := trimmed
	if !strings.HasPrefix(trimmed, "package ") {
		source = "package invariant\nfunc _() {\n" + text + "\n}\n"
	}
	return parser.ParseFile(token.NewFileSet(), "invariant.go", source, 0)
}

func isOrchestratorPromptAssignment(expr ast.Expr) bool {
	indexExpr, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	ident, ok := indexExpr.X.(*ast.Ident)
	if !ok || ident.Name != "orchestratorMap" {
		return false
	}
	key, ok := stringLiteralValue(indexExpr.Index)
	return ok && key == "prompt"
}

func isCanonicalOrchestratorPromptSource(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if isCanonicalOrchestratorComposerCall(call) {
		return true
	}
	return isLegacyOrchestratorAssetReadCall(call)
}

func isCanonicalOrchestratorComposerCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	if ident.Name == "composeOrchestratorPrompt" {
		return len(call.Args) == 2 && isCanonicalOrchestratorAgent(call.Args[0]) && isIdent(call.Args[1], "renderOptions")
	}
	if ident.Name != "renderSDDOrchestratorAsset" || len(call.Args) < 1 || len(call.Args) > 2 {
		return false
	}
	return isCanonicalOrchestratorAgent(call.Args[0]) && (len(call.Args) == 1 || isIdent(call.Args[1], "renderOptions"))
}

func isCanonicalOrchestratorAgent(expr ast.Expr) bool {
	agent, localAgent := expr.(*ast.Ident)
	return (localAgent && agent.Name == "agent") || isModelAgentOpenCodeSelector(expr)
}

func isLegacyOrchestratorAssetReadCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel == nil || selector.Sel.Name != "MustRead" {
		return false
	}
	x, ok := selector.X.(*ast.Ident)
	if !ok || x.Name != "assets" || len(call.Args) != 1 {
		return false
	}
	innerCall, ok := call.Args[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	innerIdent, ok := innerCall.Fun.(*ast.Ident)
	if !ok || innerIdent.Name != "sddOrchestratorAsset" || len(innerCall.Args) != 1 {
		return false
	}
	return isModelAgentOpenCodeSelector(innerCall.Args[0])
}

func isModelAgentOpenCodeSelector(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel == nil || selector.Sel.Name != "AgentOpenCode" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "model"
}

func stringLiteralValue(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

func compositeLitHasStringEntry(lit *ast.CompositeLit, key string, value string) bool {
	return compositeLitHasKeyMatching(lit, key, func(expr ast.Expr) bool {
		return isStringLiteral(expr, value)
	})
}

func compositeLitHasKeyMatching(lit *ast.CompositeLit, key string, matchValue func(ast.Expr) bool) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		entryKey, ok := stringLiteralValue(kv.Key)
		if !ok || entryKey != key {
			continue
		}
		if matchValue(kv.Value) {
			return true
		}
	}
	return false
}

func isIdent(expr ast.Expr, name string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == name
}

func isBinaryIdentPlusIdent(expr ast.Expr, left string, right string) bool {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != token.ADD {
		return false
	}
	return isIdent(binary.X, left) && isIdent(binary.Y, right)
}

func isStringLiteral(expr ast.Expr, expected string) bool {
	value, ok := stringLiteralValue(expr)
	return ok && value == expected
}

func unmanagedDriftStatus(count int) string {
	if count == 0 {
		return "none"
	}
	return fmt.Sprintf("%d unmapped watched file(s) changed", count)
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

func buildAuditDriftSummary(basePromptDrift, phaseOrderOK, profileNamingOK, taskScopingOK, baseAssetInjectionOK, remoteAuthorizationOK, roleTopologyOK bool) []string {
	var summary []string
	if basePromptDrift {
		if phaseOrderOK && profileNamingOK && taskScopingOK && baseAssetInjectionOK && remoteAuthorizationOK && roleTopologyOK {
			summary = append(summary, "No profile-generation or base-asset binding drift was detected alongside this prompt change, so this looks like prompt-content guidance rather than topology or materialization drift.")
		}
	}

	if !phaseOrderOK {
		summary = append(summary, "The upstream SDD profile phase order changed. Review generated phase sequencing before adopting the new baseline.")
	}
	if !profileNamingOK {
		summary = append(summary, "The upstream `sdd-orchestrator-*` naming/key-building logic changed. The overlay may no longer target the right generated agent keys.")
	}
	if !taskScopingOK {
		summary = append(summary, "The upstream profile task-permission scoping changed. Review task allow/deny assumptions before adopting the new baseline.")
	}
	if !baseAssetInjectionOK {
		summary = append(summary, "The upstream base orchestrator asset binding changed. Verify the overlay is still reading the intended upstream source prompt before sync/apply.")
	}
	if !remoteAuthorizationOK {
		summary = append(summary, "The upstream remote authorization contract no longer reaches every maintained OpenCode prompt carrier. Review its canonical source and projection chain before adoption.")
	}
	if !roleTopologyOK {
		summary = append(summary, "The upstream Judgment Day or review actor key set changed. Review global versus profile-scoped topology before adopting the new baseline.")
	}

	return summary
}

func summarizePromptDrift(snapshotText, upstreamPromptText string) []string {
	var summary []string

	addedHeadings := diffStrings(extractMarkdownHeadings(upstreamPromptText), extractMarkdownHeadings(snapshotText))
	if len(addedHeadings) == 1 {
		summary = append(summary, fmt.Sprintf("New prompt section added: %s.", addedHeadings[0]))
	} else if len(addedHeadings) > 1 {
		summary = append(summary, fmt.Sprintf("New prompt sections added: %s.", strings.Join(addedHeadings, ", ")))
	}

	if strings.Contains(upstreamPromptText, "### Language Domain Contract") && !strings.Contains(snapshotText, "### Language Domain Contract") {
		summary = append(summary, "The prompt now separates direct user conversation from generated technical artifacts, so persona tone should not leak into code, docs, or other outputs.")
	}
	if strings.Contains(upstreamPromptText, "Generated technical artifacts default to English") && !strings.Contains(snapshotText, "Generated technical artifacts default to English") {
		summary = append(summary, "Generated technical artifacts now default to English even when the active persona or chat language is different.")
	}
	if strings.Contains(upstreamPromptText, "neutral/professional Spanish") && !strings.Contains(snapshotText, "neutral/professional Spanish") {
		summary = append(summary, "Spanish artifact fallback is now explicitly neutral/professional instead of inheriting a conversational regional tone.")
	}
	if strings.Contains(snapshotText, "¿Querés ajustar algo o continuamos?") && strings.Contains(upstreamPromptText, "¿Quiere ajustar algo o continuamos?") {
		summary = append(summary, "The Spanish fallback copy shifted from Rioplatense voseo to neutral/professional wording in the base prompt.")
	}

	if len(summary) == 0 {
		summary = append(summary, "The base prompt content changed, but the audit could not reduce it to a known heuristic. Inspect the full diff before adopting the new baseline.")
	}

	return summary
}

func extractMarkdownHeadings(text string) []string {
	var headings []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		if heading == "" {
			continue
		}
		headings = append(headings, heading)
	}
	return headings
}

func diffStrings(current, previous []string) []string {
	seen := make(map[string]struct{}, len(previous))
	for _, item := range previous {
		seen[item] = struct{}{}
	}
	var diff []string
	for _, item := range current {
		if _, ok := seen[item]; ok {
			continue
		}
		diff = append(diff, item)
	}
	return diff
}
