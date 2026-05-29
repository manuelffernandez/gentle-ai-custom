---
name: gentle-ai-overlay-maintainer
description: "Trigger: gentle ai update, auditar gentle ai, depurar gentle ai, refresh overlay. Maintain the gentle-ai-custom overlay against upstream Gentle AI changes."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

# Gentle AI Overlay Maintainer

## Activation Contract

Use this skill when updating, auditing, or repairing the `gentle-ai-custom` overlay after upstream Gentle AI changes.

## Hard Rules

- Work from `gentle-ai-custom`, not from the upstream repo.
- Treat `/home/manuel/Documentos/gentle-ai` as upstream input only.
- Read semantic intent before making maintenance decisions.
- Preserve the local keep/prune baseline and the orchestrator sanitization goals.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes.
- Do not change intent, keep/prune, or sanitization behavior for new upstream changes without explicit user approval.

## Decision Gates

| If | Then |
|---|---|
| Upstream changed orchestrator structure | Update the sanitizers in both scripts before applying the overlay |
| Upstream added new skills or workflow behavior | STOP, summarize the impact, and ask the user what to keep or depure |
| The script can no longer sanitize safely | Fail closed, refresh docs, and record the blocker |

## Execution Steps

1. Read `overlay/gentle-ai/policy/maintenance-intent.md`.
2. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
3. Read `overlay/gentle-ai/state/upstream-state.json`.
4. Read `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.
5. Inspect upstream `/home/manuel/Documentos/gentle-ai` and determine the current relevant version boundary (tag and/or commit).
6. If `last_maintained_commit` exists, review the upstream change range from `last_maintained_commit` to the current upstream state, including intermediate minor releases or commits in that range.
7. Classify findings into:
   - behavior / workflow / feature changes relevant to the overlay
   - likely low-priority bugfix / chore noise
8. If relevant changes affect keep/prune intent or sanitization behavior, STOP and ask the user what to preserve or depure before editing anything.
9. After approval, update scripts, policy, docs, state, snapshots, and logs together.
10. Re-check that generated orchestrator prompts still preserve `## Model Assignments` while removing PR/budget governance.

## Output Contract

Return:
- files changed
- why the overlay needed adjustment
- whether keep/prune or sanitizer rules changed
- what upstream range was audited
- whether user approval was required and how it affected the result
- any migration note the user should know

## References

- `../../../../overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`
- `../../../../overlay/gentle-ai/policy/maintenance-intent.md`
- `../../../../overlay/gentle-ai/state/upstream-state.json`
