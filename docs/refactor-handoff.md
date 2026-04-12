# Refactor Handoff

Date: 2026-04-11

## Scope Completed

The multi-phase readability and reuse campaign is complete across core, docker integration, and TUI layers.

Completed outcomes:

- Core service timeout policy is configurable and covered by tests.
- Docker repository internals are split by concern (metrics, logs, resources fan-out).
- TUI update/message/logs/layout/table flows are decomposed into focused files and helpers.
- Shared ANSI-safe text/layout primitives are centralized and reused.
- Regression suites were expanded for logs behavior, tables rendering, browse details/loading, chrome hints, update routing, and message handling.

## Current Architecture

### Core

- `internal/core/service.go`: orchestrates data loading and timeout policy via `ServiceConfig`.
- `internal/core/filter.go`, `internal/core/sort.go`, `internal/core/format.go`, `internal/core/types.go`: pure transformation utilities with dedicated tests.

### Docker Adapter

- `internal/docker/repository.go`: repository shell, client lifecycle, and top-level orchestration.
- `internal/docker/repository_metrics.go`: metrics loading + CPU/memory math helpers.
- `internal/docker/repository_logs.go`: logs fetch/normalization + history append.
- `internal/docker/repository_resources.go`: parallel fan-out for images/networks/volumes/info.
- `internal/docker/mapper.go`: API-to-domain row mapping.

### TUI

- `internal/tui/model.go`: root model and message types.
- `internal/tui/update.go`: top-level event dispatch, key routing via mode router, and consolidated message/loading/tick handlers.
- `internal/tui/loading/orchestrator.go`: loading stage transition orchestration for async result sequencing.
- `internal/tui/selection.go`: root adapter that applies selection state to the model.
- `internal/tui/theme/*`: style ownership and domain-split style composition (chrome, browse, logs, tables, layout).
- `internal/tui/state/selection.go`: tab/scope/cursor state ownership with get/set/clamp/reconcile APIs.
- `internal/tui/mode/router.go`: screen/key routing and mode transition decisions.
- `internal/tui/logs_update.go`: root logs-mode glue only.
- `internal/tui/view.go`: page-level mode routing and frame composition.
- `internal/tui/commands.go`: service-backed command builders.
- `internal/tui/logs/*`: logs domain state, controller, helpers, rendering.
- `internal/tui/browse/view.go`: browse-domain rendering and details.
- `internal/tui/chrome/view.go`: chrome-domain header/footer rendering.
- `internal/tui/tables/*`: table schemas, rendering, and btable implementation.
- `internal/tui/util/*`: shared stateless layout/text/format helpers.

## Symbol Migration Map

### Docker Repository Split

- Legacy concern: monolithic repository implementation in `internal/docker/repository.go`.
- Current locations:
  - `LoadContainerMetrics`, `computeCPUPercent`, `computeMemoryUsage`, `effectiveMemoryUsage` -> `internal/docker/repository_metrics.go`
  - `LoadContainerLiveData`, `normalizeLogs`, `appendHistory`, `tailOption` -> `internal/docker/repository_logs.go`
  - `loadSupportingResourcesData` + result structs -> `internal/docker/repository_resources.go`

### Core Service Timeout Policy

- Legacy behavior: hardcoded timeout branch inside `LoadContainerLiveData`.
- Current locations:
  - `ServiceConfig`, `DefaultServiceConfig`, `NewServiceWithConfig`, `liveDataTimeoutForTail` -> `internal/core/service.go`

### Logs Subsystem

- Legacy behavior: one large logs update file containing input handling, state transitions, merge logic, and rendering.
- Current locations:
  - `State`, `ResultMsg`, source types -> `internal/tui/logs/types.go`
  - merge/sanitize/range helpers -> `internal/tui/logs/helpers.go`
  - state transitions (`Reset*`, `Apply*`, viewport sync) -> `internal/tui/logs/update.go`
  - key and result orchestration (`Controller`) -> `internal/tui/logs/controller.go`
  - logs rendering -> `internal/tui/logs/view.go`
  - root adapter layer -> `internal/tui/logs_update.go`

### Update Message Handling

- Legacy behavior: message handling mixed directly in `Update` path.
- Current locations:
  - window-size/loading/resource/tick/log handlers -> `internal/tui/update.go`
  - loading stage orchestration transitions -> `internal/tui/loading/orchestrator.go`
  - root key routing + browse delegation -> `internal/tui/update.go`

### Mode Routing

- Legacy behavior: screen/key routing and logs enter/exit transition decisions lived inline in root update/logs files.
- Current locations:
  - root key route classification + mode transitions -> `internal/tui/mode/router.go`
  - root adapters for screen enums/state wiring -> `internal/tui/update.go`, `internal/tui/logs_update.go`

### Selection State

- Legacy behavior: tab/scope/cursor mutation logic lived directly in root `internal/tui/selection.go`.
- Current locations:
  - state transitions and cursor APIs -> `internal/tui/state/selection.go`
  - model glue/adaptation -> `internal/tui/selection.go`

### Layout and Text Helpers

- Legacy behavior: duplicated ANSI-width and frame-size math in multiple rendering files.
- Current locations:
  - text shaping/truncation and ANSI helpers -> `internal/tui/util/text.go`
  - line clipping/padding helpers -> `internal/tui/util/lines.go`
  - layout/frame sizing helpers -> `internal/tui/util/layout.go`
  - generic formatting helpers -> `internal/tui/util/format.go`

### Style Ownership

- Legacy behavior: all style construction and ownership lived in root `internal/tui/styles.go`.
- Current locations:
  - style set contract -> `internal/tui/theme/types.go`
  - style composition entrypoint -> `internal/tui/theme/default.go`
  - domain style builders -> `internal/tui/theme/chrome.go`, `internal/tui/theme/browse.go`, `internal/tui/theme/logs.go`, `internal/tui/theme/tables.go`, `internal/tui/theme/layout.go`

### Tables

- Legacy behavior: repeated per-tab table wiring and row/cell rendering branches.
- Current locations:
  - resource specs and row shaping -> `internal/tui/tables/resource_specs.go`
  - root renderer adapter -> `internal/tui/view.go`
  - domain schema/spec/render types -> `internal/tui/tables/specs.go`, `internal/tui/tables/types.go`, `internal/tui/tables/render.go`
  - btable implementation detail -> `internal/tui/tables/btable/component.go`

## Dependency Direction Rules

- Root orchestration package (`internal/tui`) may depend on subpackages (`internal/tui/browse`, `internal/tui/chrome`, `internal/tui/logs`, `internal/tui/mode`, `internal/tui/state`, `internal/tui/tables`, `internal/tui/util`).
- Subpackages must not import the root package path `easydocker/internal/tui`.
- Dependency guardrail: `internal/tui/dependency_rules_test.go` enforces this for production code in subpackages.

## Validation Baseline

Primary validation commands used during the campaign:

```bash
gofmt -w <changed-files>
go test ./internal/tui ./internal/core ./internal/docker
go test ./...
go build ./...
```

Additional runtime smoke target used in plan:

```bash
timeout 8s go run ./cmd/easydocker
```

## Contributor Notes

- Prefer adding new pure helpers under `internal/tui/util` or `internal/tui/logs/helpers.go` before adding branch complexity to stateful update/render functions.
- Keep table behavior routed through spec builders in `internal/tui/tables/resource_specs.go` and root composition in `internal/tui/view.go`.
- Keep tab/scope/cursor transitions inside `internal/tui/state/selection.go`; root should only map state to model fields.
- Keep ANSI clipping/truncation routed through `internal/tui/util/text.go` helpers.
- For new logs interactions, preserve `follow`, history-loading, and viewport offset invariants already covered by `internal/tui/logs_update_test.go`.
