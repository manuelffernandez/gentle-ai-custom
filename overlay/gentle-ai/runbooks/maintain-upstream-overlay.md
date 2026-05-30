# Runbook — mantenimiento del overlay de Gentle AI

## Objetivo

Mantener la capa de personalización de `gentle-ai-custom` alineada con el upstream en `/home/manuel/Documentos/gentle-ai` sin volver a empezar desde cero en cada actualización.

## Modelo operativo

Este mantenimiento se apoya en cuatro capas explícitas:

- **Intento** → `overlay/gentle-ai/policy/maintenance-intent.md`
- **Política** → `overlay/gentle-ai/policy/gentle-ai-policy.json`
- **Estado** → `overlay/gentle-ai/state/upstream-state.json`
- **Log** → `overlay/gentle-ai/logs/update-log.md`

No cumplen el mismo rol:

- el **intento** describe el criterio del usuario
- la **política** alimenta los scripts runtime
- el **estado** marca desde dónde hay que auditar el upstream
- el **log** deja historial narrativo de mantenimiento

## Cuándo correr este runbook

- después de actualizar `gentle-ai`
- cuando `gentle-ai sync` vuelva a introducir convenciones de PR/budget
- cuando cambie la estructura del prompt inline de algún orchestrator
- cuando aparezcan nuevas skills upstream que puedan entrar en conflicto con la política local

## Tipos de actualización de Gentle AI y su impacto en el overlay

Gentle AI puede actualizarse de tres maneras distintas, con efectos diferentes sobre `opencode.json` y las skills instaladas.

### Mecanismos de actualización

| Mecanismo | Qué actualiza | Impacto en `opencode.json` | Skills prunidas |
|---|---|---|---|
| `brew upgrade gentle-ai` | Solo el binario | Ninguno | Sin cambio |
| `gentle-ai sync` / "Sync Configurations" | Prompts, skills, MCP configs, SDD | **Resetea** prompts de orchestrators a upstream inline | **Vuelven** (reinstala todas) |
| Reinstalación via TUI | Todo: topología de agentes, presets, skills, configuración | **Resetea todo**, puede cambiar agentes | **Vuelven** + posible topología nueva |

### Por qué `gentle-ai sync` resetea los prompts

El mecanismo es verificable en el código fuente del upstream (`internal/components/sdd/profiles.go`):

```
~/.config/opencode/profiles/ vacío o inexistente
→ HasExternalProfileFiles() = false
→ ResolveProfileStrategy() = "generated-multi"   ← no "external-single-active"
→ PreserveOpenCodeOrchestratorPrompt = false
→ sync sobreescribe el prompt del orchestrator con el asset inline upstream
```

El directorio `~/.config/opencode/profiles/` en este setup está vacío, por lo que cada `gentle-ai sync` resetea los prompts de los orchestrators al contenido inline de upstream.

### Regla operativa invariante

**Después de cualquier operación de Gentle AI que no sea solo `brew upgrade`, es obligatorio correr:**

```bash
bash apply-gentle-ai-custom.sh all
```

Esto re-aplica el overlay completo:
- poda las skills no deseadas (branch-pr, chained-pr, issue-creation, work-unit-commits)
- re-aplica los overrides de modelos (general, explore)
- actualiza `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/*.last.md` para todos los orchestrators
- mantiene además `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` como baseline versionado
- sanitiza y regenera los `*.overlay.md` en `~/.config/opencode/prompts/sdd/orchestrators/`
- reescribe `opencode.json` con referencias `{file:...}`

### Cuándo usar reinstalación vs sync

Usá **solo sync** cuando la actualización:
- cambia prompts, skills, MCP configs
- cambia modelos/variants sin cambiar la estructura de agentes
- mantiene los mismos agentes con contenido actualizado

Usá **reinstalación** cuando la actualización:
- renombra o elimina agentes existentes
- agrega nuevos agentes o presets
- cambia la forma en que Gentle AI construye `opencode.json`
- deja artefactos de una versión anterior que sync no limpia

**Señal de que necesitás reinstalar:** después de sync seguís viendo agentes viejos o no aparece un agente nuevo que debería existir según las release notes.

Durante esta auditoría, la skill maintainer debe devolver esta recomendación de forma explícita: `gentle-ai sync` para cambios sin drift de topología; reinstalación completa cuando el upstream cambió la topología o la forma en que materializa los agentes.

### Flujo completo post-actualización

```
brew upgrade gentle-ai          ← actualiza el binario
git pull (gentle-ai repo)       ← actualiza el clon local del upstream
gentle-ai sync                  ← reaplica config managed (resetea prompts + skills)
bash apply-gentle-ai-custom.sh all  ← re-aplica el overlay custom (OBLIGATORIO)
git diff overlay/gentle-ai/snapshots/  ← verificá qué cambió en los prompts upstream
git add -p && git commit        ← commitea el nuevo estado de los snapshots
```

Si la actualización requiere reinstalación, el agente de mantenimiento debe auditar primero antes de correr el script.

### Señales del script para actuar

Después de cada corrida del script, leé el bloque `Summary:` y los `topology:` warnings. Cada señal mapea a una acción concreta:

| Señal en el output | Significa | Acción |
|---|---|---|
| `orchestrators kept (already applied): N` | Estado estable: todo ya está aplicado y el script no tuvo que tocar nada. | Ninguna. Indica una corrida idempotente exitosa. |
| `SDD profiles managed: N` / `created: N` / `updated: N` / `unchanged: N` | Cuántos perfiles se reconciliaron desde el config local y cuántos agent entries se crearon/actualizaron/no cambiaron. | Ninguna si todo cuadra con lo esperado. Útil para verificar que un cambio del config local realmente se aplicó. |
| `SDD profiles unmanaged ...: N > 0` + `WARNING - unmanaged SDD profiles left untouched` | Hay perfiles SDD (`sdd-orchestrator-<name>`) en `opencode.json` que no están nombrados en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`. El script no los tocó. | Decidir si querés gestionar ese perfil (agregalo al config local) o sacarlo (borralo a mano de `opencode.json`). El script NUNCA borra perfiles automáticamente. |
| `ERROR: local SDD profile config at ... is not valid JSON` o `... unexpected top-level field ...` o `... missing required field ...` o `... unexpected fields ...` o `... must be a non-empty string` o `... must match ^[a-z0-9]...` o `... missing required phases ...` o `... unknown phases ...` | El config local de perfiles no pasa el schema V1 strict. | El script **no escribe nada** a `opencode.json` en este caso (fail-closed). Arreglá el JSON o eliminá el archivo. |
| `topology: unknown orchestrator matched by prefix only: X` | Apareció un orchestrator nuevo upstream que sólo matchea por prefijo (excluyendo `sdd-orchestrator-*` que son perfil-managed). | Auditar el cambio upstream. Si es legítimo, agregar `X` a `orchestrator_agent_keys` en la policy con aprobación del usuario. |
| `topology: expected orchestrator missing from opencode.json: X` | Un orchestrator listado en la policy ya no existe en upstream. | Auditar si fue renombrado/eliminado upstream. Actualizar policy + intent. |
| `topology: agent_override target was missing from upstream (created): X` | El override apuntaba a un agente que no existía upstream **o que existía con un tipo distinto a object**; el script creó/sobrescribió el stub. | Verificar si el agente fue renombrado o cambió de forma upstream. Ajustar el `key` del override o el intent si corresponde. |
| `repo snapshots - changed: N > 0` | Cambió el baseline versionado de `gentle-orchestrator`. | `git diff overlay/gentle-ai/snapshots/` para revisar. Si los anchors del sanitizador se movieron, actualizar ambos scripts. |
| `local snapshots - changed: N > 0` | Cambió algún snapshot operativo local (incluidos profiles) bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`. | Verificar cuál cambió en el directorio local. No esperes ver esto en git para snapshots per-perfil. |
| `local snapshot migrations from repo: N > 0` | El helper copió snapshots legacy desde el repo al directorio local operativo. | Confirmar que la migración dejó los archivos esperados en `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` y luego remover del repo los snapshots per-perfil versionados. |
| `repo snapshot backfills from local: N > 0` | El helper recreó el snapshot versionado de `gentle-orchestrator` desde la copia operativa local. | Revisar por qué faltaba el snapshot versionado en el repo antes de la corrida. |
| `orchestrators recovered from snapshot: N > 0` + el bloque `NOTE: ... may pre-date the current upstream` | Algún `.overlay.md` no existía en disco y se reconstruyó desde el snapshot. El contenido aplicado puede ser de una versión upstream anterior. | Investigar por qué se perdió el archivo (¿borrado manual? ¿bug en otro script?). Si querés capturar fresco, correr `gentle-ai sync` y re-correr el script — el snapshot se actualizará desde el inline upstream. |
| `WARNING - keep skills missing (expected but absent):` | Alguna skill que debería existir está ausente en un target. | Probablemente la skill se renombró/eliminó upstream o el usuario la quiso fuera. Revisar intent. |
| `WARNING recovering X from local snapshot - content may pre-date current upstream` | El script va a reconstruir el `.overlay.md` desde el snapshot operativo local porque el target file falta. Se imprime durante la corrida, no en el Summary. | Si fue intencional (borrado manual), todo OK. Si no, investigar quién está borrando los `.overlay.md`. |
| `ERROR: broken state for orchestrator 'X': opencode.json prompt is '{file:...}' but the target file is missing and ...` | `opencode.json` apunta a un `{file:...}` inexistente y no existe el snapshot recuperable requerido (local para profiles; local o repo versionado para `gentle-orchestrator`). | Correr `gentle-ai sync` para resetear los prompts a inline upstream, después re-correr el script. |
| `ERROR: post-write verification failed: agent X model is Y after write, expected Z` | El JSON se escribió pero al re-leer los valores no coinciden con lo esperado. | Bug serio: investigar si hay otro proceso escribiendo `opencode.json` (race) o si el script tiene un bug de serialización. |
| `ERROR: post-write verification failed: orchestrator X prompt is Y after write, expected Z` | Idem anterior pero para la referencia `{file:...}` del orchestrator. | Idem anterior. |
| `ERROR: OpenCode config at ... is not valid JSON: ...` | `opencode.json` está corrupto y no se puede parsear. | Restaurar desde `~/.gentle-ai/backups/<timestamp>/` o correr `gentle-ai sync` para regenerar. |
| `ERROR: unsafe agent key for snapshot path: 'X'` | Un agente en `opencode.json` tiene un key con `/`, `\`, `..` o caracteres nulos que harían path traversal al escribir el snapshot. | Investigar de dónde salió ese key en `opencode.json`; no debería pasar bajo flujos normales. |

### Perfiles SDD locales (config externo)

Desde el cambio del schema de perfiles SDD, los assignments de modelo/variant per-perfil (`sdd-orchestrator-<name>` + 10 phase agents `sdd-<phase>-<name>`) ya no se versionan en este repo. Se gestionan desde un archivo per-máquina fuera del repo:

```
~/.config/gentle-ai-custom/opencode-sdd-profiles.json
```

Contrato operacional:

- Si el archivo no existe, el helper no toca ningún perfil SDD en `opencode.json`. Esto es por diseño: una máquina sin config local no debe alterar configuraciones existentes.
- Si el archivo existe, el helper valida estrictamente (schema V1: sin defaults, sin herencia, los 10 phase keys requeridos, `variant` siempre requerido aunque sea `""`) y falla cerrado antes de tocar `opencode.json` si algo está mal.
- Profiles nombrados en el config local pero ausentes en `opencode.json` se crean. Los agentes orchestrator nuevos se crean como stubs sin `prompt` — la siguiente corrida de `gentle-ai sync` los materializa con el prompt upstream, y la corrida posterior del overlay los sanitiza vía prefix match.
- Profiles presentes en `opencode.json` pero ausentes del config local quedan intactos. Se reportan como `WARNING - unmanaged SDD profiles left untouched`. El helper nunca borra perfiles automáticamente.

Schema completo y reglas duras: ver `README.md` raíz, sección "Perfiles SDD locales", o `AGENTS.md` sección "SDD profile local config".

Snapshots de orchestrators:

- El repo versionado conserva solo `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`.
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` guarda la copia operativa local de `gentle-orchestrator` y todos los `sdd-orchestrator-<profile>.last.md`.
- Si todavía quedan snapshots legacy per-perfil en el repo, la próxima corrida del helper los migra al directorio local para preservar recovery sin requerir un sync inmediato.

Razón del split:

- El repo versionado solo conoce baseline portable (`gentle-orchestrator`). Esto permite compartir el repo entre máquinas sin filtrar elecciones de modelo personales.
- El config local es el único lugar donde viven los assignments por-perfil. Cambiarlo no requiere commit en este repo.

### Opción de hardening: estrategia external-single-active

Por defecto, `gentle-ai sync` resetea el prompt del orchestrator porque `~/.config/opencode/profiles/` está vacío. Esto se puede cambiar.

El upstream tiene esta lógica en `internal/components/sdd/profiles.go`:

```
~/.config/opencode/profiles/ tiene al menos un *.json directamente en la raíz
                              (los subdirectorios son ignorados por HasExternalProfileFiles())
→ HasExternalProfileFiles() = true
→ ResolveProfileStrategy() = "external-single-active"
→ PreserveOpenCodeOrchestratorPrompt = true
→ sync NO sobreescribe el prompt del orchestrator
```

Crear un archivo `*.json` directamente bajo `~/.config/opencode/profiles/` (no en subcarpetas) hace que `gentle-ai sync` respete la referencia `{file:...}` del overlay.

**Tradeoffs**:

- **A favor**: el overlay sobrevive a `gentle-ai sync` sin necesidad de re-correr el script para restaurar prompts. (La poda de skills sí sigue requiriendo el script.)
- **En contra (crítico)**: el usuario sigue ejecutando **indefinidamente** la versión sanitizada anterior del prompt upstream. El script ya no puede ni siquiera intentar resanitizar contra el nuevo upstream porque no lo ve — `opencode.json` queda fijo en `{file:...}` y los `*.last.md` snapshots dejan de refrescarse.
- **En contra**: si los anchors del sanitizador se mueven upstream, solo te enterás cuando alguien borre el profile y `sync` vuelva al comportamiento default — y para ese momento podrías llevar meses ejecutando un prompt sanitizado contra una estructura que upstream ya cambió.
- **En contra**: `git diff overlay/gentle-ai/snapshots/` deja de ser señal útil de cambios upstream.

**No activar sin pedido explícito del usuario y una conversación sobre estos tradeoffs.** El comportamiento default (sync resetea, script re-aplica) tiene la ventaja de mantener los snapshots como una bitácora viva del estado upstream y de garantizar que cada corrida del script sanitiza contra el upstream actual, no contra una foto vieja.

---

## Quick path

1. Trabajá desde `gentle-ai-custom`.
2. Leé el intento en `overlay/gentle-ai/policy/maintenance-intent.md`.
3. Leé la política en `overlay/gentle-ai/policy/gentle-ai-policy.json`.
4. Leé el estado en `overlay/gentle-ai/state/upstream-state.json`.
5. Determiná el rango a auditar desde `last_maintained_commit` / `last_maintained_tag` hasta el estado actual del upstream.
6. Revisá snapshots versionados en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` y snapshots operativos en `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.
7. Si cambió la estructura del orchestrator, ajustá el sanitizador en ambos scripts:
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`
8. Separá cambios relevantes del overlay de bugfix/chore noise.
9. Si hay cambios relevantes de comportamiento o nuevas convenciones, frená y pedile al usuario una decisión explícita.
10. Actualizá docs, skill, política y estado si cambió el workflow.
11. Registrá la decisión en `overlay/gentle-ai/logs/update-log.md`.

## Gate humana

No todo cambio upstream debe mutar el overlay.

El agente debe pedir aprobación humana cuando aparezcan cambios que puedan afectar:

- la intención keep/prune
- la conveniencia de nuevas skills upstream
- la lógica del sanitizador del orchestrator
- la interpretación de qué cambios sí importan para este repo

## Checklist de mantenimiento

- [ ] `maintenance-intent.md` sigue representando lo que el usuario quiere conservar y depurar.
- [ ] La política keep/prune sigue representando la intención del usuario.
- [ ] `upstream-state.json` apunta a la última versión/commit realmente mantenida.
- [ ] Los scripts siguen generando prompts derivados bajo `~/.config/opencode/prompts/sdd/orchestrators/`.
- [ ] `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` conserva solo `gentle-orchestrator.last.md`.
- [ ] `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` contiene `gentle-orchestrator.last.md` y los snapshots `sdd-orchestrator-<profile>.last.md` gestionados localmente.
- [ ] El sanitizador todavía remueve PR/budget/chained-PR/review-workload sin romper `## Model Assignments`.
- [ ] `agent.general` sigue en `openai/gpt-5.4` / `high`.
- [ ] `agent.explore` sigue en `google-vertex/gemini-3.1-pro-preview` / `high`.
- [ ] `apply-gentle-ai-custom.sh` y `.ps1` siguen siendo el único par de entrypoints públicos y mantienen paridad funcional.
- [ ] La policy versionada NO lista keys exactas de orchestrators per-perfil (`sdd-orchestrator-<name>`). Solo lista `gentle-orchestrator` + el prefijo `sdd-orchestrator` para sanitización.
- [ ] El config local de perfiles SDD (si existe) sigue cumpliendo el schema V1 strict; si fue modificado externamente, la próxima corrida del helper lo va a validar antes de cualquier escritura.

## Notas operativas

- El source of truth del orchestrator **no** es un archivo estático del repo: el script captura el prompt inline real desde `opencode.json`, lo sanitiza y recién después genera el `.overlay.md` operativo.
- Si el sanitizador no encuentra anchors esperados, debe fallar cerrado y no reescribir prompts automáticamente.
- El script principal de uso humano es `apply-gentle-ai-custom.*`; `apply-gentle-ai-policy.*` se mantiene como helper interno invocado por el entrypoint cuando aplica (`opencode` o `claude`).
- El log no reemplaza al estado: `update-log.md` cuenta qué se decidió; `upstream-state.json` marca cuál fue la última versión/commit mantenida.
