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
- Preserve the local keep/prune baseline and the orchestrator sanitization goals.
- Keep bash and PowerShell scripts behaviorally equivalent.
- Update `AGENTS.md`, `README.md`, and `overlay/gentle-ai/logs/update-log.md` when the workflow changes.

## Decision Gates

| If | Then |
|---|---|
| Upstream changed orchestrator structure | Update the sanitizers in both scripts before applying the overlay |
| Upstream added new skills | Decide keep vs prune and update policy JSON |
| The script can no longer sanitize safely | Fail closed, refresh docs, and record the blocker |

## Execution Steps

1. Read `overlay/gentle-ai/policy/gentle-ai-policy.json`.
2. Read `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.
3. Inspect upstream `/home/manuel/Documentos/gentle-ai` for changes affecting SDD prompts, skill inventory, or OpenCode profile behavior.
4. Update scripts, policy, docs, snapshots, and logs together when behavior changes.
5. Re-check that generated orchestrator prompts still preserve `## Model Assignments` while removing PR/budget governance.

## Output Contract

Return:
- files changed
- why the overlay needed adjustment
- whether keep/prune or sanitizer rules changed
- any migration note the user should know

## References

- `../../../../overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`
