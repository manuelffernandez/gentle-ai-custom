---
name: gentle-ai-overlay-maintainer
description: "Trigger: gentle ai update, auditar gentle ai, depurar gentle ai, refresh overlay. Maintain the gentle-ai-custom overlay against upstream Gentle AI changes."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.4"
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
- Follow this order for maintainer updates: update `gentle-ai` binary -> `git pull` upstream -> audit from `gentle-ai-custom` -> update this repo if needed -> run `gentle-ai sync` or reinstall -> run `apply-gentle-ai-custom`.
- ALWAYS triage the update type before deciding what to audit (see Update-Type Triage). The triage table itself describes what state the overlay is in for each path.
- After auditing upstream drift, ALWAYS translate the findings into an explicit adoption recommendation: `gentle-ai sync` when topology is unchanged, full reinstall when topology changed or sync cannot materialize the new upstream shape.
- Read semantic intent before making maintenance decisions.
- Preserve the local keep/prune baseline and the orchestrator sanitization goals.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes (see `AGENTS.md` rules 3 and 4).
- Do not change intent, keep/prune, or sanitization behavior for new upstream changes without explicit user approval.
- The versioned policy MUST NOT carry per-profile orchestrator/phase model+variant choices. Those live in the per-machine local config at `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`. If you need a new profile-managed assignment, edit the LOCAL file — do NOT add it back to `gentle-ai-policy.json`.
- The versioned repo MUST keep only `gentle-orchestrator.last.md` under `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`. Profile snapshots belong in `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.

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
| User updated upstream (`git pull`) but has NOT run `gentle-ai sync` yet | Run `bash audit-gentle-ai-upstream.sh` FIRST. Use it to detect base prompt drift + profile invariants before recommending sync/reinstall. |
| User just ran `gentle-ai sync` | Re-apply overlay immediately (`bash apply-gentle-ai-custom.sh opencode` minimum, or `all` if they also want to refresh every custom target). The apply entrypoint now auto-verifies the materialized `gentle-orchestrator` against the last audited baseline. |
| User just ran TUI reinstall | Audit topology BEFORE re-applying. New/renamed/removed agents may require policy updates first. |
| Script printed `topology: ...` warnings | Investigate each warning. New explicit orchestrators need policy entries; missing/created entries need maintenance-intent updates. STOP and ask the user before mutating policy. Note: `sdd-orchestrator-<name>` orchestrators are deliberately suppressed from prefix-only topology warnings; they belong to the SDD profile local config, not the versioned policy. |
| Script printed `WARNING - unmanaged SDD profiles left untouched` | A profile exists in `opencode.json` but is not named in `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`. Ask the user whether to add it to the local config (to manage) or delete its agent keys manually (to remove). NEVER delete it automatically. |
| Script raised `ERROR: local SDD profile config at ... is not valid JSON` / `... unexpected top-level field ...` / `... missing required field ...` / `... must be a non-empty string` / `... must match ^[a-z0-9]...` / `... missing required phases ...` / `... unknown phases ...` | The local profile config failed strict V1 validation. The script wrote nothing. Surface the exact error to the user; ask them to fix or remove the file. Do NOT relax the schema. |
| `bash audit-gentle-ai-upstream.sh` reports `base prompt drift: yes` | The upstream `gentle-orchestrator` base asset no longer matches the audited baseline. STOP, review the upstream delta, and update snapshot + metadata + state only after user approval. |
| `bash audit-gentle-ai-upstream.sh` reports `profile phase order: mismatch` / `profile orchestrator naming: mismatch` / `profile task scoping invariant: mismatch` / `base asset injection invariant: mismatch` | Upstream changed the mechanics that build named SDD profiles. STOP and review before recommending `sync`; the overlay assumptions may be stale even if the base prompt diff is small. |
| Script summary shows `repo snapshots - changed: N > 0` | Review `git diff overlay/gentle-ai/snapshots/`. Only the versioned `gentle-orchestrator` baseline should drift there. |
| Script summary shows `local snapshots - changed: N > 0` | A local operational snapshot changed under `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`. Mention it in verification/output, but do not expect a git diff for profile snapshots. |
| Script summary shows `local snapshot migrations from repo: N > 0` | Legacy repo snapshots were copied into the local operational directory. After verification, remove the old versioned profile snapshots from the repo so only `gentle-orchestrator.last.md` remains tracked. |
| Script printed `orchestrators recovered from snapshot: N > 0` | User-side state was broken (deleted overlay files). Now consistent again. Worth noting in the log. |
| Script raised `broken state for orchestrator X` | Run `gentle-ai sync` to reset prompts to inline, then re-run the script. Record the cause in the log. |
| Sanitizer fails (`missing required marker` / `missing expected block`) | Upstream changed orchestrator structure. Update the sanitizer in `internal/overlay/apply_policy.go` before applying the overlay. |
| Upstream topology changed (agents/presets added, removed, renamed, or upstream shape no longer matches what sync can refresh) | STOP, summarize the impact, ask the user what to preserve or depure, and explicitly recommend a full reinstall before re-applying the overlay. `gentle-ai sync` alone is not enough for topology drift. |
| Upstream added new skills or workflow behavior without topology drift | STOP, summarize the impact, ask the user what to keep or depure, and explicitly say that `gentle-ai sync` is the correct upstream refresh path before re-applying the overlay. |
| The script can no longer sanitize safely | Fail closed, refresh docs, and record the blocker. |

## Execution Steps

1. **Confirm the workflow order**: binary update -> upstream `git pull` -> maintainer audit from `gentle-ai-custom` -> overlay repo updates if needed -> `gentle-ai sync` or reinstall -> overlay apply.
2. **Triage**: determine update type (see Update-Type Triage). If unclear, ask the user.
3. Read `overlay/gentle-ai/policy/maintenance-intent.md`.
4. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
5. Read `overlay/gentle-ai/state/upstream-state.json`.
6. Read `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.
7. If the user has NOT run `gentle-ai sync` yet, work from `gentle-ai-custom`, activate the maintainer workflow, run `bash audit-gentle-ai-upstream.sh` (or `.ps1` on Windows) FIRST, and capture its full output.
8. Inspect upstream `/home/manuel/Documentos/gentle-ai` and determine the current relevant version boundary (tag and/or commit).
9. If `last_maintained_commit` exists, review the upstream change range from `last_maintained_commit` to the current upstream state, including intermediate minor releases or commits in that range.
10. Classify findings into:
   - base prompt drift (`gentle-orchestrator`)
   - profile-generation invariant drift
   - behavior / workflow / feature changes relevant to the overlay
   - topology changes (renamed/added/removed agents)
   - recommended upstream adoption path: `gentle-ai sync` vs full reinstall
   - likely low-priority bugfix / chore noise
11. If relevant changes affect keep/prune intent, sanitization behavior, or topology, STOP and ask the user what to preserve or depure before editing anything. In that same handoff, explicitly tell the user whether the audited upstream delta should be applied with `gentle-ai sync` or with a full reinstall, and why.
12. After approval, update this repo if needed: scripts, policy, docs, state, snapshots, metadata, and logs together.
13. Only after that, run the upstream refresh path recommended by the audit: `gentle-ai sync` or full reinstall.
14. **Run the apply entrypoint**: `bash apply-gentle-ai-custom.sh opencode` minimum (or `all` if the user also wants every custom target refreshed). Capture full output, including the `topology:` lines and the final `Summary:` block.
15. **Read the summary** and act on each signal (see Decision Gates).
16. **Verify post-state on disk** (read-only checks; ALL must pass):
    - For each path in `gentle-ai-policy.json` → `skills.targets`: none of `skills.prune` entries may exist as directories inside it.
    - For each `agent_overrides` entry: `~/.config/opencode/opencode.json` → `agent.<key>.model` must equal the policy value; `variant` must equal the policy value when set.
    - For each `orchestrator_agent_keys` entry: `agent.<key>.prompt` must be a `{file:...}` reference, and the referenced file must exist on disk.
    - `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` must contain only `gentle-orchestrator.last.md`, and its corresponding `*.overlay.md` must exist under `generated_orchestrators_dir`.
    - `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` must contain the operational snapshot for `gentle-orchestrator`, plus any managed `sdd-orchestrator-<name>` snapshots.
    - If `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` exists: for each profile `<name>` in it, `agent.sdd-orchestrator-<name>` and `agent.sdd-<phase>-<name>` (10 phases) must exist with matching `model` + `variant`.
    - `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml` must exist and match `upstream-state.json` + the SHA-256 of `gentle-orchestrator.last.md`.
17. Record the decision in `overlay/gentle-ai/logs/update-log.md` (include update type, topology findings, base prompt drift, profile invariant drift, recovery events, and any policy mutations).

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
- recommended upstream adoption path (`gentle-ai sync` or full reinstall) and the reason for that recommendation
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
