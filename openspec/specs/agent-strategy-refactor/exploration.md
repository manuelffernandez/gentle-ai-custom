## Exploration: agent-strategy-refactor

### Current State
El motor de instalación en `internal/overlay/apply_custom.go` soporta múltiples agentes (`opencode`, `claude`, `codex`, `gemini`, `antigravity`) mediante dos grandes bloques `switch` hardcodeados:
1. `buildCustomTarget`: Determina la ruta base y mapea los subdirectorios donde van los comandos por agente.
2. `buildCustomCommandContent`: Inyecta el frontmatter YAML necesario para cada agente.

Adicionalmente, la política de overlay (`runApplyPolicyWithOptions` en `apply_policy.go`) se dispara al final de forma condicional para OpenCode o Claude, pero internamente el struct `applyPolicyState` está 100% acoplado a la configuración de OpenCode (`policy.OpenCode.ConfigPath`, etc.).

### Affected Areas
- `internal/overlay/apply_custom.go` — Refactor principal: eliminar los switches, crear la interfaz `Agent` y la implementación concreta `OpenCodeAgent`.
- `internal/overlay/apply_policy.go` — Modificar para que su ejecución (policy de OpenCode) pertenezca o sea invocada como parte del contrato `ApplyOverlay` del `OpenCodeAgent`.

### Approaches
1. **Strategy / Agent Interface** — Crear una interfaz con los métodos que el motor de aplicación necesita para instalar skills, generar frontmatters y aplicar políticas.
   - Pros: Extensible, limpio. "Cada Agent sabe si la tiene [policy] o no". Elimina código muerto al instante.
   - Cons: Requiere repensar el hook de invocación de la política (de global en `RunApplyCustom` a por-agente).
   - Effort: Medium

### Recommendation
Implementar la interfaz `Agent` concreta:
```go
type Agent interface {
    Name() string
    BasePath() string
    InstallMessage() string
    BuildCommands(sources customSourceFiles) []customCommand
    BuildCommandContent(cmd customCommand, body string) string
    SupportsOverlay() bool
    ApplyOverlay(repoRoot string, options applyPolicyOptions) int
}
```
Eliminar todo el soporte de `claude`, `codex`, `gemini` y `antigravity`. Crear un `map[string]Agent` para el registro, e iterar sobre los agentes solicitados. Invocar `agent.ApplyOverlay(...)` si el agente responde `true` a `SupportsOverlay()`. 

### Risks
- Cambiar el hook de la política de `global` a `por-agente` significa que si en el futuro varios agentes soportan policy, las salidas estándar y los logs se entrelazarán en el orden en que se ejecuten. Esto es aceptable y deseable, pero requiere mover la impresión del sumario al nivel de agente.
- Hay dependencias leves en `apply_policy.go` que deberán moverse, pero el estado actual en realidad ya está altamente acoplado a OpenCode, por lo que encapsularlo en `OpenCodeAgent` unifica el dominio.

### Ready for Proposal
Yes — the refactor is straightforward, reduces dead code, and directly satisfies the owner's request.