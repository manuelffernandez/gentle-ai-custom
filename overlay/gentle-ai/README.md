# Gentle AI Overlay / Control Plane

Este overlay mantiene una política **persistente y reaplicable** para tu stack de Gentle AI, desacoplada del repo upstream (`/home/manuel/Documentos/gentle-ai`).

## Quick path

1. Actualizá el binario de Gentle AI.
2. Hacé `git pull` en tu clon upstream de `/home/manuel/Documentos/gentle-ai`.
3. Abrí `gentle-ai-custom` y usá la skill maintainer para revisar el delta upstream.
4. Auditá ANTES de sync:
   - Linux/macOS: `bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh`
   - Windows: `~\Documentos\gentle-ai-custom\audit-gentle-ai-upstream.ps1`
5. Si la auditoría exige cambios en este repo, actualizá primero `gentle-ai-custom`.
6. Si la auditoría da OK, corré tu `gentle-ai sync` o reinstall según el cambio upstream auditado.
7. Reaplicá la capa custom:
   - Linux/macOS: `bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode`
   - Windows: `~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 opencode`
   - Usá `all` en lugar de `opencode` si también querés refrescar las skills/wrappers custom de todos los targets soportados.
   - Agregá `--verbose` si querés ver cada archivo tocado y la modificación concreta que hizo el helper.
8. Reiniciá OpenCode si el script tocó `opencode.json`.

Cuando audités cambios del upstream con la skill maintainer, esa skill te tiene que decir además si ese delta se resuelve con `gentle-ai sync` o si requiere reinstalación completa. Si hubo cambios de topología, `sync` no alcanza.

## Qué contiene

- `policy/gentle-ai-policy.json`  
  Baseline machine-readable de keep/prune, overrides de agentes (`general`, `explore`), rutas operativas de OpenCode, y la ruta al config local de perfiles SDD (`opencode.sdd_profiles_local_config_path`).
- `policy/maintenance-intent.md`  
  Fuente de verdad semántica: explica qué se quiere conservar, qué se quiere depurar y por qué.
- `policy/orchestrator-policy.md`  
  Reglas de limpieza del prompt inline de los orchestrators.
- `state/upstream-state.json`  
  Frontera operativa de la última versión/tag/commit upstream mantenido.
- `runbooks/maintain-upstream-overlay.md`  
  Runbook humano para mantener la capa local frente a cambios del upstream.
- `logs/update-log.md`  
  Log incremental de decisiones y updates aplicados.
- `scripts/apply-gentle-ai-policy.sh` y `.ps1`  
  Wrappers internos finos hacia la CLI Go compartida. La implementación real vive en `cmd/gentle-ai-overlay` + `internal/overlay` y poda skills, aplica `agent_overrides`, reconcilia perfiles SDD desde el config local (`~/.config/gentle-ai-custom/opencode-sdd-profiles.json`), captura prompts inline y genera orchestrators derivados bajo `~/.config/opencode/prompts/sdd/orchestrators/`. Mantiene el snapshot versionado de `gentle-orchestrator` en el repo, mantiene snapshots operativos locales bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`, recupera desde snapshot si un `.overlay.md` faltó en disco, detecta drift de topología (orchestrators desconocidos/faltantes), valida estrictamente el config local de perfiles antes de cualquier escritura (fail-closed) y verifica automáticamente que el `gentle-orchestrator` materializado siga alineado con el último baseline auditado.
- `cmd/gentle-ai-overlay/main.go` + `internal/overlay/*.go`
  CLI Go compartida para `apply-custom`, `apply-policy` y `audit-upstream`.

### Config externo gestionado fuera del repo

- `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` (NO está en el repo)  
  Config per-máquina de perfiles SDD. Schema V1 strict, sin defaults ni herencia. Si no existe, el helper no toca ningún perfil SDD en `opencode.json`. Detalle del schema y reglas duras en el `README.md` raíz, sección "Perfiles SDD locales".
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` (NO está en el repo)  
  Directorio operativo per-máquina para snapshots de orchestrators. Guarda siempre `gentle-orchestrator.last.md` como copia operativa local y guarda los `sdd-orchestrator-<profile>.last.md` solo acá.

## Convenciones

- El source of truth del orchestrator **no** es un archivo estático del repo.
- El helper lee el prompt inline real desde `~/.config/opencode/opencode.json`, lo snapshottea por agente, lo sanitiza y recién después genera el `.overlay.md` operativo.
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` conserva solo el baseline versionado de `gentle-orchestrator`.
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` es la fuente operativa local para recuperación: `gentle-orchestrator` se recupera desde ahí y, si falta, puede caer al snapshot versionado del repo; `sdd-orchestrator-<profile>` se recupera solo desde el directorio local.
- Si el `.overlay.md` falta en disco pero el snapshot recuperable existe, el helper recupera desde ese snapshot. Si no existe, falla cerrado pidiendo `gentle-ai sync`.
- Si faltan anchors esperados, el sanitizador debe fallar cerrado y NO reescribir prompts automáticamente.
- El repo upstream se trata como **fuente de verdad de entrada**; este overlay como **fuente de verdad de decisiones locales**.
- El maintainer debe leer las cuatro capas en este orden: `maintenance-intent.md` → `gentle-ai-policy.json` → `upstream-state.json` → `update-log.md`.
- `update-log.md` no reemplaza al estado upstream mantenido; solo deja trazabilidad narrativa.
- Cada cambio sobre cualquier asset del overlay agrega una entrada a `update-log.md` (ver `AGENTS.md` regla 4).
