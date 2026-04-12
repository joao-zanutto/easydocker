package tui

import (
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/logs"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_UpdateCrossModeRouting(t *testing.T) {
	m := model{
		width:     120,
		height:    30,
		activeTab: tabContainers,
		showAll:   true,
		styles:    defaultStyles(),
		logs:      logs.NewState(),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
		},
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	current := updated.(model)
	if current.width != 100 || current.height != 40 {
		t.Fatalf("window size not applied: got (%d,%d)", current.width, current.height)
	}

	updated, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	current = updated.(model)
	if cmd == nil {
		t.Fatalf("enter should return logs load command when container is selected")
	}
	if current.screen != screenModeLogs {
		t.Fatalf("screen = %v, want logs", current.screen)
	}
	if current.logs.ContainerID != "ctr-1" {
		t.Fatalf("logs container = %q, want ctr-1", current.logs.ContainerID)
	}

	updated, cmd = current.Update(tea.KeyMsg{Type: tea.KeyEsc})
	current = updated.(model)
	if cmd != nil {
		t.Fatalf("esc in logs should not schedule command")
	}
	if current.screen != screenModeBrowse {
		t.Fatalf("screen = %v, want browse", current.screen)
	}
	if current.activeTab != tabContainers {
		t.Fatalf("activeTab = %d, want %d", current.activeTab, tabContainers)
	}
}

func TestIntegration_ViewRendersBrowseAndLogsModes(t *testing.T) {
	m := model{
		width:     100,
		height:    28,
		activeTab: tabContainers,
		showAll:   true,
		screen:    screenModeBrowse,
		styles:    defaultStyles(),
		logs:      logs.NewState(),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", Image: "nginx", Status: "Up"}},
		},
	}

	browseView := m.View()
	if !strings.Contains(browseView, "EasyDocker") {
		t.Fatalf("browse view missing header")
	}

	m.screen = screenModeLogs
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: []string{"line-1", "line-2"}}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())

	logsView := m.View()
	if !strings.Contains(logsView, "Logs") || !strings.Contains(logsView, "api") {
		t.Fatalf("logs view missing logs breadcrumb context")
	}
}

func TestIntegration_UpdateResultFlow(t *testing.T) {
	m := model{
		showAll:      true,
		loading:      true,
		loadingStage: loadStageContainers,
		styles:       defaultStyles(),
		logs:         logs.NewState(),
	}

	updated, cmd := m.Update(containersResultMsg{containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}})
	current := updated.(model)
	if cmd == nil || current.loadingStage != loadStageResources {
		t.Fatalf("expected transition to resources stage")
	}

	updated, cmd = current.Update(resourcesResultMsg{snapshot: core.Snapshot{Images: []core.ImageRow{{ID: "img-1"}}, Networks: []core.NetworkRow{{ID: "net-1"}}, Volumes: []core.VolumeRow{{Name: "vol-1"}}, TotalLimit: 1000}})
	current = updated.(model)
	if cmd == nil || current.loadingStage != loadStageMetrics {
		t.Fatalf("expected transition to metrics stage")
	}

	updated, cmd = current.Update(metricsResultMsg{metricsByID: map[string]core.ContainerMetrics{"ctr-1": {CPUPercent: 12.5, MemoryUsage: "10 MiB", MemoryLimit: "100 MiB", MemoryUsageBytes: 10, MemoryLimitBytes: 100, MemoryPercent: 10}}, totalCPU: 12.5, totalMem: 10})
	current = updated.(model)
	if cmd != nil {
		t.Fatalf("metrics stage should not schedule command")
	}
	if current.loading || current.loadingStage != loadStageIdle {
		t.Fatalf("expected load flow to finish at idle")
	}
	if current.snapshot.TotalCPU != 12.5 || current.snapshot.TotalMem != 10 {
		t.Fatalf("totals not applied")
	}
}
