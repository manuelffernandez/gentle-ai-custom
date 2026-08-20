# Maintenance Intent — Gentle AI Custom

This file is the semantic source of truth for how this repo maintains a small local layer on top of Gentle AI.

The current strategy is **upstream-first migration**: move the overlay toward the current Gentle AI v2 runtime behavior, keep only explicit local customizations, and avoid carrying old overlay patches forward unless they still solve a confirmed problem.

For an approved upstream asset, replacement-first means letting upstream Gentle AI materialize its finished runtime artifact. Keep the source snapshot for audit only, preserve the installed artifact, and apply only an explicit surgical customization documented in this file.

## Current baseline

- Last maintained boundary before this strategy shift: `v1.44.3`.
- New migration target: current Gentle AI v2+ upstream behavior, to be pinned in `upstream-state.json` only after the v2 audit/adoption event closes.
- The local overlay must be reduced while preserving the customizations listed in this file.
- Do not treat old v1.44-era owned runtime assets as authoritative just because they exist in the repo.

## Why this repo exists now

`gentle-ai-custom` exists to keep a small set of local preferences and owned skills installable after Gentle AI updates.

It should not be a broad fork of Gentle AI.

The goal is to:

- follow upstream SDD/runtime behavior by default
- preserve a small list of explicit local customizations
- keep maintainer decisions reviewable before runtime mutation
- avoid reintroducing old patches without fresh evidence that they are still needed

## What we keep as local overlay behavior

Keep these customizations unless the user explicitly changes this intent:

- `judgment-retrospective` as a repo-owned shared skill.
- The `judgment-day` runtime hook that invokes `judgment-retrospective` at terminal completion.
- `package-security` as a repo-owned shared skill.
- `code-design` as a repo-owned shared skill.
- `commit-planner` as the preferred local commit-planning workflow.
- A sanitized repo-owned `branch-pr` as the preferred local PR finalization workflow.
- A sanitized repo-owned runtime copy of `chained-pr` so GitHub CLI actions resolve an explicit `owner/repo`, avoid wrong-repo mutation, and keep automatic SDD/review-budget activation plan-only until the user explicitly requests branch/PR execution.
- `pr-finalizer` as a temporary deprecated compatibility skill during the `branch-pr` migration.
- Agent-Readable Markdown Standard for agent-consumed Markdown.
- `no-auto-commit`: never commit, push, or create PRs unless the user explicitly asks.
- Gemini anti-sycophancy guardrails for direct replies when a Gemini model is active.

## What we acquire from upstream by default

Acquire current upstream behavior by default for:

- SDD orchestration and phase behavior.
- Review and remediation lifecycle.
- Review receipts, review state, and native upstream review gates.
- Engram protocol updates.
- OpenCode command surfaces and runtime materialization that are not in direct conflict with the local customizations above.
- The rendered orchestrator, phase prompts, native review contracts/status, and review plugin are upstream-owned runtime artifacts. Raw upstream source templates are never custom install sources.
- Custom apply must not inspect or certify the semantics of upstream-owned prompts, contracts, plugins, or review workflows. Upstream owns that validation.
- Custom apply may validate only concrete prerequisites consumed by a local operation: ownership boundaries, base prompt references/files needed for configured profile cloning, safe obsolete-reference migration, and exact Judgment Day patch anchors.
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

## Commit and PR workflow boundary

The repo keeps `commit-planner` and the sanitized repo-owned `branch-pr` as opt-in local workflows.

`pr-finalizer` remains installable only as temporary deprecated compatibility while commands and references migrate to `branch-pr`.

These tools may replace the user's day-to-day use of upstream commit/PR skills, but they do not authorize self-initiated commits, pushes, or PR creation. State-changing actions still require explicit user invocation and the approval boundaries defined by each command.

`work-unit-commits` stays retained from upstream as the default planning criterion when commit grouping is needed.

Relationship:

- `work-unit-commits` defines the grouping standard: reviewable work units, tests/docs with the change when appropriate, rollback boundary, evidence, and awareness of chained-PR or review-budget slicing.
- `commit-planner` is the local operational layer: when the user explicitly invokes `/commit-plan`, `/commit-apply`, or `/commit-fast`, it should use that retained upstream criterion when available and then propose or execute concrete local commits.

Rules:

- Keep `commit-planner` and its commands installable.
- Keep the repo-owned sanitized `branch-pr` installable as the preferred PR workflow skill.
- Keep the sanitized repo-owned `chained-pr` installable while retaining its approved upstream snapshot for audit/sync.
- Keep upstream `issue-creation` pruned from the default OpenCode runtime surface.
- Keep `pr-finalizer` installable temporarily as deprecated compatibility.
- Preserve the `no-auto-commit` rule above them.
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

Local, per-machine choices regarding OpenCode models and SDD profiles are not part of shared repo policy.

These live outside the repo in `~/.config/gentle-ai-custom/opencode-local-config.json`:

- `model` and `variant` assignments
- custom named SDD profiles
- optional `upstream_repo_path`
- optional `opencode_config_path`
- explicit built-in agent overrides
- optional global Judgment Day and review actor assignments, including explicit dynamic inheritance
- optional named-profile Judgment Day assignments

The versioned repo owns portable runtime intent and installer behavior only.

Profile reconciliation may change `model` and `variant`, but it must preserve upstream base prompt values and cloned agent fields without interpreting their semantics. New named agents clone the corresponding installed base agent before assignments are changed.

The three Judgment Day actors may be assigned globally or cloned with a profile suffix when that named profile explicitly configures them. Omitted named JD slots delegate to the global actor. The six native review actors, including `review-validator`, remain global and shared; suffixed review actors are never created.

For global JD/review slots, omission preserves upstream/runtime assignment. Explicit `{"inherit": true}` removes both `model` and `variant` so OpenCode Task inherits the invoking primary actor's concrete assignment. A referenced missing upstream actor must fail before any write.

## Managed asset boundary

`overlay/gentle-ai/policy/managed-assets.json` is the machine-readable map for audit, sync, and apply.

It does not define intent by itself.

If the manifest conflicts with this file, this file wins semantically and the manifest must be updated after human confirmation.

## Maintainer workflow

Before mutating repo or runtime state for upstream alignment:

1. Immediately after loading the minimal maintainer skill/context, run the canonical audit release parity preflight synchronously, before broad repository exploration or repository-reading/drift-analysis delegation.
2. Require the resolved upstream checkout to be clean, including staged, unstaged, untracked, and submodule state.
3. Require exact `refs/tags/v<V>` for installed `gentle-ai --version` to exist, peel to a commit, and equal upstream `HEAD`.
4. Audit drift against the last maintained boundary only after parity succeeds.
5. Classify each meaningful change as `Adquirir`, `Sanitizar`, or `Ignorar`.
6. Produce a concise pre-mutation decision summary.
7. Stop for explicit user approval.

If release parity fails, stop immediately and delegate nothing. Do not use `last_maintained_commit`, `git describe`, nearest tags, semantic version ordering, or branch ancestry as release parity proof. `last_maintained_commit` defines only the audit range after parity succeeds. Prereleases are allowed only through their exact matching tag. Missing or stale tags, dirty state, and a branch ahead of the release block audit and sync with manual fetch/checkout guidance; automation must not fetch or check out revisions.

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
