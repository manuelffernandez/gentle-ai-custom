## Managed OpenCode assets

This directory is active runtime and maintenance input.

Current roles:

- `overlay/gentle-ai/assets/upstream/` stores approved upstream copies used for audit/review.
- `overlay/gentle-ai/assets/owned/` stores the repo-owned assets that `apply-gentle-ai-custom` installs into runtime targets.

Rules:

- Keep upstream and owned trees path-mirrored where practical.
- Store only behavior-defining assets here: orchestrator prompt, SDD phase prompts/skills, shared SDD files, and SDD commands.
- Canonical repo-owned portable skills remain in `shared/skills/`; they are not moved into this tree.

See also:

- `../policy/managed-assets.json`
- `../owned-assets-refactor.md`
