# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Este repo ya no es solo un instalador de dos skills custom. Ahora funciona como una **capa unificada de personalización y mantenimiento** sobre Gentle AI:

- instala tus skills y wrappers propios
- reaplica tu política local después de `gentle-ai sync`
- audita el baseline upstream de `gentle-orchestrator` antes de sync/reinstall de mantenimiento
- depura skills no deseadas del runtime
- fija overrides de modelo para los agentes built-in de OpenCode listados en `agent_overrides` (ver `overlay/gentle-ai/policy/gentle-ai-policy.json`)
- reconcilia perfiles SDD locales (`sdd-orchestrator-<name>` + 10 phase agents) desde un config por-máquina en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`
- captura prompts inline de orchestrators, los sanitiza y genera prompts derivados por agente/perfil.
- mantiene el snapshot versionado de `gentle-orchestrator` y snapshots operativos locales por máquina bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`
- mantiene el runbook y la skill para auditar futuras actualizaciones del upstream

## Modelo de mantenimiento

La capa de mantenimiento se apoya en cuatro piezas distintas:

- **Intento** → `overlay/gentle-ai/policy/maintenance-intent.md`
- **Política** → `overlay/gentle-ai/policy/gentle-ai-policy.json`
- **Estado** → `overlay/gentle-ai/state/upstream-state.json`
- **Log** → `overlay/gentle-ai/logs/update-log.md`

Cada una cumple un rol distinto:

- `maintenance-intent.md` explica qué quiere conservar y depurar el usuario, y por qué
- `gentle-ai-policy.json` alimenta la lógica operativa de los scripts
- `upstream-state.json` guarda desde qué versión/tag/commit hay que auditar el upstream (última versión mantenida)
- `update-log.md` deja historial narrativo de decisiones de mantenimiento

## Estructura

- `apply-gentle-ai-custom.sh` — entrypoint principal Linux/macOS
- `apply-gentle-ai-custom.ps1` — entrypoint principal Windows (PowerShell 5.1+)
- `audit-gentle-ai-upstream.sh` — auditoría read-only del baseline upstream antes de sync/reinstall
- `audit-gentle-ai-upstream.ps1` — equivalente Windows de la auditoría upstream
- `shared/skills/commit-planner/SKILL.md` — source of truth neutral para planificación/aplicación de commits
- `shared/skills/pr-finalizer/SKILL.md` — source of truth neutral para creación/regeneración de PRs
- `shared/commands/*.md` — cuerpos compartidos para wrappers/prompts por agente
- `go.mod` — módulo Go del runtime compartido
- `cmd/gentle-ai-overlay/main.go` — CLI Go compartida para `apply-custom`, `apply-policy` y `audit-upstream`
- `internal/overlay/*.go` — implementación compartida del overlay, auditoría upstream y sanitización
- `overlay/gentle-ai/README.md` — guía del control-plane de Gentle AI
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — política machine-readable del overlay
- `overlay/gentle-ai/policy/maintenance-intent.md` — intención de mantenimiento del overlay en lenguaje humano/LLM
- `overlay/gentle-ai/policy/orchestrator-policy.md` — criterio de sanitización del orchestrator
- `overlay/gentle-ai/state/upstream-state.json` — estado operativo de la última versión/commit upstream mantenido
- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` — runbook humano para mantenimiento incremental
- `overlay/gentle-ai/logs/update-log.md` — historial de decisiones del overlay
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` — wrapper bash interno fino hacia la CLI Go compartida
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1` — wrapper PowerShell interno fino equivalente
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` — skill de mantenimiento del overlay
- `AGENTS.md` — contrato operativo para agentes

## Targets soportados

- `opencode` → `~/.config/opencode`
- `claude` → `~/.claude`
- `codex` → `~/.codex`
- `gemini` → `~/.gemini`
- `antigravity` → `~/.gemini/antigravity`

## Uso

La capa custom tiene **un único par de entrypoints públicos**:

- `apply-gentle-ai-custom.sh`
- `apply-gentle-ai-custom.ps1`

Para mantenimiento upstream, hay además un par público separado:

- `audit-gentle-ai-upstream.sh`
- `audit-gentle-ai-upstream.ps1`

Todos esos wrappers son finos: delegan en la CLI Go compartida (`go run ./cmd/gentle-ai-overlay ...`) y no duplican la lógica de negocio entre shell, PowerShell y Python.

### Linux / macOS

```bash
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode --verbose
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh claude
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh codex
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh gemini
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh antigravity
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

### Windows (PowerShell 5.1+)

> **Requisito previo — política de ejecución:**
>
> ```powershell
> Set-ExecutionPolicy -Scope Process Bypass
> ```

```powershell
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 opencode
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 opencode --verbose
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 claude
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 codex
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 gemini
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 antigravity
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 all
```

## Flujo recomendado

```bash
brew upgrade gentle-ai
git -C ~/Documentos/gentle-ai pull

# trabajar desde gentle-ai-custom con la skill maintainer
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh

# si la auditoría no exige adaptar este repo primero
gentle-ai sync

# mínimo para OpenCode/policy del overlay
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode

# o refresh completo multi-target
# bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

Orden mental correcto:

1. actualizás el binario de `gentle-ai`
2. hacés `git pull` en `/home/manuel/Documentos/gentle-ai`
3. abrís `gentle-ai-custom`, usás la skill maintainer y corrés `audit-gentle-ai-upstream`
4. si hace falta, actualizás este repo antes de seguir
5. recién ahí corrés `gentle-ai sync` (o reinstall si la auditoría lo recomienda)
6. `apply-gentle-ai-custom` → responde **"¿quedó materializado en disco lo que ya auditamos?"**

Elección del target final:

- `opencode` → suficiente para re-materializar OpenCode y la policy del overlay
- `all` → además reinstala las skills/wrappers custom en Claude, Codex, Gemini y Antigravity

Si la auditoría detecta drift de prompt base, invariantes de perfiles o cambios de topología relevantes, frená ahí y adaptá el overlay antes de correr `sync` o reinstall.

El flujo completo hace, en una sola pasada:

1. auditoría read-only del baseline upstream versionado (`gentle-orchestrator.last.md` + `.meta.yaml`)
2. reinstalación de tus skills/wrappers custom
3. poda de skills Gentle AI no deseadas
4. overrides de modelo para `general` y `explore`
5. captura + sanitización de orchestrators inline de OpenCode
6. generación de prompts derivados por orchestrator bajo `~/.config/opencode/prompts/sdd/orchestrators/`
7. actualización dual de snapshots: `gentle-orchestrator` queda versionado en el repo y también copiado al snapshot local operativo; los snapshots per-perfil quedan solo en el directorio local
8. recuperación automática desde snapshot si algún `.overlay.md` fue borrado de disco
9. verificación post-write de que los overrides y las refs `{file:...}` persistieron en `opencode.json`
10. verificación automática fail-closed de que el `gentle-orchestrator` materializado sigue alineado con el último baseline auditado

> **Nota OpenCode:** si el script cambia `~/.config/opencode/opencode.json`, reiniciá OpenCode. La config no se recarga en caliente.

### Qué reporta el script

Al final de cada corrida, el script imprime un bloque `Summary:` con contadores y, si corresponde, bloques `WARNING`/`NOTE`. Los más importantes:

- `--verbose` — además del `Summary:`, imprime un bloque `Verbose changes:` con cada archivo tocado y el detalle concreto de lo que se escribió, regeneró, podó o actualizó.

- `orchestrators kept (already applied): N` — todo estaba aplicado y el script no tuvo que hacer nada. Run idempotente.
- `orchestrators recovered from snapshot: N` — algún `.overlay.md` faltaba en disco y se reconstruyó desde `*.last.md`. Aparece un `NOTE` adicional avisando que el snapshot puede pre-datar la versión actual de upstream — si querés capturar fresco, corré `gentle-ai sync` y volvé a correr el script.
- `repo snapshots - changed: N > 0` — cambió el baseline versionado de `gentle-orchestrator`. Revisalo con `git diff overlay/gentle-ai/snapshots/`.
- `local snapshots - changed: N > 0` — cambió algún snapshot operativo local bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`.
- `local snapshot migrations from repo: N > 0` — el helper copió snapshots legacy desde el repo al directorio local operativo para conservar la recuperación sin pedir un sync inmediato.
- `repo snapshot backfills from local: N > 0` — el helper recreó el snapshot versionado de `gentle-orchestrator` desde la copia operativa local.
- `topology warnings: N > 0` — apareció un orchestrator nuevo, falta uno esperado o algún `agent_override` apunta a una key inexistente. Acción concreta por warning: ver el runbook.
- `audited base baseline verification: ok` — el `gentle-orchestrator` materializado coincide con `gentle-orchestrator.last.md` + `.meta.yaml` y con el overlay sanitizado esperado.
- `SDD profiles managed: N` / `created: N` / `updated: N` / `unchanged: N` — cuántos perfiles del config local se aplicaron y cuántos agent entries se crearon/actualizaron/no cambiaron.
- `SDD profiles unmanaged (present in opencode.json, absent from local config): N` + `WARNING - unmanaged SDD profiles left untouched` — hay perfiles en `opencode.json` que el config local no menciona. El script no los toca. Para gestionarlos, agregalos al config local; para sacarlos, borralos a mano de `opencode.json`.
- `WARNING - keep skills missing` — alguna skill que debería estar conservada está ausente en un target. Probable renombramiento upstream.
- `ERROR: local SDD profile config at ... is not valid JSON` / `... missing required field ...` / `... must be a non-empty string` / `... must match ^[a-z0-9][a-z0-9._-]*$` — el config local no pasa el schema V1 strict. El script **no escribe nada** a `opencode.json` en este caso. Arreglá o eliminá el archivo y volvé a correr.
- `ERROR: broken state for orchestrator X` — `opencode.json` apunta a un archivo inexistente y no hay snapshot para recuperar. Solución: `gentle-ai sync` para resetear a inline, después re-correr el script.
- `ERROR: post-write verification failed: ...` — el script escribió `opencode.json` pero al re-leerlo los valores no coinciden con lo esperado. Suele ser otro proceso escribiendo el archivo en paralelo, o un bug serio del script.
- `ERROR: audited snapshot metadata mismatch ...` — el baseline versionado del repo quedó inconsistente entre `gentle-orchestrator.last.md`, `.meta.yaml`, policy y `upstream-state.json`. Repará el baseline auditado antes de volver a aplicar.
- `ERROR: audited baseline mismatch for orchestrator 'gentle-orchestrator' ...` — corriste `sync`/apply contra un upstream distinto del último baseline auditado, o el snapshot local quedó stale. Solución: auditá primero con `bash audit-gentle-ai-upstream.sh`, actualizá el baseline si corresponde, después `gentle-ai sync` y `apply` de nuevo.

Detalle completo de cada señal en `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.

## Política actual

### Se conservan

- `_shared`
- `cognitive-doc-design`
- `comment-writer`
- `go-testing`
- `judgment-day`
- `sdd-apply`
- `sdd-archive`
- `sdd-design`
- `sdd-explore`
- `sdd-init`
- `sdd-onboard`
- `sdd-propose`
- `sdd-spec`
- `sdd-tasks`
- `sdd-verify`
- `skill-creator`
- `skill-improver`
- `skill-registry`

### Se podan

- `branch-pr`
- `chained-pr`
- `issue-creation`
- `work-unit-commits`

### Overrides de agentes

- `general` → `openai/gpt-5.4` / `high`
- `explore` → `google-vertex/gemini-3.1-pro-preview` / `high`

### Perfiles SDD locales

Los perfiles SDD (`sdd-orchestrator-<name>` + los 10 agentes de fase `sdd-init-<name>`, …, `sdd-onboard-<name>`) **no** se versionan en este repo. Se reconcilian desde un config por-máquina en:

```
~/.config/gentle-ai-custom/opencode-sdd-profiles.json
```

Comportamiento del script:

- Si el archivo **no existe** → el helper no toca ningún perfil SDD en `opencode.json`.
- Si existe → valida estrictamente con schema V1 y **falla cerrado antes de cualquier escritura** si algo está mal.
- Para cada perfil nombrado en el config local → crea o actualiza orchestrator + 10 phase agents con `model` y `variant` exactos.
- Perfiles presentes en `opencode.json` pero **no** nombrados en el config local → quedan intactos pero se reportan como `WARNING - unmanaged SDD profiles left untouched` + contador.
- **Nunca borra perfiles automáticamente**. Si querés sacar uno: editás el config y borrás los agentes correspondientes en `opencode.json` a mano.

Schema V1 (no hay defaults ni herencia):

```jsonc
{
  "version": 1,
  "profiles": [
    {
      "name": "vertex",
      "orchestrator": { "model": "provider/model", "variant": "..." },
      "phases": {
        "sdd-init":     { "model": "provider/model", "variant": "..." },
        "sdd-explore":  { "model": "provider/model", "variant": "..." },
        "sdd-propose":  { "model": "provider/model", "variant": "..." },
        "sdd-spec":     { "model": "provider/model", "variant": "..." },
        "sdd-design":   { "model": "provider/model", "variant": "..." },
        "sdd-tasks":    { "model": "provider/model", "variant": "..." },
        "sdd-apply":    { "model": "provider/model", "variant": "..." },
        "sdd-verify":   { "model": "provider/model", "variant": "..." },
        "sdd-archive":  { "model": "provider/model", "variant": "..." },
        "sdd-onboard":  { "model": "provider/model", "variant": "..." }
      }
    }
  ]
}
```

Reglas duras:

- El top-level debe tener exactamente `version` y `profiles`. Cualquier campo extra rechaza el archivo.
- `version` debe ser exactamente `1`.
- `profiles` debe ser un array no vacío.
- Cada profile debe tener exactamente los campos `name`, `orchestrator`, `phases`. Cualquier campo extra rechaza el archivo.
- `name` debe matchear `^[a-z0-9][a-z0-9._-]*$` (sufijo seguro para agent keys) y ser único.
- Cada `orchestrator`/phase assignment debe tener exactamente `{ "model": "...", "variant": "..." }`.
- `model` debe ser un string no vacío.
- `variant` debe ser un string (puede ser `""` si no aplica), pero el campo es **requerido**.
- `phases` debe contener exactamente los 10 phase keys SDD listados arriba.

El script solo gestiona `model`/`variant`. El `prompt` del orchestrator del perfil viene de `gentle-ai sync` y la sanitización inline existente sigue corriendo igual.

### Snapshots locales de orchestrators

Los snapshots operativos por máquina viven en:

```
~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/
```

Reglas:

- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` sigue versionado en el repo como baseline portable.
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml` fija el hash, la frontera de `upstream-state.json` y las invariantes mínimas esperadas del asset upstream asociado.
- El helper mantiene además `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/gentle-orchestrator.last.md` como copia operativa local preferida para recovery.
- Los snapshots `sdd-orchestrator-<profile>.last.md` viven solo en el directorio local; ya no se versionan en este repo.
- Si el helper encuentra snapshots legacy por perfil todavía presentes en el repo, los migra al directorio local en la próxima corrida.

## Cómo se resuelve el orchestrator

El orchestrator upstream de Gentle AI queda inline por diseño. Esta capa custom **no** usa un prompt estático del repo como source of truth.

En cambio, el helper hace esto:

1. lee el prompt inline actual desde `opencode.json`
2. escribe el snapshot operativo local en `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/<agent>.last.md`
3. si el agente es `gentle-orchestrator`, además valida contra `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` y `gentle-orchestrator.last.meta.yaml`
4. elimina PR/budget/chained-PR/review-workload flow
5. escribe el prompt derivado bajo `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md`
6. cambia la referencia del agente a ese archivo generado

La auditoría previa se hace aparte, directo contra el repo upstream, con:

```bash
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh
```

Ese script NO necesita `gentle-ai sync`: compara el asset upstream real con el baseline versionado y además chequea invariantes livianas de generación de perfiles (`profilePhaseOrder`, prefijo `sdd-orchestrator-`, scoping de permisos de task y binding del asset base a `gentle-orchestrator`).

Recovery/lookup:

- `gentle-orchestrator` → usa primero el snapshot local operativo; si falta, cae al snapshot versionado del repo y lo vuelve a copiar al directorio local.
- `sdd-orchestrator-<profile>` → usa solo el snapshot local operativo. Si falta, el helper falla con mensaje accionable pidiendo `gentle-ai sync`.

Si faltan anchors esperados, el sanitizador falla cerrado y no reescribe automáticamente el prompt.

## Skill y runbook de mantenimiento

Para futuras actualizaciones del upstream:

- Skill: `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- Runbook: `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`

La skill es el punto de entrada recomendado para pedirle al agente que revise diffs del upstream y mantenga actualizados:

- maintenance intent
- scripts
- política
- estado upstream mantenido
- docs
- snapshots
- reglas de sanitización

La skill ahora debe auditar el rango entre la última versión/commit mantenido y el estado actual del upstream, separar cambios relevantes de bugfix/chore noise, decir explícitamente si para traer ese delta alcanza con `gentle-ai sync` o hace falta reinstalación completa, y frenar con gate humana antes de cambiar intención o política para nuevos comportamientos. Regla práctica: cambios de topología upstream => reinstalación; cambios de comportamiento/contenido sin drift de topología => `gentle-ai sync`.

## Comandos custom disponibles

- `/commit-plan`
- `/commit-apply`
- `/commit-fast`
- `/pr-create`
- `/pr-regenerate`

Todos se instalan desde `shared/` y generan wrappers específicos por agente en tiempo de aplicación.
