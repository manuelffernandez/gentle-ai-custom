# CI Integration Guide

## Goal

Make build automation follow the same package security baseline as local development,
regardless of the CI platform (GitHub Actions, GitLab CI, Jenkins, Cloud Build, etc.).

## Core Rules

1. Put the real policy in repo files — CI reads them automatically without depending on developer dotfiles.
2. Use deterministic installs:
   - pnpm: `pnpm install --frozen-lockfile`
   - npm: `npm ci`
3. Let the project config drive package security:
   - pnpm reads `pnpm-workspace.yaml`
   - npm reads `.npmrc`

## pnpm guidance

- Prefer `allowBuilds` + `strictDepBuilds` for legitimate native/build-script dependencies.
- This is the key difference from npm: pnpm has a **native allowlist model** for dependency build scripts.
- Do **not** add `--ignore-scripts` in CI when already using `allowBuilds` — it would also block
  the reviewed scripts you explicitly allowed.
- Discovery workflow:
  1. Run an install with scripts blocked.
  2. Identify what breaks.
  3. Review those packages.
  4. Add only the legitimate ones to `allowBuilds`.

## npm guidance

- `npm ci` automatically reads the repo `.npmrc`.
- npm has no first-class equivalent to pnpm's `allowBuilds` + `strictDepBuilds`.
- `ignore-scripts=true` is the strongest built-in baseline when the project can tolerate it.
- When it cannot: combine exact versions, `allow-git=none`, `min-release-age`, deterministic installs,
  per-package review, and if needed third-party tools (`@lavamoat/allow-scripts`, `npq`, `sfw`).

## Container and cloud builds

Repo-level config files work regardless of how the build is containerized — Docker, Buildpacks,
or any platform that reads the project directory. No CI-platform-specific changes are needed
beyond using deterministic install commands and ensuring the config file is present in the repo.

## Dependency confusion

For private packages, add scoped registry rules to prevent name hijacking:

```ini
@your-scope:registry=https://your-private-registry.example.com/
```
