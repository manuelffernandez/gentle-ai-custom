---
description: Execute an approved post-SDD commit plan, or generate one first if missing
agent: gentleman
---

Read the skill file at `~/.config/opencode/skills/commit-planner/SKILL.md` FIRST, then follow it exactly.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`
- Current project: !`echo -n "$(basename $(pwd))"`
- Mode: apply

TASK:
Execute a post-SDD local commit flow for the current changes.

Rules for this command:
- If there is already an explicitly approved commit plan in this conversation, execute that exact plan
- If there is no approved plan yet, generate one first and **stop** for approval
- Never assume approval for `git add` or `git commit`
- Use repository commit conventions first; if none are defined, fall back to Conventional Commits
- Never push or create/update PRs as part of this command
