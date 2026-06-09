---
name: gentle-ai-overlay-maintainer
description: "Trigger: gentle ai update, auditar gentle ai, depurar gentle ai, refresh overlay. Maintain the gentle-ai-custom overlay against upstream Gentle AI changes."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.9"
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
- Follow this order for maintainer updates: update `gentle-ai` binary -> `git pull` upstream -> audit from `gentle-ai-custom` -> approval gate -> repo sync via `sync-gentle-ai-upstream-assets` when a new boundary was accepted -> refresh upstream runtime (`gentle-ai sync` or reinstall) -> run `apply-gentle-ai-custom`.
- `audit-gentle-ai-upstream` discovers drift from `git diff --name-status --find-renames <last_maintained_commit>..HEAD`, filtered through `overlay/gentle-ai/policy/managed-assets.json`, while still checking structural invariants from upstream integration code.
- `sync-gentle-ai-upstream-assets` refreshes `overlay/gentle-ai/assets/upstream/...` plus the maintained boundary only after the drift has been reviewed and accepted.
- `apply-gentle-ai-custom` is now canonical: it installs repo-owned SDD/runtime assets from `overlay/gentle-ai/assets/owned/...`, installs repo-owned portable skills from `shared/skills/`, renders wrapper commands from `shared/commands/`, prunes rejected upstream skills, applies built-in overrides, and reconciles SDD profiles.
- The apply path is driven by repo-owned runtime assets declared in `overlay/gentle-ai/assets/owned/...` plus portable repo-owned skills from `shared/skills/`.
- Read semantic intent before making maintenance decisions.
- Before any repo mutation, produce a concise decision summary that states what is new upstream, what to adopt, what to discard, the rationale for each, whether repo sync is required, and the recommended upstream refresh path.
- After that summary, STOP for explicit user approval before mutating this repo, advancing the maintained upstream boundary, syncing approved upstream assets, or refreshing local runtime state.
- Preserve the local keep/prune baseline and the repo-owned orchestrator behavior goals.
- Keep rejecting upstream changes that reintroduce `chained-pr`, review-budget, or review-workload governance into the repo-owned orchestrator behavior unless the user explicitly changes maintenance intent.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes, but write to the log only for eligible closed maintenance events under `AGENTS.md` rule 4.
- Do not change intent, keep/prune, or repo-owned orchestrator behavior for new upstream changes without explicit user approval.
- After approved maintenance work, return a closing summary of what was actually adopted, what was actually discarded, and why.
- After maintenance edits, run one fresh-context reviewer/subagent consistency pass over the changed maintainer artifacts before closing.
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
| User updated upstream (`git pull`) but has NOT run `gentle-ai sync` yet | Run `bash audit-gentle-ai-upstream.sh` FIRST. Use it to detect managed-asset drift + profile invariants before recommending repo sync + runtime refresh. |
| User just ran `gentle-ai sync` | Re-apply overlay immediately with `bash apply-gentle-ai-custom.sh opencode` (minimum) or `all`. |
| User just ran TUI reinstall | Audit topology BEFORE re-applying. |
| Script printed `topology: ...` warnings | Investigate each warning. STOP and ask before mutating policy/intent. |
| Script printed `WARNING - unmanaged SDD profiles left untouched` | Ask whether to add the profile(s) to the local config or remove those agent keys manually. NEVER delete them automatically. |
| `bash audit-gentle-ai-upstream.sh` reports `base prompt drift: yes` | Review `Drift summary:` first, then inspect the upstream delta before updating approved upstream snapshots/state. |
| `bash audit-gentle-ai-upstream.sh` reports profile/base invariant mismatch | STOP and review before recommending `sync`; the overlay assumptions may be stale even if prompt drift looks small. |
| Audit is complete and repo/runtime state is still unchanged | Return the decision summary (`new upstream`, `adopt`, `discard`, `why`, `repo sync`, refresh recommendation) and STOP for approval before any mutation. |
| Adopted change affects topology, presets, or materialization for the maintained runtime target | STOP, summarize the impact, and explicitly recommend a full reinstall before re-applying the overlay. |
| Upstream only added a new agent/platform outside the maintained runtime target/materialized state | Note it in the audit, but do NOT recommend reinstall for that reason alone. |
| Upstream added new skills or workflow behavior without maintained-target topology drift | STOP, summarize the impact, and recommend `gentle-ai sync` after the repo is updated. |
| Upstream tries to reintroduce chained/stacked PR governance into the orchestrator | Recommend discard, preserve the owned depuration, and STOP before changing intent/policy. |
| Fresh-context review finds inconsistency after maintenance edits | Fix it before closing, or STOP and surface the inconsistency if the correct resolution is unclear. |

## Execution Steps

1. Confirm the workflow order: binary update -> upstream `git pull` -> maintainer audit from `gentle-ai-custom` -> approval gate -> repo sync if a new boundary was accepted -> `gentle-ai sync` or reinstall -> overlay apply.
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
    - topology/preset/materialization changes for the maintained runtime target
    - upstream additions outside the maintained runtime target/materialized state
    - recommended upstream adoption path: `gentle-ai sync` vs full reinstall
11. Produce a short plain-language drift summary before the full diff review.
12. Convert that audit into a concise decision summary with these exact buckets:
    - what is new upstream
    - recommend adopt
    - recommend discard
    - why
    - repo sync requirement: whether approval of a new boundary requires `bash sync-gentle-ai-upstream-assets.sh`
    - recommended upstream refresh path: `gentle-ai sync` vs full reinstall
13. STOP for explicit user approval before editing repo files, advancing `upstream-state.json`, running `sync-gentle-ai-upstream-assets`, or refreshing the local runtime.
14. After approval, update this repo if needed: owned assets, approved upstream snapshots, policy, docs, state, and logs together. If a new upstream boundary was accepted, run `bash sync-gentle-ai-upstream-assets.sh` as the repo sync step.
15. Only after repo sync, run the recommended upstream refresh path: `gentle-ai sync` or full reinstall.
16. Run the apply entrypoint: `bash apply-gentle-ai-custom.sh opencode` minimum (or `all`). Capture the `Summary:` block.
17. Verify post-state on disk:
   - pruned skills are absent only in the selected registered runtime targets
   - owned runtime targets declared in `managed-assets.json` exist on disk
   - `agent_overrides` resolve to the configured `model` / `variant`
   - `agent.gentle-orchestrator.prompt` points to the owned runtime prompt file
   - each managed named profile orchestrator points to the same owned runtime prompt file
   - each unsuffixed and managed suffixed SDD phase agent points to the owned runtime prompt file for that phase
   - default/named profile assignments still match the local config
   - `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` plus metadata remain consistent with `upstream-state.json`
18. Run one fresh-context reviewer/subagent pass against the changed maintainer artifacts and the final maintenance diff to confirm the workflow, docs, and summary outputs stay consistent.
19. Return a closing summary that states what was actually adopted, what was actually discarded, and why.
20. If the work closed an eligible maintenance event, record one consolidated entry in `overlay/gentle-ai/logs/update-log.md`.

## Output Contract

Return:
- update type the user performed (brew / sync / reinstall)
- files changed in the overlay (policy, assets, scripts, docs, log)
- topology drift detected and how it was resolved
- brief drift summary in plain language
- pre-mutation decision summary with `new upstream`, `adopt`, `discard`, `why`, `repo sync requirement`, and refresh recommendation
- recommended upstream adoption path (`gentle-ai sync` or full reinstall) and why
- whether a newly accepted upstream boundary requires `bash sync-gentle-ai-upstream-assets.sh`
- apply summary counts (owned asset writes, prompt ref updates, topology warnings, profile counts)
- post-state verification result (what passed / failed)
- fresh-context reviewer/subagent result (what it checked and whether it found inconsistencies)
- whether keep/prune or repo-owned orchestrator behavior changed
- what upstream range was audited
- whether user approval was required and how it affected the result
- closing summary of what was actually adopted vs discarded and why

## References

- `../../../../overlay/gentle-ai/maintenance.md`
- `../../../../overlay/gentle-ai/policy/maintenance-intent.md`
- `../../../../overlay/gentle-ai/policy/managed-assets.json`
- `../../../../overlay/gentle-ai/state/upstream-state.json`
- `../../../../AGENTS.md`
