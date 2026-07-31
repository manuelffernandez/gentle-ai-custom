---
description: Enter maintainer mode for the gentle-ai-custom overlay
---

Read the skill file at `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` FIRST, then follow it exactly.

CONTEXT:
- Working directory: !`git rev-parse --show-toplevel 2>/dev/null || pwd`
- Current project: !`basename "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"`

TASK:
Enter maintainer mode. Follow the full workflow defined in the skill. Ask what the user updated only as diagnostic context; never trust that answer as parity proof. Before drift analysis, run the canonical audit so its preflight verifies a clean upstream checkout at the exact `v<V>` tag and commit for installed `gentle-ai --version`. Treat no `gentle-ai sync` or reinstall yet as the normal pre-audit state.
