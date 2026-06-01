# Reusable Package Security Snippets

This folder captures the reusable output of a package-install security review originally done for another project.

## Quick path

1. Use `agent-prompt.md` to brief another agent that will create a global reusable skill.
2. Copy the config snippets that match your package manager and environment.
3. Adapt `allowBuilds` and CI behavior per project before enforcing them blindly.

## What belongs where

| Layer | Purpose | Files |
| --- | --- | --- |
| Machine-global | Protect ad-hoc installs on one workstation | `global-npmrc`, `global-pnpm-config.yaml` |
| Project-level | Protect all developers, CI, Docker, and Cloud Build | `project-npmrc`, `project-pnpm-workspace.yaml` |
| CI guidance | Make build automation follow the same baseline | `cloud-build-notes.md` |
| Agent behavior | Teach an LLM how to reason about package installs/updates | `agent-prompt.md` |

## Key decisions captured here

| Topic | Decision |
| --- | --- |
| npm minimum version | Use npm 11+ for `allow-git`, `min-release-age`, and `ignore-scripts` support together |
| pnpm minimum version | pnpm 10.33.0 already supports the required security settings |
| Exact versions | Use `saveExact` / `save-exact` for future installs; migrate existing caret ranges separately |
| pnpm build scripts | Prefer `allowBuilds` + `strictDepBuilds`; do not rely on deprecated `onlyBuiltDependencies` |
| Trust enforcement | Use `trustPolicy: no-downgrade` to reject weaker publish-time trust signals |
| Exotic sources | Use `blockExoticSubdeps: true` to block git/tarball subdependencies |
| Global vs project config | Keep both; global protects one machine, project config protects CI and teammates |
| CI posture | Prefer project-enforced config and deterministic installs over hidden per-machine setup |

## npm vs pnpm: script security model

| Package manager | Best native baseline | Why |
| --- | --- | --- |
| npm | `ignore-scripts=true` | npm does not have a first-class built-in allowlist equivalent to pnpm's reviewed build-script model |
| pnpm | `allowBuilds` + `strictDepBuilds: true` | pnpm lets you allow only reviewed dependency build scripts and fail on unexpected new ones |

Practical rule:

- For **npm**, `ignore-scripts=true` is usually the strongest native baseline when the project can tolerate it.
- For **pnpm**, do not blindly add `ignore-scripts` on top of `allowBuilds`, because that would also block the reviewed scripts you intentionally allowed.

## Important cautions

- Do **not** copy `allowBuilds` blindly across projects. Review which packages genuinely need install/build scripts.
- Do **not** force `--ignore-scripts` in CI for pnpm when you already rely on `allowBuilds`, unless you are intentionally using it only as a temporary discovery step.
- Treat `npx` and `pnpm dlx` as high-risk shortcuts. Prefer preinstalled, versioned tools.
- For private packages, add scoped registry rules to prevent dependency confusion.

## Next step

Use `agent-prompt.md` as the starting context for the other agent, then decide whether the final reusable skill should also ship these snippets as `assets/`.
