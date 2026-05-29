# Gentle AI Overlay Update Log

> Este archivo registra decisiones e hitos del mantenimiento del overlay. No es la fuente autoritativa del último upstream mantenido; esa responsabilidad vive en `overlay/gentle-ai/state/upstream-state.json`.

## 2026-05-29 — Apply script hardening + maintainer skill v1.1

Descubrimiento que motivó el cambio:

- `gentle-ai sync` resetea incondicionalmente los prompts inline de los orchestrators y reinstala todas las skills (verificado contra `internal/components/sdd/profiles.go` → `ResolveProfileStrategy` en upstream; con `~/.config/opencode/profiles/` vacío, devuelve `generated-multi` → `PreserveOpenCodeOrchestratorPrompt = false`).
- La versión previa del script tenía un bug silencioso: si `opencode.json` apuntaba a un `{file:...}` cuyo target no existía en disco, el script skipeaba con un mensaje informativo y dejaba OpenCode sin poder cargar el orchestrator.
- La skill del maintainer no distinguía qué tipo de update había ocurrido (brew / sync / reinstall), no tenía step explícito de re-aplicar el script ni de verificar el estado post-corrida, y no detectaba drift de topología.

Cambios en los scripts (`apply-gentle-ai-policy.sh` + `.ps1`, paridad mantenida):

- Recuperación desde snapshot: si el `{file:...}` target falta pero existe `*.last.md`, re-sanitiza desde el snapshot. Si no hay snapshot, falla con un mensaje accionable que pide `gentle-ai sync` para resetear a inline.
- Drift tracking de snapshots: contadores `new` / `changed` / `unchanged` en el summary; recordatorio de `git diff overlay/gentle-ai/snapshots/` cuando hay drift.
- Detección de drift de topología: warnings cuando un orchestrator matchea solo por prefijo (sin estar en `orchestrator_agent_keys`), cuando una key esperada falta en upstream, y cuando un `agent_override` creó un stub desde cero.
- Validación explícita del parseo JSON con error accionable hacia los backups de `~/.gentle-ai/backups/`.
- Verificación post-write: re-lee `opencode.json` después de escribir y confirma que los model overrides y las refs `{file:...}` de orchestrators persistieron.
- Bloque `Summary:` estructurado al final con todos los contadores.
- Las skills "keep" ausentes se agregan en un bloque `WARNING` agregado al final, en lugar de quedarse perdidas en el log inline.

Cambio en la política (`gentle-ai-policy.json`):

- Se listaron `sdd-orchestrator-mixed`, `sdd-orchestrator-vertex` y `sdd-orchestrator-vertex-claude` explícitamente en `orchestrator_agent_keys`. El prefijo `sdd-orchestrator` queda como tripwire real para detectar orchestrators nuevos agregados upstream que la policy todavía no conoce.

Cambio en la skill (`.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, v1.0 → v1.1):

- Nuevo step obligatorio "Update-Type Triage" al inicio: el agente debe determinar si el usuario hizo `brew upgrade`, `gentle-ai sync` o reinstalación TUI antes de cualquier otra acción.
- Decision Gates expandidas con escenarios sync / reinstall / topology / snapshot drift / broken state / sanitizer fail.
- Step explícito de correr el script + step de verificación on-disk (skills prunidas ausentes, model overrides correctos, refs de orchestrators apuntando a archivos existentes).
- Output Contract amplificado con señales de topology / verification / recovery.
- Nueva sección "Hardening option" documentando la estrategia `external-single-active` con tradeoffs.

Cambio en el runbook (`maintain-upstream-overlay.md`):

- Tabla mapeando cada señal del summary del script (topology, snapshot drift, recovery, broken state, post-write verification fail, keep skills missing) a una acción concreta.
- Sección de `external-single-active` explicando el mecanismo upstream, sus tradeoffs, y por qué se deja opt-in (no activar sin pedido explícito porque perdés visibilidad de drift upstream).

Verificación:

- Corridas idempotentes reportan `topology warnings: 0` después del update de policy.
- Test manual del fallback: borré `~/.config/opencode/prompts/sdd/orchestrators/gentle-orchestrator.overlay.md`, corrí el script, recuperó desde `gentle-orchestrator.last.md` y reportó `orchestrators recovered from snapshot: 1`.

## 2026-05-29 — Documented gentle-ai update types and mandatory script re-run rule

- Agregada sección "Tipos de actualización de Gentle AI y su impacto en el overlay" al runbook con tabla de impacto por mecanismo (brew / sync / reinstall) y regla operativa invariante.
- Agregada sección "Update flow" a `AGENTS.md` con la misma regla en forma resumida para que cualquier agente trabajando en el repo la encuentre.
- Documentación basada en lectura directa del código upstream (`internal/cli/sync.go` + `internal/components/sdd/inject.go` + `internal/components/sdd/profiles.go`).

## 2026-05-29 — Fixed apply script execute bit and bash invocation

- `apply-gentle-ai-policy.sh` estaba commiteado con modo `100644` en git. Cualquier clone fresco fallaba con "Permission denied" porque el entrypoint llamaba al helper directamente sin `bash`.
- Cambiado `apply-gentle-ai-custom.sh` para invocar el helper con `bash` explícito (inmune al bit de ejecución para siempre).
- Corregido el modo del helper a `100755` en el index de git.
- Commiteados también los snapshots `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/*.last.md` que estaban en el working tree como untracked. Son estado de referencia del upstream — versionarlos permite que `git diff` muestre el drift entre actualizaciones.

## 2026-05-29 — Intent / policy / state / log split

- Eliminado `overlay/gentle-ai/prompts/audit-gentle-ai-update.md` por quedar reemplazado por la skill del maintainer.
- Agregado `overlay/gentle-ai/policy/maintenance-intent.md` como fuente de verdad semántica del criterio del usuario.
- Agregado `overlay/gentle-ai/state/upstream-state.json` como frontera operativa de la última versión/commit mantenido.
- Seed inicial de `upstream-state.json` con el estado actual del upstream: `v1.32.0-6-g412eed3` sobre `v1.32.0` (`412eed3d39defb2f955a63e21ca13cef4df358c9`).
- La skill `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` pasó a ser version-aware y con gate humana explícita antes de mutar intención/política por cambios upstream relevantes.
- El runbook ahora distingue explícitamente intento, política, estado y log.

## 2026-05-29 — Removed compatibility alias layer

- Deleted `inject-skills.sh` and `inject-skills.ps1`.
- Moved the full installation + wrapper-rendering flow directly into `apply-gentle-ai-custom.sh` and `apply-gentle-ai-custom.ps1`.
- Kept `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh/.ps1` as internal helpers invoked only when targets include `opencode` or `claude`.
- Updated docs to reflect that the only public entrypoint pair is now `apply-gentle-ai-custom.sh/.ps1`.

## 2026-05-29 — Unified custom layer and dynamic orchestrator generation

- Added `apply-gentle-ai-custom.sh` and `.ps1` as canonical entrypoints.
- Converted `inject-skills.sh` and `.ps1` into compatibility aliases that still execute the full custom-layer workflow.
- Reframed `gentle-ai-custom` as a unified installation + depuration + maintenance layer for Gentle AI.
- Reworked the overlay helper so it no longer depends on a static repo-owned orchestrator prompt file.
- The helper now reads inline orchestrators from `opencode.json`, snapshots them per agent, sanitizes them, and generates `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md` files.
- Deleted the obsolete static derived prompt artifact and the obsolete single-snapshot placeholder.
- Added `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` as the runtime entrypoint for overlay maintenance.
- Added `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` as the human maintenance runbook.

## 2026-05-28 — Script fixes, parity alignment, and agent_overrides

- Fixed `apply-gentle-ai-policy.sh` Python guard: resolved `PYTHON_CMD="${PYTHON:-python3}"` at startup; guard and both inline Python calls now use `${PYTHON_CMD}` consistently.
- Fixed `apply-gentle-ai-policy.ps1` `Get-PromptContentForSnapshot`: added `-PathType Leaf` to `Test-Path` so directories are not mistakenly matched as prompt files.
- Fixed `apply-gentle-ai-policy.ps1` newline parity: replaced `[Environment]::NewLine` (CRLF on Windows) with `` "`n" `` in both snapshot-write and config-write paths; bash counterpart already writes LF.
- Fixed missing-keep warning separator parity: bash now uses `$(IFS=', '; echo "${missing_keep[*]}")` (comma-space), matching PS1's `$MissingKeep -join ', '`.
- Cleaned up `AGENTS.md` section 2: split the single shared parity bullet list into two explicit sub-lists, one per script pair (`inject-skills.*` vs `apply-gentle-ai-policy.*`), removing ambiguity about which items apply to which pair.
- Added `agent_overrides` to `gentle-ai-policy.json`: `general=openai/gpt-5.4/high`, `explore=google-vertex/gemini-3.1-pro-preview/high`.
- Updated both scripts to apply `agent_overrides` from policy atomically (single write) alongside the prompt redirect; bash uses `config_changed` flag, PS1 uses `$ConfigChanged` flag.
- Fixed `overlay/gentle-ai/README.md` Windows path example: changed relative `./overlay/...` to absolute `~\Documentos\gentle-ai-custom\...`.

## 2026-05-28 — Baseline overlay/control-plane bootstrap

- Created overlay structure under `overlay/gentle-ai/`.
- Persisted keep/prune baseline in `policy/gentle-ai-policy.json`.
- Added orchestrator derivation policy (`policy/orchestrator-policy.md`).
- Added upstream-audit prompt (`prompts/audit-gentle-ai-update.md`).
- Added derived OpenCode orchestrator prompt with PR/budget/chained-PR workflow removed.
- Added initial upstream snapshot placeholder; the first policy-apply run refreshes it with the effective upstream prompt content before redirect.
- Added paired apply scripts (`scripts/apply-gentle-ai-policy.sh` + `.ps1`) with parity commitments.

Baseline decisions:

- KEEP: `_shared`, SDD core phases, `skill-registry`, `skill-creator`, `skill-improver`, `cognitive-doc-design`, `comment-writer`, `judgment-day`, `go-testing`.
- PRUNE: `branch-pr`, `chained-pr`, `issue-creation`, `work-unit-commits`.
- Historical baseline: the first version of the overlay redirected `gentle-orchestrator` to a repo-owned derived prompt file. This was replaced on 2026-05-29 by dynamic per-orchestrator generation from inline prompts.
