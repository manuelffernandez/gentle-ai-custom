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

Those decisions live outside the repo, in `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`.

Operational orchestrator snapshots are also separated by scope:

- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` is versioned as a portable baseline
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` stores the local operational snapshot of `gentle-orchestrator` and all per-profile snapshots

The versioned policy only preserves the portable baseline of the overlay; local profile configuration is projected to `opencode.json` at runtime and must not be copied back into `gentle-ai-policy.json`.

## What we want to prune

We want to prune conventions that impose a specific way of collaborating in repositories:

- `branch-pr`
- `chained-pr`
- `issue-creation`
- `work-unit-commits`
- orchestrator blocks that impose:
  - PR strategy
  - review budget
  - chained/stacked PRs
  - `size:exception`
  - reviewer burnout protection as a PR policy

## Orchestrator sanitization goals

The sanitized OpenCode orchestrator layer must keep core SDD orchestration behavior while removing PR/budget workflow governance.

### Remove (hard rule)

When sanitizing an inline orchestrator captured from `~/.config/opencode/opencode.json`, remove or neutralize all content tied to:

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

- The generated `.overlay.md` prompt must remain a standalone valid prompt.
- Do not inject repo-specific hacks into core orchestration logic.
- Keep wording as close to upstream as possible unless removal requires a minimal rewrite.
- If required anchors are missing, fail closed and keep the current prompt reference untouched.

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

### Low-priority changes or noise

They normally do not require touching the overlay if they don't change observable behavior:

- internal bugfixes with no impact on prompts, skills, or generated config
- internal maintenance chores
- refactors without functional changes
- upstream docs that do not alter the runtime or assets

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
- the orchestrator sanitizer
- the interpretation of what to keep or prune

the agent must **stop and ask** before changing intent, policy, or scripts.

## What must be updated after a human decision

Once the decision is made:

- `maintenance-intent.md` if the intent changed
- `gentle-ai-policy.json` if the operational policy changed
- `upstream-state.json` when maintenance is closed
- `update-log.md` only when a closed, eligible maintenance event or contract change needs high-signal traceability beyond Git history

## Final rule

Intent overrides automation.

If the script, policy, or skill conflicts with this file, the agent must treat this document as the semantic source of truth and ask for human confirmation before proceeding.
