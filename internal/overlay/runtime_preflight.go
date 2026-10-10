package overlay

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type runtimePatchDefinition struct {
	Version  int                   `json:"version"`
	Targets  []runtimePatchTarget  `json:"targets"`
	Sections []runtimePatchSection `json:"sections"`
}

type runtimePatchSection struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type runtimePatchTarget struct {
	Path         string                    `json:"path"`
	Replacements []runtimePatchReplacement `json:"replacements"`
}

type runtimePatchReplacement struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type preparedRuntimePatch struct {
	path    string
	content string
	mode    os.FileMode
}

type runtimeFileSnapshot struct {
	path    string
	mode    os.FileMode
	content []byte
}

func (s *applyPolicyState) preflightCustomRuntimePrerequisites(repoRoot string) error {
	if err := s.rejectUnmanagedObsoletePromptRefs(); err != nil {
		return err
	}
	if s.resolvedDefaultProfile != nil || len(s.resolvedProfiles) > 0 {
		if err := s.requireProfileBasePrompts(); err != nil {
			return err
		}
	}
	if len(s.resolvedProfiles) > 0 {
		base, _ := jsonObject(s.agents[s.policy.OpenCode.BaseOrchestratorKey])
		if _, _, err := namedOrchestratorPermissionContract(base); err != nil {
			return upstreamRuntimeError("profile reconciliation requires valid base orchestrator permissions: %v", err)
		}
	}
	if err := s.requireConfiguredActorBases(); err != nil {
		return err
	}
	if err := s.prepareRemoteAuthorizationRemoval(); err != nil {
		return err
	}
	if err := s.prepareRuntimePatches(repoRoot); err != nil {
		return err
	}
	return nil
}

func (s *applyPolicyState) requireConfiguredActorBases() error {
	required := map[string]string{}
	if s.resolvedDefaultProfile != nil {
		for key := range s.resolvedDefaultProfile.JudgmentDay {
			required[key] = "default_profile.judgment_day"
		}
		for key := range s.resolvedDefaultProfile.Review {
			required[key] = "default_profile.review"
		}
	}
	for _, profile := range s.resolvedProfiles {
		for _, key := range judgmentDayActorKeys {
			required[key] = fmt.Sprintf("profiles[%q].judgment_day", profile.Name)
		}
	}
	keys := make([]string, 0, len(required))
	for key := range required {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, ok := jsonObject(s.agents[key]); !ok {
			return upstreamRuntimeError("%s references missing upstream base actor %q", required[key], key)
		}
	}
	return nil
}

func (s *applyPolicyState) requireProfileBasePrompts() error {
	keys := append([]string{s.policy.OpenCode.BaseOrchestratorKey}, s.policy.OpenCode.SDDPhases...)
	for _, key := range keys {
		agent, ok := jsonObject(s.agents[key])
		if !ok {
			return upstreamRuntimeError("profile reconciliation requires installed base agent %q", key)
		}
		prompt := jsonString(agent["prompt"])
		if strings.TrimSpace(prompt) == "" || strings.Contains(prompt, ".overlay.md") {
			return upstreamRuntimeError("profile reconciliation requires a safe upstream base prompt for %q", key)
		}
		if err := s.snapshotPromptFile(prompt); err != nil {
			return upstreamRuntimeError("profile reconciliation requires prompt for %q: %v", key, err)
		}
		s.expectedPromptRefs[key] = prompt
	}
	return nil
}

func (s *applyPolicyState) snapshotPromptFile(prompt string) error {
	path, isFile, err := s.runtimePromptFilePath(prompt)
	if err != nil {
		return err
	}
	if !isFile {
		return nil
	}
	for _, snapshot := range s.preservedUpstreamFiles {
		if snapshot.path == path {
			return nil
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", path, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	s.preservedUpstreamFiles = append(s.preservedUpstreamFiles, runtimeFileSnapshot{path: path, mode: info.Mode(), content: content})
	return nil
}

func upstreamRuntimeError(format string, args ...any) error {
	return fmt.Errorf("upstream runtime preflight failed: %s; run `gentle-ai sync` or reinstall first, then re-apply the overlay", fmt.Sprintf(format, args...))
}

func (s *applyPolicyState) resolveRuntimePrompt(prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is empty")
	}
	path, isFile, err := s.runtimePromptFilePath(prompt)
	if err != nil {
		return "", err
	}
	if !isFile {
		return prompt, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}
	return string(raw), nil
}

func (s *applyPolicyState) runtimePromptFilePath(prompt string) (string, bool, error) {
	if !strings.HasPrefix(prompt, "{file:") {
		return "", false, nil
	}
	if !strings.HasSuffix(prompt, "}") {
		return "", false, fmt.Errorf("malformed prompt file reference %q", prompt)
	}
	path := strings.TrimSuffix(strings.TrimPrefix(prompt, "{file:"), "}")
	if strings.TrimSpace(path) == "" || strings.ContainsAny(path, "{}") {
		return "", false, fmt.Errorf("malformed prompt file reference %q", prompt)
	}
	if strings.HasPrefix(path, "~") {
		path = expandUser(path)
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(s.configPath), filepath.FromSlash(path))
	}
	return filepath.Clean(path), true, nil
}

func (s *applyPolicyState) rejectUnmanagedObsoletePromptRefs() error {
	managed := map[string]bool{s.policy.OpenCode.BaseOrchestratorKey: true}
	for _, profile := range s.resolvedProfiles {
		managed[s.policy.OpenCode.ProfileOrchestratorPrefix+profile.Name] = true
		for _, phase := range s.policy.OpenCode.SDDPhases {
			managed[phase+"-"+profile.Name] = true
		}
	}
	for key, raw := range s.agents {
		agent, ok := jsonObject(raw)
		if !ok {
			continue
		}
		prompt := jsonString(agent["prompt"])
		if !strings.Contains(prompt, ".overlay.md") {
			continue
		}
		if !managed[key] {
			return upstreamRuntimeError("unmanaged agent %q still references obsolete custom prompt %q; manage or remove that profile before retrying", key, prompt)
		}
		if key == s.policy.OpenCode.BaseOrchestratorKey {
			return upstreamRuntimeError("base orchestrator still references obsolete custom prompt %q", prompt)
		}
		path, isFile, err := s.runtimePromptFilePath(prompt)
		if err != nil {
			return upstreamRuntimeError("managed agent %q has malformed obsolete custom prompt reference %q", key, prompt)
		}
		if !isFile {
			return upstreamRuntimeError("managed agent %q has malformed obsolete custom prompt reference %q", key, prompt)
		}
		s.obsoletePromptPaths = append(s.obsoletePromptPaths, path)
	}
	return nil
}

func (s *applyPolicyState) prepareRuntimePatches(repoRoot string) error {
	for _, ref := range s.managedTarget.RuntimePatches {
		definitionPath := filepath.Join(repoRoot, filepath.FromSlash(ref.PatchPath))
		var definition runtimePatchDefinition
		if err := readJSONFile(definitionPath, &definition); err != nil {
			return fmt.Errorf("runtime patch %q is invalid at %s: %w", ref.Key, definitionPath, err)
		}
		if definition.Version != 1 {
			return fmt.Errorf("runtime patch %q at %s has unsupported version %d", ref.Key, definitionPath, definition.Version)
		}
		if len(definition.Sections) != 0 {
			return fmt.Errorf("runtime patch %q uses unsupported section replacement data", ref.Key)
		}
		for _, target := range definition.Targets {
			runtimeRoot := filepath.Dir(s.configPath)
			path, err := resolveRuntimePatchTarget(runtimeRoot, target.Path)
			if err != nil {
				return upstreamRuntimeError("runtime patch %q target %q is outside the selected runtime root: %v", ref.Key, target.Path, err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return upstreamRuntimeError("runtime patch %q requires installed upstream file %s: %v", ref.Key, path, err)
			}
			content := string(raw)
			for index, replacement := range target.Replacements {
				if replacement.Before == "" || replacement.After == "" || replacement.Before == replacement.After {
					return fmt.Errorf("runtime patch %q target %s replacement %d is invalid", ref.Key, target.Path, index+1)
				}
				afterCount := strings.Count(content, replacement.After)
				contentWithoutAppliedDelta := strings.ReplaceAll(content, replacement.After, "")
				unpatchedAnchorCount := strings.Count(contentWithoutAppliedDelta, replacement.Before)
				if afterCount == 1 && unpatchedAnchorCount == 0 {
					continue
				}
				if afterCount != 0 || unpatchedAnchorCount != 1 {
					return upstreamRuntimeError("runtime patch %q target %s has incompatible anchor %d", ref.Key, path, index+1)
				}
				content = strings.Replace(content, replacement.Before, replacement.After, 1)
			}
			info, err := os.Stat(path)
			if err != nil {
				return upstreamRuntimeError("runtime patch %q requires installed upstream file %s: %v", ref.Key, path, err)
			}
			s.runtimePatches = append(s.runtimePatches, preparedRuntimePatch{path: path, content: content, mode: info.Mode()})
		}
	}
	return nil
}

func resolveRuntimePatchTarget(runtimeRoot, target string) (string, error) {
	targetPath := filepath.FromSlash(target)
	if target == "" || filepath.IsAbs(targetPath) {
		return "", fmt.Errorf("target path must be relative")
	}

	root, err := filepath.Abs(runtimeRoot)
	if err != nil {
		return "", err
	}
	path := filepath.Clean(filepath.Join(root, targetPath))
	if err := ensurePathWithinRoot(root, path); err != nil {
		return "", err
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("cannot resolve runtime root: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve target path: %w", err)
	}
	if err := ensurePathWithinRoot(resolvedRoot, resolvedPath); err != nil {
		return "", fmt.Errorf("target resolves outside runtime root: %w", err)
	}
	return resolvedPath, nil
}

func ensurePathWithinRoot(root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("path %s escapes %s", path, root)
	}
	return nil
}

func (s *applyPolicyState) applyRuntimePatches() error {
	for _, patch := range s.runtimePatches {
		status, err := writeAtomicFileWithStatus(patch.path, []byte(patch.content), patch.mode.Perm())
		if err != nil {
			return err
		}
		if shouldRecordWriteStatus(status) {
			s.recordVerbose(patch.path, fmt.Sprintf("applied runtime patch (%s)", describeWriteStatus(status)))
		}
	}
	return nil
}

func (s *applyPolicyState) removeObsoleteManagedPrompts() error {
	if len(s.obsoletePromptPaths) == 0 {
		return nil
	}
	for _, raw := range s.agents {
		agent, ok := jsonObject(raw)
		if !ok {
			continue
		}
		if strings.Contains(jsonString(agent["prompt"]), ".overlay.md") {
			return fmt.Errorf("refusing to remove obsolete prompt while an agent still references it")
		}
	}
	sort.Strings(s.obsoletePromptPaths)
	for _, path := range s.obsoletePromptPaths {
		allowedRoot := filepath.Join(filepath.Dir(s.configPath), "prompts", "sdd", "orchestrators")
		rel, err := filepath.Rel(allowedRoot, path)
		if err != nil || strings.HasPrefix(rel, "..") || !strings.HasSuffix(path, ".overlay.md") {
			return fmt.Errorf("refusing to remove unsafe obsolete prompt path %s", path)
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *applyPolicyState) verifyRuntimePatches() error {
	for _, patch := range s.runtimePatches {
		raw, err := os.ReadFile(patch.path)
		info, statErr := os.Stat(patch.path)
		if err != nil || statErr != nil || string(raw) != patch.content || info.Mode() != patch.mode {
			return fmt.Errorf("post-write verification failed for runtime patch target %s", patch.path)
		}
	}
	return nil
}

func (s *applyPolicyState) verifyPreservedUpstreamFiles() error {
	for _, snapshot := range s.preservedUpstreamFiles {
		info, err := os.Stat(snapshot.path)
		if err != nil {
			return fmt.Errorf("upstream-owned prerequisite disappeared during custom apply: %s: %w", snapshot.path, err)
		}
		content, err := os.ReadFile(snapshot.path)
		if err != nil {
			return err
		}
		if info.Mode() != snapshot.mode || !bytes.Equal(content, snapshot.content) {
			return fmt.Errorf("custom apply mutated upstream-owned prerequisite %s", snapshot.path)
		}
	}
	return nil
}
