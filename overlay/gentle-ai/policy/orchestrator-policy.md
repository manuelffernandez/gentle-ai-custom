# Orchestrator Derivation Policy

## Objective

Maintain a derived OpenCode `gentle-orchestrator` prompt that keeps core SDD orchestration behavior but **removes PR/budget workflow governance**.

## Remove (hard rule)

When deriving from upstream (`/home/manuel/Documentos/gentle-ai/internal/assets/opencode/sdd-orchestrator.md`), remove or neutralize all content tied to:

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

## Derivation constraints

1. The derived file must remain a standalone valid prompt.
2. Do not inject local repo-specific hacks into orchestration logic.
3. Keep wording close to upstream unless removal requires minimal rewrites.
4. Track the derivation in `overlay/gentle-ai/logs/update-log.md`.

## Update process

1. Read upstream orchestrator prompt.
2. Diff against current derived prompt.
3. Re-apply this policy.
4. Update `snapshots/upstream/opencode/gentle-orchestrator.last.md`.
5. Append a dated note in `logs/update-log.md`.
