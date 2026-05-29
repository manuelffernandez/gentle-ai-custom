# Orchestrator Derivation Policy

## Objective

Maintain a sanitized OpenCode orchestrator layer that keeps core SDD orchestration behavior but **removes PR/budget workflow governance**.

## Remove (hard rule)

When sanitizing an inline orchestrator captured from `~/.config/opencode/opencode.json`, remove or neutralize all content tied to:

- PR strategy selection in SDD preflight
- review budget / changed-lines budget gates
- chained/stacked PR flow control
- size-exception policy handling
- review workload forecast branching before `sdd-apply`

This includes both explicit sections and references inside other sections.

## Keep (hard rule)

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

## Sanitization constraints

1. The generated `.overlay.md` file must remain a standalone valid prompt.
2. Do not inject local repo-specific hacks into orchestration logic.
3. Keep wording close to upstream unless removal requires minimal rewrites.
4. Track the derivation in `overlay/gentle-ai/logs/update-log.md`.
5. If required anchors are missing, fail closed and keep the current prompt reference untouched.

## Update process

1. Read the current inline orchestrator prompt from OpenCode config after `gentle-ai sync`.
2. Snapshot it under `snapshots/upstream/opencode/orchestrators/<agent>.last.md`.
3. Re-apply this policy.
4. Emit the generated prompt under `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md`.
5. Append a dated note in `logs/update-log.md`.
