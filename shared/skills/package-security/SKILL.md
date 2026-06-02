---
name: package-security
description: "Trigger: package install, dependency update, npm install, pnpm install, supply chain, security audit, Dockerfile with npm/pnpm, CI pipeline with install steps, shell script or Makefile invoking npm/pnpm, package.json lifecycle hooks, dependabot or renovate config, devcontainer with node, .nvmrc or .node-version change, git hooks with install steps, justfile with npm/pnpm. Enforce package security posture across npm/pnpm projects — including any file that directly or indirectly declares, executes, or automates a package installation."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.2"
---

## Activation Contract

Load this skill whenever the agent touches a file that directly or indirectly declares, executes, or automates a package installation — regardless of whether the user explicitly asks about security.

**Direct intent**
- User installs, updates, or audits dependencies
- A project needs supply-chain hardening
- Reviewing or setting up npm/pnpm config files (`.npmrc`, `pnpm-workspace.yaml`)

**Contextual — file being edited**
- `Dockerfile` / `Containerfile` with `RUN npm install`, `RUN pnpm install`, `RUN npm ci`, `RUN pnpm add`, `RUN npx`, or `RUN pnpm dlx`
- CI config files (`.github/workflows/*.yml`, `.gitlab-ci.yml`, `Jenkinsfile`, `.circleci/config.yml`, `azure-pipelines.yml`) that include install steps
- Shell scripts (`.sh`, `.bash`) or `Makefile` targets that call `npm`, `pnpm`, `npx`, or `pnpm dlx`
- `docker-compose.yml` / `compose.yaml` referencing a build context that contains any of the above
- `package.json` when editing `scripts.postinstall`, `scripts.preinstall`, `scripts.prepare`, or any script that invokes `npx` / `pnpm dlx`
- Dependency automation configs: `.github/dependabot.yml`, `renovate.json`, `.renovaterc`, `.renovaterc.json`
- Dev container configs: `.devcontainer/devcontainer.json`, `.devcontainer/Dockerfile`
- Node version declarations: `.nvmrc`, `.node-version`, `.tool-versions` (affects which npm ships with Node)
- Git hook configs: `.husky/`, `lefthook.yml`, `.lefthook.yml`, or any file under `.git/hooks/` that calls npm/pnpm
- `justfile` targets that call `npm`, `pnpm`, `npx`, or `pnpm dlx`
- `.env` / `.env.*` files containing `NPM_TOKEN`, `NODE_AUTH_TOKEN`, or `PNPM_TOKEN`

Do NOT apply templates blindly. Detect what actually exists in the project first — see Execution Steps.

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

1. **Detect package manager and version.** Identify npm vs pnpm and confirm minimum version (npm 11+, pnpm 10.33+).
2. **Inventory what the project actually has.** Check which layers are present before surfacing any recommendation:
   - Config files: `.npmrc`, `pnpm-workspace.yaml`
   - CI config files in `.github/workflows/`, `.gitlab-ci.yml`, etc.
   - Container files: `Dockerfile`, `.devcontainer/`
   - Hook files: `.husky/`, `lefthook.yml`
   - Automation: `dependabot.yml`, `renovate.json`
   - Node version pins: `.nvmrc`, `.node-version`, `.tool-versions`
   - Only surface guidance for layers that exist OR are being introduced right now.
3. **Map the request to the right config layer(s).** Use the Decision Gate table above.
4. **Adapt the relevant template from `assets/`.** Never paste raw template values.
5. **pnpm only**: derive `allowBuilds` by asking which packages genuinely need build scripts; or run a discovery install with scripts blocked first.
6. **CI / Docker / devcontainer**: always use `pnpm install --frozen-lockfile` or `npm ci` — never bare `install`.
7. **Lifecycle hooks** (`postinstall`, `prepare`, `preinstall`): flag any invocation of `npx`/`pnpm dlx` as high risk; recommend pinned, lockfile-backed alternatives.
8. **Dependency automation** (Dependabot/Renovate): verify update scope, reviewer assignment, and that PRs cannot auto-merge without review.
9. **Node version changes** (`.nvmrc`, `.node-version`): confirm the target Node version ships with a supported npm version; flag downgrades.
10. **Private packages**: add scoped registry rules to prevent dependency confusion.
11. Flag deprecated `onlyBuiltDependencies` and recommend `allowBuilds` + `strictDepBuilds` instead.

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
