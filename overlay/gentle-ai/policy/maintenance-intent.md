# Maintenance Intent — Gentle AI Custom

This file is the semantic source of truth for how this repo maintains a small local layer on top of Gentle AI.

The current strategy is **upstream-first migration**: keep only explicit local customizations and avoid carrying old overlay patches forward unless they still solve a confirmed problem. The maintained repository snapshot boundary is Gentle AI v4.1.1 at `898d3976e8049a7238eaa7de60bff325c30d2b3d`. This is repository snapshot adoption only; the materialized OpenCode and Codex runtimes remain unverified.

For an approved upstream asset, replacement-first means letting upstream Gentle AI materialize its finished runtime artifact. Keep the source snapshot for audit only, preserve the installed artifact, and apply only an explicit surgical customization documented in this file.

## Current baseline

- Maintained repository snapshot boundary: Gentle AI `v4.1.1` at `898d3976e8049a7238eaa7de60bff325c30d2b3d`.
- The separately authorized v4.1.1 snapshot sync advanced `upstream-state.json` and refreshed 22 approved audit snapshots, leaving 33 unchanged with none added or deleted across 19 directories. It did not refresh or apply either runtime; their current versions remain unverified.
- Current v4 policy follows upstream ODD and does not retain SDD skills as current OpenCode runtime policy. The SDD names remain only as migration blockers for user-owned legacy settings and as explicitly historical evidence.
- Continue acquiring compatible Gentle AI behavior upstream-first. Pin each accepted boundary in `upstream-state.json` only after its audit, explicit approval, and snapshot sync close.
- The local overlay must be reduced while preserving the customizations listed in this file.
- Do not treat old v1.44-era owned runtime assets as authoritative just because they exist in the repo.

## Codex target boundary

Codex apply installs the eight repo-owned portable skills from `shared/skills/` under `~/.codex/skills/`; allowlisted same-name skills are intentional repo-owned replacements. It does not prune unrelated upstream Codex skills, render command wrappers, replace the full `~/.codex/AGENTS.md`, or change `config.toml`, Engram files, SDD prompts, or model profiles.

- Preserve unrelated upstream Codex instruction bytes. Insert the marker-bounded local rules after the unique upstream `<!-- /gentle-ai:engram-protocol -->` anchor and replace exactly the `odd-explorer` and `odd-worker` table rows from `gpt-6.1-luna` / `high` to `gpt-6.1-sol` / `high` in the same declared `AGENTS.md` patch. Leave `odd-verify`, Judgment Day, review, and all other rows unchanged. This is a Codex instruction-table exception, not a model profile or OpenCode override; actual account model availability is unverified.
- Keep `notion-project-knowledge` activation in a separate marker-bounded block after the local-instructions closing marker, using a second sequential replacement. Preserve replacement zero unchanged because legacy-block migration derives its recognized content from that replacement.
- Keep implementation scope, plain-language communication, Agent-Readable Markdown, credential-file safety, and the conditional Gemini response rule in that local block. Do not add the retired remote-operation authorization rule or a local no-auto-commit override.
- A stale second copy of the local block may remain between the orchestrator closing marker and the CodeGraph opening marker. During Codex patch preflight, remove only the exact legacy block containing the retired no-commit line, whether the current Engram-adjacent block is already installed or not. Reject duplicate, partial, changed, or relocated local blocks before any selected-target write; do not relax generic patch matching or claim the whole file.
- Apply only the Judgment Day retrospective delta when its exact Codex-installed skill anchor matches. Leave the upstream ODD work-unit commit instructions unchanged. Validate every declared patch before writing any target; preserve file modes and fail closed on missing, changed, or ambiguous anchors.
- Treat upstream Codex skills and `AGENTS.md` as materialized runtime inputs, never as repo-owned full-file install sources. Install only the declared portable-skill allowlist; leave unrelated upstream skills alone. Track copied portable skills as owned output; surgical patch targets are protected but not wholly owned.
- The v4 audit verifies the maintained repository boundary and OpenCode source structure, not installed Codex runtime materialization or model availability. Runtime behavior remains unverified until the user authorizes a Codex apply and its result is inspected separately.

## Historical v3.7.0 partial SDD archive acceptance

At the v3.7.0 boundary, this repo consciously accepted behavior that allowed an explicit archive request to sync a syntactically applicable spec delta while SDD tasks remained unchecked. This is historical policy evidence only. The v4 maintained target is ODD without SDD; do not treat this archived decision as current behavior or a future v4 audit watchpoint.

## Why this repo exists now

`gentle-ai-custom` exists to keep a small set of local preferences and owned skills installable after Gentle AI updates.

It should not be a broad fork of Gentle AI.

The goal is to:

- follow upstream ODD/runtime behavior by default
- preserve a small list of explicit local customizations
- keep maintainer decisions reviewable before runtime mutation
- avoid reintroducing old patches without fresh evidence that they are still needed

## What we keep as local overlay behavior

Keep these customizations unless the user explicitly changes this intent:

- `judgment-retrospective` as a repo-owned shared skill.
- The `judgment-day` runtime hook that invokes `judgment-retrospective` at terminal completion.
- `package-security` as a repo-owned shared skill.
- `code-design` as a repo-owned shared skill.
- `notion-project-knowledge` as a repo-owned shared skill with identical OpenCode and Codex activation. Current Notion guides own project structure and recording procedures; the parent retrieves progressively and proactively preserves useful project knowledge and personal activity evidence within their authorized scope.
- `commit-planner` as the preferred local commit-planning workflow.
- A sanitized repo-owned `branch-pr` as the preferred local PR finalization workflow.
- A sanitized repo-owned runtime copy of `chained-pr` so GitHub CLI actions resolve an explicit `owner/repo`, avoid wrong-repo mutation, and keep automatic SDD/review-budget activation plan-only until the user explicitly requests branch/PR execution.
- `pr-finalizer` as a temporary deprecated compatibility skill during the `branch-pr` migration.
- Agent-Readable Markdown Standard for agent-consumed Markdown.
- A concise implementation-scope rule in the repo-owned global `AGENTS.md`; that file is its single source of truth.
- A plain-language communication rule in the repo-owned global `AGENTS.md`; that file is its single source of truth.
- Follow upstream ODD's required work-unit commit at task close. This requirement does not authorize pushes or PRs, which retain their explicit user-authorization and workflow gates; commits outside an ODD task-close requirement still need an explicit user request.
- Do not carry the upstream remote-operation authorization rule into overlay-maintained OpenCode prompts or the repo-owned global `AGENTS.md`. Keep the upstream contract and its source snapshot unchanged for audit; custom apply removes only the exact recognized upstream or previously relaxed section from maintained prompt carriers. This does not change credential-file safety or unrelated runtime/tool permission guards.
- Gemini anti-sycophancy guardrails for direct replies when a Gemini model is active.

## What we acquire from upstream by default

Acquire current upstream behavior by default for:

- OpenCode ODD orchestration, routing, and agent behavior.
- The v4 ODD runtime surface without SDD skills; current OpenCode policy prunes legacy SDD skill names if they are present.
- Review and remediation lifecycle.
- Review receipts, review state, and native upstream review gates.
- Engram protocol updates.
- OpenCode command surfaces and runtime materialization that are not in direct conflict with the local customizations above.
- The canonical remote-operation authorization contract remains an audit input; its rule is intentionally removed from maintained OpenCode prompt projections.
- The rendered orchestrator, phase prompts, native review contracts/status, and review plugin are upstream-owned runtime artifacts. Raw upstream source templates are never custom install sources.
- Custom apply must not inspect or certify the semantics of upstream-owned prompts, contracts, plugins, or review workflows. Upstream owns that validation.
- Custom apply may validate only concrete prerequisites consumed by a local operation: ownership boundaries, base prompt references/files needed for configured profile cloning, safe obsolete-reference migration, exact declared runtime-patch anchors, and the exact known forms of a legacy prompt section it removes.
- Upstream-owned plugin absence or opaque content must not block unrelated custom operations. Files consumed but not owned by a custom operation must remain byte- and mode-unchanged.
- PR strategy, review-budget, chained PR, and work-unit concepts, unless they directly conflict with an explicit local customization or the user rejects them in a concrete maintenance decision.
- Approved upstream `chained-pr` snapshots remain retained for audit/sync even though runtime installation uses a sanitized repo-owned shared copy for explicit GitHub repo scoping.
- Upstream `issue-creation` is outside the default repo-owned runtime/install surface. Keep it pruned from OpenCode runtime unless the user later asks to restore an issue workflow explicitly.
- Upstream `branch-pr` remains the explicit exception: keep the upstream copy pruned because its mandatory issue-first rule conflicts with the local workflow where not every PR should require an approved issue. Replace it locally with the sanitized repo-owned `branch-pr` skill instead.

Upstream-first means these areas are not rejected globally anymore, except for explicit confirmed conflicts such as the upstream `branch-pr` copy.

## What we do not carry forward as local patches

Do not preserve these v1.44-era local patches by default:

- custom delegation thresholds
- metadata-only file-size preflight as a local orchestration rule
- scoped inline multi-file override rules
- custom review/fix anti-loop convergence guard
- old local SDD review lifecycle behavior that predates upstream v2 review/remediation

If a similar problem reappears on the v2 baseline, solve it against the v2 design instead of copying the old patch forward.

## Remote operation authorization boundary

Keep the upstream authorization contract as the audit baseline, but do not install its rule into overlay-maintained OpenCode prompts. The snapshot under `assets/upstream/` remains unchanged and is never a runtime source.

- Apply validates reachable maintained inline/file prompts and any existing legacy shared SDD prompt files still declared among owned overlay assets before its first write, then removes only the exact upstream rule or the previously installed relaxed version. Missing optional legacy files are skipped.
- Apply rejects duplicate, malformed, unknown, or partially changed sections and unsafe prompt paths before any write. It leaves unrelated prompt bytes and file modes unchanged and is idempotent once the section is absent.
- Remote-bearing unmanaged named profiles block apply rather than being changed.
- The repo-owned full-file `AGENTS.md` omits this section; credential-file safety and unrelated permission guards remain unchanged.
- Runtime refresh and overlay apply remain separately authorized. Repository changes do not establish that the installed runtime has been updated.

## Commit and PR workflow boundary

The repo keeps `commit-planner` and the sanitized repo-owned `branch-pr` as opt-in local workflows.

`pr-finalizer` remains installable only as temporary deprecated compatibility while commands and references migrate to `branch-pr`.

These tools may replace the user's day-to-day use of upstream commit/PR skills. The upstream ODD work-unit commit requirement applies at ODD task close; it does not authorize pushes or PRs. Pushes and PR creation retain the explicit user-authorization and workflow gates defined by each command. For explicitly invoked `/pr-create` and `/pr-regenerate`, generated title/body content is audit-visible rather than an approval gate; push authorization for a missing or stale remote branch, unresolved identity, required repo-policy metadata, and chain context remain blocking decisions.

`work-unit-commits` stays retained from upstream as the planning criterion and required ODD task-close commit workflow. The local OpenCode and Codex opt-out patches and OpenCode no-auto-commit instruction have been removed from the repo-owned runtime sources. This is a repository-source change only: do not claim the materialized runtime has been refreshed or verified until the user separately authorizes the relevant runtime refresh and apply.

Relationship:

- `work-unit-commits` defines the grouping standard: reviewable work units, tests/docs with the change when appropriate, rollback boundary, evidence, and awareness of chained-PR or review-budget slicing.
- `commit-planner` is the local operational layer: when the user explicitly invokes `/commit-plan`, `/commit-apply`, or `/commit-fast`, it should use that retained upstream criterion when available and then propose or execute concrete local commits.

Rules:

- Keep `commit-planner` and its commands installable.
- Keep the repo-owned sanitized `branch-pr` installable as the preferred PR workflow skill.
- Keep the sanitized repo-owned `chained-pr` installable while retaining its approved upstream snapshot for audit/sync.
- Keep upstream `issue-creation` pruned from the default OpenCode runtime surface.
- Keep `pr-finalizer` installable temporarily as deprecated compatibility.
- Keep upstream ODD task-close commits distinct from push and PR authorization; those remote actions remain user-gated.
- When upstream adds or changes commit/PR workflows, compare them against these tools instead of assuming either side wins automatically. Upstream `branch-pr` remains a confirmed conflict and stays pruned; `work-unit-commits` and `chained-pr` are retained by default, while `issue-creation` remains intentionally outside the default runtime surface.
- If command names collide, stop and ask before choosing which command owns the runtime entrypoint.

## Judgment Day boundary

The retrospective handoff is local behavior and should be preserved.

The base `judgment-day` protocol remains upstream-owned. The local handoff is an exact-anchor, idempotent runtime patch, not a full owned skill fork.

When updating against upstream:

- Start from upstream `judgment-day` behavior.
- Add only the retrospective handoff needed to invoke `judgment-retrospective`.
- Fail before runtime writes when the installed skill is absent or its anchors are incompatible.
- Do not keep older custom Judgment Day protocol differences unless the user explicitly approves them after review.

## Historical SDD snapshot boundary

The former v3.7.0 snapshots included `sdd-research` as an output-only evidence collector with orchestrator-owned persistence. This is historical snapshot evidence, not current v4 runtime behavior. The v4 snapshot sync retired those files and the current keep/prune policy no longer retains SDD skills. Existing machine-local SDD settings are user-owned and are not removed or migrated by policy or apply.

- Do not interpret these v3.7.0 snapshots as proof that SDD agents or profiles are part of the v4 overlay.
- Do not infer runtime state from the repository snapshot cleanup.

## Agent-readable Markdown boundary

The Agent-Readable Markdown Standard remains local runtime policy.

Apply it to Markdown whose primary audience is agents or LLMs, including:

- `AGENTS.md`
- `SKILL.md`
- maintainer docs
- intent and policy docs
- runtime instruction files

Keep this standard as a small, explicit policy layer. Do not use it as a reason to rewrite unrelated human docs during upstream adoption.

## Runtime and config ownership

Local, per-machine choices regarding OpenCode models are not part of shared repo policy. Legacy SDD model/profile settings are migration blockers and must be preserved for manual review; apply never maps them to ODD roles.

These live outside the repo in `~/.config/gentle-ai-custom/opencode-local-config.json`:

- `model` and `variant` assignments
- legacy `default_profile` and `profiles` values, which apply rejects without modifying
- optional `upstream_repo_path`
- optional `opencode_config_path`
- explicit built-in agent overrides
- optional global Judgment Day and review actor assignments, including explicit dynamic inheritance

The versioned repo owns portable runtime intent and installer behavior only.

The current apply does not generate SDD profiles or named SDD agents. It rejects legacy profile settings and SDD agent keys before writes, preserving those user-owned values for manual review.

`global_actors` is the only supported Judgment Day and review assignment surface. It assigns only the three global JD actors and six global review actors, including `review-validator`; the local configuration no longer supports named profiles or suffixed actors.

For global JD/review slots, omission preserves upstream/runtime assignment. Explicit `{"inherit": true}` removes both `model` and `variant` so OpenCode Task inherits the invoking primary actor's concrete assignment. A referenced missing upstream actor must fail before any write.

## Managed asset boundary

`overlay/gentle-ai/policy/managed-assets.json` is the machine-readable map for audit, sync, and apply.

It does not define intent by itself.

If the manifest conflicts with this file, this file wins semantically and the manifest must be updated after human confirmation.

## Maintainer workflow

Before mutating repo or runtime state for upstream alignment:

1. Immediately after loading the minimal maintainer skill/context, run the canonical audit synchronously, before broad repository exploration or repository-reading/drift-analysis delegation. By default it checks the release matching the installed binary. When the user explicitly selects another release, pass its exact tag with `--release v<V>`.
2. Require the resolved upstream checkout to be clean, including staged, unstaged, untracked, and submodule state.
3. Require exact `refs/tags/v<V>` for the selected release to exist, peel to a commit, and equal upstream `HEAD`.
4. Audit drift against the last maintained boundary only after parity succeeds.
5. Classify each meaningful change as `Adquirir`, `Sanitizar`, or `Ignorar`.
6. Produce a concise pre-mutation decision summary.
7. Stop for explicit user approval.

If selected-release verification fails, stop immediately and delegate nothing. An explicit release selector is preview-only: it does not change the installed binary or runtime, adopt the release, or authorize repository snapshot sync. Repo sync requires separate explicit authorization and must use the same selected release. Do not use `last_maintained_commit`, `git describe`, nearest tags, semantic version ordering, or branch ancestry as release parity proof. `last_maintained_commit` defines only the audit range after selected-release verification succeeds. Prereleases are allowed only through their exact matching tag. Missing or stale tags, dirty state, and a branch ahead of the release block audit and sync with manual fetch/checkout guidance; automation must not fetch or check out revisions.

Not having run `gentle-ai sync` or a reinstall yet is the normal pre-audit state. Runtime refresh is human-owned by default: the maintainer may audit, classify, recommend, and print exact instructions, but must not run `gentle-ai sync` or a reinstall unless the human explicitly requests that specific action. Audit/adoption approval, repo snapshot sync, and overlay-edit approval do not authorize runtime refresh.

Repo snapshot sync via `sync-gentle-ai-upstream-assets` remains a distinct, separately approved repository action. Apply also remains separately authorized: runtime-refresh approval does not authorize `apply-gentle-ai-custom`, and apply continues to use its own installed-binary versus maintained-overlay preflight without requiring an upstream clone.

After approval:

- update owned assets, policy, docs, and state together when needed
- keep Bash and PowerShell entrypoints behaviorally equivalent when either side changes
- run the relevant verification
- run one fresh-context consistency review before closing maintenance work

## Maintenance log scope

`overlay/gentle-ai/logs/update-log.md` is a closed-event maintenance record, not a mirror of repository history.

Log only:

- closed upstream audit outcomes
- adopted, rejected, or postponed upstream changes or ranges
- maintenance contract or policy decisions tied to upstream alignment
- tooling/runtime changes that affect audit, apply, recovery, or verification against upstream
- maintenance incidents and recoveries

Do not log:

- documentation wording cleanups
- repo-local refactors with no upstream-maintenance impact
- cosmetic edits or intermediate iterations already explained by Git history

Keep one consolidated entry per closed maintenance event.

## Human gate

Stop and ask before changing:

- the retained local customization list
- the upstream-first acquisition policy
- the commit/PR workflow boundary
- the Judgment Day boundary
- managed asset ownership or runtime targets
- install/apply behavior

## Final rule

Intent overrides automation.

If a script, policy file, managed-assets manifest, runtime asset, or skill conflicts with this file, treat this file as the semantic source of truth and ask for human confirmation before proceeding.
