package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ManagedAssetsManifest struct {
	Version     int                            `json:"version"`
	Status      string                         `json:"status"`
	Description string                         `json:"description"`
	Targets     map[string]ManagedAssetsTarget `json:"targets"`
}

type ManagedAssetsTarget struct {
	WatchRoots                 []string               `json:"watch_roots"`
	UpstreamPathInterpretation string                 `json:"upstream_path_interpretation"`
	StructuralInvariantSources []string               `json:"structural_invariant_sources"`
	OwnedOverlayAssets         []OwnedOverlayAsset    `json:"owned_overlay_assets"`
	RetainedUpstreamSkills     []UpstreamSkillRef     `json:"retained_upstream_skills"`
	PrunedUpstreamSkills       []UpstreamSkillRef     `json:"pruned_upstream_skills"`
	RepoOwnedSharedSkills      RepoOwnedSharedSkills  `json:"repo_owned_shared_skills"`
	RepoOwnedCommandBodies     RepoOwnedCommandBodies `json:"repo_owned_command_bodies"`
}

type OwnedOverlayAsset struct {
	Key                  string                     `json:"key"`
	Class                string                     `json:"class"`
	Kind                 string                     `json:"kind"`
	UpstreamPath         string                     `json:"upstream_path"`
	RepoUpstreamPath     string                     `json:"repo_upstream_path"`
	RepoOwnedPath        string                     `json:"repo_owned_path"`
	RuntimeSyncMode      string                     `json:"runtime_sync_mode"`
	RuntimeTargets       []string                   `json:"runtime_targets"`
	MaterializedMarkdown *MaterializedMarkdownAsset `json:"materialized_markdown,omitempty"`
}

type MaterializedMarkdownAsset struct {
	BaseSectionID  string                       `json:"base_section_id"`
	SectionSources []MaterializedMarkdownSource `json:"section_sources"`
}

type MaterializedMarkdownSource struct {
	SectionID  string `json:"section_id"`
	SourcePath string `json:"source_path"`
}

type UpstreamSkillRef struct {
	Key              string `json:"key"`
	UpstreamPath     string `json:"upstream_path"`
	RepoUpstreamPath string `json:"repo_upstream_path"`
}

type RepoOwnedSharedSkills struct {
	Root        string   `json:"root"`
	RuntimeRoot string   `json:"runtime_root"`
	InstallMode string   `json:"install_mode"`
	Allowlist   []string `json:"allowlist"`
}

type RepoOwnedCommandBodies struct {
	Root        string                  `json:"root"`
	RuntimeRoot string                  `json:"runtime_root"`
	RenderMode  string                  `json:"render_mode"`
	Entries     []RepoOwnedCommandEntry `json:"entries"`
}

type RepoOwnedCommandEntry struct {
	Key         string `json:"key"`
	BodyPath    string `json:"body_path"`
	OwnerSkill  string `json:"owner_skill"`
	Mode        string `json:"mode"`
	CommandType string `json:"command_type"`
	Description string `json:"description"`
}

func loadManagedAssetsManifest(repoRoot string) (ManagedAssetsManifest, string, error) {
	manifestPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "managed-assets.json")
	var manifest ManagedAssetsManifest
	if err := readJSONFile(manifestPath, &manifest); err != nil {
		return ManagedAssetsManifest{}, manifestPath, fmt.Errorf("managed-assets file is not valid JSON at %s: %w", manifestPath, err)
	}
	return manifest, manifestPath, nil
}

func loadManagedAssetsTarget(repoRoot, targetName string) (ManagedAssetsTarget, string, error) {
	manifest, manifestPath, err := loadManagedAssetsManifest(repoRoot)
	if err != nil {
		return ManagedAssetsTarget{}, manifestPath, err
	}
	target, ok := manifest.Targets[targetName]
	if !ok {
		return ManagedAssetsTarget{}, manifestPath, fmt.Errorf("managed-assets manifest at %s does not define target %q", manifestPath, targetName)
	}
	return target, manifestPath, nil
}

func findOwnedOverlayAsset(assets []OwnedOverlayAsset, key string) (OwnedOverlayAsset, bool) {
	for _, asset := range assets {
		if asset.Key == key {
			return asset, true
		}
	}
	return OwnedOverlayAsset{}, false
}

func ownedAssetUpstreamPaths(asset OwnedOverlayAsset) []string {
	paths := []string{}
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		for _, existing := range paths {
			if existing == path {
				return
			}
		}
		paths = append(paths, path)
	}
	add(asset.UpstreamPath)
	if asset.MaterializedMarkdown != nil {
		for _, section := range asset.MaterializedMarkdown.SectionSources {
			add(section.SourcePath)
		}
	}
	return paths
}

func runtimePromptTarget(asset OwnedOverlayAsset) (string, bool) {
	for _, target := range asset.RuntimeTargets {
		expanded := expandUser(target)
		clean := filepath.ToSlash(expanded)
		if strings.Contains(clean, "/prompts/") {
			return expanded, true
		}
	}
	return "", false
}

func copyDirectoryContentsWithStatus(srcRoot, dstRoot string, prune bool, recorder *verboseRecorder, counters *writeCounters) error {
	info, err := os.Stat(srcRoot)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source directory missing: %s", srcRoot)
	}
	seen := map[string]bool{".": true}
	err = filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return ensureDir(dstRoot)
		}
		seen[rel] = true
		dstPath := filepath.Join(dstRoot, rel)
		if info.IsDir() {
			return ensureDir(dstPath)
		}
		status, err := copyFileWithStatus(path, dstPath)
		if err != nil {
			return err
		}
		trackWriteStatus(status, counters)
		if shouldRecordWriteStatus(status) {
			recorder.record(dstPath, fmt.Sprintf("installed owned directory file (%s)", describeWriteStatus(status)))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !prune {
		return nil
	}
	return pruneMissingOwnedDirectoryEntries(dstRoot, seen, recorder, counters)
}

func pruneMissingOwnedDirectoryEntries(dstRoot string, seen map[string]bool, recorder *verboseRecorder, counters *writeCounters) error {
	if !pathExists(dstRoot) {
		return nil
	}
	var paths []string
	err := filepath.Walk(dstRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dstRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	for _, rel := range paths {
		if seen[rel] {
			continue
		}
		full := filepath.Join(dstRoot, rel)
		if err := os.RemoveAll(full); err != nil {
			return err
		}
		trackWriteStatus("deleted", counters)
		recorder.record(full, "removed stale owned runtime entry during directory sync")
	}
	return nil
}
