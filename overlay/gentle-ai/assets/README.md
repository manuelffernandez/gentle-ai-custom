## Managed OpenCode assets

This directory stores maintenance snapshots and true repo-owned runtime assets.

Current roles:

- `overlay/gentle-ai/assets/upstream/` stores approved upstream copies used for audit/review.
- `overlay/gentle-ai/assets/owned/` stores only true repo-owned assets that `apply-gentle-ai-custom` installs.

Rules:

- Never copy raw upstream source snapshots into runtime from the owned tree.
- Keep the materialized OpenCode `AGENTS.md` baseline here; keep surgical runtime deltas under `../patches/`.
- Canonical repo-owned portable skills remain in `shared/skills/`; they are not moved into this tree.

See also:

- `../policy/managed-assets.json`
- `../owned-assets-refactor.md`
