---
description: Enter maintainer mode for the gentle-ai-custom overlay
---

Read the skill file at `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` FIRST, then follow it exactly.

CONTEXT:
- Working directory: !`git rev-parse --show-toplevel 2>/dev/null || pwd`
- Current project: !`basename "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"`

TASK:
Enter maintainer mode. Immediately after loading the skill and its minimal context, run `bash audit-gentle-ai-upstream.sh` synchronously. Do this before broad repository exploration or any repository-reading/drift-analysis delegation. If release parity fails, STOP immediately and delegate nothing. The user's update description is diagnostic only; `last_maintained_commit` defines the audit range, not release parity.

Follow the full workflow defined in the skill. Treat no `gentle-ai sync` or reinstall yet as the normal pre-audit state. You may recommend and print exact runtime-refresh instructions, but MUST NOT run `gentle-ai sync` or a reinstall unless the human explicitly requests that specific runtime action. Approval of findings, adoption, repo snapshot sync, or overlay edits is not runtime-refresh authorization. Keep `bash sync-gentle-ai-upstream-assets.sh` separate and approval-gated. Overlay apply also requires separate explicit authorization; runtime-refresh approval does not authorize apply.
