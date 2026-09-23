---
name: package-security
description: "Trigger: npm/pnpm install, dependency/lockfile update, supply-chain review, Docker/CI/devcontainer install, package hooks, install scripts, dependency automation, or Node version pins. Harden package supply chains."
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
- Dependency manifests and lockfiles when adding, removing, or updating packages

Never read, search, or inspect credential files, including `.env` files or real `.npmrc` files. Use placeholder templates, documentation, schemas, and redacted user-run output instead.

Do NOT apply templates blindly. Detect what actually exists in the project first — see Execution Steps.

## Hard Rules

1. **Detect first**: identify npm vs pnpm and its effective version. Do not assume a major-version label guarantees every setting or command; verify each required feature against that exact CLI.
2. **Never force `allowBuilds`**: it is project-specific and requires manual review — never copy from another repo
3. **Refuse blind advice**: if the user says "just install X", surface the risk before complying
4. **Three layers**: distinguish global (one machine), project (repo/CI/Docker), and CI command
5. **pnpm ≠ npm**: pnpm has `allowBuilds` + `strictDepBuilds`; npm has `allowScripts`, `npm approve-scripts`, and `strict-allow-scripts` in supported releases. npm's reviewed-script policy is advisory by default; verify support and enable strict enforcement deliberately.
6. **`npx` / `pnpm dlx` are high risk**: prefer preinstalled, versioned, lockfile-backed tools

## Decision Gate

| Layer | File | Protects |
|---|---|---|
| Machine-global | `~/.npmrc` / `~/.config/pnpm/config.yaml` | Local ad-hoc installs |
| Project | `.npmrc` / `pnpm-workspace.yaml` | All devs, Docker, CI |
| CI command | Step config / Dockerfile | Deterministic reproducibility |

## Execution Steps

1. **Detect package manager and effective version.** Check which CLI actually runs in local, CI, and container contexts. Verify support for each setting/command; do not treat npm 11+ or pnpm 10.33+ as a guarantee for every feature.
2. **Inventory what the project actually has.** Check which layers are present before surfacing any recommendation:
   - Config files: `.npmrc`, `pnpm-workspace.yaml`
   - CI config files in `.github/workflows/`, `.gitlab-ci.yml`, etc.
   - Container files: `Dockerfile`, `.devcontainer/`
   - Hook files: `.husky/`, `lefthook.yml`
   - Automation: `dependabot.yml`, `renovate.json`
   - Node version pins: `.nvmrc`, `.node-version`, `.tool-versions`
   - Only surface guidance for layers that exist OR are being introduced right now.
3. **Map the request to the right config layer(s).** Use the Decision Gate table above.
4. **Adapt the relevant template from `assets/`.** Never paste raw template values. The npm baseline uses `ignore-scripts=true`; an opt-in reviewed `allowScripts` policy requires a supported CLI and strict enforcement. Do not combine it with `ignore-scripts=true`, which takes precedence and blocks the approved scripts too.
5. **Build-script approvals**: require human review before granting scripts for either manager. For npm, approve only reviewed packages through `allowScripts` / `npm approve-scripts`, then enforce with supported `strict-allow-scripts`; never blanket-approve pending packages. For pnpm, derive `allowBuilds` by asking which packages genuinely need scripts or run a discovery install with scripts blocked. Feature floors are setting-specific (`allowBuilds`: pnpm 10.26; `strictDepBuilds`: pnpm 10.3); verify effective support and never invent grants.
6. **CI / Docker / devcontainer**: always use `pnpm install --frozen-lockfile` or `npm ci` — never bare `install`.
7. **Lifecycle hooks** (`postinstall`, `prepare`, `preinstall`): flag any invocation of `npx`/`pnpm dlx` as high risk; recommend pinned, lockfile-backed alternatives.
8. **Dependency automation** (Dependabot/Renovate): verify update scope, reviewer assignment, and that PRs cannot auto-merge without review.
9. **Node version changes** (`.nvmrc`, `.node-version`): confirm the target Node version ships with a supported npm version; flag downgrades.
10. **Private packages**: add scoped registry rules to prevent dependency confusion.
11. Flag deprecated pnpm `onlyBuiltDependencies` and recommend `allowBuilds` + `strictDepBuilds` only when the effective pnpm supports both.

## Output Contract

- Adapted config snippet (not a raw template copy)
- Explanation of each setting added
- Exact CLI feature/version checks and any unsupported settings
- Reminder that npm `allowScripts` and pnpm `allowBuilds` entries require human review

## Assets

- `assets/global-npmrc` — machine-global npm baseline template
- `assets/global-pnpm-config.yaml` — machine-global pnpm baseline template
- `assets/project-npmrc` — project-level npm template
- `assets/project-pnpm-workspace.yaml` — project-level pnpm template
- `assets/ci-guide.md` — CI integration guidance (platform-agnostic)
