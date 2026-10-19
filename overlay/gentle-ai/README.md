# Gentle AI overlay assets

This directory contains the operational assets for the Gentle AI overlay.

## Read first

- Main repository guide: `../../README.md`
- Human maintenance guide: `maintenance.md`
- Maintainer skill: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Key paths

- `maintenance.md` — centralized human maintenance guide
- `assets/upstream/` — approved upstream copies for audit and diff review
- `assets/owned/` — true repo-owned assets installed by `apply-gentle-ai-custom`; the current full-file instruction asset is OpenCode-only
- `patches/` — target-specific surgical local deltas applied to compatible upstream runtime files
- `policy/` — semantic intent and machine-readable overlay policy
- `state/` — last maintained upstream boundary
- `logs/` — closed maintenance-event ledger
- `../../shared/skills/` — canonical portable repo-owned skills installed for OpenCode and Codex
- `../../shared/commands/` — OpenCode wrapper source; Codex does not receive these commands

## Current model

- `policy/managed-assets.json` is the canonical map between approved upstream inputs and repo-owned runtime outputs.
- Audit is an activation gate. It runs synchronously after the maintainer loads minimal context and before broad exploration or delegation.
- Audit and repo sync default to the installed binary's exact release tag. An explicit `--release v<V>` can select another clean exact tag without changing the binary or runtime. Neither command fetches tags or changes revisions automatically.
- An explicit-release audit is preview only, not adoption or repo-sync approval. Repo snapshot sync remains separately approval-gated and must use the same selected release.
- Upstream resolution reads only `upstream_repo_path` from the local config. This lets audit/sync inspect a new topology before per-machine profile assignments are migrated; apply still validates the complete runtime config before writing.
- `last_maintained_commit` defines the audit range after release parity succeeds; it is not parity proof.
- `sync-gentle-ai-upstream-assets` refreshes approved snapshots and the maintained boundary. It does not refresh runtime.
- `apply-gentle-ai-custom` preserves Gentle AI-materialized ODD prompts, contracts, commands, and plugins. Snapshots under `assets/upstream/` are audit inputs, not custom install sources.
- `assets/owned/opencode/AGENTS.md` remains a full-file repo-owned install source and the single source of truth for global agent rules. Its persona and OpenCode full Engram sections must match the approved source snapshots semantically; focused tests allow only the six declared overlay-only sections.
- Audit follows the composed OpenCode orchestrator call chain and watches the upstream remote-authorization contract plus its maintained projection paths.
- Apply removes only the exact recognized upstream or previously relaxed remote-authorization section from maintained OpenCode inline/file prompts, optional legacy shared SDD prompts, and the prepared Codex `AGENTS.md` patch. Missing optional legacy files are skipped; malformed, duplicate, or unknown sections and unsafe paths fail before writes. Unrelated content and modes are preserved, and upstream audit snapshots and structural projection checks remain unchanged.
- The owned global `AGENTS.md` intentionally omits that section. Credential-file safety and unrelated runtime/tool permission guards are unchanged.
- Runtime refresh, repo snapshot sync, and overlay apply are separate authorization boundaries.
- `internal/overlay/target_parity.go` fails closed when a registered agent lacks a manifest target or explicit host-specific coverage. Shared skill allowlists and the Judgment Day runtime hook must agree; command wrappers on OpenCode require installed owner skills as Codex entrypoints. Public isolated apply tests prove current target-specific effects; source audit alone does not verify installed runtimes.

`shared/skills/notion-project-knowledge/SKILL.md` is the canonical Notion module for both targets. Matching activation blocks route Notion knowledge requests and proactive capture of meaningful behavioral and communicative interaction evidence in Me, including situated subjective impressions without a technical outcome; technical milestones alone do not trigger Me records. The current Me guide owns recording details, while thematic knowledge stays in its project. Codex preflight migrates only the exact previous activation next to the current local block and rejects unknown or relocated variants. This repository change does not install the skill or write to Notion; runtime apply remains separately authorized. Repository identity lookup and explicitly requested project creation follow the current global Notion index.

## Current adoption status

- The maintained repository snapshot boundary is v4.1.1 at `898d3976e8049a7238eaa7de60bff325c30d2b3d`.
- The separately authorized sync added both approved Codex orchestrator and shared authorization snapshots under `assets/upstream/codex/`, left 55 files unchanged, and did not write state. The fresh v4.1.1 audit passed for both repository targets: zero drift, six OpenCode structural flags `ok`, and Codex snapshots and source-to-delivery `ok`. The earlier OpenCode-only post-sync result remains historical evidence; installed runtimes were not audited.
- The authorized v4.1.1 snapshot sync changed 22 files, left 33 unchanged, and added or deleted none across 19 directories. The canonical post-sync audit passed with no managed or watched-unmanaged drift and six structural flags `ok`.
- The audit verifies repository snapshots and source structure only. OpenCode and Codex materialized runtime versions remain unverified; no runtime refresh, reinstall, or overlay apply is implied.
- The earlier v4.0.0 sync added 11 upstream-only OpenCode orchestrator and agent snapshots and deleted 35 snapshots: 18 exact SDD paths declared by `retired_upstream_snapshots` and 17 stale files in still-declared synced directories (11 commands, 2 plugins, and 4 `_shared` files). The manifest removals are bounded to exact files and reject unsafe paths, active-path overlap, symlink parents, and directories; directory pruning applies only within declared synced assets. Neither mechanism changes runtime files.
- Shared skill policy no longer keeps SDD skills and prunes explicitly rejected names in either selected runtime; Codex preserves every other upstream skill. No apply ran, and user-owned machine-local SDD settings or SDD runtime files were not removed.
- The v4 audit preserves all six OpenCode checks and independently verifies Codex source-to-`AGENTS.md` delivery through agentguidance, its adapter, and install/sync routing. Both targets' watched drift and structural failures block a future repository snapshot sync before boundary writes; installed runtimes are not verified.
- All five `plugins-v2` files are audit/sync-only. OpenCode 1.18.32 uses the V1 plugin path; this repository does not claim V2 runtime materialization support.
- The v3.7.0 `sdd-research` skill and command provide output-only collection with orchestrator-owned persistence. Its former local precedence patch is retired; the materialized runtime has not been verified against this source contract.
- Custom apply no longer creates SDD agents or profiles. It rejects legacy SDD assignments and keys before writes while preserving them for manual migration; global actor assignments use `global_actors`.
- Upstream ODD requires a work-unit commit at task close. The local opt-out patches are removed from both targets, but their installed runtime state is unverified until separately authorized refresh and apply. Task-close commits do not authorize pushes or PRs.
- Codex apply reuses the eight canonical `shared/skills/` sources, prunes only explicit rejected names after path preflight, inserts the local rules block after the upstream Engram marker and the separate Notion activation after that block, replaces only the `odd-explorer` and `odd-worker` instruction-table rows from Luna/high to Sol/high in the same exact-anchor `AGENTS.md` patch, and applies the declared Judgment Day retrospective delta. It leaves other ODD, JD, and review rows and the upstream work-unit skill unchanged, preserves unrelated Codex content, and does not render commands, replace the full `AGENTS.md`, or change Engram, `config.toml`, or model profiles. Account model availability and installed Codex runtime state remain unverified.
- Codex patch preparation also removes an exact stale second local block in the orchestrator-to-CodeGraph gap. Missing legacy is valid; duplicate, partial, changed, or relocated blocks fail before selected-target writes. This repository-only migration does not apply to the installed runtime until separately authorized.
- For selected OpenCode/Codex sanitized `chained-pr` installs, apply preflights the single obsolete `skills/chained-pr/references/chaining-details.md` against the approved upstream snapshot, rejecting unknown bytes, symlinks, or unsafe parents before selected-target writes. It removes only an exact match after installing the sanitized skill; the actively referenced `~/.agents/skills/` copy remains untouched. This repository change has not run against the installed runtimes.
- The repo's v3.7.0 partial SDD archive acceptance is historical policy evidence, not current v4 behavior or a future audit watchpoint.
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

- At the historical v2.5.0 boundary, repository policy modeled 11 SDD phases, including `sdd-research` after `sdd-explore`.
- At that time, the managed-assets policy covered the upstream `sdd-research` phase skill and its related shared and command files.
- Audit invariants validate semantic managed-profile detection and deny-by-default `permission.task` delegation. Deprecated agent-local `tools.task` is not required.
- The v2.5.0 repository snapshot sync, user-run full SDD component installation, and later separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully. This evidence is historical; the maintained repository boundary and independently verified runtime have since advanced to v2.9.1.
- Apply updated config; pruned 2 skills; changed and verified 1 owned asset; reported 0 topology warnings; managed 4 profiles; and reported agents created 3, updated 4, unchanged 49, with 0 unmanaged profiles.
- Final verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. The sanitized repo-owned `branch-pr`, all 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, configured overrides, and 4 managed profiles match canonical or local config. Conflicting upstream `branch-pr`, `issue-creation`, and Hermes are absent.
- All profiles have the ordered 11-phase topology and the 3 named `sdd-research` actors. Review actors remain global with no suffixed review actors, and Judgment Day differs only by the retrospective patch. `general` is `openai/gpt-5.6-terra` / `high`; `explore` is `openai/gpt-5.6-sol` / `medium`.
- At the v2.5.0 runtime verification point, managed background intent and effective policy were `on`; the launcher was `~/.gentle-ai/bin/opencode`, and Desktop/plain processes remained foreground. `~/.gentle-ai/state.json` was observed still reporting `installed_binary_version` as `2.4.0`; no root cause was asserted and external state was not edited.
- Historical RDD context: v2.2.2 had a confirmed correction-handoff bug, v2.5.0 supplied source and test evidence for the pre-edit plan, and a temporary local patch was assessed and deferred. The former mandatory canary is now an optional regression diagnostic, not an activation gate; see `maintenance.md` for the preserved analysis and current contracts.

## Preserved local boundaries

- Codex materialization is upstream-owned. Its local instruction and skill patches fail closed on missing or changed anchors, preserve unrelated bytes and file modes, and remain protected from owned-state cleanup without claiming whole-file ownership. The upstream repository audit now covers Codex inputs but not installed Codex runtime state.

- `assets/owned/opencode/AGENTS.md` is the source of truth for global credential-file safety. It applies to the primary agent and every delegated subagent or tool, favors non-secret alternatives, and requires a stop plus explicit user direction before any exceptional access.
- `shared/skills/package-security/SKILL.md` prohibits credential-file reads. Its templates use blocked scripts as npm's conservative baseline; reviewed npm approvals require supported strict enforcement, while pnpm `allowBuilds` / `strictDepBuilds` require exact CLI feature checks. Editing this canonical source does not update the installed global copy; `apply-gentle-ai-custom` is the propagation step.
- Policy-pruned skills without a repo-owned same-name replacement are denied through OpenCode `permission.skill` rules, globally and for configured agents. This blocks external copies without deleting or changing another runtime's files; the sanitized repo-owned `branch-pr` remains available.
- `shared/skills/chained-pr/` remains the sanitized runtime source. It adopts upstream anti-code-golf and bounded single-pass slicing rules while preserving explicit `owner/repo` scoping and plan-only automatic activation.
- `branch-pr` remains pruned upstream and replaced by the repo-owned sanitized copy. Explicit `/pr-create` and `/pr-regenerate` render title/body for audit visibility and then mutate through GitHub CLI without a content-approval pause; push authorization and unresolved required metadata, identity, or chain context remain blocking.
- `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` remain retained snapshots, not forced runtime selections.
- Global Judgment Day and review assignments are configured independently of SDD profiles through `global_actors`; custom apply creates no suffixed actors or SDD agents.
- The Judgment Day retrospective handoff remains a declared surgical patch and still requires exact compatible anchors.
- Upstream ODD task-close work-unit commits are retained for both targets; the local opt-out patches and OpenCode no-auto-commit runtime instruction are removed from repo sources. This does not authorize pushes or PRs, and the materialized runtime remains unverified until separately authorized refresh and apply.
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
