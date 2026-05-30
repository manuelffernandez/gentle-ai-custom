---
name: gentle-ai-overlay-maintainer
description: "Trigger: gentle ai update, auditar gentle ai, depurar gentle ai, refresh overlay. Maintain the gentle-ai-custom overlay against upstream Gentle AI changes."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.1"
---

# Gentle AI Overlay Maintainer

## Activation Contract

Use this skill when:
- Auditing or repairing the `gentle-ai-custom` overlay after upstream Gentle AI changes.
- The user ran `gentle-ai sync` or a TUI reinstall and the overlay needs to be re-applied.
- Topology drift is suspected (renamed/removed/added agents upstream).
- The script reported topology warnings, snapshot drift, or a broken-state error.

## Hard Rules

- Work from `gentle-ai-custom`, not from the upstream repo.
- Treat `/home/manuel/Documentos/gentle-ai` as upstream input only.
- ALWAYS triage the update type before deciding what to audit (see Update-Type Triage). The triage table itself describes what state the overlay is in for each path.
- Read semantic intent before making maintenance decisions.
- Preserve the local keep/prune baseline and the orchestrator sanitization goals.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes (see `AGENTS.md` rules 3 and 4).
- Do not change intent, keep/prune, or sanitization behavior for new upstream changes without explicit user approval.

## Update-Type Triage (MANDATORY first step)

Before doing anything else, determine what the user actually did. Ask if unclear.

| User did | State of overlay on disk | Audit needed | Re-apply script |
|---|---|---|---|
| `brew upgrade gentle-ai` only | Intact | Only if upstream changed | No |
| `gentle-ai sync` / "Sync Configurations" | **Broken**: prompts reset to upstream inline, pruned skills restored | Yes | **Yes — mandatory** |
| TUI reinstallation | **Broken + topology may have shifted** | Yes + topology check | **Yes — audit topology first** |

Re-apply paths are mandatory regardless of whether upstream content changed — because `gentle-ai sync` unconditionally wipes the overlay state in the filesystem.

## Decision Gates

| If | Then |
|---|---|
| User just ran `gentle-ai sync` | Re-apply overlay immediately (`bash apply-gentle-ai-custom.sh all`). Audit drift afterwards via snapshot diff. |
| User just ran TUI reinstall | Audit topology BEFORE re-applying. New/renamed/removed agents may require policy updates first. |
| Script printed `topology: ...` warnings | Investigate each warning. New explicit orchestrators need policy entries; missing/created entries need maintenance-intent updates. STOP and ask the user before mutating policy. |
| Script summary shows `snapshots - changed: N > 0` | Review `git diff overlay/gentle-ai/snapshots/`. If sanitizer anchors moved, update both scripts. |
| Script printed `orchestrators recovered from snapshot: N > 0` | User-side state was broken (deleted overlay files). Now consistent again. Worth noting in the log. |
| Script raised `broken state for orchestrator X` | Run `gentle-ai sync` to reset prompts to inline, then re-run the script. Record the cause in the log. |
| Sanitizer fails (`missing required marker` / `missing expected block`) | Upstream changed orchestrator structure. Update the sanitizers in both scripts before applying the overlay. |
| Upstream added new skills or workflow behavior | STOP, summarize the impact, and ask the user what to keep or depure. |
| The script can no longer sanitize safely | Fail closed, refresh docs, and record the blocker. |

## Execution Steps

1. **Triage**: determine update type (see Update-Type Triage). If unclear, ask the user.
2. Read `overlay/gentle-ai/policy/maintenance-intent.md`.
3. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
4. Read `overlay/gentle-ai/state/upstream-state.json`.
5. Read `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.
6. Inspect upstream `/home/manuel/Documentos/gentle-ai` and determine the current relevant version boundary (tag and/or commit).
7. If `last_maintained_commit` exists, review the upstream change range from `last_maintained_commit` to the current upstream state, including intermediate minor releases or commits in that range.
8. Classify findings into:
   - behavior / workflow / feature changes relevant to the overlay
   - topology changes (renamed/added/removed agents)
   - likely low-priority bugfix / chore noise
9. If relevant changes affect keep/prune intent, sanitization behavior, or topology, STOP and ask the user what to preserve or depure before editing anything.
10. After approval, update scripts, policy, docs, state, snapshots, and logs together.
11. **Run the script**: `bash apply-gentle-ai-custom.sh all` (or `.ps1` on Windows). Capture full output, including the `topology:` lines and the final `Summary:` block.
12. **Read the summary** and act on each signal (see Decision Gates).
13. **Verify post-state on disk** (read-only checks; ALL must pass):
    - For each path in `gentle-ai-policy.json` → `skills.targets`: none of `skills.prune` entries may exist as directories inside it.
    - For each `agent_overrides` entry: `~/.config/opencode/opencode.json` → `agent.<key>.model` must equal the policy value; `variant` must equal the policy value when set.
    - For each `orchestrator_agent_keys` entry: `agent.<key>.prompt` must be a `{file:...}` reference, and the referenced file must exist on disk.
    - For each `*.last.md` in `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`: a corresponding `*.overlay.md` must exist under `generated_orchestrators_dir`.
14. Record the decision in `overlay/gentle-ai/logs/update-log.md` (include update type, topology findings, snapshot drift, recovery events, and any policy mutations).

## Hardening option: external-single-active strategy

By default, `gentle-ai sync` resets the orchestrator prompts because `~/.config/opencode/profiles/` is empty. Creating any `*.json` file **directly under** that directory (subdirectories are ignored by upstream's `HasExternalProfileFiles`) flips upstream's profile strategy detection to `external-single-active`, which preserves the existing `{file:...}` reference during sync. Reference: `internal/components/sdd/profiles.go` → `ResolveProfileStrategy` in upstream (verify against the commit pinned in `overlay/gentle-ai/state/upstream-state.json` → `last_maintained_commit` before relying on it).

**Tradeoffs**:
- Pro: the overlay survives `gentle-ai sync` without needing to re-run the script for prompt restoration. Skills pruning still requires the script.
- Con (critical): the user keeps executing the **previous** sanitized version of the upstream prompt indefinitely. The script can no longer even attempt to re-sanitize against the new upstream because it never sees the new inline content.
- Con: `*.last.md` snapshots stop refreshing — `git diff overlay/gentle-ai/snapshots/` is no longer a useful drift signal.
- Con: when upstream sanitizer anchors move, you only notice the next time someone deletes the profile and triggers sync's default behavior, by which point you may have been running a stale-sanitized prompt for a long time.

This is opt-in only. Do not enable it without explicit user request and a full discussion of the tradeoffs above. The default behavior (sync resets, script re-applies) has the advantage of keeping snapshots as a live log of upstream state and guaranteeing every script run sanitizes against current upstream, not against an old capture.

## Output Contract

Return:
- update type the user performed (brew / sync / reinstall)
- files changed in the overlay (policy, scripts, docs, log)
- topology drift detected and how it was resolved (entries added to policy, intent updates, deferred to user)
- snapshot drift detected (which prompts changed since the previous run)
- script summary counts (generated, recovered, skipped, snapshots new/changed/unchanged, topology warnings)
- post-state verification result (which checks passed / failed)
- whether keep/prune or sanitizer rules changed
- what upstream range was audited
- whether user approval was required and how it affected the result
- any migration note the user should know

## References

- `../../../../overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`
- `../../../../overlay/gentle-ai/policy/maintenance-intent.md`
- `../../../../overlay/gentle-ai/state/upstream-state.json`
- `../../../../AGENTS.md` (Update flow table)
