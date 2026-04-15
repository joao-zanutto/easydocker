package tui

import (
	"strconv"
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/logs"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_UpdateCrossModeRouting(t *testing.T) {
	m := model{
		width:          120,
		height:         30,
		activeTab:      tabContainers,
		showAll:        true,
		styles:         defaultStyles(),
		logs:           logs.NewState(),
		metricsSpinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		logsSpinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
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
		width:          100,
		height:         28,
		activeTab:      tabContainers,
		showAll:        true,
		screen:         screenModeBrowse,
		styles:         defaultStyles(),
		logs:           logs.NewState(),
		metricsSpinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		logsSpinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
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
		showAll:        true,
		loading:        true,
		loadingStage:   loadStageContainers,
		styles:         defaultStyles(),
		logs:           logs.NewState(),
		metricsSpinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		logsSpinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
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
	if !current.metricsLoaded {
		t.Fatalf("metricsLoaded = false, want true after first metrics result")
	}
}

func TestIntegration_ContainerRefreshPreservesRunningMetrics(t *testing.T) {
	m := model{
		width:          120,
		height:         30,
		activeTab:      tabContainers,
		showAll:        true,
		loading:        false,
		loadingStage:   loadStageIdle,
		styles:         defaultStyles(),
		logs:           logs.NewState(),
		metricsSpinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		logsSpinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{
				FullID:           "ctr-1",
				Name:             "api",
				State:            "running",
				CPUPercent:       12.5,
				MemoryPercent:    10,
				MemoryUsage:      "10 MiB",
				MemoryLimit:      "100 MiB",
				MemoryUsageBytes: 10,
				MemoryLimitBytes: 100,
			}},
		},
	}

	updated, _ := m.Update(containersResultMsg{containers: []core.ContainerRow{{
		FullID:      "ctr-1",
		Name:        "api",
		State:       "running",
		CPUPercent:  -1,
		MemoryUsage: "loading",
		MemoryLimit: "-",
	}}})
	current := updated.(model)

	container := current.snapshot.Containers[0]
	if container.CPUPercent != 12.5 || container.MemoryUsage != "10 MiB" {
		t.Fatalf("running metrics were not preserved during refresh: %+v", container)
	}
}

func TestIntegration_LoadingIndicatorOnlyBeforeInitialMetrics(t *testing.T) {
	m := model{
		width:          120,
		height:         30,
		activeTab:      tabContainers,
		showAll:        true,
		loading:        true,
		loadingStage:   loadStageMetrics,
		styles:         defaultStyles(),
		logs:           logs.NewState(),
		metricsSpinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		logsSpinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"}},
		},
	}

	before := m.View()
	if !strings.Contains(before, "loading metrics") {
		t.Fatalf("expected pre-initial metrics view to include loading stage indicator, got %q", before)
	}

	m.metricsLoaded = true
	after := m.View()
	if strings.Contains(after, "loading metrics") {
		t.Fatalf("expected post-initial metrics view to avoid loading indicator, got %q", after)
	}
}

func TestIntegration_BackspaceDoesNotQuitOrExitFilter(t *testing.T) {
	m := New(nil).(model)
	m.width = 100
	m.height = 30
	m.screen = screenModeBrowse

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	current := updated.(model)
	if cmd != nil {
		t.Fatalf("backspace in browse mode should not trigger a command")
	}
	if current.browseFilterActive {
		t.Fatalf("backspace in browse mode should not activate filter mode")
	}

	current.browseFilterActive = true
	current.browseFilterInput.Focus()
	current.browseFilterInput.SetValue("abc")
	current.browseFilterQuery = "abc"

	updated, _ = current.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	after := updated.(model)
	if !after.browseFilterActive {
		t.Fatalf("backspace in filter mode should not exit filter mode")
	}
	if after.browseFilterQuery != "ab" {
		t.Fatalf("backspace in filter mode should edit text, got %q", after.browseFilterQuery)
	}
}

func TestIntegration_FilterPromptIcon(t *testing.T) {
	m := New(nil).(model)
	if m.browseFilterInput.Prompt != "🔎︎ " {
		t.Fatalf("filter prompt = %q, want %q", m.browseFilterInput.Prompt, "🔎︎ ")
	}
}

func TestIntegration_FilterMode_AllowsVerticalNavigation(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.containerCursor = 0
	m.browseFilterActive = true
	m.browseFilterInput.Focus()
	m.browseFilterInput.SetValue("api")
	m.browseFilterQuery = "api"
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api-1", State: "running"},
			{FullID: "ctr-2", Name: "api-2", State: "running"},
		},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	after := updated.(model)
	if after.containerCursor != 1 {
		t.Fatalf("filter mode down should move cursor to 1, got %d", after.containerCursor)
	}
	if !after.browseFilterActive {
		t.Fatalf("filter mode should remain active while navigating")
	}
	if after.browseFilterQuery != "api" {
		t.Fatalf("filter query should remain unchanged while navigating, got %q", after.browseFilterQuery)
	}
}

func TestIntegration_HorizontalTabSwitchClearsFilter(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.browseFilterQuery = "redis"
	m.browseFilterInput.SetValue("redis")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	after := updated.(model)

	if after.activeTab != tabImages {
		t.Fatalf("active tab = %d, want %d", after.activeTab, tabImages)
	}
	if after.browseFilterQuery != "" {
		t.Fatalf("filter query should be cleared on horizontal tab switch, got %q", after.browseFilterQuery)
	}
	if after.browseFilterInput.Value() != "" {
		t.Fatalf("filter input value should be cleared on horizontal tab switch, got %q", after.browseFilterInput.Value())
	}
}

func TestIntegration_LogsFiltering_ByContainsAndClearOnEsc(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: []string{"alpha line", "quick match", "zeta line"}}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	current := updated.(model)
	if !current.logs.FilterActive {
		t.Fatalf("slash should activate logs filter mode")
	}

	updated, _ = current.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	current = updated.(model)
	if current.logs.FilterQuery != "q" {
		t.Fatalf("logs filter query = %q, want q", current.logs.FilterQuery)
	}
	filtered := current.logs.Viewport.View()
	if !strings.Contains(filtered, "quick match") {
		t.Fatalf("matching log line should be visible, got %q", filtered)
	}
	if strings.Contains(filtered, "alpha line") || strings.Contains(filtered, "zeta line") {
		t.Fatalf("non-matching log lines should be hidden, got %q", filtered)
	}

	updated, _ = current.Update(tea.KeyMsg{Type: tea.KeyEsc})
	after := updated.(model)
	if after.logs.FilterActive {
		t.Fatalf("esc should exit logs filter mode")
	}
	if after.logs.FilterQuery != "" {
		t.Fatalf("esc should clear logs filter query, got %q", after.logs.FilterQuery)
	}
	restored := after.logs.Viewport.View()
	if !strings.Contains(restored, "alpha line") || !strings.Contains(restored, "quick match") || !strings.Contains(restored, "zeta line") {
		t.Fatalf("logs viewport should restore full lines after clearing filter, got %q", restored)
	}
}

func TestIntegration_LogsFilterMode_AllowsVerticalNavigation(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"

	lines := make([]string, 0, 80)
	for i := 0; i < 80; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.logs.Data = core.ContainerLiveData{Logs: lines}
	m.logs.FilterActive = true
	m.logs.FilterInput.Focus()
	m.logs.FilterQuery = ""
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	m.logs.SetFollow(false)
	m.logs.Viewport.GotoTop()

	before := m.logs.Viewport.YOffset
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	after := updated.(model)

	if after.logs.Viewport.YOffset <= before {
		t.Fatalf("expected vertical navigation in logs filter mode to move viewport, before=%d after=%d", before, after.logs.Viewport.YOffset)
	}
	if !after.logs.FilterActive {
		t.Fatalf("logs filter mode should remain active while navigating")
	}
	if after.logs.FilterQuery != "" {
		t.Fatalf("logs filter query should remain unchanged while navigating, got %q", after.logs.FilterQuery)
	}
}

func TestIntegration_BrowseFilterInputView_UsesDynamicLineWidth(t *testing.T) {
	m := New(nil).(model)
	m.browseFilterInput.SetValue("abc")
	view := m.renderBrowseFilterInputView(20)
	if !strings.Contains(view, "🔎︎ abc") {
		t.Fatalf("expected prompt and value in browse filter input view, got %q", view)
	}
	if m.browseFilterInput.Width != 0 {
		t.Fatalf("render helper should not mutate model input width, got %d", m.browseFilterInput.Width)
	}
}
