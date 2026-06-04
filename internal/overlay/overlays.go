package overlay

import (
	"fmt"
	"path/filepath"
	"strings"
)

// generateOverlays iterates over all orchestrator agents in the config, captures
// their inline prompts into local (and optionally repo) snapshots, sanitizes the
// content, writes the generated overlay file, and rewrites the agent prompt to a
// {file:...} reference.
func (s *applyPolicyState) generateOverlays() error {
	for _, dir := range []string{s.generatedDir, s.repoSnapshotDir, s.localSnapshotDir} {
		if err := ensureDir(dir); err != nil {
			return err
		}
	}

	for _, key := range sortedKeys(s.agents) {
		if !s.isOrchestrator(key) {
			continue
		}
		agent, ok := jsonObject(s.agents[key])
		if !ok {
			fmt.Printf("  skip %s: agent entry is not an object\n", key)
			s.metrics.skippedCount++
			continue
		}
		prompt, ok := agent["prompt"].(string)
		if !ok || strings.TrimSpace(prompt) == "" {
			fmt.Printf("  skip %s: prompt missing or not a string\n", key)
			s.metrics.skippedCount++
			continue
		}

		safeKey, err := safeSnapshotKey(key)
		if err != nil {
			return err
		}
		generatedPath := filepath.Join(s.generatedDir, safeKey+".overlay.md")
		desiredPrompt := "{file:" + generatedPath + "}"
		repoSnapshotPath := filepath.Join(s.repoSnapshotDir, safeKey+".last.md")
		localSnapshotPath := filepath.Join(s.localSnapshotDir, safeKey+".last.md")

		s.migrateRepoSnapshotToLocal(key, repoSnapshotPath, localSnapshotPath)
		s.backfillRepoSnapshotFromLocal(key, localSnapshotPath, repoSnapshotPath)

		if prompt == desiredPrompt && pathExists(generatedPath) {
			if !pathExists(localSnapshotPath) {
				return fmt.Errorf("local operational snapshot missing for orchestrator %q at %s. Run `gentle-ai sync` to reset the orchestrator prompt to inline content, then re-run this script to capture a fresh snapshot.", key, localSnapshotPath)
			}
			if s.shouldWriteRepoSnapshot(key) && !pathExists(repoSnapshotPath) {
				s.backfillRepoSnapshotFromLocal(key, localSnapshotPath, repoSnapshotPath)
				if !pathExists(repoSnapshotPath) {
					return fmt.Errorf("versioned repo snapshot missing for orchestrator %q at %s. Run `gentle-ai sync` to capture fresh upstream, then re-run this script.", key, repoSnapshotPath)
				}
			}
			fmt.Printf("  keep %s: already points to generated overlay prompt\n", key)
			if key == s.policy.OpenCode.BaseOrchestratorKey {
				content, err := readText(localSnapshotPath)
				if err != nil {
					return fmt.Errorf("Cannot read local operational snapshot for audited base orchestrator at %s: %v", localSnapshotPath, err)
				}
				s.baseRuntimePrompt = strings.TrimRight(content, "\r\n")
				s.baseGeneratedPath = generatedPath
			}
			s.writtenOrchestrators[key] = true
			s.metrics.keptCount++
			continue
		}

		recovered := false
		inlinePrompt := prompt
		if strings.HasPrefix(prompt, "{file:") && strings.HasSuffix(prompt, "}") {
			if !pathExists(localSnapshotPath) {
				s.migrateRepoSnapshotToLocal(key, repoSnapshotPath, localSnapshotPath)
			}
			if !pathExists(localSnapshotPath) {
				missingDetail := fmt.Sprintf("no local operational snapshot exists at %s", localSnapshotPath)
				if s.shouldWriteRepoSnapshot(key) {
					missingDetail = fmt.Sprintf("no local operational snapshot exists at %s and no repo snapshot exists at %s", localSnapshotPath, repoSnapshotPath)
				}
				return fmt.Errorf("broken state for orchestrator %q: opencode.json prompt is %q but the target file is missing and %s. Run `gentle-ai sync` to reset the orchestrator prompt to inline content, then re-run this script.", key, prompt, missingDetail)
			}
			if s.shouldWriteRepoSnapshot(key) && !pathExists(repoSnapshotPath) {
				s.backfillRepoSnapshotFromLocal(key, localSnapshotPath, repoSnapshotPath)
			}
			content, err := readText(localSnapshotPath)
			if err != nil {
				return err
			}
			inlinePrompt = strings.TrimRight(content, "\r\n")
			recovered = true
			fmt.Printf("  WARNING recovering %s from local snapshot - content may pre-date current upstream; run `gentle-ai sync` then re-run this script to capture fresh upstream into the snapshot\n", key)
		}

		sanitized, err := sanitizePrompt(inlinePrompt, s.policy)
		if err != nil {
			return err
		}
		snapshotStatus := "recovered"
		if !recovered {
			localStatus, err := writeSnapshotWithStatus(localSnapshotPath, inlinePrompt, &s.metrics.localSnapshots)
			if err != nil {
				return err
			}
			if shouldRecordWriteStatus(localStatus) {
				s.recordVerbose(localSnapshotPath, fmt.Sprintf("local snapshot for %s (%s)", key, describeWriteStatus(localStatus)))
			}
			if s.shouldWriteRepoSnapshot(key) {
				repoStatus, err := writeSnapshotWithStatus(repoSnapshotPath, inlinePrompt, &s.metrics.repoSnapshots)
				if err != nil {
					return err
				}
				if shouldRecordWriteStatus(repoStatus) {
					s.recordVerbose(repoSnapshotPath, fmt.Sprintf("versioned repo snapshot for %s (%s)", key, describeWriteStatus(repoStatus)))
				}
				snapshotStatus = fmt.Sprintf("local: %s, repo: %s", localStatus, repoStatus)
			} else {
				snapshotStatus = fmt.Sprintf("local: %s", localStatus)
			}
		}

		overlayStatus, err := writeTextFileWithStatus(generatedPath, sanitized)
		if err != nil {
			return err
		}
		oldPrompt := jsonString(agent["prompt"])
		if oldPrompt != desiredPrompt {
			agent["prompt"] = desiredPrompt
			s.configChanged = true
			s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.prompt: %s -> %s", key, summarizePromptValue(oldPrompt), desiredPrompt))
		}
		if key == s.policy.OpenCode.BaseOrchestratorKey {
			s.baseRuntimePrompt = inlinePrompt
			s.baseGeneratedPath = generatedPath
		}
		s.writtenOrchestrators[key] = true
		if recovered {
			s.metrics.recoveredCount++
			if shouldRecordWriteStatus(overlayStatus) {
				s.recordVerbose(generatedPath, fmt.Sprintf("recovered sanitized overlay for %s from local snapshot (%s)", key, describeWriteStatus(overlayStatus)))
			}
			fmt.Printf("  recovered %s -> %s (from snapshot)\n", key, generatedPath)
		} else {
			s.metrics.generatedCount++
			if shouldRecordWriteStatus(overlayStatus) {
				s.recordVerbose(generatedPath, fmt.Sprintf("sanitized overlay for %s (%s)", key, describeWriteStatus(overlayStatus)))
			}
			fmt.Printf("  generated %s -> %s (snapshot: %s)\n", key, generatedPath, snapshotStatus)
		}
	}
	return nil
}

// isOrchestrator reports whether key identifies an orchestrator agent, either by
// exact match against the policy's known keys or by prefix match.
func (s *applyPolicyState) isOrchestrator(key string) bool {
	for _, exact := range s.policy.OpenCode.OrchestratorAgentKeys {
		if key == exact {
			return true
		}
	}
	for _, prefix := range s.policy.OpenCode.OrchestratorAgentPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// isProfileOrchestrator reports whether key is a profile-scoped orchestrator
// (i.e. matches the profile orchestrator prefix from policy).
func (s *applyPolicyState) isProfileOrchestrator(key string) bool {
	prefix := s.policy.OpenCode.ProfileOrchestratorPrefix
	return prefix != "" && strings.HasPrefix(key, prefix)
}

// shouldWriteRepoSnapshot reports whether the versioned repo snapshot directory
// should be updated for this agent key (only exact-match orchestrators, not
// profile-prefixed ones).
func (s *applyPolicyState) shouldWriteRepoSnapshot(key string) bool {
	for _, exact := range s.policy.OpenCode.OrchestratorAgentKeys {
		if key == exact {
			return true
		}
	}
	return false
}
