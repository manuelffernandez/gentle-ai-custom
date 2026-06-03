# Gentle AI Overlay Update Log

> Este archivo registra decisiones e hitos del mantenimiento del overlay. No es la fuente autoritativa del último upstream mantenido; esa responsabilidad vive en `overlay/gentle-ai/state/upstream-state.json`.

## 2026-06-02 — Reframed root README around an AI-managed Gentle AI customization layer

Razón del cambio:

- El README raíz explicaba bien la mecánica del repo, pero seguía sonando demasiado interno y poco inspirador para alguien que llegue desde afuera.
- Además, parte del wording seguía empujando la idea de un “overlay personalizable”, cuando la realidad actual es más específica: una customización profunda de Gentle AI que se vuelve mantenible gracias a automatización y agentes de IA, no un producto fácil de reconfigurar por cualquiera.

WHAT cambió:

- `README.md`:
  - agregado un posicionamiento inicial más amigable/comercial del repo como capa de customización mantenida con IA sobre `gentle-ai`
  - agregada referencia explícita al repo oficial `https://github.com/Gentleman-Programming/gentle-ai`
  - incorporada una explicación breve de para qué sirve Gentle AI y por qué mejora la experiencia de desarrollo con IA
  - aclarado que este repo conserva la base upstream, depura lo que no encaja con el workflow diario y suma skills/wrappers propios
  - ajustado el mensaje para no sobreprometer reutilización inmediata por terceros ni configurabilidad sencilla: hoy el repo sigue orientado principalmente al flujo del autor y se presenta como una capa mantenible, no como una solución plug-and-play
  - explicitada la visión futura de evolucionar hacia un instalador TUI más personalizado y expandir funcionalidades
  - documentada la razón de usar Go como única fuente de verdad entre wrappers `.sh` y `.ps1`, reutilizando además una dependencia ya presente en el stack por Engram

WHY:

- La presentación tenía que volverse más precisa: este repo no vende configurabilidad inmediata para terceros, sino mantenibilidad operativa sobre una customización compleja de Gentle AI.
- Explicar por qué la automatización vive en Go evita que parezca una elección arbitraria: la decisión responde a paridad cross-platform y reutilización coherente del stack existente.

Verificación:

- Revisión manual de `README.md` para confirmar que la nueva introducción referencia correctamente al upstream, mantiene el contrato operativo del repo y mejora el tono de presentación sin cambiar comportamiento.

## 2026-06-01 — Audited v1.33.2 post-tag docs fix (21634526)

Tipo de update: `git pull` upstream solamente — sin `gentle-ai sync`, sin reinstalación.

WHAT cambió upstream:

- Commit `21634526` (`docs: correct recent release documentation drift`) modifica `README.md`, `docs/agents.md`, `docs/components.md`, `docs/non-interactive.md`, `docs/opencode-profiles.md`, `docs/platforms.md`, `docs/usage.md`, `internal/cli/run.go`, `internal/cli/scope.go`.
- Los cambios en Go son puramente de comentarios: actualizan la terminología para reemplazar referencias a "CLAUDE.md / ~/.claude/" por lenguaje agnóstico al agente ("system prompts / global config directory").

WHY:

- Corrección de documentación post-release. No hay cambio de comportamiento, no hay drift en el base prompt del orchestrator, no hay cambio en invariantes de perfil.

Audit result:

- `base prompt drift: no` — snapshot y metadata siguen alineados.
- `profile phase order`, `profile orchestrator naming`, `profile task scoping invariant`, `base asset injection invariant`: todos `ok`.
- Topología: sin cambios.

Acción tomada:

- Actualizado `upstream-state.json` → `last_maintained_commit` al nuevo HEAD `21634526`.
- `gentle-ai sync` NO ejecutado — overlay intacto en disco, nada requiere re-apply.
- Sin cambios en policy, scripts, ni documentación del overlay.

Verificación:

- `bash audit-gentle-ai-upstream.sh` retornó `base prompt drift: no` y todos los invariantes `ok` post-update del state.

## 2026-06-01 — Added verbose file-level output for apply-custom

Razón del cambio:

- El entrypoint `apply-gentle-ai-custom` ya mostraba bien los contadores del `Summary:`, pero faltaba trazabilidad humana cuando el usuario necesitaba saber QUÉ archivos se tocaron y QUÉ cambió concretamente en cada uno.
- Ese gap hacía más difícil auditar corridas reales del overlay, sobre todo cuando el apply reescribía snapshots, prompts generados o `opencode.json` sin dejar un detalle explícito por archivo.

WHAT cambió:

- `internal/overlay/apply_custom.go`, `internal/overlay/apply_policy.go`, `internal/overlay/overlays.go`, `internal/overlay/profiles.go`, `internal/overlay/snapshots.go`, `internal/overlay/summary.go`, `internal/overlay/util.go`, `internal/overlay/verbose.go`, `cmd/gentle-ai-overlay/main.go`:
  - nuevo flag `--verbose` para `apply-custom` y `apply-policy`
  - recorder compartido de cambios por archivo
  - output adicional `Verbose changes:` con paths tocados y detalle de escrituras, regeneraciones, poda de skills, snapshots y updates en `opencode.json`
  - el `Summary:` previo se mantiene backward compatible
- `apply-gentle-ai-custom.sh/.ps1`:
  - ahora propagan el nombre real del entrypoint para que la ayuda del CLI muestre el wrapper canónico y no el subcomando interno
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` y `.ps1`:
  - ahora forwardean args al subcomando Go y preservan ayuda/errores consistentes para `--verbose`
- `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `AGENTS.md`:
  - documentado el nuevo modo verbose y su contrato operativo

WHY:

- Los contadores sirven para ver magnitud, pero NO alcanzan para entender el efecto real de una corrida. Cuando mantenés overlays, necesitás ver el archivo exacto y la mutación concreta para auditar sin adivinar.
- Mantener el `Summary:` intacto evita romper hábitos o parsing existente, mientras que `--verbose` agrega el detalle solo cuando realmente se necesita.

Verificación:

- `gofmt -w cmd/gentle-ai-overlay/main.go internal/overlay/apply_custom.go internal/overlay/apply_policy.go internal/overlay/overlays.go internal/overlay/profiles.go internal/overlay/snapshots.go internal/overlay/summary.go internal/overlay/util.go internal/overlay/verbose.go`
- `go test ./...`
- `go run ./cmd/gentle-ai-overlay --help`
- `go run ./cmd/gentle-ai-overlay apply-policy --help`
- `go run ./cmd/gentle-ai-overlay apply-custom --help`

## 2026-06-01 — Clarified maintainer workflow ordering and apply target choice

Razón del cambio:

- El flujo operativo ya distinguía bien auditoría pre-sync vs apply post-sync, pero la documentación todavía no dejaba explícitos dos detalles importantes para mantenimiento real: trabajar desde `gentle-ai-custom` con la skill maintainer y actualizar este repo si la auditoría upstream lo exige antes de correr `sync`.
- Además, la documentación trataba `bash apply-gentle-ai-custom.sh all` como único cierre válido, cuando en realidad `opencode` ya alcanza para re-materializar OpenCode + la policy del overlay; `all` solo agrega el refresh de skills/wrappers custom en el resto de targets.

WHAT cambió:

- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`:
  - el camino recomendado ahora explicita el orden completo: update binario → `git pull` upstream → abrir `gentle-ai-custom`/activar maintainer → auditar → actualizar este repo si hace falta → `sync` o reinstall → `apply ... opencode|all`
  - la regla operativa post-update ya no fuerza `all`; documenta `opencode` como mínimo y `all` como refresh multi-target
- `README.md`, `overlay/gentle-ai/README.md`, `AGENTS.md` y `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - alineados con el mismo orden operativo y con la distinción `opencode` vs `all`

WHY:

- El mantenimiento del overlay tiene dos preguntas distintas y ambas deben quedar visibles en el runbook: primero si el upstream se puede adoptar, después cómo materializar localmente lo ya auditado.
- Si la documentación obliga siempre a `all`, mezcla la necesidad real del overlay de OpenCode con una conveniencia adicional de distribución multi-target.

Verificación:

- Revisión cruzada de coherencia entre `runbooks/maintain-upstream-overlay.md`, `README.md`, `overlay/gentle-ai/README.md`, `AGENTS.md` y `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`.
- Confirmado en `internal/overlay/apply_custom.go` que `opencode` ya dispara `RunApplyPolicy()` y que `all` solo amplía la instalación de skills/wrappers custom a todos los targets soportados.

## 2026-06-01 — Completed the shared Go overlay CLI refactor

Razón del cambio:

- El refactor a Go había quedado a mitad de camino: la CLI nueva ya existía, pero `audit-upstream` seguía viviendo solo en Python y los wrappers públicos e internos seguían cargando lógica pesada o referencias stale.
- Eso dejaba el repo en un estado inconsistente: no compilaba, no había una única implementación compartida y la documentación seguía describiendo la arquitectura anterior.

WHAT cambió:

- `internal/overlay/audit_upstream.go`:
  - nuevo comando Go `audit-upstream`
  - valida alineación de `gentle-orchestrator.last.md`, `.meta.yaml` y `upstream-state.json`
  - compara el prompt base upstream contra el baseline versionado
  - chequea invariantes de perfiles (`profilePhaseOrder`, naming `sdd-orchestrator-*`, deny-all task scoping, binding del asset base en `inject.go`)
- `apply-gentle-ai-custom.sh/.ps1`, `audit-gentle-ai-upstream.sh/.ps1`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh/.ps1`:
  - convertidos en wrappers finos que delegan a `go run ./cmd/gentle-ai-overlay ...`
  - sin lógica operativa duplicada en shell o PowerShell
- `overlay/gentle-ai/scripts/audit-gentle-ai-upstream.py`:
  - eliminado por quedar totalmente supersedido por la implementación Go
- `README.md`, `AGENTS.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - actualizadas para reflejar la arquitectura Go compartida y remover referencias stale a Python o a sanitizadores duplicados por script

WHY:

- El objetivo del refactor era tener una sola implementación compartida y verificable. Mientras la auditoría y los wrappers siguieran repartidos entre Go, shell, PowerShell y Python, esa promesa era falsa.
- Mover todo el comportamiento a Go reduce drift entre plataformas y hace que `go test ./...` sea una verificación real del runtime principal.

Verificación:

- `go test ./...`
- `bash audit-gentle-ai-upstream.sh`
- `bash apply-gentle-ai-custom.sh all`
- Parity PowerShell por inspección: los wrappers `.ps1` ahora delegan al mismo entrypoint Go que los `.sh`.

## 2026-06-01 — Separate pre-sync upstream audit from apply helper

Razón del cambio:

- El helper `apply-gentle-ai-policy` estaba cargando demasiadas responsabilidades a la vez: aplicar la capa custom, reconstruir overlays, recuperar desde snapshots y, al mismo tiempo, actuar como detector principal de drift upstream.
- Eso generaba un problema operativo: antes de `gentle-ai sync`, `opencode.json` todavía refleja el estado previamente overlay-applied, así que el helper no puede detectar de forma confiable el drift del prompt inline nuevo sin materializar primero el upstream. Necesitábamos un auditor pre-sync separado.

WHAT cambió:

- Nuevos entrypoints públicos:
  - `audit-gentle-ai-upstream.sh`
  - `audit-gentle-ai-upstream.ps1`
- Nuevo motor compartido:
  - `overlay/gentle-ai/scripts/audit-gentle-ai-upstream.py`
  - lee directamente el repo upstream de Gentle AI
  - compara el prompt base upstream (`internal/assets/opencode/sdd-orchestrator.md`) contra `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`
  - valida el sidecar metadata file `gentle-orchestrator.last.meta.yaml`
  - chequea invariantes upstream de perfiles (`profilePhaseOrder`, prefix `sdd-orchestrator-`, task scoping deny-all-then-allow-profile-phases-and-JD, binding del asset base en `inject.go`)
- Nuevo sidecar versionado:
  - `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml`
  - endurece la alineación entre snapshot base, `upstream-state.json` y la fuente upstream auditada
- `apply-gentle-ai-policy.sh` y `.ps1`:
  - ya no son el auditor semántico principal del upstream
  - siguen materializando, sanitizando, reconciliando perfiles y recuperando desde snapshots
  - agregan verificación automática fail-closed del baseline auditado: el `gentle-orchestrator` materializado después de `sync`/apply debe coincidir con `gentle-orchestrator.last.md` + `.meta.yaml`
- `apply-gentle-ai-custom.sh` y `.ps1`:
  - el flujo normal de apply ahora incluye esa verificación automáticamente; no hay paso manual extra de verificación
- `README.md`, `AGENTS.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - documentado el flujo nuevo: audit pre-sync → decidir/adaptar → sync/reinstall → apply con auto-verificación fail-closed

WHY:

- El `gentle-orchestrator` base es la única fuente de verdad necesaria para auditar drift del prompt. Los prompts `sdd-orchestrator-<perfil>` son derivados de esa base por lógica de upstream; no hacía falta diffear sus snapshots para auditar cambios semánticos.
- Separar auditoría pre-sync de materialización post-sync simplifica el mantenimiento y hace explícita la pregunta correcta en cada fase:
  - `audit-gentle-ai-upstream` => “¿es seguro avanzar con sync/reinstall?”
  - `apply-gentle-ai-custom` => “¿quedó materializado en disco lo que ya auditamos?”

Verificación:

- `bash audit-gentle-ai-upstream.sh` => OK; metadata alineada, hash verificado, base prompt sin drift, invariantes de perfiles OK.
- `bash apply-gentle-ai-custom.sh all` => OK; `audited base baseline verification: ok`, 0 topology warnings, 3 perfiles SDD gestionados, sin drift local/repo.
- Parity PowerShell por inspección: wrapper `.ps1` invoca el mismo motor Python y el helper `.ps1` implementa la misma verificación fail-closed del baseline auditado.

## 2026-05-30 — Profile orchestrator snapshots moved out of the repo

Razón del cambio:

- Después de mover los perfiles SDD nombrados a un config per-máquina (`~/.config/gentle-ai-custom/opencode-sdd-profiles.json`), seguía habiendo una inconsistencia: los snapshots `sdd-orchestrator-<perfil>.last.md` seguían versionándose dentro del repo aunque representan estado operativo local derivado de perfiles también locales.
- Eso ensuciaba el historial con artefactos redundantes y mezclaba en el árbol versionado algo que ya no forma parte del baseline portable del overlay.

WHAT cambió:

- `overlay/gentle-ai/policy/gentle-ai-policy.json`:
  - Nuevo campo `opencode.local_orchestrator_snapshot_dir = "~/.config/gentle-ai-custom/opencode-orchestrator-snapshots"`.
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` y `.ps1`:
  - Split de storage para snapshots de orchestrators.
  - `gentle-orchestrator` ahora mantiene DOS copias:
    1. snapshot operativo local en `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/gentle-orchestrator.last.md`
    2. snapshot versionado en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`
  - Los `sdd-orchestrator-<perfil>.last.md` pasan a vivir SOLO en el directorio local per-máquina.
  - Recovery/lookup actualizado:
    - para `gentle-orchestrator`: preferir snapshot local; si falta, fallback al versionado del repo y mirror de vuelta al local
    - para `sdd-orchestrator-<perfil>`: usar solo el directorio local; si falta, fail closed pidiendo `gentle-ai sync`
  - Nueva migración automática: si todavía existe un snapshot legado de perfil en el repo, copiarlo al directorio local y seguir desde ahí.
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`:
  - removidos del repo `sdd-orchestrator-mixed.last.md`, `sdd-orchestrator-vertex.last.md`, `sdd-orchestrator-vertex-claude.last.md`
  - queda versionado únicamente `gentle-orchestrator.last.md`
- `AGENTS.md`, `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/policy/orchestrator-policy.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - documentada la nueva frontera: el repo conserva solo el snapshot base portable; los snapshots de perfiles son estado operativo local.

WHY:

- Los perfiles SDD nombrados ya no son parte del baseline compartido. Mantener sus snapshots en git rompía esa frontera conceptual.
- `gentle-orchestrator.last.md` sí conserva valor histórico/reviewable porque es el baseline común del overlay contra upstream. Los snapshots per-profile no agregan señal útil al historial del repo.

Verificación:

- Revisión de scripts bash/PS1: ambos implementan la misma regla de dual snapshot para `gentle-orchestrator` y snapshot local-only para perfiles.
- Confirmado que el directorio local `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` contiene `gentle-orchestrator.last.md` + los snapshots de perfiles.
- Confirmado que el árbol versionado de snapshots queda reducido al snapshot base una vez removidos los tres `sdd-orchestrator-*.last.md` del repo.

## 2026-05-30 — Maintainer now recommends sync vs reinstall explicitly

Razón del cambio:

- La skill `gentle-ai-overlay-maintainer` ya distinguía drift de topología versus cambios de comportamiento, pero no convertía siempre ese hallazgo en una recomendación operativa explícita para el usuario.
- Eso dejaba un hueco peligroso: ante cambios de topología upstream, alguien podía asumir que `gentle-ai sync` alcanzaba aunque en este repo ya sabemos que no materializa bien ese tipo de cambios.

WHAT cambió:

- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - Nueva hard rule para traducir toda auditoría upstream en una recomendación explícita de adopción: `gentle-ai sync` cuando no hay drift de topología, reinstalación completa cuando sí lo hay.
  - Nuevos decision gates separando cambios de topología upstream de cambios de workflow/skills sin drift de topología.
  - Execution steps y output contract actualizados para exigir que la recomendación `sync vs reinstall` salga en el handoff al usuario.
- `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`:
  - Documentada la regla operativa para que la recomendación de la skill coincida con el runbook: topología cambió => reinstalación; contenido/comportamiento sin drift => `gentle-ai sync`.

WHY:

- El problema no era técnico en scripts/policy sino de guidance operativa: el criterio existía en el runbook, pero no estaba lo bastante cerca del punto donde el agente le responde al usuario qué hacer.
- Mover esa decisión al output explícito de la skill reduce el riesgo de aplicar mal una actualización upstream y quedar con un estado parcialmente refrescado.

Verificación:

- Revisión manual de consistencia entre skill, README raíz, overlay README y runbook.
- Confirmado que no fue necesario tocar scripts ni policy: el cambio es solo de instrucciones/documentación para surfacing de la decisión ya existente.

## 2026-05-30 — Closed upstream v1.33.2 maintenance audit

Razón del cambio:

- El upstream `gentle-ai` avanzó desde `412eed3d39defb2f955a63e21ca13cef4df358c9` (`v1.32.0`) hasta `0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc` (`v1.33.2`) y había que cerrar formalmente la auditoría posterior al upstream git pull en este repo.

WHAT cambió:

- `overlay/gentle-ai/state/upstream-state.json`:
  - Actualizado el boundary mantenido a `v1.33.2` / `0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc`.
  - Actualizado `last_reviewed_at` al momento de cierre de la auditoría.
  - Reescrita la nota para reflejar que esta entrada ya no es solo el seed inicial sino un mantenimiento cerrado contra un rango auditado.
- `overlay/gentle-ai/logs/update-log.md`:
  - Ajustada esta entrada top-level para dejar explícitos el rango auditado, el no-cambio en policy/scripts/docs y la verificación puntual sobre anclas/bloques del prompt upstream.

WHY:

- La auditoría del rango `412eed3d39defb2f955a63e21ca13cef4df358c9..0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc` mostró cambios relevantes para awareness, pero no para mutación del overlay.
- Upstream agregó los subagentes `jd-judge-a`, `jd-judge-b` y `jd-fix-agent`, y sumó deduplicación de sub-agent launches en el orchestrator.
- Ninguno de esos cambios exige tocar keep/prune, sanitizador, scripts o documentación operativa: los anchors requeridos siguen intactos y los bloques de PR/review que este overlay depura siguen presentes donde se espera.

Verificación:

- Revisión del rango upstream con `git log --oneline`, `git diff --name-only` y `git show --stat` entre los commits auditados.
- Inspección puntual de `internal/assets/opencode/sdd-orchestrator.md`, auditando las anclas y bloques prohibidos relevantes para el sanitizer.
- Confirmación de que `Chained PR strategy`, `Review budget`, `C. PRs`, `D. Review`, `### Delivery Strategy`, `### Chain Strategy` y `### Review Workload Guard (MANDATORY)` siguen presentes en el prompt upstream y, por lo tanto, el sanitizador actual sigue siendo válido.
- Revisión de los commits/archivos upstream relevantes a este overlay: JD agents (`1f13b8d`, `b5351d3`, `22485fb`) y deduplicación del orchestrator (`28fe11d`), sin necesidad de cambios en policy, scripts o docs locales.

## 2026-05-30 — Follow-up review fixes for local SDD profile config

Razón del cambio:

- La revisión fresh-context posterior a la implementación encontró 2 bugs reales de parity bash↔PowerShell en la validación del config local de perfiles SDD y una omisión semántica en `maintenance-intent.md`.

WHAT cambió:

- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`:
  - Fix del caso `profiles` con un solo elemento: el helper ya no unwrappea silenciosamente un array JSON de 1 elemento a `PSCustomObject`. Ahora exige que `profiles` sea array real y luego lo force-wrappea con `@(...)`, igualando el comportamiento del bash/Python.
  - Validación del `name` del perfil ahora usa `-cnotmatch` para que la regex sea case-sensitive. Antes PowerShell aceptaba uppercase mientras bash lo rechazaba.
  - Renombrado `$Profile` → `$ProfileEntry` para evitar shadowing del automatic variable.
  - Nuevo rechazo explícito de campos extra en el top-level del config local: solo `version` y `profiles` son válidos.
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`:
  - Nuevo rechazo explícito de campos extra en el top-level del config local: solo `version` y `profiles` son válidos. Con esto el schema V1 queda simétricamente estricto arriba y dentro de cada profile.
- `overlay/gentle-ai/policy/maintenance-intent.md`:
  - Nueva sección `## Qué NO se versiona` aclarando que las elecciones locales de `model`/`variant` por perfil SDD nombrado viven fuera del repo en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` y no deben volver a filtrarse a `gentle-ai-policy.json`.
- `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`:
  - Documentado el rechazo de campos top-level extra y actualizado el catálogo de errores esperables del schema strict.

WHY:

- Sin el fix del single-element array, Windows/PowerShell fallaba cerrado con un mensaje engañoso para un JSON perfectamente válido con un solo profile.
- Sin el fix de case-sensitivity, PowerShell podía crear suffixes con mayúsculas que bash jamás aceptaría, rompiendo parity cross-platform.
- Sin actualizar `maintenance-intent.md`, el repo quedaba con la policy/runtime reflejando una decisión semántica nueva sin haberla declarado en la fuente de verdad del intent.

Verificación:

- Revisión fresh-context confirmó que ambos bugs eran reales y localizó el punto exacto de divergencia.
- Inspección manual posterior de ambos scripts: ahora los dos rechazan campos top-level extra, tratan `profiles` como array real, y validan `name` con semántica case-sensitive consistente.
- Chequeo documental: `maintenance-intent.md` ahora explicita la nueva frontera entre policy portable y config per-máquina.

## 2026-05-29 — SDD profile local config + policy depersonalization

Razón del cambio:

- La versionada `gentle-ai-policy.json` traía hardcodeadas las keys exactas `sdd-orchestrator-mixed`, `sdd-orchestrator-vertex`, `sdd-orchestrator-vertex-claude` en `orchestrator_agent_keys`. Esto es información per-máquina (qué perfiles SDD usa este usuario, con qué modelos/variants) filtrada al repo compartido. Sumado a eso, `gentle-ai sync` deja en `opencode.json` agent entries con `model`+`variant` per perfil y per fase (~33 keys) que tampoco deberían vivir en el repo.
- Solución: introducir un config per-máquina fuera del repo, con schema strict, y mover toda la gestión de perfiles SDD a ese archivo. El repo solo retiene baseline portable (`gentle-orchestrator` + prefijo `sdd-orchestrator` como sanitization tripwire).

WHAT cambió:

- `overlay/gentle-ai/policy/gentle-ai-policy.json`:
  - Removidas las keys exactas `sdd-orchestrator-mixed`, `sdd-orchestrator-vertex`, `sdd-orchestrator-vertex-claude` de `orchestrator_agent_keys`. La policy ahora solo lista `gentle-orchestrator` ahí.
  - El prefijo `sdd-orchestrator` se conserva en `orchestrator_agent_prefixes` para que la sanitización siga matcheando per-profile orchestrators con prompt inline upstream.
  - Nuevo campo `opencode.profile_orchestrator_prefix = "sdd-orchestrator-"` — usado por ambos helpers para distinguir profile-managed orchestrators de orchestrators desconocidos, y suprimir el `topology: unknown orchestrator matched by prefix only:` para esos casos.
  - Nuevo campo `opencode.sdd_profiles_local_config_path = "~/.config/gentle-ai-custom/opencode-sdd-profiles.json"`.
  - Nuevo campo `opencode.sdd_phases` listando las 10 phases SDD canónicas (`sdd-init`, `sdd-explore`, …, `sdd-onboard`).
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` y `.ps1`:
  - Nueva fase de reconciliación SDD-profile entre el override loop y el topology check.
  - Validación strict V1 del config local: `version=1`, `profiles` array no vacío, cada profile con exactamente `name`/`orchestrator`/`phases`, `name` matcheando `^[a-z0-9][a-z0-9._-]*$` y único, `phases` con exactamente las 10 phase keys, cada assignment con exactamente `model` (non-empty string) y `variant` (string, may be ""). Sin defaults, sin herencia, sin campos extra.
  - **Fail-closed antes de cualquier escritura a `opencode.json`** si la validación falla.
  - Para cada profile managed: crea/actualiza `sdd-orchestrator-<name>` + 10 phase agents `sdd-<phase>-<name>` con `model`/`variant` exactos del config. NO toca `prompt` (sigue siendo dominio de `gentle-ai sync` + sanitización inline existente).
  - Detecta perfiles unmanaged (en `opencode.json` pero no en el config local), los reporta con `WARNING - unmanaged SDD profiles left untouched` + contador, los deja intactos. **NUNCA borra automáticamente.**
  - Nuevos contadores en el `Summary:`: `SDD profiles managed`, `SDD profile agents created`/`updated`/`unchanged`, `SDD profiles unmanaged`.
  - Post-write verification extendida: cada profile managed debe tener su orchestrator + 10 phase agents persistidos en disco.
  - Topology drift check actualizado: profile-managed orchestrators ahora se EXCLUYEN del warning `unknown orchestrator matched by prefix only:` (eran ruido inevitable bajo el nuevo modelo).
- `AGENTS.md`: nueva sección `## SDD profile local config` con schema completo, hard rules y nota sobre por qué los prompts no están en el schema. Actualizada la sección `## Overlay policy baseline` para explicar que profile orchestrators NO se versionan.
- `README.md`: nueva subsección `### Perfiles SDD locales` debajo de `## Política actual` con schema, contrato operacional y reglas duras. Actualizado bloque `### Qué reporta el script` con los nuevos counters y el set de ERROR de validation strict.
- `overlay/gentle-ai/README.md`: descripción del helper actualizada para mencionar la reconciliación de perfiles y la validación strict fail-closed. Nueva subsección `### Config externo gestionado fuera del repo`.
- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`: nuevas filas en la tabla de señales (counters de SDD profiles, WARNING unmanaged, ERRORs de validación). Nueva sección `### Perfiles SDD locales (config externo)`. Checklist actualizada con dos items nuevos sobre la policy depersonalizada y el config local.
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`: nueva hard rule prohibiendo agregar profile-managed assignments a `gentle-ai-policy.json` (deben vivir en el config local). Nuevas Decision Gates para el WARNING unmanaged y los ERRORs de validación. Nuevo step de post-state verification que chequea la reconciliación SDD profile.

WHY (detalle adicional):

- El `vertex` profile actual en `opencode.json` tiene `sdd-onboard.model = null variant = null`. Bajo el schema V1 strict esto es inválido (`model` requiere non-empty string). La generación inicial del config local lo respeta: el profile `vertex` quedó EXCLUIDO del archivo activo `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` y se guardó en cuarentena en `~/.config/gentle-ai-custom/opencode-sdd-profiles.invalid.json` con un `_note` top-level explicando cómo activarlo (fix el model vacío + mover al archivo activo). Esto deja al usuario consciente del estado roto sin hacerlo invisible y sin inventar valores. Mientras tanto, el script lista a `vertex` como "unmanaged" en cada corrida — visibilidad permanente.

Verificación:

- Run con NO local config: 0 perfiles managed, 0 unmanaged, 0 topology warnings, opencode.json sin cambios. Idempotente.
- Run con local config válido (mixed + vertex-claude): 2 managed, 22 unchanged, 1 unmanaged (vertex), WARNING listado, opencode.json sin cambios (estado ya consistente).
- 8 negative tests sobre el config local (version inválida, version distinta, profile sin phases, model vacío, name con `/`, phase desconocida, campo extra, JSON malformado): TODOS fallan con `ERROR:` accionable y `opencode.json` queda byte-idéntico al estado previo (fail-closed verificado).
- Test de update positivo: cambié `sdd-init-mixed.model` en el config local a un valor fake, corrí el script, opencode.json se actualizó con el fake, post-write verification passed; revertí, opencode.json volvió a estado original byte-idéntico al baseline.
- Test de create-from-scratch: agregué un profile `test-create` brand-new al config local; el script creó las 11 agent keys (1 orchestrator + 10 phases), reportó `created: 11`, post-write verification passed. Cleanup manual restauró opencode.json byte-idéntico.
- Parity bash↔PS1 verificada por inspección (no hay pwsh en este host). El PS1 mirrorea exactamente bash: misma fase de reconciliación, mismas reglas de validación, mismos mensajes de error, mismo orden (validate-all-then-mutate), misma post-write verification, mismo set de counters en el summary.

## 2026-05-29 — Judgment Day round 1 fixes

Adversarial dual review (Judge A + Judge B en paralelo) sobre los 5 commits de hardening (`bfbc363..HEAD`). Confirmados 2 CRITICAL y 11 WARNING (real). Esta entrada documenta los fixes aplicados.

Fixes de código en `apply-gentle-ai-policy.{sh,ps1}`:

- **CRITICAL — PS1 `Remove-ExactOnce` reemplazaba TODAS las ocurrencias**: usaba `String.Replace($Old, $New)` (multi-occurrence) mientras bash usaba `text.replace(old, new, 1)` (single-occurrence). Reescrita con `IndexOf` + `Substring` para reemplazar solo la primera ocurrencia, mirroring la semántica de Python.
- **CRITICAL — Preflight sanitization: parity divergente entre bash y PS1**: bash eliminaba las dos líneas (Chained PR strategy + Review budget) como un único bloque con un `replace_once`; PS1 las eliminaba como dos llamadas independientes a `Remove-ExactOnce`. Reorganizado PS1 para mirrorear bash: una sola llamada con el bloque concatenado por un newline PowerShell (`"`n"`), lo que iguala las semánticas de fallo (ambos ahora fallan con `missing expected text: preflight PR/review choices` si upstream las reordena).
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
