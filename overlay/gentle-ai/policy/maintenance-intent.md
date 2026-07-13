# Maintenance Intent — Gentle AI Custom

This file is the semantic source of truth for how this repo maintains a small local layer on top of Gentle AI.

The current strategy is **upstream-first migration**: move the overlay toward the current Gentle AI v2 runtime behavior, keep only explicit local customizations, and avoid carrying old overlay patches forward unless they still solve a confirmed problem.

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
- Sanitized repo-owned runtime copies of `chained-pr` and `issue-creation` so GitHub CLI actions resolve an explicit `owner/repo` and avoid wrong-repo mutation.
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
- PR strategy, review-budget, chained PR, and work-unit concepts, unless they directly conflict with an explicit local customization or the user rejects them in a concrete maintenance decision.
- `issue-creation` as an optional upstream issue workflow when it helps the user's process.
- Approved upstream `chained-pr` and `issue-creation` snapshots remain retained for audit/sync even though runtime installation uses sanitized repo-owned shared copies for explicit GitHub repo scoping.
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
- Keep sanitized repo-owned `chained-pr` and `issue-creation` installable while retaining their approved upstream snapshots for audit/sync.
- Keep `pr-finalizer` installable temporarily as deprecated compatibility.
- Preserve the `no-auto-commit` rule above them.
- When upstream adds or changes commit/PR workflows, compare them against these tools instead of assuming either side wins automatically. Upstream `branch-pr` remains a confirmed conflict and stays pruned; `work-unit-commits`, `chained-pr`, and `issue-creation` are retained by default, with `issue-creation` remaining contextual rather than universal.
- If command names collide, stop and ask before choosing which command owns the runtime entrypoint.

## Judgment Day boundary

The retrospective handoff is local behavior and should be preserved.

The base `judgment-day` protocol should track upstream unless the hook requires a minimal owned runtime wrapper.

When updating against upstream:

- Start from upstream `judgment-day` behavior.
- Add only the retrospective handoff needed to invoke `judgment-retrospective`.
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

The versioned repo owns portable runtime intent and installer behavior only.

## Managed asset boundary

`overlay/gentle-ai/policy/managed-assets.json` is the machine-readable map for audit, sync, and apply.

It does not define intent by itself.

If the manifest conflicts with this file, this file wins semantically and the manifest must be updated after human confirmation.

## Maintainer workflow

Before mutating repo or runtime state for upstream alignment:

1. Confirm the current upstream version and dirty state.
2. Audit drift against the last maintained boundary.
3. Classify each meaningful change as `Adquirir`, `Sanitizar`, or `Ignorar`.
4. Produce a concise pre-mutation decision summary.
5. Stop for explicit user approval.

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
