# gentle-ai-custom

Configuración custom **fuera del árbol gestionado por `gentle-ai sync`**.

## Objetivo

Guardar acá el source of truth de extensiones propias para GentleAI/OpenCode, de modo que:

- `gentle-ai sync` pueda seguir actualizando `~/.config/opencode`
- las customizaciones no se pierdan
- la reaplicación sea explícita y repetible

## Estructura

- `opencode/skills/commit-planner/SKILL.md` — skill post-SDD para planificar/ejecutar commits
- `opencode/commands/commit-plan.md` — slash command read-only para proponer el plan
- `opencode/commands/commit-apply.md` — slash command para ejecutar un plan aprobado o generar uno primero
- `apply-opencode-overrides.sh` — copia estas customizaciones a `~/.config/opencode`

## Uso

```bash
bash ~/Documentos/gentle-ai-custom/apply-opencode-overrides.sh
```

## Flujo recomendado

```bash
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/apply-opencode-overrides.sh
```

## Nota importante

No se parchea `~/.config/opencode/AGENTS.md` automáticamente porque es un asset claramente gestionado por GentleAI y es más frágil frente a updates.

La integración durable queda apoyada en:

- skill custom
- slash commands explícitos
- reaplicación manual post-sync

Si más adelante querés reintroducir auto-load por contexto, conviene hacerlo como overlay/patch separado o directamente upstream.
