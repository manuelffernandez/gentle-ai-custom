## Repo-owned OpenCode assets

This tree is the canonical source of repo-owned SDD/runtime behavior for OpenCode.

Contents are compared against the approved upstream snapshot set in `../upstream/opencode/`, even though that tree stores only the retained upstream inputs/assets and not a full upstream-owned mirror.

These assets are distinct from `shared/skills/`, which remains the canonical source of portable repo-owned skills.

The tree now also owns the runtime `AGENTS.md` installed into `~/.config/opencode/AGENTS.md`.

That file is the materialized upstream OpenCode AGENTS baseline plus repo-owned local overlay semantics.
