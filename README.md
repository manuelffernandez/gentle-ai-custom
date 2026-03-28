# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Guardar acá el source of truth de overlays propios para OpenCode, Claude y Codex, de modo que:

- `gentle-ai sync` pueda seguir actualizando `~/.config/opencode`
- las customizaciones no se pierdan
- la reaplicación sea explícita y repetible

## Estructura

- `shared/skills/commit-planner/SKILL.md` — source of truth neutral de la skill
- `shared/commands/commit-plan-body.md` — cuerpo compartido para wrappers/prompts en modo `plan`
- `shared/commands/commit-apply-body.md` — cuerpo compartido para wrappers/prompts en modo `apply`
- `inject-skills.sh` — valida targets, renderiza wrappers finos por agente y los instala en los directorios gestionados

Los wrappers específicos de OpenCode, Claude y Codex **ya no se versionan** en este repo. Se generan durante la instalación a partir de las fuentes compartidas.

## Targets soportados

- `opencode` → `~/.config/opencode`
- `claude` → `~/.claude`
- `codex` → `~/.codex`

## Uso

```bash
bash ~/Documentos/gentle-ai-custom/inject-skills.sh opencode
bash ~/Documentos/gentle-ai-custom/inject-skills.sh claude
bash ~/Documentos/gentle-ai-custom/inject-skills.sh codex
bash ~/Documentos/gentle-ai-custom/inject-skills.sh claude codex
bash ~/Documentos/gentle-ai-custom/inject-skills.sh all
```

El script exige targets explícitos para evitar mutaciones innecesarias por default.
Si los archivos de destino ya existen, se reemplazan. Eso es intencional: permite reaplicar overlays tras un sync sin intervención manual.

## Flujo recomendado

```bash
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/inject-skills.sh all
```

Para Claude y Codex no se hace auto-mutation de assets gestionados upstream. La idea sigue siendo la misma: **actualización del agente primero, reaplicación manual después**.

## Nota importante

No se parchean automáticamente archivos gestionados upstream como `~/.config/opencode/AGENTS.md`, `~/.claude/CLAUDE.md` o equivalentes. Son assets frágiles frente a updates y esta repo se limita a reinstalar overlays explícitos.

La integración durable queda apoyada en:

- skill custom compartida (`commit-planner`) como fuente neutral
- cuerpos compartidos para `plan` y `apply`
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
4. La skill se copia desde la misma fuente compartida a `skills/commit-planner/SKILL.md` en cada target.

Esto mantiene el workflow manual post-sync, elimina duplicación authored y deja las diferencias por agente encapsuladas en el instalador.
