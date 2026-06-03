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
- audita el baseline upstream de `gentle-orchestrator` antes de sync/reinstall
- depura skills no deseadas del runtime
- fija overrides de modelo para los agentes built-in de OpenCode listados en `agent_overrides` (ver `overlay/gentle-ai/policy/gentle-ai-policy.json`)
- reconcilia perfiles SDD locales (`sdd-orchestrator-<name>` + 10 phase agents) desde un config por-máquina en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`
- captura los prompts inline de los orchestrators inyectados por **gentle-ai**, y genera nuevos prompts derivados por agente/perfil y sanitizados.
- mantiene el snapshot versionado de `gentle-orchestrator` y snapshots operativos locales por máquina bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`
- mantiene el runbook y la skill para auditar futuras actualizaciones del upstream

## Por qué Go

La automatización principal vive en Go porque permite mantener **un solo lugar de verdad** para la lógica compartida entre los entrypoints `.sh` y `.ps1`. En vez de duplicar comportamiento entre Bash y PowerShell, ambos wrappers delegan en la misma CLI.

Además, Go ya forma parte natural del stack porque es una dependencia directa de [Engram](https://github.com/Gentleman-Programming/engram). Reutilizarlo acá simplifica el ecosistema, reduce drift entre plataformas y hace más sostenible la evolución del overlay.

## Agentes soportados

- `opencode` → `~/.config/opencode`

## Uso

Para el uso normal, pensalo así: **hay un solo comando importante** y sirve para reaplicar esta capa sobre OpenCode.

Los entrypoints públicos son estos:

- `apply-gentle-ai-custom.sh`
- `apply-gentle-ai-custom.ps1`

Ese es el comando que vas a usar normalmente. El target recomendado es `opencode`. `all` hoy es equivalente porque `opencode` es el único agente soportado.

El script `audit-gentle-ai-upstream.*` existe, pero no está pensado como punto de entrada normal para una persona. Su uso recomendado es a través del agente en modo mantenimiento; salvo debugging puntual o un caso excepcional, no hace falta correrlo manualmente.

Todos estos wrappers delegan en la CLI Go compartida (`go run ./cmd/gentle-ai-overlay ...`) y no duplican lógica entre shell y PowerShell.

### Uso rápido

### Linux / macOS

```bash
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode
```

Opcional: agregá `--verbose` si querés ver el detalle de archivos tocados.

### Windows (PowerShell 5.1+)

> **Requisito previo — política de ejecución:**
>
> ```powershell
> Set-ExecutionPolicy -Scope Process Bypass
> ```

```powershell
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 opencode
```

Opcional: agregá `--verbose` si querés ver el detalle de archivos tocados.

## Mantenimiento del overlay

El mantenimiento de este repo sigue una secuencia fija. Lo variable no es el orden, sino la evaluación del agente sobre si el delta del upstream exige adaptar primero este overlay.

### Secuencia recomendada

1. Actualizar el binario `gentle-ai`.
2. Hacer `git pull` en `/home/manuel/Documentos/gentle-ai`.
3. Desde `gentle-ai-custom`, pedir al agente que entre en modo mantenimiento y ejecute la auditoría upstream.
4. Si la auditoría detecta drift relevante para el overlay, adaptar este repo antes de continuar.
5. Si no hace falta adaptar nada, ejecutar `gentle-ai sync` o reinstall completo si la auditoría lo recomienda.
6. Reaplicar el overlay con `apply-gentle-ai-custom`.
7. Reiniciar OpenCode si cambió `~/.config/opencode/opencode.json`.

```bash
brew upgrade gentle-ai
git -C ~/Documentos/gentle-ai pull

# desde gentle-ai-custom, normalmente vía el agente maintainer
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh

# solo después de una auditoría satisfactoria
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode

# o equivalente con all (único agente soportado)
# bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

### Qué decide la auditoría

La auditoría upstream responde tres preguntas concretas:

- si el drift detectado parece relevante para el overlay o si es solo ruido de baja prioridad
- si primero hay que adaptar este repo a esa nueva versión del upstream
- si alcanza con ejecutar `gentle-ai sync` o si hace falta un reinstall completo desde la TUI

El script `audit-gentle-ai-upstream.*` es público, pero el uso recomendado es a través del agente con la skill de mantenimiento, para recibir esa salida ya interpretada.

### Qué reaplica `apply-gentle-ai-custom`

- reinstala skills y wrappers propios
- poda skills upstream no deseadas
- aplica overrides de modelo definidos en la policy
- vuelve a materializar prompts derivados de orchestrators en OpenCode
- actualiza snapshots y validaciones necesarias para sostener el overlay

`opencode` re-materializa OpenCode y la policy del overlay. `all` hoy es equivalente porque `opencode` es el único agente soportado.

`~/.config/gentle-ai-custom/opencode-sdd-profiles.json` funciona como resguardo local de los perfiles SDD que querés conservar. Sirve para volver a aplicar tus elecciones de `model` y `variant` cuando un reinstall de Gentle AI las pisa, cuando un perfil desaparece o cuando simplemente querés evitar depender de recordar esa configuración manualmente en cada máquina.

### Artefactos base del mantenimiento

| Artefacto                                        | Rol                                                                           |
| ------------------------------------------------ | ----------------------------------------------------------------------------- |
| `overlay/gentle-ai/policy/maintenance-intent.md` | Fuente semántica de qué conservar, qué depurar y qué comportamiento proteger. |
| `overlay/gentle-ai/policy/gentle-ai-policy.json` | Policy operativa consumida por la CLI y los wrappers.                         |
| `overlay/gentle-ai/state/upstream-state.json`    | Última frontera upstream mantenida que sirve como referencia para auditar.    |
| `overlay/gentle-ai/logs/update-log.md`           | Historial de decisiones y cambios del overlay.                                |

### Archivos y directorios clave

| Path                                                                                             | Función                                                                  |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`        | Baseline versionado del orchestrator upstream auditado.                  |
| `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml` | Metadata e invariantes mínimas del baseline versionado.                  |
| `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`                                          | Config local por máquina para perfiles SDD.                              |
| `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`                                    | Snapshots operativos locales usados durante la reaplicación.             |
| `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`                                        | Procedimiento formal y troubleshooting profundo.                         |
| `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`                                           | Capacidad del agente para digerir la auditoría y guiar el mantenimiento. |

### Herramientas de mantenimiento incluidas

- **Skill maintainer**: es la forma recomendada de operar este mantenimiento desde un agente.
- **Runbook**: concentra el detalle técnico, señales de error y recovery manual.
- **Auditoría pública**: existe como script separable, pero no hace falta interpretarla a mano salvo debugging puntual.

El detalle fino de implementación sobre cómo se sanitiza o reinyecta el orchestrator vive en el runbook.

> **Nota OpenCode:** si el script cambia `~/.config/opencode/opencode.json`, reinicie OpenCode. La configuración no se recarga en caliente.

## Política actual

Esta sección resume la policy en lenguaje humano. El detalle más preciso, técnico y normativo vive en `overlay/gentle-ai/policy/maintenance-intent.md`.

La idea general es simple: esta capa conserva lo que suma capacidad técnica real al trabajo con IA y poda lo que intenta imponer una manera específica de manejar ramas, PRs y carga de revisión. El valor que quiero mantener de Gentle AI está en SDD, testing, documentación, skill tooling y revisión técnica; no en gobernanza de colaboración que no aplica a mi flujo personal.

### Skills del overlay

| Skill                  | Estado      | Qué hace                                                                       | Por qué se queda / se va                                                                                                                     |
| ---------------------- | ----------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `_shared`              | Se conserva | Expone referencias compartidas para skills SDD.                                | Sostiene el resto del flujo sin imponer proceso extra.                                                                                       |
| `cognitive-doc-design` | Se conserva | Ayuda a escribir documentación más clara y escaneable.                         | Mejora la calidad de los artifacts sin meter gobernanza de repo.                                                                             |
| `comment-writer`       | Se conserva | Ayuda a redactar comentarios de colaboración.                                  | Suma claridad comunicacional sin forzar una estrategia de PR.                                                                                |
| `go-testing`           | Se conserva | Da patrones prácticos para tests en Go.                                        | Aporta calidad técnica directa.                                                                                                              |
| `judgment-day`         | Se conserva | Hace revisión adversarial/dual review.                                         | Refuerza criterio técnico y validación real.                                                                                                 |
| `sdd-init`             | Se conserva | Inicializa contexto, testing y capacidades SDD.                                | Es parte del núcleo útil del flujo SDD.                                                                                                      |
| `sdd-explore`          | Se conserva | Explora ideas antes de proponer cambios.                                       | Mejora análisis y reduce implementación impulsiva.                                                                                           |
| `sdd-propose`          | Se conserva | Convierte exploración en propuesta concreta.                                   | Mantiene trazabilidad y foco de cambio.                                                                                                      |
| `sdd-spec`             | Se conserva | Escribe requisitos y escenarios.                                               | Conserva la disciplina del flujo SDD.                                                                                                        |
| `sdd-design`           | Se conserva | Desarrolla el diseño técnico del cambio.                                       | Aporta arquitectura y decisiones explícitas.                                                                                                 |
| `sdd-tasks`            | Se conserva | Baja diseño/especificación a tareas ejecutables.                               | Hace operativo el trabajo sin perder estructura.                                                                                             |
| `sdd-apply`            | Se conserva | Implementa tareas definidas por SDD.                                           | Conserva la parte más útil del flujo técnico.                                                                                                |
| `sdd-verify`           | Se conserva | Verifica que lo aplicado cumpla con spec y tasks.                              | Mantiene validación técnica al final del ciclo.                                                                                              |
| `sdd-archive`          | Se conserva | Cierra el cambio y deja persistencia final.                                    | Completa el circuito SDD con trazabilidad.                                                                                                   |
| `sdd-onboard`          | Se conserva | Guía un uso completo del flujo SDD.                                            | Facilita adopción sin tocar gobernanza de repo.                                                                                              |
| `skill-creator`        | Se conserva | Ayuda a crear nuevas skills.                                                   | Hace extensible la capa custom.                                                                                                              |
| `skill-improver`       | Se conserva | Ayuda a revisar y mejorar skills existentes.                                   | Permite evolucionar tooling propio con criterio.                                                                                             |
| `skill-registry`       | Se conserva | Indexa las skills disponibles y sus triggers.                                  | Hace mantenible el ecosistema de skills.                                                                                                     |
| `branch-pr`            | Se poda     | Empuja un workflow de branch + PR predefinido.                                 | Esa gobernanza no forma parte de mi flujo local.                                                                                             |
| `chained-pr`           | Se poda     | Empuja PRs encadenadas/stacked como mecánica de trabajo.                       | No quiero que el agente imponga ese modelo por defecto.                                                                                      |
| `issue-creation`       | Se poda     | Empuja la creación de issues dentro del proceso.                               | No necesito que el flujo técnico dependa de ese ritual.                                                                                      |
| `work-unit-commits`    | Se poda     | Impone una estrategia particular de commits por unidad.                        | Prefiero decidir ese criterio según contexto, no como policy fija.                                                                           |
| `code-design`          | Se agrega   | Ayuda a pensar estructura, separación de responsabilidades y diseño de código. | Es una mejora concreta para calidad técnica diaria.                                                                                          |
| `commit-planner`       | Se agrega   | Ordena cambios en commits coherentes.                                          | Sí impone una forma de agrupar commits, pero esa gobernanza responde a mi criterio personal de trabajo y por eso la quiero disponible.       |
| `package-security`     | Se agrega   | Revisa riesgos al instalar o actualizar paquetes.                              | Agrega una capa práctica de seguridad de supply chain.                                                                                       |
| `pr-finalizer`         | Se agrega   | Ayuda a preparar o regenerar PRs a partir de cambios ya hechos.                | Sí introduce una manera concreta de cerrar PRs, pero en este caso refleja mi forma elegida de prepararlas y por eso se incorpora al overlay. |

### Sanitización del orchestrator

El overlay NO rompe el corazón técnico del orchestrator. Lo que hace es sacarle la parte de gobernanza de PRs y carga de revisión para dejar un coordinador SDD útil, pero sin preguntas ni gates que no aplican a este flujo.

| Qué se modifica                                                                                      | Intención                                                                                                   |
| ---------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| Se eliminan las preguntas de preflight sobre estrategia de PR encadenadas y presupuesto de revisión. | No quiero que el flujo arranque preguntando cómo gestionar PRs o cuántas líneas aceptar en review.          |
| Se remueven los bloques del preflight dedicados a PRs y review.                                      | Esa conversación de gobernanza no es parte del valor que busco conservar.                                   |
| Se quitan secciones como `Delivery Strategy`, `Chain Strategy` y `Review Workload Guard`.            | No quiero que el orchestrator me imponga stacked PRs, budgets o forecast de carga como condición del flujo. |
| Se elimina la exigencia de “pasar” un guard de workload antes de `sdd-apply`.                        | La implementación no debe depender de un gate pensado para proteger un proceso de review que acá no uso.    |
| Se neutralizan referencias a `size:exception`, reviewer burden y políticas similares.                | Son reglas válidas en otros contextos, pero no son la fuente de verdad de este entorno.                     |

Lo que se preserva es la parte valiosa: delegación, routing SDD, preflight básico, init guard, dependency graph, TDD forwarding, continuidad de apply y protocolos de contexto. En resumen: se conserva la capacidad técnica y se saca la gobernanza de colaboración.

### Overrides de agentes

Estos overrides existen por cómo OpenCode delega trabajo. Verificado contra la documentación oficial de OpenCode: el sistema trae agentes built-in propios (`Build`, `Plan`, `General`, `Explore`, `Scout`) y, si un subagente no tiene modelo explícito, usa el modelo del agente primario que lo invocó. Esta capa agrega overrides porque, en este setup, quiero fijar de forma explícita qué modelo usan algunos agentes built-in de OpenCode cuando entran en juego fuera del sistema de perfiles SDD.

Tabla operativa:

| Caso                            | Qué agentes intervienen                                                                                                                  | Cómo se resuelve el modelo                                                                                                                   | Dónde se cambia                                         |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Delegación built-in de OpenCode | `general`, `explore`                                                                                                                     | Este overlay les fija modelo/variant explícitos mediante `agent_overrides`, en vez de dejar que hereden el modelo del agente que los invoca. | `overlay/gentle-ai/policy/gentle-ai-policy.json`        |
| Delegación del flujo SDD        | `sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard` | Se resuelve por la configuración de perfiles SDD, no por `agent_overrides`.                                                                  | `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` |

En otras palabras: estos overrides no duplican la configuración SDD. Cubren una capa distinta: los agentes built-in de OpenCode que el orchestrator puede usar al delegar fuera del sistema de perfiles SDD.

Valores versionados hoy:

- `general` → `openai/gpt-5.4` / `high`
- `explore` → `google-vertex/gemini-3.1-pro-preview` / `high`

## Comandos custom disponibles

- `/commit-plan`
- `/commit-apply`
- `/commit-fast`
- `/pr-create`
- `/pr-regenerate`

Todos se instalan desde `shared/` y generan wrappers específicos por agente en tiempo de aplicación.
