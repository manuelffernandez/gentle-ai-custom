---
name: package-security
description: "Trigger: package install, dependency update, npm install, pnpm install, supply chain, security audit. Enforce package security posture across npm/pnpm projects."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## Activation Contract

Use when:
- User installs, updates, or audits dependencies
- A project needs supply-chain hardening
- Reviewing or setting up npm/pnpm config files
- Evaluating CI pipeline install commands

Do NOT apply templates blindly. Verify package manager, version, and project context first.

## Hard Rules

1. **Detect first**: identify npm vs pnpm and check minimum version (npm 11+, pnpm 10.33+)
2. **Never force `allowBuilds`**: it is project-specific and requires manual review — never copy from another repo
3. **Refuse blind advice**: if the user says "just install X", surface the risk before complying
4. **Three layers**: distinguish global (one machine), project (repo/CI/Docker), and CI command
5. **pnpm ≠ npm**: pnpm has `allowBuilds` + `strictDepBuilds`; npm does not — strategies differ
6. **`npx` / `pnpm dlx` are high risk**: prefer preinstalled, versioned, lockfile-backed tools

## Decision Gate

| Layer | File | Protects |
|---|---|---|
| Machine-global | `~/.npmrc` / `~/.config/pnpm/config.yaml` | Local ad-hoc installs |
| Project | `.npmrc` / `pnpm-workspace.yaml` | All devs, Docker, CI |
| CI command | Step config / Dockerfile | Deterministic reproducibility |

## Execution Steps

1. Detect package manager and version.
2. Map the request to the right config layer(s).
3. Copy the relevant template from `assets/` and adapt — never paste raw template values.
4. **pnpm only**: derive `allowBuilds` by asking which packages genuinely need build scripts; or run a discovery install with scripts blocked first.
5. **CI**: always use `pnpm install --frozen-lockfile` or `npm ci` — never bare `install`.
6. **Private packages**: add scoped registry rules to prevent dependency confusion.
7. Flag deprecated `onlyBuiltDependencies` and recommend `allowBuilds` + `strictDepBuilds` instead.

## Output Contract

- Adapted config snippet (not a raw template copy)
- Explanation of each setting added
- Warning if version requirements are not met
- Reminder about `allowBuilds` entries that still need human review

## Assets

- `assets/global-npmrc` — machine-global npm baseline template
- `assets/global-pnpm-config.yaml` — machine-global pnpm baseline template
- `assets/project-npmrc` — project-level npm template
- `assets/project-pnpm-workspace.yaml` — project-level pnpm template
- `assets/ci-guide.md` — CI integration guidance (platform-agnostic)
