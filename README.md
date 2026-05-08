# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Guardar acá el source of truth de overlays propios para OpenCode, Claude, Codex, Antigravity y Gemini CLI, de modo que:

- `gentle-ai sync` pueda seguir actualizando `~/.config/opencode`
- las customizaciones no se pierdan
- la reaplicación sea explícita y repetible

## Estructura

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
```

```powershell
# Windows
Set-ExecutionPolicy -Scope Process Bypass
# (gentle-ai sync si aplica)
.\inject-skills.ps1 all
```

Para Claude, Codex y Gemini CLI no se hace auto-mutation de assets gestionados upstream. La idea sigue siendo la misma: **actualización del agente primero, reaplicación manual después**.

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

**Intención**: velocidad cuando confiás en el agente. Un solo paso en lugar de dos.

**Cuándo usarlo**: cambios chicos y claros donde no necesitás revisar el plan antes de que se aplique.

---

### `/pr-create`

**Qué hace**: genera título y body de PR a partir del diff comprometido de la rama actual, refresca refs remotas con `git fetch` de forma automática y verifica el head remoto con comandos read-only antes de generar el contenido. Respeta la plantilla del repositorio (`.github/PULL_REQUEST_TEMPLATE.md`, `CONTRIBUTING.md`, etc.) o usa una estructura genérica como fallback. La única pausa normal es la aprobación del contenido; después de eso, crea la PR en GitHub sin pedir una segunda confirmación.

**Intención**: producir contenido de PR preciso basado solo en lo que está comprometido, sin inventar cambios ni reutilizar borradores anteriores, con una sola aprobación visible en el flujo normal.

**Cuándo usarlo**: cuando tenés commits locales listos y querés abrir una PR nueva. Si ya existe una PR abierta para la misma rama, el comando te indica que uses `/pr-regenerate` en su lugar.

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
