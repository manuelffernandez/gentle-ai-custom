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
	upstreamHead, gitErr := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if gitErr != nil {
		failures = append(failures, fmt.Sprintf("cannot inspect upstream git state: %v", gitErr))
	}
	upstreamDescribe, _ := runGit(upstreamRepo, true, "describe", "--tags", "--always")
	if upstreamDescribe == "" {
		notes = append(notes, "upstream git describe returned nothing; repo may have no tags (HEAD hash still captured via rev-parse)")
	}
	upstreamExactTag, _ := runGit(upstreamRepo, true, "describe", "--tags", "--exact-match")

	if upstreamHead != "" && state.LastMaintainedCommit != "" && upstreamHead != state.LastMaintainedCommit {
		notes = append(notes, fmt.Sprintf("upstream HEAD %s differs from last maintained commit %s; prompt/invariant drift checks below show whether the baseline still holds", upstreamHead, state.LastMaintainedCommit))
	}
	if upstreamExactTag != "" && state.LastMaintainedTag != "" && upstreamExactTag != state.LastMaintainedTag {
		notes = append(notes, fmt.Sprintf("upstream exact tag %s differs from last maintained tag %s; review state/log if you are closing a new upstream audit", upstreamExactTag, state.LastMaintainedTag))
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
	phaseOrderOK := structuralResult.PhaseOrderOK
	driftSummary = append(driftSummary, buildAuditDriftSummary(basePromptDrift, phaseOrderOK, profileNamingOK, taskScopingOK, baseAssetInjectionOK)...)

	fmt.Println("Auditing Gentle AI upstream baseline...")
	fmt.Printf("- Repo root: %s\n", repoRoot)
	fmt.Printf("- Upstream repo: %s\n", upstreamRepo)
	fmt.Printf("- Upstream source: %s\n", upstreamRepoSource)
	if upstreamDescribe != "" {
		fmt.Printf("- Upstream HEAD: %s (%s)\n", upstreamDescribe, upstreamHead)
	}
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
	PhaseOrderOK         bool
	ProfileNamingOK      bool
	TaskScopingOK        bool
	BaseAssetInjectionOK bool
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
	phaseOrder, phaseErr := extractProfilePhaseOrder(upstreamProfilesText)
	if phaseErr != nil {
		failures = append(failures, phaseErr.Error())
	}
	if len(phaseOrder) > 0 && !sameStrings(phaseOrder, policy.OpenCode.SDDPhases) {
		failures = append(failures, fmt.Sprintf("upstream profilePhaseOrder is %#v, expected %#v from policy", phaseOrder, policy.OpenCode.SDDPhases))
	}
	const profilePrefixSnippet = `const orchPrefix = "sdd-orchestrator-"`
	const profileKeyBuilderSnippet = `keys = append(keys, "sdd-orchestrator"+suffix)`
	if !strings.Contains(upstreamProfilesText, profilePrefixSnippet) {
		failures = append(failures, "upstream profiles.go no longer declares DetectProfiles prefix 'sdd-orchestrator-'")
	}
	if !strings.Contains(upstreamProfilesText, profileKeyBuilderSnippet) {
		failures = append(failures, "upstream ProfileAgentKeys no longer builds profile orchestrator keys from 'sdd-orchestrator'+suffix")
	}
	if !profileTaskScopingInvariantOK(upstreamProfilesText) {
		failures = append(failures, "upstream profiles.go no longer preserves a deny-by-default task permission map for named profile orchestrators")
	}
	if !profileTaskToolInvariantOK(upstreamProfilesText) {
		failures = append(failures, `upstream profiles.go no longer enables the task tool in the profile orchestrator entry`)
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
	if !baseAssetInjectionInvariantOK(upstreamInjectText) {
		failures = append(failures, "upstream inject.go no longer binds the OpenCode orchestrator prompt from the canonical SDD asset render path")
	}
	result.ProfileNamingOK = strings.Contains(upstreamProfilesText, profilePrefixSnippet) && strings.Contains(upstreamProfilesText, profileKeyBuilderSnippet)
	result.TaskScopingOK = profileTaskToolInvariantOK(upstreamProfilesText) && profileTaskScopingInvariantOK(upstreamProfilesText)
	result.BaseAssetInjectionOK = containsAll(upstreamInjectText, requiredInjectSnippets) && baseAssetInjectionInvariantOK(upstreamInjectText)
	result.PhaseOrderOK = len(phaseOrder) > 0 && sameStrings(phaseOrder, policy.OpenCode.SDDPhases)
	return result, failures
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

func profileTaskToolInvariantOK(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	state := profileTaskToolState{}
	ast.Inspect(file, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		inspectProfileTaskToolAssignment(assign, &state)
		return true
	})
	return state.wiredToProfileTools
}

type profileTaskToolState struct {
	wiredToProfileTools bool
}

func inspectProfileTaskToolAssignment(assign *ast.AssignStmt, state *profileTaskToolState) {
	for i, lhs := range assign.Lhs {
		if i >= len(assign.Rhs) {
			continue
		}
		if profileToolsAssignmentUsesTaskTrue(lhs, assign.Rhs[i]) {
			state.wiredToProfileTools = true
		}
	}
}

func profileToolsAssignmentUsesTaskTrue(lhs ast.Expr, rhs ast.Expr) bool {
	if index, ok := lhs.(*ast.IndexExpr); ok && isProfileMapIdent(index.X) {
		if key, ok := stringLiteralValue(index.Index); ok && key == "tools" {
			return toolMapEnablesTask(rhs)
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
	return compositeLitHasKeyMatching(lit, "tools", toolMapEnablesTask)
}

func toolMapEnablesTask(expr ast.Expr) bool {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	if compositeLitHasKeyMatching(lit, "task", isBoolTrueLiteral) {
		return true
	}
	return compositeLitHasKeyMatching(lit, "__replace__", func(value ast.Expr) bool {
		replaceLit, ok := value.(*ast.CompositeLit)
		if !ok {
			return false
		}
		return compositeLitHasKeyMatching(replaceLit, "task", isBoolTrueLiteral)
	})
}

func baseAssetInjectionInvariantOK(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	file, err := parseInvariantGoSource(text)
	if err != nil {
		return false
	}
	matched := false
	ast.Inspect(file, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i := range assign.Lhs {
			if i >= len(assign.Rhs) {
				continue
			}
			if isOrchestratorPromptAssignment(assign.Lhs[i]) && isCanonicalOrchestratorPromptSource(assign.Rhs[i]) {
				matched = true
				return false
			}
		}
		return true
	})
	return matched
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
	return parser.ParseFile(token.NewFileSet(), "invariant.go", source, parser.SkipObjectResolution)
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
	if isRenderSDDOrchestratorAssetCall(call) {
		return true
	}
	return isLegacyOrchestratorAssetReadCall(call)
}

func isRenderSDDOrchestratorAssetCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "renderSDDOrchestratorAsset" || len(call.Args) != 1 {
		return false
	}
	return isModelAgentOpenCodeSelector(call.Args[0])
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

func isBoolTrueLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "true"
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

func buildAuditDriftSummary(basePromptDrift, phaseOrderOK, profileNamingOK, taskScopingOK, baseAssetInjectionOK bool) []string {
	var summary []string
	if basePromptDrift {
		if phaseOrderOK && profileNamingOK && taskScopingOK && baseAssetInjectionOK {
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
