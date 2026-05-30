# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Este repo ya no es solo un instalador de dos skills custom. Ahora funciona como una **capa unificada de personalización y mantenimiento** sobre Gentle AI:

- instala tus skills y wrappers propios
- reaplica tu política local después de `gentle-ai sync`
- depura skills no deseadas del runtime
- fija overrides de modelo para los agentes built-in de OpenCode listados en `agent_overrides` (ver `overlay/gentle-ai/policy/gentle-ai-policy.json`)
- reconcilia perfiles SDD locales (`sdd-orchestrator-<name>` + 10 phase agents) desde un config por-máquina en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`
- captura prompts inline de orchestrators, los sanitiza y genera prompts derivados por agente/perfil.
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
- `shared/skills/commit-planner/SKILL.md` — source of truth neutral para planificación/aplicación de commits
- `shared/skills/pr-finalizer/SKILL.md` — source of truth neutral para creación/regeneración de PRs
- `shared/commands/*.md` — cuerpos compartidos para wrappers/prompts por agente
- `overlay/gentle-ai/README.md` — guía del control-plane de Gentle AI
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — política machine-readable del overlay
- `overlay/gentle-ai/policy/maintenance-intent.md` — intención de mantenimiento del overlay en lenguaje humano/LLM
- `overlay/gentle-ai/policy/orchestrator-policy.md` — criterio de sanitización del orchestrator
- `overlay/gentle-ai/state/upstream-state.json` — estado operativo de la última versión/commit upstream mantenido
- `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` — runbook humano para mantenimiento incremental
- `overlay/gentle-ai/logs/update-log.md` — historial de decisiones del overlay
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` — helper bash interno para depurar Gentle AI
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1` — helper PowerShell interno equivalente
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

### Linux / macOS

```bash
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode
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
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 claude
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 codex
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 gemini
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 antigravity
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 all
```

## Flujo recomendado

```bash
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

Este flujo hace, en una sola pasada:

1. reinstalación de tus skills/wrappers custom
2. poda de skills Gentle AI no deseadas
3. overrides de modelo para `general` y `explore`
4. captura + sanitización de orchestrators inline de OpenCode
5. generación de prompts derivados por orchestrator bajo `~/.config/opencode/prompts/sdd/orchestrators/`
6. recuperación automática desde snapshot si algún `.overlay.md` fue borrado de disco
7. verificación post-write de que los overrides y las refs `{file:...}` persistieron en `opencode.json`

> **Nota OpenCode:** si el script cambia `~/.config/opencode/opencode.json`, reiniciá OpenCode. La config no se recarga en caliente.

### Qué reporta el script

Al final de cada corrida, el script imprime un bloque `Summary:` con contadores y, si corresponde, bloques `WARNING`/`NOTE`. Los más importantes:

- `orchestrators kept (already applied): N` — todo estaba aplicado y el script no tuvo que hacer nada. Run idempotente.
- `orchestrators recovered from snapshot: N` — algún `.overlay.md` faltaba en disco y se reconstruyó desde `*.last.md`. Aparece un `NOTE` adicional avisando que el snapshot puede pre-datar la versión actual de upstream — si querés capturar fresco, corré `gentle-ai sync` y volvé a correr el script.
- `snapshots - changed: N > 0` — los prompts inline upstream cambiaron desde la última corrida. Revisalo con `git diff overlay/gentle-ai/snapshots/`.
- `topology warnings: N > 0` — apareció un orchestrator nuevo, falta uno esperado o algún `agent_override` apunta a una key inexistente. Acción concreta por warning: ver el runbook.
- `SDD profiles managed: N` / `created: N` / `updated: N` / `unchanged: N` — cuántos perfiles del config local se aplicaron y cuántos agent entries se crearon/actualizaron/no cambiaron.
- `SDD profiles unmanaged (present in opencode.json, absent from local config): N` + `WARNING - unmanaged SDD profiles left untouched` — hay perfiles en `opencode.json` que el config local no menciona. El script no los toca. Para gestionarlos, agregalos al config local; para sacarlos, borralos a mano de `opencode.json`.
- `WARNING - keep skills missing` — alguna skill que debería estar conservada está ausente en un target. Probable renombramiento upstream.
- `ERROR: local SDD profile config at ... is not valid JSON` / `... missing required field ...` / `... must be a non-empty string` / `... must match ^[a-z0-9][a-z0-9._-]*$` — el config local no pasa el schema V1 strict. El script **no escribe nada** a `opencode.json` en este caso. Arreglá o eliminá el archivo y volvé a correr.
- `ERROR: broken state for orchestrator X` — `opencode.json` apunta a un archivo inexistente y no hay snapshot para recuperar. Solución: `gentle-ai sync` para resetear a inline, después re-correr el script.
- `ERROR: post-write verification failed: ...` — el script escribió `opencode.json` pero al re-leerlo los valores no coinciden con lo esperado. Suele ser otro proceso escribiendo el archivo en paralelo, o un bug serio del script.

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

## Cómo se resuelve el orchestrator

El orchestrator upstream de Gentle AI queda inline por diseño. Esta capa custom **no** usa un prompt estático del repo como source of truth.

En cambio, el helper hace esto:

1. lee el prompt inline actual desde `opencode.json`
2. genera un snapshot por orchestrator en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`
3. elimina PR/budget/chained-PR/review-workload flow
4. escribe el prompt derivado bajo `~/.config/opencode/prompts/sdd/orchestrators/<agent>.overlay.md`
5. cambia la referencia del agente a ese archivo generado

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
