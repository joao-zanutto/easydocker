package tui

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func TestIntegration_UpdateCrossModeRouting(t *testing.T) {
	m := model{
		width:            120,
		height:           30,
		activeTab:        tabContainers,
		showAll:          true,
		styles:           defaultStyles(),
		logs:             logs.NewState(),
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		logsSpinner:      spinner.New(spinner.WithSpinner(spinner.Dot)),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
		},
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	current := updated.(model)
	if current.width != 100 || current.height != 40 {
		t.Fatalf("window size not applied: got (%d,%d)", current.width, current.height)
	}

	updated, cmd := current.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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

	updated, cmd = current.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
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
		width:            100,
		height:           28,
		activeTab:        tabContainers,
		showAll:          true,
		screen:           screenModeBrowse,
		styles:           defaultStyles(),
		logs:             logs.NewState(),
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		logsSpinner:      spinner.New(spinner.WithSpinner(spinner.Dot)),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", Image: "nginx", Status: "Up"}},
		},
	}

	browseView := m.View().Content
	if !strings.Contains(browseView, "EasyDocker") {
		t.Fatalf("browse view missing header")
	}

	m.screen = screenModeLogs
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: []string{"line-1", "line-2"}}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())

	logsView := m.View().Content
	if !strings.Contains(logsView, "Logs") || !strings.Contains(logsView, "api") {
		t.Fatalf("logs view missing logs breadcrumb context")
	}
}

func TestIntegration_UpdateResultFlow(t *testing.T) {
	m := model{
		showAll:          true,
		loading:          true,
		loadingStage:     loadStageContainers,
		styles:           defaultStyles(),
		logs:             logs.NewState(),
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		logsSpinner:      spinner.New(spinner.WithSpinner(spinner.Dot)),
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
		width:            120,
		height:           30,
		activeTab:        tabContainers,
		showAll:          true,
		loading:          false,
		loadingStage:     loadStageIdle,
		styles:           defaultStyles(),
		logs:             logs.NewState(),
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		logsSpinner:      spinner.New(spinner.WithSpinner(spinner.Dot)),
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
		width:            120,
		height:           30,
		activeTab:        tabContainers,
		showAll:          true,
		loading:          true,
		loadingStage:     loadStageMetrics,
		styles:           defaultStyles(),
		logs:             logs.NewState(),
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		logsSpinner:      spinner.New(spinner.WithSpinner(spinner.Dot)),
		snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"}},
		},
	}

	before := m.View().Content
	if !strings.Contains(before, "loading metrics") {
		t.Fatalf("expected pre-initial metrics view to include loading stage indicator, got %q", before)
	}

	m.metricsLoaded = true
	after := m.View().Content
	if strings.Contains(after, "loading metrics") {
		t.Fatalf("expected post-initial metrics view to avoid loading indicator, got %q", after)
	}
}

func TestIntegration_BackspaceDoesNotQuitOrExitFilter(t *testing.T) {
	m := New(nil).(model)
	m.width = 100
	m.height = 30
	m.screen = screenModeBrowse

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
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

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	after := updated.(model)
	if !after.browseFilterActive {
		t.Fatalf("backspace in filter mode should not exit filter mode")
	}
	if after.browseFilterQuery != "ab" {
		t.Fatalf("backspace in filter mode should edit text, got %q", after.browseFilterQuery)
	}
}

func TestIntegration_LogsWrapToggleWithW(t *testing.T) {
	m := New(nil).(model)
	m.width = 80
	m.height = 24
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: []string{"abcdefghijklmnopqrstuvwxyz"}}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	current := updated.(model)
	if !current.logs.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}
	if current.logs.WrapXOffset != 0 {
		t.Fatalf("wrap should preserve zero offset by default, got %d", current.logs.WrapXOffset)
	}

	wrappedView := current.logs.Viewport.View()
	if !strings.Contains(wrappedView, "\n") {
		t.Fatalf("wrapped viewport should render on multiple lines, got %q", wrappedView)
	}

	if !strings.Contains(current.View().Content, "lines:") {
		t.Fatalf("wrapped line rows should not inflate total log count, view=%q", current.View().Content)
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	after := updated.(model)
	if after.logs.HorizontalOffset != current.logs.HorizontalOffset {
		t.Fatalf("horizontal scroll should be ignored while wrapped, got %d want %d", after.logs.HorizontalOffset, current.logs.HorizontalOffset)
	}
}

func TestIntegration_LogsWrapTogglePreservesRawLineAnchorWhenNotFollowing(t *testing.T) {
	m := New(nil).(model)
	m.width = 80
	m.height = 24
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"
	logsData := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		logsData = append(logsData, strconv.Itoa(i)+" "+strings.Repeat("x", 48))
	}
	m.logs.Data = core.ContainerLiveData{Logs: logsData}
	m.logs.SetFollow(false)
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())

	nearBottom := max(0, len(logsData)-m.logVisibleRows()-1)
	m.logs.Viewport.SetYOffset(nearBottom)

	beforeList := logs.FilterLogLines(m.logs.Data.Logs, m.logs.FilterQuery)
	beforeStart, _ := logs.VisibleLogRange(m.logs, beforeList)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	after := updated.(model)
	if !after.logs.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}

	afterList := logs.FilterLogLines(after.logs.Data.Logs, after.logs.FilterQuery)
	afterStart, _ := logs.VisibleLogRange(after.logs, afterList)
	if afterStart != beforeStart {
		t.Fatalf("visible raw log anchor changed across wrap toggle, before=%d after=%d", beforeStart, afterStart)
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

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
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

func TestIntegration_ContainersComposeRow_CollapsesAndExpands(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}

	if got := m.itemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should show one row, got %d", got)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterExpand := updated.(model)
	if got := afterExpand.itemCountForTab(tabContainers); got != 3 {
		t.Fatalf("expanded compose list should show project + 2 containers, got %d", got)
	}

	updated, _ = afterExpand.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterCollapse := updated.(model)
	if got := afterCollapse.itemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should return to one row, got %d", got)
	}
}

func TestIntegration_ContainersComposeRow_EnterDoesNotOpenLogs(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := updated.(model)
	if cmd != nil {
		t.Fatalf("enter on compose project row should not open logs command")
	}
	if after.screen != screenModeBrowse {
		t.Fatalf("screen = %v, want browse", after.screen)
	}
	if got := after.itemCountForTab(tabContainers); got != 2 {
		t.Fatalf("enter on compose project should expand row, got item count %d", got)
	}
}

func TestIntegration_ContainersComposeFooterShowsContextualEnterHelp(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}

	composeView := m.View().Content
	if !strings.Contains(composeView, "expand") {
		t.Fatalf("collapsed compose row should advertise expand action, got %q", composeView)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := updated.(model)
	expandedView := after.View().Content
	if !strings.Contains(expandedView, "collapse") {
		t.Fatalf("expanded compose row should advertise collapse action, got %q", expandedView)
	}
}

func TestIntegration_ContainersTabCount_UsesTotalContainersWhenComposeCollapsed(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}

	view := m.View().Content
	if !strings.Contains(util.StripANSI(view), "Containers") {
		t.Fatalf("header should show containers, got %q", view)
	}
	if got := m.itemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should still render one row, got %d", got)
	}
}

func TestIntegration_HorizontalTabSwitchClearsFilter(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeBrowse
	m.activeTab = tabContainers
	m.showAll = true
	m.browseFilterQuery = "redis"
	m.browseFilterInput.SetValue("redis")

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
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

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	current := updated.(model)
	if !current.logs.FilterActive {
		t.Fatalf("slash should activate logs filter mode")
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
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

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
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

	before := m.logs.Viewport.YOffset()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	after := updated.(model)

	if after.logs.Viewport.YOffset() <= before {
		t.Fatalf("expected vertical navigation in logs filter mode to move viewport, before=%d after=%d", before, after.logs.Viewport.YOffset())
	}
	if !after.logs.FilterActive {
		t.Fatalf("logs filter mode should remain active while navigating")
	}
	if after.logs.FilterQuery != "" {
		t.Fatalf("logs filter query should remain unchanged while navigating, got %q", after.logs.FilterQuery)
	}
}

func TestIntegration_LogsFilterFooterShowsNavigationHelp(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.logs.FilterActive = true
	m.logs.FilterInput.Focus()

	view := m.View().Content
	if !strings.Contains(view, "navigate") {
		t.Fatalf("logs filter footer should show navigation help, got %q", view)
	}
	if strings.Contains(view, "←") || strings.Contains(view, "→") {
		t.Fatalf("logs filter footer should only show vertical navigation hints, got %q", view)
	}
}

func TestIntegration_LogsFilterOpen_ReducesRowsFromTop(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.logs.Data = core.ContainerLiveData{Logs: lines}
	m.logs.SetFollow(false)
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	m.logs.Viewport.SetYOffset(10)

	beforeRows := m.logVisibleRows()
	beforeYOffset := m.logs.Viewport.YOffset()
	beforeBottom := beforeYOffset + beforeRows - 1

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	after := updated.(model)

	if !after.logs.FilterActive {
		t.Fatalf("slash should activate logs filter mode")
	}
	afterRows := after.logVisibleRows()
	afterBottom := after.logs.Viewport.YOffset() + afterRows - 1
	if afterRows >= beforeRows {
		t.Fatalf("expected fewer visible rows after opening filter, before=%d after=%d", beforeRows, afterRows)
	}
	if afterBottom != beforeBottom {
		t.Fatalf("opening filter should trim rows from top (preserve bottom), beforeBottom=%d afterBottom=%d", beforeBottom, afterBottom)
	}
}

func TestIntegration_LogsFilterOpenClose_NoViewportDrift(t *testing.T) {
	m := New(nil).(model)
	m.width = 120
	m.height = 34
	m.screen = screenModeLogs
	m.activeTab = tabContainers
	m.snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.logs.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.logs.Data = core.ContainerLiveData{Logs: lines}
	m.logs.SetFollow(false)
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	m.logs.Viewport.SetYOffset(20)

	baseRows := m.logVisibleRows()
	baseBottom := m.logs.Viewport.YOffset() + baseRows - 1

	current := m
	for i := 0; i < 3; i++ {
		updated, _ := current.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		current = updated.(model)
		if !current.logs.FilterActive {
			t.Fatalf("cycle %d: slash should activate logs filter mode", i)
		}

		updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		current = updated.(model)
		if current.logs.FilterActive {
			t.Fatalf("cycle %d: enter should close logs filter mode", i)
		}

		bottom := current.logs.Viewport.YOffset() + current.logVisibleRows() - 1
		if bottom != baseBottom {
			t.Fatalf("cycle %d: viewport drift detected, bottom=%d want=%d", i, bottom, baseBottom)
		}
	}
}

func TestIntegration_BrowseFilterInputView_UsesDynamicLineWidth(t *testing.T) {
	m := New(nil).(model)
	m.browseFilterInput.SetValue("abc")
	view := m.renderBrowseFilterInputView(20)
	if !strings.Contains(view, "🔎︎") || !strings.Contains(view, "abc") {
		t.Fatalf("expected prompt and value in browse filter input view, got %q", view)
	}
	if m.browseFilterInput.Width() != 0 {
		t.Fatalf("render helper should not mutate model input width, got %d", m.browseFilterInput.Width())
	}
}

func TestIntegration_ShouldPollLogsOnTick_GatedByLogLoadingState(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeLogs
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: make([]string, 220)}
	for i := range m.logs.Data.Logs {
		m.logs.Data.Logs[i] = "line"
	}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	m.logs.Viewport.GotoTop()

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("should request history when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while viewport is at top and history is available")
	}

	m.logs.InitialLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while initial load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while initial logs load is in progress")
	}

	m.logs.InitialLoad = false
	m.logs.HistoryLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while history load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while history logs load is in progress")
	}

	m.logs.HistoryLoad = false
	m.logs.HistoryDone = true
	m.logs.Viewport.GotoBottom()
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history after history is exhausted")
	}
	if !m.shouldPollLogsOnTick() {
		t.Fatalf("should poll when not at top and no load is active")
	}

	m.screen = screenModeBrowse
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll outside logs screen")
	}
}

func TestIntegration_TickPrefersHistoryLoadAtTop(t *testing.T) {
	m := New(nil).(model)
	m.screen = screenModeLogs
	m.logs.ContainerID = "ctr-1"
	m.logs.Data = core.ContainerLiveData{Logs: make([]string, 220)}
	for i := range m.logs.Data.Logs {
		m.logs.Data.Logs[i] = "line"
	}
	m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	m.logs.Viewport.GotoTop()
	m.logs.InitialLoad = false
	m.logs.HistoryLoad = false
	m.logs.HistoryDone = false

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("expected history load to be scheduled when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("polling should stay disabled at top while history loading is available")
	}

	updated, cmd := m.Update(tickMsg(time.Now()))
	current := updated.(model)
	if cmd == nil {
		t.Fatalf("tick at top should schedule a history load command")
	}
	if current.logs.HistoryLoad {
		t.Fatalf("tick handling should not mark history loading without result handling")
	}
}
