# Prompt for another agent

Use this prompt in another session to help an agent create a **global reusable skill** for package-install security.

```text
I want you to create a reusable global skill for package-install and dependency-update security across multiple projects.

Context:
- The source material comes from a previous security hardening review done in another repository, but I want to extract the reusable behavior into a global skill that can be applied across projects.
- The final result should NOT be “skill only”. I want a hybrid outcome: the skill should teach agent behavior and decision-making, while reusable config snippets/templates should provide real enforcement in repos and CI.

What this previous review established:

1. Package manager compatibility
- pnpm 10.33.0 already supports the important supply-chain security settings we need.
- npm should be 11+ if we want to rely on `allow-git`, `min-release-age`, and `ignore-scripts` together.

2. Global machine-level baseline
- Global npm config should include:
  - `ignore-scripts=true`
  - `allow-git=none`
  - `min-release-age=3`
  - `save-exact=true`
- Global pnpm config should include:
  - `minimumReleaseAge: 4320` (3 days, minutes)
  - `trustPolicy: no-downgrade`
  - `blockExoticSubdeps: true`
  - `saveExact: true`
- Important: `allowBuilds` should NOT be treated as a machine-global default because it is project-specific.

3. Project-level baseline
- Project config must duplicate the relevant security posture because global dotfiles only protect one workstation, while repo files protect:
  - teammates
  - Docker builds
  - CI/CD
  - Cloud Build / Buildpacks
- For pnpm projects, the repo-level policy should live in `pnpm-workspace.yaml`.
- For npm projects, the repo-level policy should live in `.npmrc`.

4. pnpm-specific decisions
- Prefer `allowBuilds` + `strictDepBuilds`.
- Do not use deprecated `onlyBuiltDependencies` as the main recommendation.
- We treated `allowBuilds` as a reviewed allowlist for packages like `sharp` or `esbuild` that genuinely need install/build scripts.
- `strictDepBuilds: true` is important because it turns unexpected build-script attempts into hard failures.
- `trustPolicy: no-downgrade` should be explained clearly in the skill: it blocks installation when a newer version of a package has weaker publish-time trust evidence than older versions.
- `blockExoticSubdeps: true` should also be explained clearly: it blocks transitive dependencies from coming from git URLs or raw tarball URLs instead of the normal registry path.

5. Exact versioning policy
- We concluded that `saveExact` / `save-exact` is a good default for apps/services/tools.
- The goal is to make future installs save exact versions instead of `^` ranges.
- Existing caret ranges in old `package.json` files should be treated as a separate migration step, not silently rewritten as part of every install.

6. CI / Cloud Build behavior
- CI should use deterministic installs:
  - pnpm: `pnpm install --frozen-lockfile`
  - npm: `npm ci`
- Repo files must be the main enforcement mechanism in CI.
- Do not blindly recommend `--ignore-scripts` for pnpm CI if the project genuinely needs packages such as `sharp` or `esbuild` to build. Prefer `allowBuilds` + `strictDepBuilds`.
- Buildpacks are not automatically less secure than Dockerfiles. The recommendation was: do not migrate from Buildpacks to Docker only for “consistency” unless more control is actually needed.

7. Risk areas the skill should explicitly address
- `npx` and `pnpm dlx` are risky because they execute code fetched on the fly.
- The skill should teach the agent to prefer preinstalled, versioned, lockfile-backed tools over ad-hoc execution.
- The skill should also consider dependency confusion and recommend scoped private registries when relevant.

8. What the skill should do
- Detect whether the repo uses pnpm or npm.
- Verify version compatibility before recommending settings.
- Explain what belongs in global config vs project config vs CI.
- Refuse blind dependency installation/update advice.
- Recommend exact versions for future installs.
- Warn when a project is using deprecated pnpm config patterns.
- Generate the right file patches or reusable snippets.

9. What the skill should NOT do
- It should not pretend that guidance alone is enough.
- It should not replace real repo config.
- It should not force the same `allowBuilds` map on every project.
- It should not advise copy/paste security without compatibility review.

What I want from you:
1. Design the reusable skill itself.
2. Keep the SKILL.md concise and operational.
3. Put supporting config templates/snippets into assets.
4. Add any references/docs needed for explanation.
5. Make the skill portable across projects, not tied to the original repo.
6. Prefer English for the skill artifacts.

If helpful, use these reusable snippet files as source material:
- global-npmrc
- global-pnpm-config.yaml
- project-npmrc
- project-pnpm-workspace.yaml
- cloud-build-notes.md

Important framing:
- This came from a real hardening exercise in another repo.
- I want the reusable abstraction now.
- I care about both awareness (agent behavior) and enforcement (actual config files).
```
