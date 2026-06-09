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

Hoy este repo funciona como una **capa unificada de personalización y mantenimiento** sobre Gentle AI. En lugar de modificar el proyecto original a mano cada vez que se actualiza, este repositorio automatiza la adaptación del entorno para que se ajuste a un flujo de trabajo específico.

Su meta principal es permitir consumir las actualizaciones del proyecto original (`gentle-ai sync` o reinstalaciones) sin perder las configuraciones propias, manteniendo una base segura y auditable. El detalle técnico de las modificaciones exactas que se aplican (qué skills se conservan o podan, y cómo se altera el orquestador) se encuentra documentado en la sección **Política actual**, más abajo en este documento.

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

Los scripts `audit-gentle-ai-upstream.*` y `sync-gentle-ai-upstream-assets.*` existen, pero no están pensados como punto de entrada normal para una persona. Su uso recomendado es a través del agente en modo mantenimiento; salvo debugging puntual o un caso excepcional, no hace falta correrlos manualmente.

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

## Overlay maintenance

Maintenance in this repo follows a fixed sequence. What changes is not the order, but the maintainer's judgment about whether the upstream delta requires overlay updates before runtime refresh.

### Recommended sequence

1. Update the `gentle-ai` binary.
2. Run `git pull` in your local `gentle-ai` clone.
3. From `gentle-ai-custom`, ask the maintainer agent to run the upstream audit.
4. Before any repo mutation, the maintainer must return a concise decision summary: what is new upstream, what it recommends adopting, what it recommends discarding, and why.
5. Stop for explicit approval before updating this repo, advancing the upstream boundary, syncing approved upstream assets, or refreshing runtime.
6. If you approved a new upstream boundary, run `sync-gentle-ai-upstream-assets` to refresh the approved upstream copies and the audited baseline.
7. Run `gentle-ai sync` or a full reinstall if the audit recommends it.
8. Re-apply the overlay with `apply-gentle-ai-custom`.
9. Finish with a fresh-context consistency review plus a closing summary of what was actually adopted vs discarded and why.
10. Restart OpenCode if `~/.config/opencode/opencode.json` changed.

```bash
brew upgrade gentle-ai
git -C /path/to/gentle-ai pull

# from gentle-ai-custom, the maintainer agent should run this
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh

# only after approving the new upstream boundary
bash ~/Documentos/gentle-ai-custom/sync-gentle-ai-upstream-assets.sh

# only after auditing and updating this repo if needed
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode

# or the equivalent with all (the only supported agent today)
# bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

### What the audit must decide

The upstream audit must answer these operator-level questions before any repo mutation:

- what is actually new upstream and whether it matters to this overlay
- what should be adopted vs explicitly discarded, with rationale for each side
- whether this repo must change first, and whether the correct runtime path is `gentle-ai sync` or a full TUI reinstall

Today the audit discovers drift mainly with `git diff --name-status --find-renames <last_maintained_commit>..HEAD`, filtered through `overlay/gentle-ai/policy/managed-assets.json`, while still keeping structural checks for upstream changes that could break integration even when they are not markdown assets.

The `audit-gentle-ai-upstream.*` and `sync-gentle-ai-upstream-assets.*` scripts are public, but the recommended path is still the maintainer skill so the audit result is turned into an approval-gated decision summary before anything mutates. `sync-gentle-ai-upstream-assets` refreshes `overlay/gentle-ai/assets/upstream/` and the audited `gentle-orchestrator` baseline.

### What `apply-gentle-ai-custom` re-applies

- reinstalls custom skills and wrappers
- prunes rejected skills only for the selected registered CLI targets; unregistered environments remain untouched
- applies local `agent_overrides` when present
- installs repo-owned SDD/runtime assets from `overlay/gentle-ai/assets/owned/...`
- rewrites `opencode.json` so the base and SDD profiles point to those owned files

`opencode` reinstalls OpenCode and the overlay policy from canonical repo-owned sources. `all` expands to every registered agent; today it is equivalent to `opencode` because that is the only supported agent.

The configuration that decides which AI models to use or where the upstream Gentle AI clone lives is stored in a private machine-local file: `~/.config/gentle-ai-custom/opencode-local-config.json`. A later section of this README explains how to build it.

### Version preflight

Before any write, `apply-gentle-ai-custom` compares the installed `gentle-ai` binary version with `overlay/gentle-ai/state/upstream-state.json -> last_maintained_version`.

- exact match -> continue
- older/newer/unknown -> warn first
- interactive runs may continue after confirmation
- in non-interactive mode, if the version does not match or is unknown -> fail immediately

Upstream resolution:

1. `upstream_repo_path` in `opencode-local-config.json`
2. `GENTLE_AI_CUSTOM_UPSTREAM_REPO`
3. fallback to `../gentle-ai` relative to this repo
4. clear error if none of those exist

Behavior when fields are omitted:

- if `profiles` is omitted, the helper applies no named profiles
- if `agent_overrides` is omitted, the helper applies no explicit `general` / `explore` overrides
- if `default_profile` is omitted, the helper leaves the base `gentle-orchestrator` family untouched

### Core maintenance artifacts

| Artifact                                         | Role |
| ------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| `overlay/gentle-ai/policy/maintenance-intent.md` | Semantic source of truth for what to keep, what to depure, and what behavior to protect. |
| `overlay/gentle-ai/policy/gentle-ai-policy.json` | Runtime policy consumed by the CLI and wrappers. |
| `overlay/gentle-ai/policy/managed-assets.json`   | Canonical map of managed assets (owned/upstream) consumed by audit, sync, and apply. |
| `overlay/gentle-ai/state/upstream-state.json`    | Last maintained upstream boundary used as the audit reference. |
| `overlay/gentle-ai/logs/update-log.md`           | Ledger of closed maintenance/alignment events; Git keeps the implementation detail. |

Git already preserves the implementation detail. `update-log.md` is reserved for closed audits, adoption/rejection/postponement decisions, maintenance-contract changes, and incidents/recoveries that affect how this repo stays aligned with Gentle AI upstream.

### Key files and directories

| Path                                                                                             | Role |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`        | Versioned baseline of the audited upstream orchestrator. |
| `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml` | Metadata and minimum invariants for that versioned baseline. |
| `overlay/gentle-ai/assets/`                                                                      | Canonical tree of approved upstream copies plus owned overlay assets. |
| `~/.config/gentle-ai-custom/opencode-local-config.json`                                          | Canonical local config: upstream path, optional `opencode.json` override, `agent_overrides`, `default_profile`, and `profiles`. |
| `overlay/gentle-ai/maintenance.md`                                                               | Central human maintenance guide, signals, and technical notes. |
| `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`                                           | Agent capability that interprets the audit and guides maintenance. |

### Included maintenance tools

- **Maintainer skill**: the recommended way to operate this maintenance flow through an agent.
- **Maintenance**: centralizes the human workflow, high-signal indicators, and useful technical notes.
- **Public audit**: available as a standalone script, but usually does not need manual interpretation outside debugging.

Human maintenance operation lives in `overlay/gentle-ai/maintenance.md`. Maintainer agent behavior lives in its `SKILL.md`.

> **OpenCode note:** if the script changes `~/.config/opencode/opencode.json`, restart OpenCode. Configuration does not hot-reload.

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

### Comportamiento del orchestrator repo-owned

El overlay NO rompe el corazón técnico del orchestrator. Lo que hace es mantener una versión repo-owned que ya excluye la parte de gobernanza de PRs y carga de revisión, para dejar un coordinador SDD útil pero sin preguntas ni gates que no aplican a este flujo.

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

| Caso                            | Qué agentes intervienen                                                                                                                                         | Cómo se resuelve el modelo                                                                                                                   | Dónde se cambia                                         |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Delegación built-in de OpenCode | `general`, `explore`                                                                                                                                            | Este overlay les fija modelo/variant explícitos mediante `agent_overrides`, en vez de dejar que hereden el modelo del agente que los invoca. | `~/.config/gentle-ai-custom/opencode-local-config.json` |
| Familia base SDD                | `gentle-orchestrator`, `sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard` | Se resuelve por `default_profile`, no por `agent_overrides`.                                                                                 | `~/.config/gentle-ai-custom/opencode-local-config.json` |
| Perfiles SDD nombrados          | `sdd-orchestrator-<name>` + `sdd-<phase>-<name>`                                                                                                                | Se resuelve por `profiles`, no por `agent_overrides`.                                                                                        | `~/.config/gentle-ai-custom/opencode-local-config.json` |

En otras palabras: estos overrides no duplican la configuración SDD. Cubren una capa distinta: los agentes built-in de OpenCode que el orchestrator puede usar al delegar fuera del sistema de perfiles SDD.

Si querés fijar modelos para `general` o `explore`, declaralos explícitamente en `agent_overrides` dentro del config local canónico.

### Configuración avanzada

Si necesitás ir más allá de los overrides básicos y querés configurar familias enteras de perfiles SDD, este es el esquema completo que soporta el archivo:

```json
{
  "version": 1,
  "upstream_repo_path": "/path/to/gentle-ai",
  "opencode_config_path": "/path/to/opencode.json",
  "agent_overrides": [
    { "key": "general", "model": "openai/gpt-5.4", "variant": "high" },
    {
      "key": "explore",
      "model": "google-vertex/gemini-3.1-pro-preview",
      "variant": "high"
    }
  ],
  "default_profile": {
    "orchestrator": { "model": "openai/gpt-5.4", "variant": "high" },
    "phases": {
      "sdd-init": { "model": "openai/gpt-5.4", "variant": "medium" },
      "sdd-explore": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-propose": { "model": "openai/gpt-5.5", "variant": "xhigh" },
      "sdd-spec": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-design": { "model": "openai/gpt-5.5", "variant": "high" },
      "sdd-tasks": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-apply": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-verify": { "model": "openai/gpt-5.4", "variant": "xhigh" },
      "sdd-archive": { "model": "openai/gpt-5.4-mini", "variant": "medium" },
      "sdd-onboard": { "model": "openai/gpt-5.4", "variant": "medium" }
    }
  },
  "profiles": [
    {
      "name": "cheap",
      "orchestrator": { "model": "openai/gpt-5.4-mini", "variant": "low" },
      "phases": {
        "sdd-init": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-explore": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-propose": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-spec": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-design": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-tasks": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-apply": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-verify": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-archive": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-onboard": { "model": "openai/gpt-5.4-mini", "variant": "low" }
      }
    }
  ]
}
```

- **`agent_overrides`:** Fija modelos para agentes específicos fuera de SDD (como `general` o `explore`).
- **`default_profile`:** Fija los modelos base para toda la familia de SDD por defecto.
- **`profiles`:** Te permite crear perfiles SDD con nombres personalizados (ej: un perfil "cheap" que use modelos más baratos).

## Comandos custom disponibles

- `/commit-plan`
- `/commit-apply`
- `/commit-fast`
- `/pr-create`
- `/pr-regenerate`

Todos se instalan desde `shared/` y generan wrappers específicos por agente en tiempo de aplicación.
