# EasyDocker — AGENTS.md

## Project

TUI for Docker. Go 1.25, BubbleTea v2, Docker SDK v27.

## Entrypoint

`cmd/easydocker/main.go` — wires `docker.Repository` → `core.Service` → `tui.New(model)`.

## Directory layout

```
cmd/easydocker/          # main
internal/
  core/                  # domain types, Service, compose, sort, filter, format
  docker/                # Docker SDK wrapper (Repository impl)
  tui/                   # BubbleTea model, update, view, commands, keymap
    screens/{browse,viewer,menu}/
    shared/              # tabs, shell, selection, routing
    ui/{chrome,components,tables,theme}/
docs/                    # VitePress docs (separate CI trigger)
install/                 # install scripts (sh + ps1)
```

## Key commands

| Action | Command |
|--------|---------|
| Test all | `go test -v -race -coverprofile=coverage.out ./...` |
| Single package | `go test -v -race ./internal/core/` |
| Build | `go build -v ./cmd/easydocker` |
| Lint | `golangci-lint run` (config: `.golangci.yml` v2, standard + gosec/misspell/nilnesserr) |
| Dead code | `deadcode -test ./...` (install: `go install golang.org/x/tools/cmd/deadcode@latest`) |
| Generate mocks | `go generate ./...` (`//go:generate mockgen` in `internal/core/service.go`) |
| Docs dev | `npm run docs:dev` |
| Docs build | `npm run docs:build` |

## Architecture and testing quirks

- **Mocks**: `go.uber.org/mock/gomock` (not stdlib). Regenerate when `Repository` interface changes via `go generate ./...`.
- **Integration tests** (`internal/tui/integration_test.go`): no Docker daemon. Construct `model` struct literals inline, send `containersResultMsg`/`resourcesResultMsg`/`metricsResultMsg` directly via `Update`, invoke returned `cmd()` synchronously for multi-step transitions. Type-assert via `unwrapModel()`.
- **Dependency rules** (`dependency_rules_test.go`): `screens/`, `shared/`, `ui/` subpackages must never import `internal/tui` — enforced by test.
- **3-stage initial load**: `containersCmd` → `resourcesCmd` → `metricsCmd` (in `model.go:Init()`). Subsequent ticks reload full snapshot in one pass (`Service.LoadSnapshot()`). Poll interval: 1s (`pollInterval` in `model.go`).
- **Timeouts** (`core/service.go`): 5s default, 20s for tails >500 lines, 60s for >2000 lines or `tail=0` (all logs).
- **Spinners**: `metricsSpinner` = `spinner.Points`, `containerSpinner` = `spinner.Points`, `viewer.Spinner` = `spinner.Dot` — three separate spinner instances.
- **Partial-failure**: metrics/resources failures degrade gracefully (`core.Service`).

## Feature conventions

- **Logs viewer**: history loads backward (scroll to top), poll forward (scroll to bottom). Wrap toggle (`w`), follow toggle (`f`), filter (`/`).
- **Compose**: projects auto-collapsed. Enter expands/collapses, does not open logs for project rows.

## Release flow

1. PRs merge to `main` → `release-please` opens/updates a release PR.
2. Release PR merged → GitHub Release published → `release.yml` runs (GoReleaser cross-platform binaries + Docker Hub multi-arch image).
3. PR titles must follow conventional commits (`feat:`, `fix:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:`, `chore:`, `ci:`, `revert:`) — enforced by CI.

## Import paths

Module name is `easydocker` (no github prefix). BubbleTea v2 is imported from `charm.land/bubbletea/v2`, lipgloss v2 from `charm.land/lipgloss/v2`.

## CI

- **Go changes** (*.go, go.mod, go.sum) → test + lint + build
- **Docs changes** → vitepress build
- **Any PR targeting main** → conventional commits title validation
- CGO disabled for builds. Release ldflags: `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`.
