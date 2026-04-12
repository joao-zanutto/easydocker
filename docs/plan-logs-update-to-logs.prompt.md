## Plan: Push logs_update Into logs Package

### Goal

Move logs lifecycle orchestration out of root `internal/tui/logs_update.go` into `internal/tui/logs` while keeping root as a thin model/app wiring layer.

### Target Ownership

1. `internal/tui/logs` owns all logs lifecycle decisions and state transitions.
2. Root `internal/tui` only applies returned transitions to model fields and wires commands.
3. Root keeps only model-specific viewport dimension calculations if they must depend on root header/footer/frame composition.

### API Direction

Add intent/transition output from logs package so it can drive root behavior without importing root.

Suggested types:

```go
type LoadRequest struct {
    ContainerID string
    SessionID   int
    PrevCPU     []float64
    PrevMem     []float64
    Tail        int
    Src         Source
}

type Transition struct {
    ExitToBrowse bool
    ForceTab     int
    Load         *LoadRequest
    Err          error
}
```

Suggested controller API:

```go
func (Controller) Enter(state *State, containerID string) Transition
func (Controller) HandleKey(state *State, key string, containersTab int) Transition
func (Controller) HandleResult(state *State, msg ResultMsg, visibleWidth int, visibleRows int) Transition
```

### Root Integration Shape

`internal/tui/update.go` (or minimal bridge) should:

1. Call logs controller methods.
2. Apply transition fields to model (`screen`, `activeTab`, `err`, etc.).
3. Build logs load commands from `Transition.Load` via existing `loadLogsDataCmd(...)`.

Example flow:

```go
tr := logsController.HandleKey(&m.logs, key, tabContainers)
if tr.ExitToBrowse {
    m.screen = screenModeBrowse
    m.activeTab = tr.ForceTab
}
if tr.Load != nil {
    return m, m.loadLogsDataCmd(
        tr.Load.ContainerID,
        tr.Load.SessionID,
        tr.Load.PrevCPU,
        tr.Load.PrevMem,
        tr.Load.Tail,
        tr.Load.Src,
    )
}
```

### File-Level Refactor Steps

1. Add transition/intents in `internal/tui/logs` (`types.go` or new `orchestrator.go`).
2. Move enter/exit/history-load and result-acceptance orchestration from `internal/tui/logs_update.go` into `internal/tui/logs/controller.go`.
3. Update root handlers to consume transitions and only wire model + commands.
4. Remove or collapse `internal/tui/logs_update.go` once no domain logic remains.
5. Keep only root-specific dimension helpers if still needed by root-owned view composition.

### Test Migration Plan

1. Move logs lifecycle behavior tests from `internal/tui/logs_update_test.go` to `internal/tui/logs/*_test.go` where possible.
2. Keep root tests focused on routing and app integration (`update_test.go`, `integration_test.go`, `update_messages_test.go`).
3. Preserve dependency rule guard in `dependency_rules_test.go`.

### Constraints

1. `internal/tui/logs` must not import `easydocker/internal/tui`.
2. Logs package returns transitions/intents; root applies them.
3. Avoid introducing wrappers that only rename behavior without reducing root responsibility.

### Verification

1. `go test ./internal/tui/...`
2. `go test ./...`
3. `go build ./cmd/easydocker`
4. Optional smoke: `TERM=dumb timeout 8s go run ./cmd/easydocker`

### Tradeoff Summary

Benefits:

1. Logs behavior becomes fully local and easier to evolve/test.
2. Root update path becomes cleaner orchestration.

Costs:

1. One extra translation layer (transition -> root command/model mutation).
2. Slightly more indirection when tracing full event flow end-to-end.
