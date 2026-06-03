# gentle-ai-custom

Una capa de customización mantenida con IA para [Gentle AI](https://github.com/Gentleman-Programming/gentle-ai): extiende la experiencia base con skills, prompts, wrappers y política operativa propios, y hace más mantenible el trabajo de reaplicar y auditar lo que `gentle-ai sync` vuelve a materializar.

## Qué es Gentle AI

[Gentle AI](https://github.com/Gentleman-Programming/gentle-ai) es el proyecto original de Gentleman Programming para mejorar de forma MUY fuerte la experiencia de desarrollo con IA: agentes, skills, orquestación SDD, perfiles y tooling real para trabajar mejor con asistentes en el código.

Este repo existe sobre esa base, no en reemplazo. La idea es dar el crédito que corresponde al proyecto upstream y, al mismo tiempo, construir una capa más mantenible para workflows concretos.

## Por qué existe este repo

Gentle AI resuelve gran parte de la experiencia, pero hay decisiones del upstream que no siempre encajan con mi flujo diario. Por eso este repo actúa como una **capa de customización mantenida con IA**:

- conserva lo mejor del upstream
- depura lo que no se adapta a mi forma de trabajar
- agrega skills y wrappers que sí uso todos los días
- convierte una customización profunda de Gentle AI en algo más mantenible a través de automatización, auditoría y agentes

Hoy sigue estando orientado principalmente a mi flujo de trabajo. La idea es seguir iterándolo para que, con el tiempo, resulte más simple de adaptar y usar también en otros contextos.

## Visión

La dirección de este repo es seguir mejorando la experiencia con una capa cada vez más amigable: un instalador TUI más personalizado, mejor ergonomía operativa y más posibilidades de expansión para skills, overlays y flujos de trabajo reales.

## Objetivo

Hoy este repo funciona como una **capa unificada de personalización y mantenimiento** sobre Gentle AI:

- instala skills y wrappers propios
- reaplica la política local luego de `gentle-ai sync` o un reinstall completo
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

- `maintenance-intent.md` explica qué quiero conservar y depurar, por qué, y qué reglas de sanitización debe respetar el orchestrator derivado
- `gentle-ai-policy.json` alimenta la lógica operativa de los scripts
- `upstream-state.json` guarda desde qué versión/tag/commit hay que auditar el upstream (última versión mantenida)
- `update-log.md` deja historial narrativo de decisiones de mantenimiento

## Por qué Go

La automatización principal vive en Go porque permite mantener **un solo lugar de verdad** para la lógica compartida entre los entrypoints `.sh` y `.ps1`. En vez de duplicar comportamiento entre Bash y PowerShell, ambos wrappers delegan en la misma CLI.

Además, Go ya forma parte natural del stack porque es una dependencia directa de [Engram](https://github.com/Gentleman-Programming/engram). Reutilizarlo acá simplifica el ecosistema, reduce drift entre plataformas y hace más sostenible la evolución del overlay.

## Agentes soportados

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

Todos esos wrappers son finos: delegan en la CLI Go compartida (`go run ./cmd/gentle-ai-overlay ...`) y no duplican la lógica de negocio entre shell y PowerShell.

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

## Flujo de mantenimiento recomendado

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
4. leés `Summary:` y, si hubo drift, el bloque breve `Drift summary:` para decidir si hay algo nuevo que realmente te importe
5. si hace falta, actualizás este repo antes de seguir
6. recién ahí corrés `gentle-ai sync` (o reinstall si la auditoría lo recomienda)
7. `apply-gentle-ai-custom` → responde **"¿quedó materializado en disco lo que ya auditamos?"**

Elección del target final:

- `opencode` → suficiente para re-materializar OpenCode y la policy del overlay
- `all` → además reinstala las skills/wrappers custom en Claude, Codex, Gemini y Antigravity

Si la auditoría detecta drift de prompt base, invariantes de perfiles o cambios de topología relevantes, frená ahí y adaptá el overlay antes de correr `sync` o reinstall. El auditor ahora también imprime un `Drift summary:` corto en lenguaje humano para ayudarte a distinguir si el delta parece relevante para el overlay o si probablemente es ruido de baja prioridad.

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

Al final de cada corrida, el script imprime un bloque `Summary:` con contadores y, si corresponde, bloques `WARNING`/`NOTE`.

Señales clave para leer rápido:

- `--verbose` — además del `Summary:`, imprime `Verbose changes:` con cada archivo tocado.
- `orchestrators kept (already applied): N` — corrida idempotente; ya estaba todo aplicado.
- `repo snapshots - changed: N > 0` o `topology warnings: N > 0` — revisá antes de seguir.
- cualquier `ERROR:` — frená y seguí el runbook.

Detalle completo de señales, recovery y acciones correctivas en `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`.

### Qué reporta la auditoría upstream

`audit-gentle-ai-upstream` sigue siendo read-only, pero ya no te deja solo con `base prompt drift: yes/no`.

- Siempre imprime `Summary:` con el estado del baseline, metadata e invariantes.
- Si detecta drift, imprime además `Drift summary:` con bullets cortos en lenguaje humano.
- Para drift del prompt base, ese resumen debe decir qué cambió y por qué puede importarte: secciones nuevas, cambios de contrato de lenguaje/tono, separación entre conversación directa y artifacts técnicos, y si el delta parece afectar el overlay o más bien ser ruido menor.
- Ejemplo del tipo de drift que ahora resume: puede aparecer una sección nueva como `Language Domain Contract`, cambiar el contrato entre voz conversacional y artifacts técnicos, o moverse el fallback de español hacia wording neutral/profesional. Ese tipo de cambio puede importarte por policy/tono, aunque no sugiera por sí mismo drift de topología o de generación de perfiles.

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

Reglas prácticas:

- el archivo es local por máquina y no se versiona en este repo
- si existe, el helper lo valida en modo strict/fail-closed antes de escribir nada
- el helper solo gestiona `model`/`variant`; el `prompt` sigue viniendo de `gentle-ai sync`

Para el contrato completo, troubleshooting y reglas detalladas, ver `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md` y `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`.

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
