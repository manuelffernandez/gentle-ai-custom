# Gentle AI Overlay / Control Plane

Este overlay mantiene una política **persistente y reaplicable** para tu stack de Gentle AI, desacoplada del repo upstream (`/home/manuel/Documentos/gentle-ai`).

## Quick path (post-sync)

1. Ejecutá tu sync normal de Gentle AI.
2. Reaplicá este overlay:
   - Linux/macOS: `bash ~/Documentos/gentle-ai-custom/overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`
   - Windows: `~\Documentos\gentle-ai-custom\overlay\gentle-ai\scripts\apply-gentle-ai-policy.ps1`
3. Verificá que `agent.gentle-orchestrator.prompt` en `~/.config/opencode/opencode.json` apunte al prompt derivado de este repo.
4. Si el script tocó `opencode.json`, reiniciá OpenCode para que tome la nueva config.

## Qué contiene

- `policy/gentle-ai-policy.json`  
  Baseline machine-readable de skills **keep/prune** + rutas de snapshot/prompt derivado + `agent_overrides` (model y variant explícitos para agentes built-in de OpenCode como `general` y `explore`).
- `policy/orchestrator-policy.md`  
  Criterios de limpieza del prompt del orquestador (qué se elimina y qué se conserva).
- `prompts/audit-gentle-ai-update.md`  
  Prompt de auditoría para futuras actualizaciones upstream sin perder decisiones locales.
- `derived/opencode/gentle-orchestrator.md`  
  Prompt derivado (standalone) para `gentle-orchestrator`, sin flujo PR/budget.
- `snapshots/upstream/opencode/gentle-orchestrator.last.md`  
  Snapshot del último prompt upstream/base antes de redirigir a la versión derivada. Si todavía contiene el placeholder inicial, el primer `apply-gentle-ai-policy` lo refresca con el valor real.
- `logs/update-log.md`  
  Log incremental de decisiones y updates aplicados.
- `scripts/apply-gentle-ai-policy.sh` y `.ps1`  
  Scripts con paridad funcional para aplicar política de skills + redirección de prompt.

## Convenciones

- Los scripts `.sh` y `.ps1` son un par: cualquier cambio de comportamiento en uno implica cambio equivalente en el otro.
- No se parchea texto del prompt inline por regex: se snapshottea el valor previo y se cambia únicamente la referencia `agent.gentle-orchestrator.prompt`.
- El repo upstream se trata como **fuente de verdad de entrada**; este overlay como **fuente de verdad de decisiones locales**.
