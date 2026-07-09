# Maintenance Intent — Gentle AI Custom

## Why this repo exists

`gentle-ai-custom` exists to preserve the valuable technical capabilities of Gentle AI without automatically accepting workflow conventions that do not apply to the user's personal flow.

The goal is not to fork the upstream or replace it. The goal is to interpose a local layer that:

- preserves SDD and useful utilities
- prunes unwanted PR/branch/review-budget conventions
- maintains explicit and reviewable local criteria

## What we want to keep

We want to keep anything that improves structure, reasoning, and technical quality:

- the full SDD flow
- skill resolver / registry
- documentation and commenting utilities
- useful testing
- skill creation and improvement
- adversarial review

## What is NOT versioned

Local, per-machine choices regarding OpenCode SDD profiles are not part of the shared repository policy.

This includes:

- `model` and `variant` assignments per named SDD profile
- custom user profile names (`sdd-orchestrator-<profile>` and their associated phases)
- per-profile orchestrator snapshots (`sdd-orchestrator-<profile>.last.md`)
- any local combination of providers/models intended for a single machine or personal preference

Those decisions live outside the repo, in `~/.config/gentle-ai-custom/opencode-local-config.json`.

That canonical local config also owns:

- optional `upstream_repo_path`
- optional `opencode_config_path`
- optional explicit `agent_overrides` for built-in agent keys

Runtime and audit sources are separated by scope:

- `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md` is the canonical repo-owned runtime source
- `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md` is the approved upstream audit baseline
- There is no stored upstream `AGENTS.md` in the OpenCode snapshot tree; the runtime `AGENTS.md` surface is materialized directly from the upstream `persona-gentleman.md` + `engram-protocol.md` inputs in that order.
- `overlay/gentle-ai/assets/owned/opencode/AGENTS.md` is the repo-owned runtime copy of that surface, extended with local overlay semantics (`no-auto-commit`, Gemini anti-sycophancy, etc.)

The versioned policy preserves portable runtime intent plus the approved upstream baseline; local profile configuration is projected to `opencode.json` at runtime and must not be copied back into `gentle-ai-policy.json`.

The global OpenCode `AGENTS.md` surface is overlay-owned completely. Its runtime content is materialized directly from the upstream OpenCode persona + engram sources into the repo-owned runtime copy, then extended with repo-owned local overlay semantics; it is not maintained through ad-hoc post-apply mutation.

## What we want to prune

We want to prune conventions that impose a specific way of collaborating in repositories:

- `branch-pr`
- `chained-pr`
- `issue-creation`
- `work-unit-commits`
- `hermes-ephemeral-delegation` and other Hermes-only upstream additions outside the maintained OpenCode runtime target
- orchestrator blocks that impose:
  - PR strategy
  - review budget
  - chained/stacked PRs
  - `size:exception`
  - reviewer burnout protection as a PR policy

## Repo-owned orchestrator behavior goals

The repo-owned OpenCode orchestrator asset must keep core SDD orchestration behavior while removing PR/budget workflow governance.

### Remove (hard rule)

When updating the repo-owned orchestrator asset, remove or neutralize all content tied to:

- PR strategy selection in SDD preflight
- review budget / changed-lines budget gates
- chained/stacked PR flow control
- size-exception policy handling
- review workload forecast branching before `sdd-apply`

This includes both explicit sections and references embedded inside other sections.

### Preserve (hard rule)

Preserve as much as possible of:

- coordinator/delegation role boundaries
- SDD command map and routing
- session preflight concept (execution mode + artifact store)
- init guard
- dependency graph
- result contract
- skill resolver protocol
- sub-agent context protocol
- strict TDD forwarding
- apply-progress continuity
- topic-key conventions

### Guardrails

- The repo-owned orchestrator prompt must remain a standalone valid prompt.
- Do not inject repo-specific hacks into core orchestration logic.
- Keep wording as close to upstream as possible unless removal requires a minimal rewrite.
- Prefer explicit owned-file edits over dynamic transformation logic.

### Scoped inline delegation overrides

- Hard gates remain non-bypassable by chat: safety, permission, data-loss, security, commit/push/PR, review-after-code-changes, and incident safeguards.
- Default behavior: delegate exploration that needs 4+ files and multi-file writes that touch 2+ non-trivial files.
- If the user explicitly asks to keep one specific 4+ file exploration or one specific scoped multi-file write inline, honor that request for that task only, acknowledge the context/reliability tradeoff once, keep the scope tight, and do not keep resisting it.
- Do not split a logically multi-file change into artificial smaller edits just to evade the default delegation preference; the scoped override exists so the rule can be honored directly when the task remains safe and manageable.
- These overrides do not weaken the review-after-code-changes gate: any inline multi-file code change still requires fresh review immediately after that write batch, before continuing toward commit/push/PR.
- These overrides do not weaken incident safeguards or any other hard safety/permission/data-loss/security/commit-push-PR rule.

### Review/fix convergence guard

- The repo-owned orchestrator must not alternate delegated review and delegated fix rounds indefinitely.
- Default convergence path: one delegated fix round after a fresh review, then one scoped re-review when the change is non-trivial code or otherwise needs fresh judgment.
- If re-review returns small, local, or already-understood residual findings, the orchestrator may fix them inline when safe instead of delegating another writer.
- Another delegated fix round is allowed only for a new high-risk issue, security/data-loss exposure, broad behavior change, unclear implementation context, or a fix that is no longer safe/manageable inline.
- If the same issue pattern survives the fix round, the orchestrator stops the loop and asks the user or escalates to judgment-day with a concise summary of what was attempted.
- This guard caps automatic delegation loops; it does not weaken fresh-review, incident, security, permission, data-loss, or commit/push/PR gates.

## Why these conventions do not apply

They might be valid for the upstream project or for other teams, but they are not the source of truth for the local workflow.

In this environment:

- the value lies in technical capabilities, not in PR governance
- repository collaboration is handled with our own tools and criteria
- the agent must not impose a branch/PR workflow unless explicitly requested by the user
- the orchestrator should keep SDD coordination value without reintroducing repository-governance policy by default

## How to evaluate upstream changes

### Relevant changes for the overlay

They are relevant when they affect observable behavior, local user experience, or assets managed by this repo. Examples:

- new skills or changes to existing skills
- changes in orchestrator or subagent prompts
- changes in install/sync or asset generation
- new workflow conventions imposed by default
- changes in OpenCode profiles, agent references, or model tables
- changes that require adding, removing, or reclassifying entries in `policy/managed-assets.json`, including owned/upstream paths, runtime targets, sync modes, or structural invariant coverage

### Low-priority changes or noise

They normally do not require touching the overlay if they don't change observable behavior:

- internal bugfixes with no impact on prompts, skills, or generated config
- internal maintenance chores
- refactors without functional changes
- upstream docs that do not alter the runtime or assets
- manifest-only churn in `policy/managed-assets.json` that does not change owned assets, approved upstream copies, runtime targets, or audit/apply behavior

## Managed-assets boundary

`policy/managed-assets.json` is the machine-readable asset ownership and installation map for audit/sync/apply.

It does not define intent by itself.

If the manifest and this file diverge, this file defines the meaning of what should be kept, pruned, staged, or installed, and the manifest must be brought back into alignment after human confirmation.

## Maintenance log scope

`overlay/gentle-ai/logs/update-log.md` is a closed-event maintenance record, not a mirror of repository history.

Log only:

- closed upstream audit outcomes
- adopted, rejected, or postponed upstream changes or ranges
- maintenance contract or policy decisions tied to upstream alignment
- tooling/runtime changes that affect audit/apply/recover/verify behavior against upstream
- maintenance incidents and recoveries

Do not log:

- documentation wording cleanups
- repo-local refactors with no upstream-maintenance impact
- new local features or skills unrelated to upstream maintenance
- cosmetic edits or intermediate iterations already explained by Git history

Keep one consolidated entry per closed maintenance event. If no maintenance decision or incident was closed, do not update the log.

## Mandatory human gate

If relevant changes appear during the audit that could modify:

- keep/prune
- the repo-owned orchestrator behavior
- the interpretation of what to keep or prune

the agent must **stop and ask** before changing intent, policy, or scripts.

## What must be updated after a human decision

Once the decision is made:

- `maintenance-intent.md` if the intent changed
- `gentle-ai-policy.json` if the operational policy changed
- `managed-assets.json` if the approved upstream asset map, owned runtime assets, runtime targets, sync modes, or structural invariant coverage changed
- `upstream-state.json` when maintenance is closed
- `update-log.md` only when a closed, eligible maintenance event or contract change needs high-signal traceability beyond Git history

## Final rule

Intent overrides automation.

If a script, `gentle-ai-policy.json`, `managed-assets.json`, or a skill conflicts with this file, the agent must treat this document as the semantic source of truth and ask for human confirmation before proceeding.
