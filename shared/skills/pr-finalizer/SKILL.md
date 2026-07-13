---
name: pr-finalizer
description: "Deprecated compatibility skill for PR finalization; prefer `branch-pr` for PR create/regenerate from committed changes. Trigger: legacy `pr-finalizer` references."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.4"
---

## Status

`branch-pr` is now the preferred PR finalization skill in this overlay.

Keep `pr-finalizer` only as a temporary compatibility surface while older references migrate. When both skills are available, use `branch-pr` for new work.

## Compatibility Behavior

If this skill is loaded, do not execute an independent PR workflow.

Instead:

1. Load `branch-pr`.
2. Forward the user's requested mode and context:
   - create / draft PR content
   - regenerate / update existing PR content
   - committed diff only
   - explicit approval boundaries
3. Follow `branch-pr` as the source of truth.

## Hard Rules

- Do not duplicate `branch-pr` policy or command execution rules here.
- Do not require issue links, approval labels, or PR labels unless `branch-pr` finds an explicit repo policy requiring them.
- Do not run `gh`, `git push`, or any PR mutation from this compatibility skill directly.
- Do not claim PR creation or update success; `branch-pr` owns those terminal actions.

## Output Contract

Return a short compatibility notice and continue under `branch-pr`.

Example:

```text
`pr-finalizer` is deprecated in this overlay. Continuing with `branch-pr`, which owns PR create/regenerate behavior now.
```
