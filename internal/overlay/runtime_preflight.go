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
	for key := range s.resolvedGlobalActors {
		required[key] = "global_actors"
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
			if ref.Key == "codex-local-instructions" && target.Path == "AGENTS.md" {
				content, err = prepareCodexLocalInstructionsMigration(content, target.Replacements)
				if err != nil {
					return upstreamRuntimeError("runtime patch %q target %s has incompatible local instructions: %v", ref.Key, path, err)
				}
			}
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

func prepareCodexLocalInstructionsMigration(content string, replacements []runtimePatchReplacement) (string, error) {
	const anchor = "<!-- /gentle-ai:engram-protocol -->\n"
	const legacyLine = "- Never commit, push, or create a pull request unless the user explicitly asks. An implementation request, task completion, or review gate does not authorize delivery.\n"
	const rule = "- When responding,"
	const marker = "gentle-ai-custom:codex-local-instructions"
	if len(replacements) == 0 || replacements[0].Before != anchor || !strings.HasPrefix(replacements[0].After, anchor+"\n") {
		return "", fmt.Errorf("unrecognized Codex local instruction patch")
	}
	current := strings.TrimPrefix(replacements[0].After, anchor+"\n")
	if strings.Count(current, marker) != 2 || strings.Count(current, rule) != 1 || strings.Contains(current, legacyLine) || !strings.HasSuffix(current, "<!-- /"+marker+" -->\n") {
		return "", fmt.Errorf("unrecognized Codex local instruction block")
	}
	legacy := strings.Replace(current, rule, legacyLine+rule, 1)
	const orchestratorStart = "<!-- gentle-ai:orchestrator -->"
	const orchestratorEnd = "<!-- /gentle-ai:orchestrator -->"
	const codegraphStart = "<!-- gentle-ai:codegraph-guidance -->"
	gap := orchestratorEnd + "\n\n" + legacy + "\n" + codegraphStart
	if gapIndex := strings.Index(content, gap); gapIndex >= 0 {
		startIndex := strings.Index(content, orchestratorStart)
		anchorIndex := strings.Index(content, anchor)
		currentIndex := strings.Index(content, replacements[0].After)
		if strings.Count(content, orchestratorStart) != 1 || strings.Count(content, orchestratorEnd) != 1 ||
			strings.Count(content, codegraphStart) != 1 || strings.Count(content, anchor) != 1 ||
			anchorIndex+len(anchor) > startIndex || startIndex+len(orchestratorStart) > gapIndex ||
			(currentIndex >= 0 && currentIndex+len(replacements[0].After) > startIndex) {
			return "", fmt.Errorf("ambiguous Codex orchestrator-to-CodeGraph boundary")
		}
		content = strings.Replace(content, gap, "<!-- /gentle-ai:orchestrator -->\n\n<!-- gentle-ai:codegraph-guidance -->", 1)
	}
	currentCount := strings.Count(content, replacements[0].After)
	if currentCount > 1 || strings.Count(content, marker) != 2*currentCount || strings.Contains(content, legacyLine) ||
		strings.Count(content, "gentle-ai-custom:codex-local") != 2*currentCount {
		return "", fmt.Errorf("duplicate, misplaced, or unknown Codex local instruction block")
	}
	return content, nil
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
