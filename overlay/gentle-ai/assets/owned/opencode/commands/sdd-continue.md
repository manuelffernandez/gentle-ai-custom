---
description: Continue the next SDD phase in the dependency chain
agent: gentle-orchestrator
---

Follow the SDD orchestrator workflow to continue the active change.

HARD GATE:
SDD Session Preflight must already be complete for this session. It must include execution mode, artifact store, chained PR strategy, and review budget. If missing, ask the exact orchestrator preflight prompt and STOP. Do not launch the next phase in the same turn.

WORKFLOW:

1. If the `gentle-ai` binary is available, run `gentle-ai sdd-continue [change] --cwd <repo>` and treat its dispatcher/status output as authoritative — but only when the session artifact store is `openspec` or `hybrid`. When the session artifact store is `engram`, do NOT invoke the native dispatcher at all — it cannot see the change (it reads only `openspec/changes/`); resolve status entirely from Engram (`mem_search` + `mem_get_observation` on the change's topic keys) using the manual status schema in `~/.config/opencode/skills/_shared/sdd-status-contract.md` (the same schema used when the binary is unavailable). The dispatcher is authoritative only for `openspec`/`hybrid`. If unavailable, resolve the active change using the status contract. If `$ARGUMENTS` is missing and more than one active change exists, ask the user to choose and STOP. Do not guess.
2. Produce or consume structured status before acting: schemaName, planningHome/changeRoot, artifactPaths/contextFiles, task progress, dependency states, next recommended action, blocked reasons, and actionContext.
3. Check which artifacts already exist for the active change (proposal, specs, design, tasks)
4. Use the dependency graph only as a reference for expected phase order:
   proposal → [specs ∥ design] → tasks → apply → review → verify/remediate → archive
   Never use it as a fallback for recognized review/remediation status tokens.
5. Launch the appropriate sub-agent(s) for the next phase only if authoritative status says the dependency is ready. Route only by `nextRecommended` and dependency states; never infer from free text.
   - If `blockedReasons` is non-empty, do not proceed to apply, archive, or terminal work.
   - If `nextRecommended` is a planning token (`propose`, `spec`, `design`, or `tasks`), launch only that planning phase.
   - If `nextRecommended` is `apply`, launch apply only when apply dependencies are ready.
   - If `nextRecommended` is `review`, do NOT fall back to the dependency graph. Report that an explicit review operation or receipt recovery is required and stop unless the proper review path is available.
   - If `nextRecommended` is `verify`, verification may run only to refresh evidence after review gates are satisfied.
   - If `nextRecommended` is `remediate`, continue remediation only through the proper remediation path using the persisted remediation state. If remediation eligibility or scope is unclear, report the blocker and stop.
   - If `nextRecommended` is `resolve-review`, report the review blockers, receipt/transaction state, and any manual resolution needed, then stop.
   - If `nextRecommended` is `resolve-blockers`, report `blockedReasons` and stop.
   - If `nextRecommended` is `archive`, launch archive only when archive dependencies are ready.
   - If `nextRecommended` is `sdd-new` or `select-change`, report that state and stop.
   - If `nextRecommended` is unknown or unsupported, fail closed with a clear blocker. Do not guess and do not fall back to dependency inference.
6. Present the result and ask the user to proceed

CONTEXT:

- Working directory: before doing anything else, run `git rev-parse --show-toplevel 2>/dev/null || pwd` with your bash tool and use the returned path as the authoritative workspace. In OpenCode Desktop (Electron) the parse-time interpolation resolves to the app data directory, not the project.
- Current project: the `basename` of the detected workspace above.
- Change name: $ARGUMENTS
- Execution mode: ask/cache per orchestrator
- Artifact store mode: ask/cache per orchestrator; do not hardcode Engram
- Delivery strategy: ask/cache per orchestrator
- Review budget: ask/cache per orchestrator

ENGRAM NOTE:
To check which artifacts exist in engram/hybrid, search: mem_search(query: "sdd/$ARGUMENTS/", project: "{project}") to list all artifacts for this change.
Sub-agents handle persistence automatically using the selected artifact store.

Read the orchestrator instructions to coordinate this workflow. Do NOT execute phase work inline — delegate to sub-agents.

STATUS CONTRACT:

Prefer `gentle-ai sdd-continue [change] --cwd <repo>` when available — but only when the session artifact store is `openspec` or `hybrid`; when the store is `engram`, do NOT invoke the binary and resolve status from Engram using the manual status schema. Otherwise read the installed shared status contract from this agent's skills directory and follow it. Use `~/.config/opencode/skills/_shared/sdd-status-contract.md` for OpenCode, `~/.config/kilo/skills/_shared/sdd-status-contract.md` for Kilo Code, `~/.qwen/skills/_shared/sdd-status-contract.md` for Qwen, or the equivalent configured skills directory for the current adapter. Do not use a workspace-relative `skills/_shared/...` path. Carry `actionContext` and allowed edit roots into any sub-agent launch. If status reports `workspace-planning` with no allowed edit roots, do not launch apply/verify/archive work that would infer repo-local ownership.
