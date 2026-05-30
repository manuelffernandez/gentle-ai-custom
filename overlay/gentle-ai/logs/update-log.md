# Gentle AI Overlay Update Log

> Este archivo registra decisiones e hitos del mantenimiento del overlay. No es la fuente autoritativa del último upstream mantenido; esa responsabilidad vive en `overlay/gentle-ai/state/upstream-state.json`.

## 2026-05-29 — Judgment Day round 1 fixes

Adversarial dual review (Judge A + Judge B en paralelo) sobre los 5 commits de hardening (`bfbc363..HEAD`). Confirmados 2 CRITICAL y 11 WARNING (real). Esta entrada documenta los fixes aplicados.

Fixes de código en `apply-gentle-ai-policy.{sh,ps1}`:

- **CRITICAL — PS1 `Remove-ExactOnce` reemplazaba TODAS las ocurrencias**: usaba `String.Replace($Old, $New)` (multi-occurrence) mientras bash usaba `text.replace(old, new, 1)` (single-occurrence). Reescrita con `IndexOf` + `Substring` para reemplazar solo la primera ocurrencia, mirroring la semántica de Python.
- **CRITICAL — Preflight sanitization: parity divergente entre bash y PS1**: bash eliminaba las dos líneas (Chained PR strategy + Review budget) como un único bloque con un `replace_once`; PS1 las eliminaba como dos llamadas independientes a `Remove-ExactOnce`. Reorganizado PS1 para mirrorear bash: una sola llamada con el bloque concatenado por `"`n"`, lo que iguala las semánticas de fallo (ambos ahora fallan con `missing expected text: preflight PR/review choices` si upstream las reordena).
- **PS1 `ConvertTo-Json` escapaba non-ASCII a `\uXXXX`**: bash/Python usa `ensure_ascii=False` y escribe UTF-8 crudo. Esto causaba diff byte-por-byte en `opencode.json` según qué helper corriera último, dañando trazabilidad e idempotencia cross-platform. Agregada función `Unescape-NonAsciiUnicode` que post-procesa la salida de `ConvertTo-Json` con regex para devolver los `\uXXXX` a sus caracteres UTF-8, produciendo output byte-idéntico al bash.
- **Snapshot escrito antes de sanitizar**: si `sanitize_prompt` fallaba, el `*.last.md` ya había sido sobreescrito con contenido que no se pudo procesar, perdiendo el last-known-good. Reordenado en ambos scripts: sanitización primero, escritura del snapshot solo si la sanitización pasó.
- **Em-dash (U+2014) en bash vs hyphen ASCII en PS1**: el bash imprimía `snapshots — new:` y `WARNING — keep skills missing` con em-dash; PS1 con hyphen; runbook y README con em-dash. Tres convenciones inconsistentes. Estandarizado a hyphen ASCII en bash, runbook y README para parity con PS1 y mejor portabilidad de grep/terminal.
- **`SystemExit(...)` no producía el prefijo `ERROR:` que docs prometían**: agregada función `die()` en Python y `Die` en PowerShell que imprimen `ERROR: <msg>` a stderr y salen con código 1. Reemplazadas todas las llamadas a `SystemExit` / `throw` user-facing por `die()` / `Die`. Ahora docs y output realmente coinciden.
- **PS1 no flageaba sobrescritura de agentes non-object como `created_overrides`**: bash agregaba a `created_overrides` siempre que el agente no fuera dict (key faltante O key con tipo distinto a object), pero PS1 solo agregaba cuando el key directamente no existía. Parity fix: PS1 ahora siempre pushea cuando el agente no es `PSCustomObject`, igualando bash.
- **Recovery silenciosa de snapshot stale**: cuando un `.overlay.md` faltaba en disco y se recuperaba del `*.last.md`, el script lo reportaba como éxito normal sin avisar que el snapshot puede pre-datar el upstream actual. Agregado un `WARNING` per-orchestrator durante la corrida + un bloque `NOTE` en el Summary cuando `RECOVERED_COUNT > 0` que recomienda correr `gentle-ai sync` para refrescar.
- **Summary engañoso en steady-state**: cuando todos los orchestrators ya estaban aplicados, el summary mostraba `generated: 0, recovered: 0, skipped: 0, snapshots 0/0/0` — indistinguible de "no procesé nada". Agregado contador `kept_count` / `KeptCount` que se incrementa en la rama keep, surfaceado como `orchestrators kept (already applied): N`.
- **Path traversal teórico via agent key**: agregada función `safe_snapshot_key` / `Assert-SafeSnapshotKey` que rechaza keys con `/`, `\`, `..`, o caracteres nulos antes de usarlos en paths de snapshot/overlay. Defensa en profundidad ante upstream malformado.

Fixes en docs:

- `README.md`: removido hardcodeo de `("General" y "Explore")`; ahora referencia `agent_overrides` en la policy. Actualizado bloque "Qué reporta el script" para usar hyphens, agregar la fila de `kept`, mencionar el `NOTE` de stale recovery, y agregar la fila de `post-write verification failed`.
- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`: tabla de señales reescrita con hyphens, agregadas filas para `kept`, `WARNING recovering ... may pre-date current upstream`, `ERROR: post-write verification failed: orchestrator ...`, `ERROR: OpenCode config ... is not valid JSON`, y `ERROR: unsafe agent key`. Sección de `external-single-active` reforzada: la "Con" crítica ahora dice explícitamente que el usuario ejecuta indefinidamente la versión sanitizada anterior, no solo que pierde "drift detection".
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`: removida la duplicación entre "Hard Rules" y "Update-Type Triage" (la asercion sobre overlay roto post-sync ahora vive solo en la triage table). Sección `Hardening option: external-single-active strategy` ahora aclara "directly under (subdirectories are ignored)" y enfatiza el riesgo crítico de quedar ejecutando sanitización vieja.
- `AGENTS.md`: agregado one-liner "Rule 3 = live state. Rule 4 = decision history. Both deliverables are required..." arriba del párrafo explicativo, para que un lector que scanea rápido capte la distinción sin tener que leer la prosa.
- `overlay/gentle-ai/logs/update-log.md`: amenda implícitamente la entrada anterior — el "Apply script hardening" del mismo día había omitido mencionar el cambio en root `README.md` (commit `4dab640`), violando rule 4 el día que se introdujo. Esta entrada lo registra.

Findings descartados o diferidos:

- StrictMode property access en PS1 (Judge A CRITICAL): contradicho por Judge B con razonamiento técnico correcto (PS 5.1 StrictMode v3+ no lanza para propiedades inexistentes en `PSCustomObject`). Downgrade a INFO. No requiere fix.
- `gentle-orchestrator` en `orchestrator_agent_keys` como "user-specific" (Judge B WARNING): falso. Verificado que upstream emite `gentle-orchestrator` para todo usuario via `internal/components/sdd/inject.go:741` (`agentsMap["gentle-orchestrator"]`). La policy es correcta.
- Fsync, CRLF en Windows, hash check del overlay file, stderr routing, commit pin en SKILL: registrados como SUGGESTION/theoretical, no aplicados en esta ronda.

Verificación:

- `bash apply-gentle-ai-custom.sh all` corre limpio idempotente: `kept: 4`, `topology warnings: 0`, hyphens consistentes en summary.
- Recovery fallback verificado con re-borrado de `gentle-orchestrator.overlay.md` y observando el nuevo WARNING + NOTE de stale.

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
