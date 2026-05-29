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
- captura el nuevo prompt inline en los `*.last.md` (snapshot actualizado)
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

---

## Quick path

1. Trabajá desde `gentle-ai-custom`.
2. Leé el intento en `overlay/gentle-ai/policy/maintenance-intent.md`.
3. Leé la política en `overlay/gentle-ai/policy/gentle-ai-policy.json`.
4. Leé el estado en `overlay/gentle-ai/state/upstream-state.json`.
5. Determiná el rango a auditar desde `last_maintained_commit` / `last_maintained_tag` hasta el estado actual del upstream.
6. Revisá snapshots en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`.
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
- [ ] Los snapshots por orchestrator se siguen escribiendo en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`.
- [ ] El sanitizador todavía remueve PR/budget/chained-PR/review-workload sin romper `## Model Assignments`.
- [ ] `agent.general` sigue en `openai/gpt-5.4` / `high`.
- [ ] `agent.explore` sigue en `google-vertex/gemini-3.1-pro-preview` / `high`.
- [ ] `apply-gentle-ai-custom.sh` y `.ps1` siguen siendo el único par de entrypoints públicos y mantienen paridad funcional.

## Notas operativas

- El source of truth del orchestrator **no** es un archivo estático del repo: el script captura el prompt inline real desde `opencode.json`, lo sanitiza y recién después genera el `.overlay.md` operativo.
- Si el sanitizador no encuentra anchors esperados, debe fallar cerrado y no reescribir prompts automáticamente.
- El script principal de uso humano es `apply-gentle-ai-custom.*`; `apply-gentle-ai-policy.*` se mantiene como helper interno invocado por el entrypoint cuando aplica (`opencode` o `claude`).
- El log no reemplaza al estado: `update-log.md` cuenta qué se decidió; `upstream-state.json` marca cuál fue la última versión/commit mantenida.
