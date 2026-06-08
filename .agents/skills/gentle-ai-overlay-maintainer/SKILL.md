---
name: gentle-ai-overlay-maintainer
description: "Trigger: gentle ai update, auditar gentle ai, depurar gentle ai, refresh overlay. Maintain the gentle-ai-custom overlay against upstream Gentle AI changes."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.8"
---

# Gentle AI Overlay Maintainer

## Activation Contract

Use this skill when:
- Auditing or repairing the `gentle-ai-custom` overlay after upstream Gentle AI changes.
- The user ran `gentle-ai sync` or a TUI reinstall and the overlay needs to be re-applied.
- Topology drift is suspected (renamed/removed/added agents upstream).
- The maintainer needs to refresh approved upstream snapshots or verify the repo-owned runtime assets.

## Hard Rules

- Work from `gentle-ai-custom`, not from the upstream repo.
- Treat the resolved upstream `gentle-ai` clone as input only. Resolution order: `~/.config/gentle-ai-custom/opencode-local-config.json` (`upstream_repo_path`) -> `$GENTLE_AI_CUSTOM_UPSTREAM_REPO` -> `../gentle-ai` fallback.
- Follow this order for maintainer updates: update `gentle-ai` binary -> `git pull` upstream -> audit from `gentle-ai-custom` -> update this repo if needed -> refresh upstream runtime (`gentle-ai sync` or reinstall) -> run `apply-gentle-ai-custom`.
- `audit-gentle-ai-upstream` discovers drift from `git diff --name-status --find-renames <last_maintained_commit>..HEAD`, filtered through `overlay/gentle-ai/policy/managed-assets.json`, while still checking structural invariants from upstream integration code.
- `sync-gentle-ai-upstream-assets` refreshes `overlay/gentle-ai/assets/upstream/...` plus the maintained boundary only after the drift has been reviewed and accepted.
- `apply-gentle-ai-custom` is now canonical: it installs repo-owned SDD/runtime assets from `overlay/gentle-ai/assets/owned/...`, installs repo-owned portable skills from `shared/skills/`, renders wrapper commands from `shared/commands/`, prunes rejected upstream skills, applies built-in overrides, and reconciles SDD profiles.
- The apply path is driven by repo-owned runtime assets declared in `overlay/gentle-ai/assets/owned/...` plus portable repo-owned skills from `shared/skills/`.
- Read semantic intent before making maintenance decisions.
- Preserve the local keep/prune baseline and the repo-owned orchestrator behavior goals.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes, but write to the log only for eligible closed maintenance events under `AGENTS.md` rule 4.
- Do not change intent, keep/prune, or repo-owned orchestrator behavior for new upstream changes without explicit user approval.
- The versioned policy MUST NOT carry per-profile orchestrator/phase model+variant choices. Those live in the per-machine local config at `~/.config/gentle-ai-custom/opencode-local-config.json` under `default_profile` and `profiles`.

## Update-Type Triage

| User did | State of overlay on disk | Audit needed | Re-apply script |
|---|---|---|---|
| `brew upgrade gentle-ai` only | Intact | Only if upstream changed | No |
| `gentle-ai sync` / "Sync Configurations" | Prompts/skills reset to upstream runtime state | Yes | Yes |
| TUI reinstallation | Runtime reset and topology may have shifted | Yes | Yes, after topology review |

Re-apply is mandatory after `gentle-ai sync` or reinstall because upstream rewrites runtime prompt refs and restores pruned skills.

## Decision Gates

| If | Then |
|---|---|
| User updated upstream (`git pull`) but has NOT run `gentle-ai sync` yet | Run `bash audit-gentle-ai-upstream.sh` FIRST. Use it to detect managed-asset drift + profile invariants before recommending sync/reinstall. |
| User just ran `gentle-ai sync` | Re-apply overlay immediately with `bash apply-gentle-ai-custom.sh opencode` (minimum) or `all`. |
| User just ran TUI reinstall | Audit topology BEFORE re-applying. |
| Script printed `topology: ...` warnings | Investigate each warning. STOP and ask before mutating policy/intent. |
| Script printed `WARNING - unmanaged SDD profiles left untouched` | Ask whether to add the profile(s) to the local config or remove those agent keys manually. NEVER delete them automatically. |
| `bash audit-gentle-ai-upstream.sh` reports `base prompt drift: yes` | Review `Drift summary:` first, then inspect the upstream delta before updating approved upstream snapshots/state. |
| `bash audit-gentle-ai-upstream.sh` reports profile/base invariant mismatch | STOP and review before recommending `sync`; the overlay assumptions may be stale even if prompt drift looks small. |
| Upstream topology changed | STOP, summarize the impact, and explicitly recommend a full reinstall before re-applying the overlay. |
| Upstream added new skills or workflow behavior without topology drift | STOP, summarize the impact, and recommend `gentle-ai sync` after the repo is updated. |

## Execution Steps

1. Confirm the workflow order: binary update -> upstream `git pull` -> maintainer audit from `gentle-ai-custom` -> overlay repo updates if needed -> `gentle-ai sync` or reinstall -> overlay apply.
2. Determine update type (see Update-Type Triage). If unclear, ask.
3. Read `overlay/gentle-ai/policy/maintenance-intent.md`.
4. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
5. Read `overlay/gentle-ai/policy/managed-assets.json`.
6. Read `overlay/gentle-ai/state/upstream-state.json`.
7. Read `overlay/gentle-ai/maintenance.md`.
8. If the user has NOT run `gentle-ai sync` yet, run `bash audit-gentle-ai-upstream.sh` first and capture its output.
9. Review the upstream range from `last_maintained_commit` to the current upstream head.
10. Classify findings into:
   - base prompt drift (`gentle-orchestrator`)
   - profile-generation invariant drift
   - behavior/workflow changes relevant to the overlay
   - topology changes
   - recommended upstream adoption path: `gentle-ai sync` vs full reinstall
11. Produce a short plain-language drift summary before the full diff review.
12. If relevant changes affect keep/prune intent, repo-owned orchestrator behavior, or topology, STOP and ask the user what to preserve or depure before editing anything.
13. After approval, update this repo if needed: owned assets, approved upstream snapshots, policy, docs, state, and logs together. Run `bash sync-gentle-ai-upstream-assets.sh` when the approved upstream copy in `overlay/gentle-ai/assets/upstream/` should move forward.
14. Only after that, run the recommended upstream refresh path: `gentle-ai sync` or full reinstall.
15. Run the apply entrypoint: `bash apply-gentle-ai-custom.sh opencode` minimum (or `all`). Capture the `Summary:` block.
16. Verify post-state on disk:
   - pruned skills are absent only in the selected registered runtime targets
   - owned runtime targets declared in `managed-assets.json` exist on disk
   - `agent_overrides` resolve to the configured `model` / `variant`
   - `agent.gentle-orchestrator.prompt` points to the owned runtime prompt file
   - each managed named profile orchestrator points to the same owned runtime prompt file
   - each unsuffixed and managed suffixed SDD phase agent points to the owned runtime prompt file for that phase
   - default/named profile assignments still match the local config
   - `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` plus metadata remain consistent with `upstream-state.json`
17. If the work closed an eligible maintenance event, record one consolidated entry in `overlay/gentle-ai/logs/update-log.md`.

## Output Contract

Return:
- update type the user performed (brew / sync / reinstall)
- files changed in the overlay (policy, assets, scripts, docs, log)
- topology drift detected and how it was resolved
- brief drift summary in plain language
- recommended upstream adoption path (`gentle-ai sync` or full reinstall) and why
- apply summary counts (owned asset writes, prompt ref updates, topology warnings, profile counts)
- post-state verification result (what passed / failed)
- whether keep/prune or repo-owned orchestrator behavior changed
- what upstream range was audited
- whether user approval was required and how it affected the result

## References

- `../../../../overlay/gentle-ai/maintenance.md`
- `../../../../overlay/gentle-ai/policy/maintenance-intent.md`
- `../../../../overlay/gentle-ai/policy/managed-assets.json`
- `../../../../overlay/gentle-ai/state/upstream-state.json`
- `../../../../AGENTS.md`
