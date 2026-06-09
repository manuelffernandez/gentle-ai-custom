---
description: Enter maintainer mode for the gentle-ai-custom overlay
---

Read the skill file at `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` FIRST, then follow it exactly.

CONTEXT:
- Working directory: !`git rev-parse --show-toplevel 2>/dev/null || pwd`
- Current project: !`basename "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"`

TASK:
Enter maintainer mode. Follow the full workflow defined in the skill — starting with determining what the user did (brew upgrade, git pull, gentle-ai sync, or TUI reinstall) before taking any action.
