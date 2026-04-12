# TUI Baseline Snapshot

Date: 2026-04-11

## Goal

Capture an objective baseline for internal/tui complexity and symbol ownership before/while refactoring so progress can be measured over time.

## Metrics Snapshot

- Total Go lines under internal/tui (including subpackages/tests): 5050
- Root package Go lines under internal/tui/\*.go: 2541

### Largest Root Files (current)

- internal/tui/logs_update_test.go: 542
- internal/tui/update_messages_test.go: 347
- internal/tui/tables_view_test.go: 213
- internal/tui/update_test.go: 167
- internal/tui/table_specs_test.go: 153
- internal/tui/selection.go: 134

## Ownership Map (Current)

### Root Orchestration (internal/tui)

- model bootstrap + message dispatch: internal/tui/model.go, internal/tui/update.go
- message routing adapters: internal/tui/update_messages.go, internal/tui/resource_handlers.go, internal/tui/loading_handlers.go, internal/tui/tick_handlers.go
- browse key adapter: internal/tui/browse_keys.go
- root state glue: internal/tui/selection.go
- root mode transitions: internal/tui/logs_update.go
- root render composition: internal/tui/view.go, internal/tui/browse_view.go, internal/tui/chrome_view.go, internal/tui/tables_view.go

### Subpackage Ownership

- mode routing contracts: internal/tui/mode/router.go
- selection state transitions: internal/tui/state/selection.go
- logs domain state/controller/rendering: internal/tui/logs/\*
- browse domain rendering: internal/tui/browse/\*
- chrome domain rendering: internal/tui/chrome/\*
- tables schema/spec/rendering: internal/tui/tables/\*
- shared stateless helpers: internal/tui/util/\*

## Initial Interpretation

- Root still contains substantial surface area, but mutation-heavy concerns are now increasingly delegated.
- Remaining bloat is concentrated in root test files and orchestration adapters, which aligns with current refactor phase focus.
- This snapshot should be compared again after phase 4 and phase 6 completion.
