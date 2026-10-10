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
- Follow this order for maintainer updates: update `gentle-ai` binary -> update tags/checkout in the upstream clone -> immediate audit from `gentle-ai-custom` -> decision/approval gate -> separately approved repo sync via `sync-gentle-ai-upstream-assets` when a new boundary was accepted -> human-owned runtime refresh (`gentle-ai sync` or reinstall) -> separately authorized `apply-gentle-ai-custom`.
- Immediately after loading this skill and the minimal context needed to invoke it, run the canonical audit synchronously. Run nothing that performs broad repository exploration, repository-reading delegation, or drift-analysis delegation first. By default use `bash audit-gentle-ai-upstream.sh`; when the user has selected a different release from the installed binary, use `bash audit-gentle-ai-upstream.sh --release v<V>`. If selected-release verification fails, STOP immediately and delegate nothing.
- Treat the user's account of what was updated as diagnostic context only. Before every audit, run the canonical audit entrypoint and trust its release parity preflight, not the answer.
- Audit and sync default to the exact `refs/tags/v<V>` for installed `gentle-ai --version`. An explicit `--release v<V>` may select a different release, but that exact tag must exist, peel to a commit, and equal `HEAD` in a completely clean upstream checkout before drift analysis, copying, or writes. This selector does not change the binary or runtime. Never use nearest-tag, version ordering, branch ancestry, or `git describe` as proof. Never fetch or check out automatically.
- Audit/sync upstream resolution reads only `upstream_repo_path` from the local config. Do not require machine-local profile assignments to match a newly adopted phase schema before the repository audit; apply retains full runtime-config validation before writes.
- Treat `last_maintained_commit` only as the drift-range boundary after parity succeeds. It is never release parity proof.
- `audit-gentle-ai-upstream` discovers drift from `git diff --name-status --find-renames <last_maintained_commit>..<verified-release-commit>`, filtered through `overlay/gentle-ai/policy/managed-assets.json`, while still checking structural invariants from upstream integration code. An audit selected with `--release` is preview evidence only; it does not adopt the release or authorize repo snapshot sync.
- `sync-gentle-ai-upstream-assets` refreshes `overlay/gentle-ai/assets/upstream/...` plus the maintained boundary only after the drift has been reviewed and accepted and repo sync is separately authorized. Pass the same `--release v<V>` selector to sync only after that approval.
- `apply-gentle-ai-custom` installs true repo-owned assets and portable skills, renders wrappers, prunes rejected skills, applies overrides, and reconciles profiles while preserving upstream-materialized SDD/review runtime assets.
- Codex is a separate apply target: it installs the allowlisted shared portable skills, leaves unrelated upstream skills and configuration untouched, inserts only the marker-bounded local instruction block, and applies declared exact skill deltas. It does not prune Codex skills, render commands, or run OpenCode permission/model/profile reconciliation. The current upstream audit does not verify installed Codex runtime materialization.
- For OpenCode, derive external-skill denies from policy-pruned names minus repo-owned same-name skills intentionally installed by the managed target. Apply them globally and to configured agents. Always persist the canonical permission order when effective denies exist: OpenCode uses the last matching rule, and parsed JSON loses the original order needed to detect an exact deny before a matching wildcard allow.
- Before any runtime write, apply validates only its own prerequisites: ownership boundaries, base prompts consumed by configured profiles, safe obsolete-reference migration, and compatible declared runtime-patch anchors.
- Apply preflights maintained inline prompts, referenced files, and installed shared SDD prompt files before the first write; it removes only the exact recognized upstream or previously relaxed remote-authorization section. Missing referenced files, duplicate/malformed/unknown sections, unsafe paths, and remote-bearing unmanaged named profiles fail closed. Keep the upstream snapshot audit-only and its structural projection audit intact.
- The repo-owned global `AGENTS.md` intentionally omits the remote-operation authorization section. This removal does not alter credential-file safety or unrelated runtime/tool permission guards.
- Apply must not inspect or certify upstream plugin, contract, prompt, or review semantics. Opaque or absent upstream plugins do not block unrelated custom operations.
- Owned-state cleanup must track every custom-installed output while protecting upstream runtime targets, retained upstream skill trees, and surgical patch targets without claiming whole-file ownership.
- Read semantic intent before making maintenance decisions.
- For approved upstream assets, replacement-first means upstream Gentle AI materializes the finished runtime artifact. Keep raw snapshots for audit only and apply explicit surgical deltas from `maintenance-intent.md` against a compatible installed artifact.
- When editing Markdown primarily consumed by agents or LLMs, follow the runtime Agent-Readable Markdown Standard and keep only local reminders here; do not restate the full rule set.
- Before any repo mutation, produce a concise decision summary that states what is new upstream, the `Scope` / `Impact` classification, the recommended decision (`Adquirir`, `Sanitizar`, or `Ignorar`), the rationale, whether repo sync is required, and the recommended upstream refresh path.
- After that summary, STOP for explicit user approval before mutating this repo, advancing the maintained upstream boundary, or syncing approved upstream assets.
- Runtime refresh is human-owned by default. Recommend the correct path and print exact instructions, but never run `gentle-ai sync` or a reinstall unless the human explicitly requests that specific runtime action. Approval of audit findings, upstream adoption, repo snapshot sync, or overlay edits is not runtime-refresh authorization.
- Overlay apply is separately authorized. Never infer apply permission from approval of findings, repo sync, runtime refresh, or other maintenance edits.
- Preserve the local keep/prune baseline and upstream/runtime ownership boundaries defined in `maintenance-intent.md`.
- Treat v3.7.0 explicit partial SDD archive as a conscious local acceptance, not confirmed upstream intent or proof of completed implementation; apply the future-audit watchpoint in `overlay/gentle-ai/maintenance.md` and do not imply an RDD escalation was retroactively approved.
- Distinguish preview, repository snapshot adoption, and runtime state. An explicit-release audit alone is preview only; after an authorized snapshot sync, report the new maintained boundary without implying a runtime refresh or apply.
- OpenCode V2 plugin snapshots are audit/sync inputs only, not custom apply install sources. OpenCode 1.18.32 uses the V1 plugin path; V2 materialization requires custom apply schema work and must not be claimed as supported.
- Upstream ODD work-unit commit wording does not override local `no-auto-commit`; never create a commit without explicit user authorization. A prepared or declared `work-unit-commits` patch is not active runtime protection until the relevant target receives a separately authorized apply and verification. Current runtime versions are unverified. If a target's installed skill does not match the exact patch anchor, its apply must fail closed; do not claim the safeguard is active.
- Treat upstream PR strategy, chained PR, review-budget, and review-workload governance as adopt-by-default under the current upstream-first intent; sanitize only when a concrete conflict with explicit local customizations is confirmed.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes, but write to the log only for eligible closed maintenance events under `AGENTS.md` rule 4.
- Do not change intent, keep/prune, runtime ownership, or local patch behavior for new upstream changes without explicit user approval.
- After approved maintenance work, return a closing summary of what was actually `Adquirir`, `Sanitizar`, or `Ignorar`, and why.
- After maintenance edits, run one fresh-context reviewer/subagent consistency pass over the changed maintainer artifacts before closing.
- The versioned policy MUST NOT carry per-profile orchestrator/phase model+variant choices. Those live in the per-machine local config at `~/.config/gentle-ai-custom/opencode-local-config.json` under `default_profile` and `profiles`.
- The same local config may manage the three global Judgment Day actors, all six global review actors (including `review-validator`), and optional named-profile JD actors. Review actors are always global; never invent suffixed review topology.
- Omitted JD/review slots preserve runtime assignment. Only explicit `{"inherit": true}` removes both `model` and `variant` for OpenCode Task inheritance.
- Treat RDD as maintainer-approved, active, and functional based on the owner's cross-repository operational validation. Effective mode remains user- and machine-owned: check it with `gentle-ai review mode status --cwd .` and never infer it from repository policy or upstream control. The correction-required canary is an optional regression diagnostic, not an activation gate. Preserve the kill switch, candidate-bound receipts, bounded correction lifecycle, and ordinary delivery policy.
- Treat `sdd-research` as the phase immediately after `sdd-explore`. Profile audits must validate all 11 phases, semantic managed-profile detection, and deny-by-default `permission.task` delegation without requiring deprecated agent-local `tools.task`.

## Audit report contract

Use this contract in the user-facing maintainer report:

| Axis | Values | Meaning |
|---|---|---|
| Scope | `Managed` / `Unmanaged` | `Managed` is inside the overlay/runtime surface this repo maintains. `Unmanaged` is outside that surface or not relevant to it. |
| Impact | `Behavioral` / `Runtime` / `Housekeeping` | `Behavioral` changes agent or orchestrator behavior. `Runtime` changes maintained-target wiring, install, config, or materialization. `Housekeeping` covers docs, unrelated agents, or internal fixes with no maintained-target effect. |

Decision labels:

| Decision | Use when |
|---|---|
| `Adquirir` | The upstream change should be incorporated into the overlay/runtime scope. |
| `Sanitizar` | The upstream change matters, but it must be adapted so local `maintenance-intent` still wins. |
| `Ignorar` | The change was evaluated but is not relevant to the maintained target/scope. |

Report rows use this shape:

| Upstream change | Files | Scope | Impact | Decision | Why | Follow-up |
|---|---|---|---|---|---|---|

- `Upstream change` is a concise human summary of the upstream delta and why it matters; do not use a file-path list as the main content.
- `Follow-up` is optional; leave it empty when no further action is needed.
- One row represents one atomic decision. `Decision` applies exactly to the change named in `Upstream change`.
- Put distinct local remediation in `Follow-up` or a separate row. Runtime sanitization always receives its own row.
- Map every required adjustment named in the summary to an explicit row.
- Never merge rows when `Scope`, `Impact`, or `Decision` differ.
- Reconcile the summary, report rows, and final recommendation before presenting them.

Do not use `descartar` as the main decision label for report rows.

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
| Only the binary was updated | By default, the audit requires its exact matching upstream tag. For a separately selected newer release, pass `--release v<V>` and still require that exact clean tag at `HEAD`. |
| Only the upstream clone was updated | By default, the audit requires parity with the installed binary. To inspect a different exact release while preserving the binary, explicitly pass `--release v<V>`. |
| Binary and clone were updated, but the branch is ahead of the release | STOP on parity failure. Check out exact `v<V>` manually; do not audit unreleased commits. |
| Binary and clone are aligned at a clean exact release tag | Continue to drift analysis. A prerelease is valid only when its exact `v<V>` tag and commit match. |
| Neither was updated | Run the same preflight. Continue only if the existing binary and clone still match exactly. |
| Exact tag is missing or stale | STOP. Fetch tags manually, verify the exact tag, and check it out manually; do not fetch or switch revisions for the user. |
| Upstream checkout has staged, unstaged, untracked, or submodule dirt | STOP. Require a completely clean checkout before audit or sync. |
| The user has not run `gentle-ai sync` or reinstalled yet | Treat this as the normal pre-audit state. Audit first; recommend runtime refresh only after audit and repo sync when required. Do not execute it without a specific request. |
| User updated upstream (`git pull`) but has NOT run `gentle-ai sync` yet | Run `bash audit-gentle-ai-upstream.sh` immediately after skill load and before broad reads or delegation. Use it to detect managed-asset drift + profile invariants before recommending repo sync + runtime refresh. |
| User just ran `gentle-ai sync` | Recommend `bash apply-gentle-ai-custom.sh opencode` (minimum) or `all`, but run it only with separate explicit authorization; if runtime preflight fails, explain that reinstall/sync must restore the missing upstream surface first. |
| User just ran TUI reinstall | Audit topology BEFORE re-applying. |
| Script printed `topology: ...` warnings | Investigate each warning. STOP and ask before mutating policy/intent. |
| Script printed `WARNING - unmanaged SDD profiles left untouched` | Ask whether to add the profile(s) to the local config or remove those agent keys manually. NEVER delete them automatically. |
| `bash audit-gentle-ai-upstream.sh` reports `base prompt drift: yes` | Review `Drift summary:` first, then inspect the upstream delta before updating approved upstream snapshots/state. |
| `bash audit-gentle-ai-upstream.sh` reports profile/base invariant mismatch | STOP and review before recommending `sync`; the overlay assumptions may be stale even if prompt drift looks small. |
| Audit is complete and repo/runtime state is still unchanged | Return the decision summary (`new upstream`, `Scope`, `Impact`, `Decision`, `why`, `repo sync`, refresh recommendation) and STOP for approval before any mutation. |
| Adopted change affects topology, presets, or materialization for the maintained runtime target | STOP, summarize the impact, and explicitly recommend a full reinstall before re-applying the overlay. |
| Accepted Gentle AI v2.5.0 adoption | Because `sdd-research` changes maintained OpenCode profile topology, recommend a full SDD component reinstall after separately authorized repo snapshot sync. Keep runtime reinstall and overlay apply separately authorized. |
| Upstream only added a new agent/platform outside the maintained runtime target/materialized state | Note it in the audit, but do NOT recommend reinstall for that reason alone. |
| Upstream added new skills or workflow behavior without maintained-target topology drift | STOP, summarize the impact, and recommend `gentle-ai sync` after the repo is updated. |
| Upstream changes PR strategy, chained/stacked PR governance, review budget, or review workload behavior | Compare against explicit local customizations in `maintenance-intent.md`; recommend `Sanitizar` only for concrete conflicts, otherwise recommend `Adquirir`. |
| Reviewing RDD policy or effective state | Treat RDD as accepted and functional by owner decision, then query `gentle-ai review mode status --cwd .` for the effective machine/clone state. Preserve exact receipt and correction contracts; use the correction-required canary only as an optional regression diagnostic. |
| Fresh-context review finds inconsistency after maintenance edits | Fix it before closing, or STOP and surface the inconsistency if the correct resolution is unclear. |

## Execution Steps

1. Immediately after loading this skill and minimal invocation context, run `bash audit-gentle-ai-upstream.sh` synchronously, or pass `--release v<V>` when the user has selected a release other than the installed binary's. Do not broadly explore the repo or delegate repository reading/drift analysis first.
2. If selected-release verification fails, STOP immediately, delegate nothing, and return the manual recovery instructions. Do not read the drift range.
3. Capture the installed binary version and the selected exact release tag/commit separately. Treat `last_maintained_commit` only as the later audit-range boundary. An explicit-release audit is preview only and is not adoption approval.
4. Ask what was updated only when it helps diagnosis; the answer never replaces the completed preflight.
5. Read `overlay/gentle-ai/policy/maintenance-intent.md`.
6. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
7. Read `overlay/gentle-ai/policy/managed-assets.json`.
8. Read `overlay/gentle-ai/state/upstream-state.json`.
9. Read `overlay/gentle-ai/maintenance.md`.
10. Review the upstream range from `last_maintained_commit` to the verified release commit.
11. Classify findings into:
    - base prompt drift (`gentle-orchestrator`)
    - profile-generation invariant drift
    - behavior/workflow changes relevant to the overlay
    - topology/preset/materialization changes for the maintained runtime target
    - upstream additions outside the maintained runtime target/materialized state
    - recommended upstream adoption path: `gentle-ai sync` vs full reinstall
12. Produce a short plain-language drift summary before the full diff review.
13. Convert that audit into a concise decision summary with these exact buckets:
     - what is new upstream
     - `Scope` (`Managed` / `Unmanaged`)
     - `Impact` (`Behavioral` / `Runtime` / `Housekeeping`)
     - `Decision` (`Adquirir` / `Sanitizar` / `Ignorar`)
     - why
     - `Follow-up`
     - repo sync requirement: whether approval of a new boundary requires `bash sync-gentle-ai-upstream-assets.sh`
     - recommended upstream refresh path: `gentle-ai sync` vs full reinstall
14. STOP for explicit user approval before editing repo files, advancing `upstream-state.json`, or running `sync-gentle-ai-upstream-assets`.
15. After approval, update this repo if needed: owned assets, approved upstream snapshots, policy, docs, state, and logs together. If a new upstream boundary was accepted and repo sync was separately approved, run `bash sync-gentle-ai-upstream-assets.sh --release v<V>` when the audit used an explicit release selector; otherwise preserve the installed-binary default.
16. Recommend the runtime refresh path and print exact instructions. Run `gentle-ai sync` or a full reinstall only when the human explicitly requests that specific runtime action.
17. Recommend overlay apply after a runtime refresh when needed. Select `opencode`, `codex`, or `all` to match the requested runtime targets; run `bash apply-gentle-ai-custom.sh <target>` only with separate explicit authorization, then capture `Summary:`. Codex apply never refreshes its upstream runtime and does not verify its installed materialization.
18. Verify post-state on disk only for actions actually authorized and executed:
     - pruned skills are absent only in the selected registered runtime targets
     - effective external-skill denies appear after matching wildcard rules in persisted OpenCode config; repo-owned same-name replacements receive no overlay deny
     - owned runtime targets declared in `managed-assets.json` exist on disk
     - `agent_overrides` resolve to the configured `model` / `variant`
     - configured global JD assignments resolve exactly; inherited review actors have neither `model` nor `variant`
     - named profiles create only explicitly configured suffixed JD actors, delegate omitted JD slots globally, and have no suffixed review actors
     - `agent.gentle-orchestrator.prompt` remains the upstream-owned prompt or reference
     - each managed named profile orchestrator reuses the upstream base orchestrator prompt
     - each of the 11 unsuffixed and managed suffixed SDD phase agents, including `sdd-research`, reuses its upstream base phase prompt/reference
     - representative upstream-owned prompt, plugin, and contract files remain byte- and mode-unchanged by custom apply
     - no managed agent references an obsolete `.overlay.md` prompt
     - Judgment Day differs from its compatible installed upstream baseline only by the declared retrospective handoff
     - maintained inline prompt values and reachable prompt files contain no recognized upstream or previously relaxed remote-authorization section; unrelated bytes and file modes are preserved
     - upstream `sdd-research` remains output-only and orchestrator-persisted; the retired local precedence patch is not reapplied
     - default/named profile assignments still match the local config
     - `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md` and related upstream metadata remain consistent with `upstream-state.json`
19. Run one fresh-context consistency pass against the changed maintainer artifacts and the final maintenance diff. Use a reviewer/subagent when delegation is authorized; otherwise perform the pass directly.
20. Return a closing summary that states what was actually `Adquirir`, `Sanitizar`, or `Ignorar`, and why.
21. If the work closed an eligible maintenance event, record one consolidated entry in `overlay/gentle-ai/logs/update-log.md`.

## Output Contract

Return:
- update type the user performed (brew / sync / reinstall)
- files changed in the overlay (policy, assets, scripts, docs, log)
- topology drift detected and how it was resolved
- brief drift summary in plain language
- pre-mutation decision summary with `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`, `repo sync requirement`, and refresh recommendation
- `Follow-up` is optional; leave it empty when no further action is needed.
- recommended upstream adoption path (`gentle-ai sync` or full reinstall) and why
- whether a newly accepted upstream boundary requires `bash sync-gentle-ai-upstream-assets.sh`
- apply summary counts for the selected target (owned asset writes, prompt ref updates, topology warnings, and OpenCode profile counts where applicable)
- post-state verification result (what passed / failed)
- fresh-context reviewer/subagent result (what it checked and whether it found inconsistencies)
- whether keep/prune, runtime ownership, or local patch behavior changed
- what upstream range was audited
- verified installed binary version and selected exact release tag/commit
- release parity failure and manual recovery instructions when audit did not start
- whether user approval was required and how it affected the result
- which separate actions, if any, the human explicitly authorized: repo sync, runtime refresh, and overlay apply
- closing summary of what was actually `Adquirir`, `Sanitizar`, or `Ignorar`, and why

## References

- `../../../../overlay/gentle-ai/maintenance.md`
- `../../../../overlay/gentle-ai/policy/maintenance-intent.md`
- `../../../../overlay/gentle-ai/policy/managed-assets.json`
- `../../../../overlay/gentle-ai/state/upstream-state.json`
- `../../../../AGENTS.md`
