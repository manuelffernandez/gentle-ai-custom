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
- Runtime refresh, repo snapshot sync, and overlay apply are separate authorization boundaries.

## v2.5.0 adaptation status

- Repository policy now models 11 SDD phases, inserting `sdd-research` after `sdd-explore`.
- Managed-assets policy covers the upstream `sdd-research` phase skill. Existing shared-skill and command-directory entries cover related shared and command files.
- Audit invariants validate semantic managed-profile detection and deny-by-default `permission.task` delegation. Deprecated agent-local `tools.task` is not required.
- The separately authorized repository snapshot sync, user-run full SDD component installation, and later separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully. Approved snapshots, `state/upstream-state.json`, and installed runtime now match v2.5.0.
- Apply updated config; pruned 2 skills; changed and verified 1 owned asset; reported 0 topology warnings; managed 4 profiles; and reported agents created 3, updated 4, unchanged 49, with 0 unmanaged profiles.
- Final verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. The sanitized repo-owned `branch-pr`, all 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, configured overrides, and 4 managed profiles match canonical or local config. Conflicting upstream `branch-pr`, `issue-creation`, and Hermes are absent.
- All profiles have the ordered 11-phase topology and the 3 named `sdd-research` actors. Review actors remain global with no suffixed review actors, and Judgment Day differs only by the retrospective patch. `general` is `openai/gpt-5.6-terra` / `high`; `explore` is `openai/gpt-5.6-sol` / `medium`.
- Managed background intent and effective policy are `on`; the launcher is `~/.gentle-ai/bin/opencode`. Current Desktop/plain processes remain foreground. Restart OpenCode once for config, and use the explicit launcher for background-enabled CLI sessions. `~/.gentle-ai/state.json` was observed still reporting `installed_binary_version` as `2.4.0` despite verified v2.5.0 binary/tag/runtime; no root cause is asserted and external state was not edited.
- RDD remains blocked. v2.5.0 supplies source and test evidence for a pre-edit correction-plan handoff, but global enable must not be recommended until the correction-required canary passes.

## Preserved local boundaries

- Policy-pruned skills without a repo-owned same-name replacement are denied through OpenCode `permission.skill` rules, globally and for configured agents. This blocks external copies without deleting or changing another runtime's files; the sanitized repo-owned `branch-pr` remains available.
- `shared/skills/chained-pr/` remains the sanitized runtime source. It adopts upstream anti-code-golf and bounded single-pass slicing rules while preserving explicit `owner/repo` scoping and plan-only automatic activation.
- `branch-pr` remains pruned upstream and replaced by the repo-owned sanitized copy.
- `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` remain retained snapshots, not forced runtime selections.
- Named profiles may create suffixed Judgment Day actors only for explicit assignments. Review actors remain global.
- The Judgment Day retrospective handoff remains the only declared surgical patch and still requires exact compatible anchors.

## Maintainer report contract

- `Scope`: `Managed` or `Unmanaged`
- `Impact`: `Behavioral`, `Runtime`, or `Housekeeping`
- `Decision`: `Adquirir`, `Sanitizar`, or `Ignorar`
- Columns: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`
- `Follow-up` is optional and stays empty when no action is needed.

For usage, start with the root `README.md`. For upstream maintenance, follow `maintenance.md`.
