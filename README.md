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

## Mantenimiento del overlay

El mantenimiento en este repositorio sigue siempre los mismos pasos. Lo único que cambia es decidir si las novedades del proyecto original (upstream) necesitan que actualicemos nuestra capa antes de aplicarlas.

### Pasos recomendados

1. Actualizá el ejecutable de `gentle-ai`.
2. Ejecutá `git pull` en tu carpeta local de `gentle-ai`.
3. Desde la carpeta de `gentle-ai-custom`, pedile al agente de mantenimiento que revise qué cambió en el proyecto original (auditoría).
4. Antes de cambiar cualquier archivo acá, el agente te tiene que dar un resumen claro: qué hay de nuevo, qué sugiere `Adquirir`, `Sanitizar` o `Ignorar`, por qué, y qué comando recomienda para actualizar.
5. Confirmá que estás de acuerdo antes de que el agente modifique cosas en este repositorio o actualice el sistema.
6. Si aceptaste los cambios del proyecto original, ejecutá `sync-gentle-ai-upstream-assets.sh` para copiar esos cambios aprobados a este repositorio.
7. Después de eso, ejecutá el comando recomendado para aplicar todo al sistema: `gentle-ai sync` o bien una reinstalación completa.
8. Volvé a aplicar nuestra capa customizada usando `apply-gentle-ai-custom.sh`.
9. Terminá con una revisión general para asegurar que todo quedó bien y dejá un resumen final de lo que se adquirió, sanitizó o ignoró.
10. Reiniciá OpenCode si el archivo `~/.config/opencode/opencode.json` tuvo algún cambio.

```bash
brew upgrade gentle-ai
git -C /path/to/gentle-ai pull

# desde gentle-ai-custom, el agente de mantenimiento debería ejecutar esto
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh

# solo después de aprobar los cambios
bash ~/Documentos/gentle-ai-custom/sync-gentle-ai-upstream-assets.sh

# solo después de revisar y actualizar este repositorio si hace falta
gentle-ai sync
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode

# o el equivalente para todo (hoy opencode es el único agente soportado)
# bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

### Qué tiene que decidir la auditoría

Antes de tocar cualquier archivo, la auditoría tiene que responder esto de forma simple y fácil de leer:

| Eje | Valores |
| --- | --- |
| `Scope` | `Managed` / `Unmanaged` |
| `Impact` | `Behavioral` / `Runtime` / `Housekeeping` |
| `Decision` | `Adquirir` / `Sanitizar` / `Ignorar` |

- `Adquirir`: incorporar el cambio al overlay/runtime.
- `Sanitizar`: adaptar el cambio para que siga mandando `maintenance-intent`.
- `Ignorar`: se evaluó, pero no aplica al alcance mantenido.
- El informe debe usar estas columnas: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`.
- `Upstream change` tiene que ser un resumen humano corto del delta upstream y de por qué importa; no una lista de rutas de archivos.
- `Follow-up` es opcional; dejalo vacío cuando no haga falta ninguna acción extra.
- `Runtime` incluye wiring, instalación, configuración o materialización del target mantenido, aunque no cambie el comportamiento de los agentes.
- `Housekeeping` cubre documentación irrelevante, agentes no relacionados o fixes internos sin efecto en el target mantenido.
- No usar `descartar` como etiqueta principal del informe.

Hoy en día, la auditoría busca diferencias usando comandos como `git diff` y las filtra para separar lo que sí importa de lo que no, pero sin dejar de revisar cosas importantes de estructura.

Los scripts `audit-gentle-ai-upstream` y `sync-gentle-ai-upstream-assets` se pueden usar a mano, pero es mejor pedirle al agente de mantenimiento que los use. Así te puede explicar qué cambió y pedirte permiso antes de tocar nada. El comando `sync-gentle-ai-upstream-assets` solo copia cosas aprobadas a este repositorio; no actualiza la configuración de tu computadora.

### Cuándo usar `gentle-ai sync` y cuándo reinstalar

- Es preferible usar `gentle-ai sync` si los cambios que adoptamos sirven, pero no rompen de forma profunda cómo están configurados los agentes y perfiles.
- Conviene hacer una reinstalación completa cuando los cambios del proyecto original afectan la estructura básica de los perfiles que usamos, o cuando un simple `gentle-ai sync` ya no alcanza para armar bien la configuración.
- Si el proyecto original agrega soporte para un nuevo agente o plataforma que no usamos, no hace falta reinstalar.
- Las reglas del proyecto original sobre estrategia de PR, PRs encadenadas y presupuesto de revisión se adoptan por defecto. Solo se sanitizan si chocan con una personalización local explícita.

### Qué hace el comando `apply-gentle-ai-custom`

- vuelve a instalar nuestras skills y comandos personalizados.
- borra las skills que no queremos usar, pero solo para los agentes donde estamos trabajando.
- aplica nuestras preferencias de modelos de IA si las configuramos.
- instala los archivos fundamentales para el flujo de trabajo SDD que mantenemos nosotros.
- modifica la configuración `opencode.json` para que el sistema apunte a nuestros archivos.

Ejecutarlo apuntando a `opencode` reinstala todo para OpenCode. Si usás `all` hace lo mismo, porque hoy es el único entorno soportado.

La configuración para elegir qué modelos de IA vas a usar o dónde está clonado el proyecto original se guarda en un archivo solo tuyo en tu computadora: `~/.config/gentle-ai-custom/opencode-local-config.json`. Más abajo se explica cómo armarlo.

### Chequeo previo de versiones

Antes de modificar nada, el comando `apply-gentle-ai-custom` revisa que la versión de `gentle-ai` que tenés instalada coincida con la que estamos manteniendo.

- si son idénticas -> sigue adelante.
- si son distintas o no se sabe -> te avisa primero.
- si lo estás corriendo a mano, podés confirmar y seguir.
- si se ejecuta de forma automática y hay diferencias -> frena todo por seguridad.

Dónde busca el proyecto original:

1. En la ruta `upstream_repo_path` configurada en tu `opencode-local-config.json`.
2. En la variable de entorno `GENTLE_AI_CUSTOM_UPSTREAM_REPO`.
3. Por defecto, asume que está en `../gentle-ai`, al lado de esta misma carpeta.
4. Si no lo encuentra en ningún lado, te tira un error claro.

Si falta algo en tu configuración:

- si no ponés `profiles`, no agrega perfiles especiales.
- si no ponés `agent_overrides`, no cambia los modelos que vienen por defecto.
- si no ponés `default_profile`, no altera la familia base de perfiles SDD.

### Archivos clave del mantenimiento

| Archivo | Para qué sirve |
| --- | --- |
| `overlay/gentle-ai/policy/maintenance-intent.md` | Es el documento donde explicamos claramente nuestra intención: qué cosas queremos conservar, cuáles quitar y cómo queremos que funcione todo. |
| `overlay/gentle-ai/policy/gentle-ai-policy.json` | Son las reglas exactas que leen los scripts cuando se ejecutan. |
| `overlay/gentle-ai/policy/managed-assets.json` | Es el mapa oficial que dice qué archivos estamos controlando y adaptando, incluyendo los overlay JSON, los comandos retenidos de OpenCode, los plugins y la copia aprobada de `engram-protocol.md` que seguimos materializando desde Claude. |
| `overlay/gentle-ai/state/upstream-state.json` | Guarda cuál fue la última versión del proyecto original que revisamos. |
| `overlay/gentle-ai/logs/update-log.md` | Es un registro donde anotamos las decisiones importantes sobre qué cosas nuevas aceptamos o rechazamos. |

Para ver el código exacto que cambió, siempre podés mirar el historial de Git. El archivo `update-log.md` lo usamos solo para anotar las decisiones y eventos importantes que cambian cómo nos alineamos con el proyecto original.

### Carpetas y archivos importantes

| Ruta | Para qué sirve |
| --- | --- |
| `overlay/gentle-ai/assets/` | Carpeta principal con las copias aprobadas del proyecto original y nuestros propios archivos, como los overlay JSON, los comandos retenidos de OpenCode, los plugins y la copia aprobada de `engram-protocol.md` que mantenemos desde Claude. |
| `~/.config/gentle-ai-custom/opencode-local-config.json` | Tu configuración personal: dónde está el código original, qué modelos preferís usar, etc. |
| `overlay/gentle-ai/maintenance.md` | Guía y notas técnicas pensadas para que las lea una persona que esté manteniendo esto. |
| `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` | Instrucciones para que el agente entienda cómo hacer la auditoría y guiarnos en el mantenimiento. |

### Herramientas de mantenimiento incluidas

- **Agente de mantenimiento (Maintainer skill)**: es la forma recomendada de trabajar, pidiéndole al agente que nos asista.
- **Documentación de mantenimiento**: concentra las explicaciones y notas técnicas.
- **Scripts de auditoría**: se pueden usar a mano, pero normalmente no hace falta salvo para buscar algún problema puntual.

Toda la guía para humanos está en `overlay/gentle-ai/maintenance.md`. El comportamiento del agente está en su propio `SKILL.md`.

> **Nota sobre OpenCode:** si el script llega a cambiar tu `~/.config/opencode/opencode.json`, tenés que reiniciar OpenCode porque no toma los cambios al vuelo.

## Política actual

Esta sección resume la policy en lenguaje humano. El detalle más preciso, técnico y normativo vive en `overlay/gentle-ai/policy/maintenance-intent.md`.

La idea general ahora es **upstream-first**: esta capa arranca desde el comportamiento actual de Gentle AI v2+, conserva unas pocas personalizaciones locales explícitas y evita seguir cargando parches viejos si ya no hacen falta. El foco está en mantener un overlay chico y entendible.

### Skills del overlay

| Skill                  | Estado      | Qué hace                                                                       | Por qué se queda / se va                                                                                                                     |
| ---------------------- | ----------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `_shared`              | Se conserva | Expone referencias compartidas para skills SDD.                                | Sostiene el resto del flujo sin imponer proceso extra.                                                                                       |
| `cognitive-doc-design` | Se conserva | Ayuda a escribir documentación más clara y escaneable.                         | Mejora la calidad de los artifacts sin meter gobernanza de repo.                                                                             |
| `comment-writer`       | Se conserva | Ayuda a redactar comentarios de colaboración.                                  | Suma claridad comunicacional sin forzar una estrategia de PR.                                                                                |
| `go-testing`           | Se conserva | Da patrones prácticos para tests en Go.                                        | Aporta calidad técnica directa.                                                                                                              |
| `judgment-day`         | Se conserva | Hace revisión adversarial/dual review y dispara la retrospectiva automática al cerrar desde el hook runtime del overlay. | Refuerza criterio técnico, valida de verdad y deja aprendizaje reutilizable.                                                                |
| `judgment-retrospective` | Se agrega   | Resume el juicio, guarda patrones reutilizables e historial de intervenciones.  | Diseñada para ser invocada por el hook runtime de `judgment-day`; también se puede cargar manualmente. Evita guardar el output crudo y ayuda a ver si una corrección realmente funcionó. |
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
| `branch-pr`            | Se sanitiza | La copia upstream se poda, pero se reinstala como skill repo-owned para crear o regenerar PRs desde cambios ya commiteados. | Conserva el nombre y lo útil del flujo de PR, pero elimina el issue-first universal: issue link, labels y checks quedan como policy contextual del repo, no como regla global. |
| `chained-pr`           | Se retiene + se sanitiza en runtime | El snapshot upstream sigue auditado, pero `apply` instala una copia repo-owned sanitizada para PRs encadenadas o stacked. | Conserva el workflow upstream y agrega protecciones operativas: resolver `owner/repo`, usar `--repo` en comandos `gh`, y permitir planificación automática sin crear branches/PRs hasta que haya pedido explícito. |
| `issue-creation`       | Se poda     | Queda fuera de la instalación por defecto del overlay y no se reinstala como skill repo-owned. | Hoy este repo no necesita imponer por defecto un workflow de issues orientado a colaborar con el upstream; si más adelante vuelve a hacer falta, se puede reevaluar como decisión explícita. |
| `hermes-ephemeral-delegation` | Se poda     | Queda fuera del boundary mantenido de OpenCode.                                 | No forma parte del comportamiento mantenido local.                                                                                           |
| `work-unit-commits`    | Se adquiere | Trae el workflow upstream para dividir cambios en unidades de trabajo.         | Se conserva como criterio upstream de agrupación: work units revisables, evidencia, rollback boundary y awareness de chained/review-budget. |
| `code-design`          | Se agrega   | Ayuda a pensar estructura, separación de responsabilidades y diseño de código. | Es una mejora concreta para calidad técnica diaria.                                                                                          |
| `commit-planner`       | Se agrega   | Ordena cambios en commits coherentes.                                          | Queda como capa operativa local: cuando la invoco, propone o ejecuta commits explícitos y usa `work-unit-commits` como criterio si existe. |
| `package-security`     | Se agrega   | Revisa riesgos al instalar o actualizar paquetes.                              | Agrega una capa práctica de seguridad de supply chain.                                                                                       |
| `pr-finalizer`         | Compat temporal | Sigue instalada de forma transitoria para referencias viejas mientras migra la capa final de PR a `branch-pr`. | Ya no es la skill preferida: queda como compatibilidad/deprecada hasta terminar la migración de comandos y referencias. |

### Comportamiento del orchestrator repo-owned

El runtime owned ya quedó reacoplado a la base upstream de SDD/orchestrator de Gentle AI v2+ por defecto. La idea ya no es mantener un fork grande con reglas locales viejas, sino conservar solo una capa chica y explícita. La frontera mantenida exacta se fija recién cuando se cierre la auditoría/adopción en `upstream-state.json`.

Se mantiene localmente:

- el hook `judgment-day` -> `judgment-retrospective`
- las skills repo-owned `code-design`, `package-security`, `commit-planner` y la `branch-pr` sanitizada, con `pr-finalizer` como compatibilidad temporal
- la copia repo-owned sanitizada de `chained-pr` para runtime, mientras su snapshot upstream sigue retenido para auditoría
- los assets de soporte locales estrictamente necesarios para SDD v2, como `sdd-design/references/threat-matrix.md` y `sdd-verify/references/report-format.md`
- las reglas `no-auto-commit`, Agent-Readable Markdown Standard y Gemini anti-sycophancy

Se adquiere desde upstream por defecto:

- el comportamiento SDD/orchestrator
- el lifecycle moderno de review y remediation
- los conceptos de PR/chained/work-unit/review-budget, salvo conflicto concreto con la capa local

Excepción concreta:

- la copia upstream de `branch-pr` se poda, pero se reemplaza por una `branch-pr` repo-owned sanitizada que no obliga issue aprobado para toda PR

En otras palabras: la fuente semántica ya no es una lista de parches heredados de v1.44, sino `overlay/gentle-ai/policy/maintenance-intent.md`, que define qué queda local y qué vuelve a upstream.

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

`/pr-create` y `/pr-regenerate` mantienen el mismo nombre, pero ahora pasan a estar gobernados por `branch-pr`.
