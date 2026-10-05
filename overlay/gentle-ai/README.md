# Gentle AI overlay assets

This directory contains the operational assets for the Gentle AI overlay.

## Read first

- Main repository guide: `../../README.md`
- Human maintenance guide: `maintenance.md`
- Maintainer skill: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Key paths

- `maintenance.md` — centralized human maintenance guide
- `assets/upstream/` — approved upstream copies for audit and diff review
- `assets/owned/` — true repo-owned assets installed by `apply-gentle-ai-custom`
- `patches/` — surgical local deltas applied to compatible upstream runtime files
- `policy/` — semantic intent and machine-readable overlay policy
- `state/` — last maintained upstream boundary
- `logs/` — closed maintenance-event ledger
- `../../shared/skills/package-security/` — canonical package-security skill and templates

## Current model

- `policy/managed-assets.json` is the canonical map between approved upstream inputs and repo-owned runtime outputs.
- Audit is an activation gate. It runs synchronously after the maintainer loads minimal context and before broad exploration or delegation.
- Audit and repo sync default to the installed binary's exact release tag. An explicit `--release v<V>` can select another clean exact tag without changing the binary or runtime. Neither command fetches tags or changes revisions automatically.
- An explicit-release audit is preview only, not adoption or repo-sync approval. Repo snapshot sync remains separately approval-gated and must use the same selected release.
- Upstream resolution reads only `upstream_repo_path` from the local config. This lets audit/sync inspect a new topology before per-machine profile assignments are migrated; apply still validates the complete runtime config before writing.
- `last_maintained_commit` defines the audit range after release parity succeeds; it is not parity proof.
- `sync-gentle-ai-upstream-assets` refreshes approved snapshots and the maintained boundary. It does not refresh runtime.
- `apply-gentle-ai-custom` preserves Gentle AI-materialized SDD prompts, contracts, commands, and plugins. Snapshots under `assets/upstream/` are audit inputs, not custom install sources.
- `assets/owned/opencode/AGENTS.md` remains a full-file repo-owned install source and the single source of truth for global agent rules. Its persona and OpenCode full Engram sections must match the approved source snapshots semantically; focused tests allow only the six declared overlay-only sections.
- Audit follows the composed OpenCode orchestrator call chain and watches the upstream remote-authorization contract plus its maintained projection paths.
- Apply removes only the exact recognized upstream or previously relaxed remote-authorization section from maintained inline prompts, referenced prompt files, and installed shared SDD prompts. Malformed, duplicate, or unknown sections and unsafe paths fail before writes; unrelated content and modes are preserved. The upstream audit snapshot and structural projection audit remain unchanged.
- The owned global `AGENTS.md` intentionally omits that section. Credential-file safety and unrelated runtime/tool permission guards are unchanged.
- Runtime refresh, repo snapshot sync, and overlay apply are separate authorization boundaries.

## Current adoption status

- The current repository snapshot boundary is v3.7.0 at `6dee8f833aec9e46015759c5065a9035795d9af1`.
- The authorized snapshot sync reported 36 managed changes: 5 new files, 31 changed, 43 unchanged, 0 deleted, and 19 directories synced; `upstream-state.json` changed.
- The canonical post-sync audit passed at that exact v3.7.0 commit with no managed or watched-unmanaged drift and all structural checks `ok`.
- The installed Gentle AI binary and runtime remain v2.9.1, with OpenCode 1.18.32. No runtime refresh or overlay apply was authorized or performed for v3.7.0. Snapshot adoption does not verify runtime behavior.
- All five `plugins-v2` files are audit/sync-only. OpenCode 1.18.32 uses the V1 plugin path; this repository does not claim V2 runtime materialization support.
- The v3.7.0 `sdd-research` skill and command provide output-only collection with orchestrator-owned persistence. Its former local precedence patch is retired; the installed v2.9.1 runtime has not been verified against this source contract.
- The ODD commit authorization patch is prepared in the repo but is not installed in the current runtime. The current v2.9.1 `apply` is intentionally incompatible with its v3.7.0 anchor. Treat the safeguard as inactive until a separately authorized runtime refresh and apply complete and are verified; until then, commits require explicit user authorization.
- The repo consciously accepts v3.7.0 explicit partial SDD archive as local policy, not as confirmed upstream intent or proof of implementation completion. Future-audit checks are in [`maintenance.md`](maintenance.md#partial-sdd-archive-watchpoint).
- The remote-authorization removal is implemented in the repository only. The installed runtime remains unchanged until a separately authorized overlay apply; this work does not claim the runtime is already clear of the rule.

### Historical v2.9.1 verification

- The separately authorized repository sync reported 0 new files, 0 changed files, 74 unchanged files, 0 deletions, and 18 synced directories; only `state/upstream-state.json` changed.
- The post-sync canonical audit passed at exact v2.9.1 / `9ec0cf443622f20fe511815ff5a83088c7467bff`, with no managed or watched-unmanaged drift and all profile, injection, remote-authorization, and JD/review topology invariants `ok`.
- Compatible OpenCode review consent and preserved-prompt migration are acquired upstream-first. Claude Code transcript-derived dispatch authority is outside the maintained OpenCode target. No policy, patch, topology, profile, preset, or materialization changes were required.
- After the repository-only closure, the user reports manually running both `gentle-ai sync` and `bash apply-gentle-ai-custom.sh opencode`. Their original summaries and exit output were not captured, so no new apply counts are asserted.
- Independent read-only post-state verification passed 602 selected assertions over runtime assignments, named profiles, prompt parity, task scoping, skill permissions, absence of suffixed review actors, and obsolete prompt references. The installed orchestrator includes the compatible `gentle-ai.review-integration.consent/v3` semantics.
- All 7 repo-owned skill sources/assets, all 5 wrapper bodies, and the owned `AGENTS.md` matched their installed targets. `issue-creation` and `hermes-ephemeral-delegation` were absent, `branch-pr` was the repo-owned replacement, and all 3 installed runtime patch targets matched approved upstream snapshots after reversing only their declared local patches.
- RDD remains globally `on`, but the user explicitly excluded this v2.9.1 maintenance candidate from RDD. Effective mode must always be queried rather than inferred from repository policy.
- In the earlier v2.9.1 state, the owned global `AGENTS.md` and installed skill sources matched their targets and independent read-only post-state verification passed. This is historical v2.9.1 evidence, not verification of v3.7.0 runtime materialization.

## Historical v2.5.0 runtime evidence

- Repository policy now models 11 SDD phases, inserting `sdd-research` after `sdd-explore`.
- Managed-assets policy covers the upstream `sdd-research` phase skill. Existing shared-skill and command-directory entries cover related shared and command files.
- Audit invariants validate semantic managed-profile detection and deny-by-default `permission.task` delegation. Deprecated agent-local `tools.task` is not required.
- The v2.5.0 repository snapshot sync, user-run full SDD component installation, and later separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully. This evidence is historical; the maintained repository boundary and independently verified runtime have since advanced to v2.9.1.
- Apply updated config; pruned 2 skills; changed and verified 1 owned asset; reported 0 topology warnings; managed 4 profiles; and reported agents created 3, updated 4, unchanged 49, with 0 unmanaged profiles.
- Final verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. The sanitized repo-owned `branch-pr`, all 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, configured overrides, and 4 managed profiles match canonical or local config. Conflicting upstream `branch-pr`, `issue-creation`, and Hermes are absent.
- All profiles have the ordered 11-phase topology and the 3 named `sdd-research` actors. Review actors remain global with no suffixed review actors, and Judgment Day differs only by the retrospective patch. `general` is `openai/gpt-5.6-terra` / `high`; `explore` is `openai/gpt-5.6-sol` / `medium`.
- At the v2.5.0 runtime verification point, managed background intent and effective policy were `on`; the launcher was `~/.gentle-ai/bin/opencode`, and Desktop/plain processes remained foreground. `~/.gentle-ai/state.json` was observed still reporting `installed_binary_version` as `2.4.0`; no root cause was asserted and external state was not edited.
- Historical RDD context: v2.2.2 had a confirmed correction-handoff bug, v2.5.0 supplied source and test evidence for the pre-edit plan, and a temporary local patch was assessed and deferred. The former mandatory canary is now an optional regression diagnostic, not an activation gate; see `maintenance.md` for the preserved analysis and current contracts.

## Preserved local boundaries

- `assets/owned/opencode/AGENTS.md` is the source of truth for global credential-file safety. It applies to the primary agent and every delegated subagent or tool, favors non-secret alternatives, and requires a stop plus explicit user direction before any exceptional access.
- `shared/skills/package-security/SKILL.md` prohibits credential-file reads. Its templates use blocked scripts as npm's conservative baseline; reviewed npm approvals require supported strict enforcement, while pnpm `allowBuilds` / `strictDepBuilds` require exact CLI feature checks. Editing this canonical source does not update the installed global copy; `apply-gentle-ai-custom` is the propagation step.
- Policy-pruned skills without a repo-owned same-name replacement are denied through OpenCode `permission.skill` rules, globally and for configured agents. This blocks external copies without deleting or changing another runtime's files; the sanitized repo-owned `branch-pr` remains available.
- `shared/skills/chained-pr/` remains the sanitized runtime source. It adopts upstream anti-code-golf and bounded single-pass slicing rules while preserving explicit `owner/repo` scoping and plan-only automatic activation.
- `branch-pr` remains pruned upstream and replaced by the repo-owned sanitized copy. Explicit `/pr-create` and `/pr-regenerate` render title/body for audit visibility and then mutate through GitHub CLI without a content-approval pause; push authorization and unresolved required metadata, identity, or chain context remain blocking.
- `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` remain retained snapshots, not forced runtime selections.
- Named profiles may create suffixed Judgment Day actors only for explicit assignments. Review actors remain global.
- The Judgment Day retrospective handoff remains a declared surgical patch and still requires exact compatible anchors.
- The prepared `work-unit-commits` authorization patch is intended to replace the v3.7.0 automatic ODD task-close commit instruction. It is not active in the current runtime; its exact anchor is incompatible with the current v2.9.1 `apply`. Do not treat the safeguard as active until separately authorized refresh/apply and verification. Without explicit commit authorization, do not commit.
- The `sdd-research` v3.7.0 source skill and OpenCode command define output-only collection with orchestrator-owned persistence. Its local precedence patch is retired and is not an install source.

## Maintainer report contract

- `Scope`: `Managed` or `Unmanaged`
- `Impact`: `Behavioral`, `Runtime`, or `Housekeeping`
- `Decision`: `Adquirir`, `Sanitizar`, or `Ignorar`
- Columns: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`
- `Follow-up` is optional and stays empty when no action is needed.
- One row equals one atomic decision. Separate rows are required when scope, impact, or decision differ, and runtime sanitization is never hidden inside an acquisition row.
- Every required summary adjustment must appear in a row; reconcile the summary, rows, and final recommendation before presenting them.

For usage, start with the root `README.md`. For upstream maintenance, follow `maintenance.md`.
