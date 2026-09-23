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
   - verify the config is copied into the install step's working directory/build context before installation
4. Pin the package-manager version and verify that exact CLI supports every configured policy feature. Major-version labels alone do not guarantee support for all settings.

## pnpm guidance

- Prefer `allowBuilds` + `strictDepBuilds` for legitimate native/build-script dependencies when the effective pnpm supports both (`allowBuilds`: 10.26+; `strictDepBuilds`: 10.3+).
- Both managers support reviewed script policies in supported releases, but their defaults differ: pnpm blocks unlisted dependency builds and `strictDepBuilds` can fail the install; npm's `allowScripts` policy is advisory unless `strict-allow-scripts` is supported and enabled.
- Verify the effective CLI version and support for each setting. Do not assume the same approval policy or defaults across npm and pnpm.
- Do **not** set `ignore-scripts=true` or pass `--ignore-scripts` when relying on `allowBuilds` — it would also block the reviewed scripts you explicitly allowed.
- Discovery workflow:
  1. Run an install with scripts blocked.
  2. Identify what breaks.
  3. Review those packages.
  4. Add only the legitimate ones to `allowBuilds`.

## npm guidance

- `npm ci` automatically reads the repo `.npmrc`.
- npm provides a reviewed `allowScripts` policy, `npm approve-scripts`, and `strict-allow-scripts` in supported releases. The policy is advisory by default, so verify exact CLI support and enable strict enforcement deliberately.
- `ignore-scripts=true` is a conservative baseline. It takes precedence over reviewed grants, so remove it if adopting the strict reviewed policy; do not describe these as composable controls.
- If a project cannot use either mode, document the exception and retain deterministic installs, per-package review, and any suitable additional controls.

## Container and cloud builds

Repo-level config only works when the install process can see it. For containers and buildpacks,
verify the relevant config is copied or mounted into the build context and is present in the
package manager's working directory before the install command runs.

## Dependency confusion

For private packages, add scoped registry rules to prevent name hijacking:

```ini
@your-scope:registry=https://your-private-registry.example.com/
```
