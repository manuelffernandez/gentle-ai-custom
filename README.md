# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Guardar acá el source of truth de overlays propios para OpenCode, Claude, Codex, Antigravity y Gemini CLI, de modo que:

- `gentle-ai sync` pueda seguir actualizando `~/.config/opencode`
- las customizaciones no se pierdan
- la reaplicación sea explícita y repetible

## Estructura

- `overlay/gentle-ai/README.md` — guía humana del overlay/control-plane para Gentle AI
- `overlay/gentle-ai/policy/gentle-ai-policy.json` — baseline keep/prune + paths para reaplicación mecánica
- `overlay/gentle-ai/policy/orchestrator-policy.md` — criterio de derivación del prompt limpio del SDD orchestrator
- `overlay/gentle-ai/prompts/audit-gentle-ai-update.md` — prompt reutilizable para auditar updates futuras de Gentle AI
- `overlay/gentle-ai/derived/opencode/gentle-orchestrator.md` — prompt derivado para OpenCode sin flujo PR/budget
- `overlay/gentle-ai/snapshots/upstream/opencode/gentle-orchestrator.last.md` — último prompt upstream capturado antes del redirect al derivado
- `overlay/gentle-ai/logs/update-log.md` — historial incremental de decisiones y auditorías
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh` — script bash para podar skills y redirigir el orchestrator
- `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1` — equivalente PowerShell 5.1+
- `shared/skills/commit-planner/SKILL.md` — source of truth neutral para planificación/aplicación de commits
- `shared/skills/pr-finalizer/SKILL.md` — source of truth neutral para creación/regeneración de PRs
- `shared/commands/commit-plan-body.md` — cuerpo compartido para wrappers/prompts en modo `plan`
- `shared/commands/commit-apply-body.md` — cuerpo compartido para wrappers/prompts en modo `apply`
- `shared/commands/commit-fast-body.md` — cuerpo compartido para wrappers/prompts en modo `auto` (sin pausa)
- `shared/commands/pr-create-body.md` — cuerpo compartido para wrappers/prompts en modo `create`
- `shared/commands/pr-regenerate-body.md` — cuerpo compartido para wrappers/prompts en modo `regenerate`
- `inject-skills.sh` — instalador para Linux/macOS (bash)
- `inject-skills.ps1` — instalador equivalente para Windows (PowerShell 5.1+)
- `AGENTS.md` — instrucciones operativas para agentes de IA
- `CLAUDE.md` — delegación a `AGENTS.md` para Claude Code

Los wrappers específicos de OpenCode, Claude, Codex y Gemini CLI **ya no se versionan** en este repo. Se generan durante la instalación a partir de las fuentes compartidas.

## Targets soportados

- `opencode` → `~/.config/opencode`
- `claude` → `~/.claude`
- `codex` → `~/.codex`
- `gemini` → `~/.gemini`
- `antigravity` → `~/.gemini/antigravity`

## Uso

**Linux / macOS:**

```bash
bash ~/Documentos/gentle-ai-custom/inject-skills.sh opencode
bash ~/Documentos/gentle-ai-custom/inject-skills.sh claude
bash ~/Documentos/gentle-ai-custom/inject-skills.sh codex
bash ~/Documentos/gentle-ai-custom/inject-skills.sh gemini
bash ~/Documentos/gentle-ai-custom/inject-skills.sh antigravity
bash ~/Documentos/gentle-ai-custom/inject-skills.sh claude codex gemini antigravity
bash ~/Documentos/gentle-ai-custom/inject-skills.sh all
```

**Windows (PowerShell 5.1+):**

> **Requisito previo — política de ejecución:** Windows bloquea la ejecución de scripts por defecto. Antes de correr el instalador, desactivá la restricción para el proceso actual:
>
> ```powershell
> Set-ExecutionPolicy -Scope Process Bypass
> ```
>
> Esto aplica solo a la sesión de PowerShell en curso; no modifica la política global del sistema.

```powershell
.\inject-skills.ps1 opencode
.\inject-skills.ps1 claude
.\inject-skills.ps1 codex
.\inject-skills.ps1 gemini
.\inject-skills.ps1 antigravity
.\inject-skills.ps1 claude codex gemini antigravity
.\inject-skills.ps1 all
```

Ambos scripts exigen targets explícitos para evitar mutaciones innecesarias por default.
Si los archivos de destino ya existen, se reemplazan. Eso es intencional: permite reaplicar overlays tras un sync sin intervención manual.

> **Nota Windows — OpenCode:** si OpenCode en tu sistema usa `%APPDATA%\opencode` en lugar de `~\.config\opencode`, ajustá la variable `$targetDir` en `Apply-OpenCode` dentro del PS1.

## Flujo recomendado

```bash
# Linux / macOS
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/inject-skills.sh all
bash ~/Documentos/gentle-ai-custom/overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh
```

```powershell
# Windows
Set-ExecutionPolicy -Scope Process Bypass
# (gentle-ai sync si aplica)
.\inject-skills.ps1 all
.\overlay\gentle-ai\scripts\apply-gentle-ai-policy.ps1
```

Para Claude, Codex y Gemini CLI no se hace auto-mutation de assets gestionados upstream. La idea sigue siendo la misma: **actualización del agente primero, reaplicación manual después**.

> **Nota OpenCode:** si `apply-gentle-ai-policy` cambia `~/.config/opencode/opencode.json`, reiniciá OpenCode. La config no se hot-reloadéa.

## Overlay Gentle AI: para qué existe

El subárbol `overlay/gentle-ai/` es una **capa de control local** sobre el upstream en `/home/manuel/Documentos/gentle-ai`.

No copia el codebase upstream. En cambio, guarda lo que sí te pertenece a vos:

- la política de qué skills conservar y cuáles podar
- el prompt derivado para `gentle-orchestrator`
- el criterio de derivación de ese prompt
- un snapshot del último prompt upstream visto
- un log incremental de decisiones
- scripts con paridad para reaplicar todo después de `gentle-ai sync`

## Política actual de keep/prune

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

Además, el prompt derivado de OpenCode elimina **todo** el flujo de PR/budget del `gentle-orchestrator`.

## Cómo funciona la automatización

### 1. Política

`overlay/gentle-ai/policy/gentle-ai-policy.json` es la baseline machine-readable. Ahí viven:

- listas keep/prune
- rutas del upstream relevante
- ruta del prompt derivado
- ruta del snapshot

### 2. Prompt derivado

`overlay/gentle-ai/derived/opencode/gentle-orchestrator.md` es la versión local que querés ejecutar en OpenCode.

El script **no parchea texto del prompt upstream**. Solo hace dos cosas:

1. snapshottea el prompt previo si todavía no apunta al derivado
2. redirige `agent.gentle-orchestrator.prompt` en `~/.config/opencode/opencode.json` a tu archivo derivado

Eso evita regex frágiles y hace que la adaptación del prompt siga siendo una tarea semántica del agente, no del script.

### 3. Auditoría incremental de updates

Cuando sube una versión nueva de Gentle AI, no hace falta reanalizar todo desde cero.

Usá como punto de entrada:

- `overlay/gentle-ai/prompts/audit-gentle-ai-update.md`

Y pedile al agente que trabaje **desde este repo** (`gentle-ai-custom`) leyendo el upstream como fuente externa en:

- `/home/manuel/Documentos/gentle-ai`

La idea es que cada auditoría nueva actualice solamente lo necesario:

- `gentle-ai-policy.json`
- `orchestrator-policy.md`
- `derived/opencode/gentle-orchestrator.md`
- `snapshots/upstream/opencode/gentle-orchestrator.last.md`
- `logs/update-log.md`
- scripts `.sh` / `.ps1` si cambia la mecánica

## Quick path para futuras iteraciones

1. Actualizás `gentle-ai`.
2. Corrés `gentle-ai sync`.
3. Reaplicás:
   - `inject-skills.*`
   - `apply-gentle-ai-policy.*`
4. Si notás drift o si la update tocó SDD/skills/orchestrator:
   - abrís `gentle-ai-custom`
   - reutilizás `overlay/gentle-ai/prompts/audit-gentle-ai-update.md`
   - actualizás policy, prompt derivado, snapshot, log y scripts

## Comandos disponibles

Los siguientes comandos se instalan en cada agente durante la ejecución del instalador. Todos leen primero la `SKILL.md` correspondiente antes de actuar.

---

### `/commit-plan`

**Qué hace**: inspecciona el working tree y propone un plan de commits agrupados coherentemente, respetando las convenciones del repositorio (o Conventional Commits como fallback).

**Intención**: darte visibilidad y control sobre cómo quedará el historial antes de escribir nada. No toca git.

**Cuándo usarlo**: cuando terminaste una tarea y querés revisar cómo agrupar los cambios antes de commitear. Siempre antes de `/commit-apply` si querés aprobación explícita del plan.

---

### `/commit-apply`

**Qué hace**: ejecuta un plan de commits aprobado. Si no hay un plan aprobado en la conversación, genera uno primero y se detiene para que lo apruebes.

**Intención**: aplicar el plan con control total — nunca commitea sin que hayas visto y aprobado el plan.

**Cuándo usarlo**: después de aprobar el output de `/commit-plan`, o cuando querés generar y aprobar el plan en un solo flujo pero sin ejecución automática.

---

### `/commit-fast`

**Qué hace**: genera el plan de commits y lo ejecuta inmediatamente sin pausar para aprobación. Muestra el plan antes de ejecutar (para auditoría), pero no espera confirmación. Se detiene si encuentra un blocker real: mismo archivo en múltiples commits, posible secreto, cambios no relacionados que no puede separar limpiamente, o fallo en algún `git commit`.

**Nota OpenCode**: el wrapper generado ya no fija `agent:` en el frontmatter; usa la resolución de agente por defecto del entorno.

**Intención**: velocidad cuando confiás en el agente. Un solo paso en lugar de dos.

**Cuándo usarlo**: cambios chicos y claros donde no necesitás revisar el plan antes de que se aplique.

---

### `/pr-create`

**Qué hace**: genera título y body de PR a partir del diff comprometido de la rama actual, refresca refs remotas con `git fetch` de forma automática y verifica el head remoto con comandos read-only antes de generar el contenido. Para detectar PRs existentes, consulta solo PRs abiertas de la misma rama head y, cuando ya está resuelta, exige coincidencia también en la base. PRs cerradas o mergeadas no bloquean la creación. Respeta la plantilla del repositorio (`.github/PULL_REQUEST_TEMPLATE.md`, `CONTRIBUTING.md`, etc.) o usa una estructura genérica como fallback. La única pausa normal es la aprobación del contenido; después de eso, crea la PR en GitHub sin pedir una segunda confirmación.

**Intención**: producir contenido de PR preciso basado solo en lo que está comprometido, sin inventar cambios ni reutilizar borradores anteriores, con una sola aprobación visible en el flujo normal.

**Cuándo usarlo**: cuando tenés commits locales listos y querés abrir una PR nueva. Si ya existe una PR abierta para la misma rama —y, cuando la base ya está resuelta, para la misma base— el comando te indica que uses `/pr-regenerate` en su lugar. PRs cerradas o mergeadas no bloquean.

---

### `/pr-regenerate`

**Qué hace**: regenera desde cero el título y body de una PR existente, usando el diff comprometido actual como única fuente de verdad, refrescando refs remotas con `git fetch` y validando el head remoto antes de editar. No reutiliza el contenido anterior de la PR. La única pausa normal es la aprobación del contenido; después de eso, actualiza la PR en GitHub sin pedir una segunda confirmación.

**Intención**: mantener la PR sincronizada con el estado real de la rama después de nuevos commits o rebase, con menos fricción operativa y una sola decisión del usuario.

**Cuándo usarlo**: cuando la PR ya existe pero su descripción quedó desactualizada respecto a los commits actuales.

---

## Nota importante

No se parchean automáticamente archivos gestionados upstream como `~/.config/opencode/AGENTS.md`, `~/.claude/CLAUDE.md` o equivalentes. Son assets frágiles frente a updates y esta repo se limita a reinstalar overlays explícitos.

La integración durable queda apoyada en:

- skills custom compartidas (`commit-planner`, `pr-finalizer`) como fuentes neutrales
- cuerpos compartidos para `plan`, `apply`, `create` y `regenerate`
- wrappers/slash commands nativos renderizados por agente durante la instalación
- reaplicación manual post-sync

Si más adelante querés reintroducir auto-load por contexto, conviene hacerlo como overlay/patch separado o directamente upstream.

## Arquitectura de render

1. El repo mantiene sólo contenido agent-agnostic en `shared/`.
2. `inject-skills.sh` valida **todos** los targets pedidos antes de copiar nada.
3. Después renderiza wrappers finos con el path de skill y el frontmatter que cada superficie espera:
   - OpenCode → `~/.config/opencode/commands/*.md`
   - Claude → `~/.claude/commands/*.md`
   - Codex → `~/.codex/prompts/*.md`
   - Gemini CLI → `~/.gemini/skills/*/SKILL.md`
4. Las skills se copian desde la misma fuente compartida a `skills/<skill-name>/SKILL.md` en cada target.

Esto mantiene el workflow manual post-sync, elimina duplicación authored y deja las diferencias por agente encapsuladas en el instalador.
