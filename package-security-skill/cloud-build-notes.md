# CI Integration Notes

## Goal

Make build automation follow the same package security baseline as local development.

## Rules

1. Put the real policy in repo files so CI can read it without depending on developer dotfiles.
2. Use deterministic installs:
   - pnpm: `pnpm install --frozen-lockfile`
   - npm: `npm ci`
3. Let the project config drive package security:
   - pnpm reads `pnpm-workspace.yaml`
   - npm reads `.npmrc`

## pnpm guidance

- Prefer `allowBuilds` + `strictDepBuilds` for legitimate native/build-script dependencies.
- This is the key difference from npm: pnpm has a **native allowlist model** for dependency build scripts.
- Do **not** add `--ignore-scripts` by default in pnpm CI if you are already using `allowBuilds`, because it would also block the reviewed scripts you explicitly intended to allow.
- A good migration/discovery flow is:
  1. try an install with scripts blocked,
  2. identify what breaks,
  3. review those packages,
  4. add only the legitimate ones to `allowBuilds`.
- Buildpacks or Docker-based builds still benefit as long as the repo contains the config file.

## npm guidance

- `npm ci` automatically reads the repo `.npmrc`.
- npm does **not** have a first-class equivalent to pnpm's `allowBuilds` + `strictDepBuilds` for dependency lifecycle scripts.
- Because of that, `ignore-scripts=true` is the strongest built-in baseline when the project can tolerate it.
- If the project cannot tolerate `ignore-scripts=true`, the honest answer is that npm's built-in controls are coarser than pnpm's. In that case, combine:
  - exact versions,
  - `allow-git=none`,
  - `min-release-age`,
  - deterministic installs via `npm ci`,
  - package review before allowing lifecycle scripts,
  - and, if needed, third-party controls such as `@lavamoat/allow-scripts`, `npq`, or `sfw`.
- Practical decision rule for npm:
  - if the app builds fine with `ignore-scripts=true`, keep it;
  - if it breaks and the dependency scripts are truly required, document the exception and compensate with stronger review.

## Dependency confusion

For private packages, add scoped registry rules such as:

```ini
@your-scope:registry=https://your-private-registry.example.com/
```
