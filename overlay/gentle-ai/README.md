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

## Current model

- `policy/managed-assets.json` is the canonical map between approved upstream inputs and repo-owned runtime outputs.
- Audit is an activation gate. It runs synchronously after the maintainer loads minimal context and before broad exploration or delegation.
- Audit and repo sync require a clean upstream checkout at the exact release tag matching the installed binary. They never fetch tags or change revisions automatically.
- Upstream resolution reads only `upstream_repo_path` from the local config. This lets audit/sync inspect a new topology before per-machine profile assignments are migrated; apply still validates the complete runtime config before writing.
- `last_maintained_commit` defines the audit range after release parity succeeds; it is not parity proof.
- `sync-gentle-ai-upstream-assets` refreshes approved snapshots and the maintained boundary. It does not refresh runtime.
- `apply-gentle-ai-custom` preserves Gentle AI-materialized SDD prompts, contracts, commands, and plugins. Snapshots under `assets/upstream/` are audit inputs, not custom install sources.
- Audit follows the composed OpenCode orchestrator call chain and watches the upstream remote-authorization contract plus its maintained projection paths.
- Runtime refresh, repo snapshot sync, and overlay apply are separate authorization boundaries.

## Current adoption status

- Approved snapshots and `state/upstream-state.json` now match verified v2.9.0 at `be49554794917ae92a6dc9dbfa2eb3db5cf70084`.
- The separately authorized repository sync reported 28 managed range changes, 3 new files, 25 changed files, 46 unchanged files, 0 deletions, and 18 synced directories.
- The post-sync canonical audit passed at exact v2.9.0 / `be49554794917ae92a6dc9dbfa2eb3db5cf70084`, with no managed or watched-unmanaged drift and all structural, profile, and topology invariants `ok`.
- The installed OpenCode runtime remains at its last verified v2.5.0 state. No v2.9 runtime refresh, reinstall, or overlay apply was authorized or run.
- The recommended runtime refresh is `gentle-ai sync`, followed by a separately authorized overlay apply. RDD is maintainer-approved as active and functional; this machine was observed global `on` with clone-local unset, but effective mode must always be queried rather than inferred from repository policy.

## Historical v2.5.0 runtime evidence

- Repository policy now models 11 SDD phases, inserting `sdd-research` after `sdd-explore`.
- Managed-assets policy covers the upstream `sdd-research` phase skill. Existing shared-skill and command-directory entries cover related shared and command files.
- Audit invariants validate semantic managed-profile detection and deny-by-default `permission.task` delegation. Deprecated agent-local `tools.task` is not required.
- The v2.5.0 repository snapshot sync, user-run full SDD component installation, and later separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully. This remains the last verified installed-runtime state; the repository snapshots and maintained boundary have since advanced to v2.9.0.
- Apply updated config; pruned 2 skills; changed and verified 1 owned asset; reported 0 topology warnings; managed 4 profiles; and reported agents created 3, updated 4, unchanged 49, with 0 unmanaged profiles.
- Final verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. The sanitized repo-owned `branch-pr`, all 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, configured overrides, and 4 managed profiles match canonical or local config. Conflicting upstream `branch-pr`, `issue-creation`, and Hermes are absent.
- All profiles have the ordered 11-phase topology and the 3 named `sdd-research` actors. Review actors remain global with no suffixed review actors, and Judgment Day differs only by the retrospective patch. `general` is `openai/gpt-5.6-terra` / `high`; `explore` is `openai/gpt-5.6-sol` / `medium`.
- At the v2.5.0 runtime verification point, managed background intent and effective policy were `on`; the launcher was `~/.gentle-ai/bin/opencode`, and Desktop/plain processes remained foreground. `~/.gentle-ai/state.json` was observed still reporting `installed_binary_version` as `2.4.0`; no root cause was asserted and external state was not edited.
- Historical RDD context: v2.2.2 had a confirmed correction-handoff bug, v2.5.0 supplied source and test evidence for the pre-edit plan, and a temporary local patch was assessed and deferred. The former mandatory canary is now an optional regression diagnostic, not an activation gate; see `maintenance.md` for the preserved analysis and current contracts.

## Preserved local boundaries

- `assets/owned/opencode/AGENTS.md` is the source of truth for global credential-file safety. It applies to the primary agent and every delegated subagent or tool, favors non-secret alternatives, and requires a stop plus explicit user direction before any exceptional access.
- Policy-pruned skills without a repo-owned same-name replacement are denied through OpenCode `permission.skill` rules, globally and for configured agents. This blocks external copies without deleting or changing another runtime's files; the sanitized repo-owned `branch-pr` remains available.
- `shared/skills/chained-pr/` remains the sanitized runtime source. It adopts upstream anti-code-golf and bounded single-pass slicing rules while preserving explicit `owner/repo` scoping and plan-only automatic activation.
- `branch-pr` remains pruned upstream and replaced by the repo-owned sanitized copy. Explicit `/pr-create` and `/pr-regenerate` render title/body for audit visibility and then mutate through GitHub CLI without a content-approval pause; push authorization and unresolved required metadata, identity, or chain context remain blocking.
- `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` remain retained snapshots, not forced runtime selections.
- Named profiles may create suffixed Judgment Day actors only for explicit assignments. Review actors remain global.
- The Judgment Day retrospective handoff remains a declared surgical patch and still requires exact compatible anchors.
- The `sdd-research` output-only guard is a second surgical patch. It prevents contradictory inline launcher text from assigning persistence to the research executor and fails closed on incompatible anchors.

## Maintainer report contract

- `Scope`: `Managed` or `Unmanaged`
- `Impact`: `Behavioral`, `Runtime`, or `Housekeeping`
- `Decision`: `Adquirir`, `Sanitizar`, or `Ignorar`
- Columns: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`
- `Follow-up` is optional and stays empty when no action is needed.
- One row equals one atomic decision. Separate rows are required when scope, impact, or decision differ, and runtime sanitization is never hidden inside an acquisition row.
- Every required summary adjustment must appear in a row; reconcile the summary, rows, and final recommendation before presenting them.

For usage, start with the root `README.md`. For upstream maintenance, follow `maintenance.md`.
