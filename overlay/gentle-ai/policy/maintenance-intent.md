# Intento de mantenimiento — Gentle AI Custom

## Por qué existe este repo

`gentle-ai-custom` existe para conservar las capacidades técnicas valiosas de Gentle AI sin aceptar automáticamente convenciones de workflow que no aplican al flujo personal del usuario.

El objetivo no es forkear el upstream ni reemplazarlo. El objetivo es interponer una capa local que:

- preserve SDD y utilidades útiles
- depure convenciones de PR/branch/review-budget no deseadas
- mantenga criterios locales explícitos y revisables

## Qué se quiere conservar

Se quiere conservar todo lo que mejore estructura, razonamiento y calidad técnica:

- el flujo SDD completo
- skill resolver / registry
- utilidades de documentación y comentarios
- testing útil
- mejora y creación de skills
- revisión adversarial

## Qué NO se versiona

Las elecciones locales por máquina sobre perfiles SDD de OpenCode no forman parte de la policy compartida del repo.

Eso incluye:

- asignaciones de `model` y `variant` por perfil SDD nombrado
- nombres de perfiles personalizados del usuario (`sdd-orchestrator-<perfil>` y sus fases asociadas)
- cualquier combinación local de proveedores/modelos pensada para una máquina o preferencia personal

Esas decisiones viven fuera del repo, en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`.

La policy versionada solo conserva el baseline portable del overlay; la configuración de perfiles locales se proyecta a `opencode.json` en runtime y no debe volver a copiarse dentro de `gentle-ai-policy.json`.

## Qué se quiere depurar

Se quieren depurar convenciones que imponen una forma específica de colaborar en repositorios:

- `branch-pr`
- `chained-pr`
- `issue-creation`
- `work-unit-commits`
- bloques del orchestrator que impongan:
  - estrategia de PR
  - budget de review
  - chained/stacked PRs
  - `size:exception`
  - reviewer burnout protection como política de PR

## Por qué esas convenciones no aplican

Pueden ser válidas para el proyecto upstream o para otros equipos, pero no son la fuente de verdad del workflow local.

En este entorno:

- el valor está en las capacidades técnicas, no en la gobernanza de PRs
- la colaboración de repositorio se resuelve con herramientas y criterios propios
- el agente no debe imponer branch/PR workflow salvo pedido explícito del usuario

## Cómo evaluar cambios upstream

### Cambios relevantes para el overlay

Son relevantes cuando afectan comportamiento observable, experiencia de uso local o assets gestionados por este repo. Ejemplos:

- nuevas skills o cambios en skills existentes
- cambios en prompts de orchestrators o subagentes
- cambios en install/sync o generación de assets
- nuevas convenciones de workflow impuestas por defecto
- cambios en perfiles OpenCode, referencias a agentes o tablas de modelos

### Cambios de baja prioridad o ruido

Normalmente no requieren tocar el overlay si no cambian comportamiento observable:

- bugfixes internos sin impacto en prompts, skills o config generada
- chores de mantenimiento interno
- refactors sin cambio funcional
- docs upstream que no alteran runtime ni assets

## Gate humana obligatoria

Si durante la auditoría aparecen cambios relevantes que puedan modificar:

- keep/prune
- el sanitizador del orchestrator
- la interpretación de qué conservar o depurar

el agente debe **frenar y preguntar** antes de cambiar intención, política o scripts.

## Qué debe actualizarse después de la decisión humana

Una vez tomada la decisión:

- `maintenance-intent.md` si cambió la intención
- `gentle-ai-policy.json` si cambió la política operativa
- `upstream-state.json` cuando el mantenimiento queda cerrado
- `update-log.md` para dejar trazabilidad narrativa

## Regla final

La intención manda sobre la automatización.

Si el script, la policy o la skill entran en conflicto con este archivo, el agente debe tratar este documento como la fuente de verdad semántica y pedir confirmación humana antes de seguir.
