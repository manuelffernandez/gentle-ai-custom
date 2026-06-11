package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type writeCounters struct {
	New       int
	Changed   int
	Unchanged int
	Deleted   int
}

func trackWriteStatus(status string, counters *writeCounters) {
	if counters == nil {
		return
	}
	switch status {
	case "new":
		counters.New++
	case "changed":
		counters.Changed++
	case "unchanged":
		counters.Unchanged++
	case "deleted":
		counters.Deleted++
	}
}

func (s *applyPolicyState) installOwnedAssets(repoRoot string) error {
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		if asset.RepoOwnedPath == "" {
			// Upstream-source-only asset: no owned file to install; skipped by apply.
			continue
		}
		source := ownedAssetSourcePath(repoRoot, asset)
		if strings.HasSuffix(asset.Kind, "_directory") {
			if err := s.installOwnedDirectoryAsset(asset, source); err != nil {
				return err
			}
			continue
		}
		if err := s.installOwnedFileAsset(asset, source); err != nil {
			return err
		}
	}
	return nil
}

func (s *applyPolicyState) installOwnedDirectoryAsset(asset OwnedOverlayAsset, source string) error {
	for _, runtimeTarget := range asset.RuntimeTargets {
		destination := expandUser(runtimeTarget)
		prune := asset.RuntimeSyncMode != "merge"
		if err := copyDirectoryContentsWithStatus(source, destination, prune, s.recorder, &s.metrics.ownedAssetWrites); err != nil {
			return fmt.Errorf("cannot install owned directory asset %q from %s to %s: %w", asset.Key, source, destination, err)
		}
		fmt.Printf("  installed owned directory %s -> %s\n", asset.Key, destination)
	}
	return nil
}

func (s *applyPolicyState) installOwnedFileAsset(asset OwnedOverlayAsset, source string) error {
	for _, runtimeTarget := range asset.RuntimeTargets {
		destination := expandUser(runtimeTarget)
		status, err := copyFileWithStatus(source, destination)
		if err != nil {
			return fmt.Errorf("cannot install owned asset %q from %s to %s: %w", asset.Key, source, destination, err)
		}
		trackWriteStatus(status, &s.metrics.ownedAssetWrites)
		if shouldRecordWriteStatus(status) {
			s.recordVerbose(destination, fmt.Sprintf("installed owned asset %s (%s)", asset.Key, describeWriteStatus(status)))
		}
		fmt.Printf("  installed owned asset %s -> %s\n", asset.Key, destination)
	}
	return nil
}

func (s *applyPolicyState) rewriteManagedPromptReferences() error {
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		promptPath, ok := runtimePromptTarget(asset)
		if !ok {
			continue
		}
		ref := "{file:" + promptPath + "}"
		switch asset.Kind {
		case "orchestrator_prompt":
			if err := s.ensurePromptReference(s.policy.OpenCode.BaseOrchestratorKey, ref); err != nil {
				return err
			}
			for name := range s.managedProfiles {
				if err := s.ensurePromptReference(s.policy.OpenCode.ProfileOrchestratorPrefix+name, ref); err != nil {
					return err
				}
			}
		case "phase_skill_prompt":
			if err := s.ensurePromptReference(asset.Key, ref); err != nil {
				return err
			}
			for name := range s.managedProfiles {
				if err := s.ensurePromptReference(asset.Key+"-"+name, ref); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *applyPolicyState) ensurePromptReference(agentKey, ref string) error {
	agent, ok := jsonObject(s.agents[agentKey])
	if !ok {
		return fmt.Errorf("required SDD agent %q missing from opencode.json; run `gentle-ai sync` or reinstall before re-applying the overlay", agentKey)
	}
	oldPrompt := jsonString(agent["prompt"])
	if oldPrompt != ref {
		agent["prompt"] = ref
		s.configChanged = true
		s.metrics.promptRefsUpdated++
		s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.prompt: %s -> %s", agentKey, summarizePromptValue(oldPrompt), ref))
	} else {
		s.metrics.promptRefsUnchanged++
	}
	if s.expectedPromptRefs == nil {
		s.expectedPromptRefs = map[string]string{}
	}
	s.expectedPromptRefs[agentKey] = ref
	return nil
}

func (s *applyPolicyState) verifyOwnedAssetTargets() error {
	verified := 0
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		if asset.RepoOwnedPath == "" {
			// Upstream-source-only asset: never installed to runtime; nothing to verify.
			continue
		}
		for _, runtimeTarget := range asset.RuntimeTargets {
			path := expandUser(runtimeTarget)
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("owned runtime target missing for %q at %s: %w", asset.Key, path, err)
			}
			if strings.HasSuffix(asset.Kind, "_directory") && !info.IsDir() {
				return fmt.Errorf("owned runtime target for %q must be a directory: %s", asset.Key, path)
			}
			verified++
		}
	}
	s.metrics.ownedRuntimeTargetsVerified = verified
	return nil
}

func ownedAssetSourcePath(repoRoot string, asset OwnedOverlayAsset) string {
	return filepath.Join(repoRoot, filepath.FromSlash(asset.RepoOwnedPath))
}
