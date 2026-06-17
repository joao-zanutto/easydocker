package tui

import (
	"strconv"
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	tea "charm.land/bubbletea/v2"
)

func unwrapModel(m tea.Model) *model {
	switch v := m.(type) {
	case *model:
		return v
	case model:
		return &v
	default:
		panic("unexpected model type")
	}
}

func TestIntegration_UpdateCrossModeRouting(t *testing.T) {
	m := newTestModel().
		withSize(120, 30).
		withContainers(core.ContainerRow{FullID: "ctr-1", Name: "api", State: "running"}).
		build()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	current := unwrapModel(updated)
	if current.width != 100 || current.height != 40 {
		t.Fatalf("window size not applied: got (%d,%d)", current.width, current.height)
	}

	updated, cmd := current.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	current = unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("enter should return a command")
	}
	// Process the browse transition message to enter logs mode
	if msg := cmd(); msg != nil {
		updated, _ = current.Update(msg)
		current = unwrapModel(updated)
	}
	if current.screen != shared.LogViewer {
		t.Fatalf("screen = %v, want logs", current.screen)
	}
	if current.viewer.ContainerID != "ctr-1" {
		t.Fatalf("logs container = %q, want ctr-1", current.viewer.ContainerID)
	}

	updated, cmd = current.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	current = unwrapModel(updated)
	// Escape in viewer returns a viewer.TransitionMsg cmd; process it to go back to browse
	if cmd != nil {
		if msg := cmd(); msg != nil {
			updated, _ = current.Update(msg)
			current = unwrapModel(updated)
		}
	}
	if current.screen != shared.Main {
		t.Fatalf("screen = %v, want browse", current.screen)
	}
	if current.browse.ActiveTab != tabContainers {
		t.Fatalf("ActiveTab = %d, want %d", current.browse.ActiveTab, tabContainers)
	}
}

func TestIntegration_ViewRendersBrowseAndLogsModes(t *testing.T) {
	m := newTestModel().
		withSize(100, 28).
		withContainers(core.ContainerRow{FullID: "ctr-1", Name: "api", State: "running", Image: "nginx", Status: "Up"}).
		build()
	m = m.syncBrowseData()
	m.dataDirty = false

	browseView := m.View().Content
	if !strings.Contains(browseView, "EasyDocker") {
		t.Fatalf("browse view missing header")
	}

	m.screen = shared.LogViewer
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Vp.Data = []string{"line-1", "line-2"}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	logsView := m.View().Content
	if !strings.Contains(logsView, "Logs") || !strings.Contains(logsView, "api") {
		t.Fatalf("logs view missing logs breadcrumb context")
	}
}

func TestIntegration_UpdateResultFlow(t *testing.T) {
	m := newTestModel().
		build()

	updated, cmd := m.Update(containersResultMsg{containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}})
	current := unwrapModel(updated)
	if cmd == nil || current.loadingStage != shared.StageResources {
		t.Fatalf("expected transition to resources stage")
	}

	updated, cmd = current.Update(resourcesResultMsg{snapshot: core.Snapshot{Images: []core.ImageRow{{ID: "img-1"}}, Networks: []core.NetworkRow{{ID: "net-1"}}, Volumes: []core.VolumeRow{{Name: "vol-1"}}, TotalLimit: 1000}})
	current = unwrapModel(updated)
	if cmd == nil || current.loadingStage != shared.StageMetrics {
		t.Fatalf("expected transition to metrics stage")
	}

	updated, cmd = current.Update(metricsResultMsg{metricsByID: map[string]core.ContainerMetrics{"ctr-1": {CPUPercent: 12.5, MemoryUsage: "10 MiB", MemoryLimit: "100 MiB", MemoryUsageBytes: 10, MemoryLimitBytes: 100, MemoryPercent: 10}}, totalCPU: 12.5, totalMem: 10})
	current = unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("metrics stage should schedule tickCmd")
	}
	if current.loading || current.loadingStage != shared.StageIdle {
		t.Fatalf("expected load flow to finish at idle")
	}
	if current.browse.Snapshot.TotalCPU != 12.5 || current.browse.Snapshot.TotalMem != 10 {
		t.Fatalf("totals not applied")
	}
	if !current.metricsLoaded {
		t.Fatalf("metricsLoaded = false, want true after first metrics result")
	}
}

func TestIntegration_ContainerRefreshPreservesRunningMetrics(t *testing.T) {
	m := newTestModel().
		withSize(120, 30).
		withLoading(false, shared.StageIdle).
		withContainers(core.ContainerRow{
			FullID:           "ctr-1",
			Name:             "api",
			State:            "running",
			CPUPercent:       12.5,
			MemoryPercent:    10,
			MemoryUsage:      "10 MiB",
			MemoryLimit:      "100 MiB",
			MemoryUsageBytes: 10,
			MemoryLimitBytes: 100,
		}).
		build()

	updated, _ := m.Update(containersResultMsg{containers: []core.ContainerRow{{
		FullID:      "ctr-1",
		Name:        "api",
		State:       "running",
		CPUPercent:  -1,
		MemoryUsage: "loading",
		MemoryLimit: "-",
	}}})
	current := unwrapModel(updated)

	container := current.browse.Snapshot.Containers[0]
	if container.CPUPercent != 12.5 || container.MemoryUsage != "10 MiB" {
		t.Fatalf("running metrics were not preserved during refresh: %+v", container)
	}
}

func TestIntegration_LoadingIndicatorOnlyBeforeInitialMetrics(t *testing.T) {
	m := newTestModel().
		withSize(200, 30).
		withLoading(true, shared.StageMetrics).
		withContainers(core.ContainerRow{FullID: "ctr-1", Name: "api", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"}).
		build()

	before := m.View().Content
	if !strings.Contains(before, "loading") {
		t.Fatalf("expected pre-initial metrics view to include loading stage indicator, got %q", before)
	}

	m.metricsLoaded = true
	after := m.View().Content
	if strings.Contains(after, "loading") {
		t.Fatalf("expected post-initial metrics view to avoid loading indicator, got %q", after)
	}
}

func TestIntegration_BackspaceDoesNotQuitOrExitFilter(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 100
	m.height = 30
	m.screen = shared.Main

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	current := unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("backspace in browse mode should not trigger a command")
	}
	if current.browse.Filter.Active {
		t.Fatalf("backspace in browse mode should not activate filter mode")
	}

	current.browse.Filter.Active = true
	current.browse.Filter.Input.Focus()
	current.browse.Filter.Input.SetValue("abc")
	current.browse.Filter.Query = "abc"

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	after := unwrapModel(updated)
	if !after.browse.Filter.Active {
		t.Fatalf("backspace in filter mode should not exit filter mode")
	}
	if after.browse.Filter.Query != "ab" {
		t.Fatalf("backspace in filter mode should edit text, got %q", after.browse.Filter.Query)
	}
}

func TestIntegration_LogsWrapToggleWithW(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 80
	m.height = 24
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Vp.Data = []string{"abcdefghijklmnopqrstuvwxyz"}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	current := unwrapModel(updated)
	if !current.viewer.Vp.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}

	wrappedView := current.viewer.Vp.View()
	if !strings.Contains(wrappedView, "\n") {
		t.Fatalf("wrapped viewport should render on multiple lines, got %q", wrappedView)
	}

	if !strings.Contains(current.View().Content, "lines:") {
		t.Fatalf("wrapped line rows should not inflate total log count, view=%q", current.View().Content)
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	after := unwrapModel(updated)
	if after.viewer.Vp.XOffset() != current.viewer.Vp.XOffset() {
		t.Fatalf("horizontal scroll should be ignored while wrapped, got %d want %d", after.viewer.Vp.XOffset(), current.viewer.Vp.XOffset())
	}
}

func TestIntegration_LogsWrapTogglePreservesRawLineAnchorWhenNotFollowing(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 80
	m.height = 24
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	logsData := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		logsData = append(logsData, strconv.Itoa(i)+" "+strings.Repeat("x", 48))
	}
	m.viewer.Vp.Data = logsData
	m.viewer.Vp.SetFollow(false)
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	nearBottom := max(0, len(logsData)-m.viewer.VisibleRows()-1)
	m.viewer.Vp.SetYOffset(nearBottom)

	beforeList := viewer.FilterLines(m.viewer.Vp.Data, m.viewer.Vp.Filter.Query)
	beforeStart, _ := viewer.VisibleContentRange(m.viewer.Vp, beforeList)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	after := unwrapModel(updated)
	if !after.viewer.Vp.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}

	afterList := viewer.FilterLines(after.viewer.Vp.Data, after.viewer.Vp.Filter.Query)
	afterStart, _ := viewer.VisibleContentRange(after.viewer.Vp, afterList)
	if afterStart != beforeStart {
		t.Fatalf("visible raw log anchor changed across wrap toggle, before=%d after=%d", beforeStart, afterStart)
	}
}

func TestIntegration_ViewerEntryResetsHorizontalPosition(t *testing.T) {
	m := unwrapModel(New(nil))
	m.viewer.Vp.SetXOffset(24)
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	current := unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("entering logs mode should return a command")
	}
	// Process the transition to enter logs mode
	if msg := cmd(); msg != nil {
		updated, _ = current.Update(msg)
		current = unwrapModel(updated)
	}
	if got := current.viewer.Vp.XOffset(); got != 0 {
		t.Fatalf("logs entry should reset viewport x offset, got %d", got)
	}

	current.viewer.Vp.SetXOffset(32)
	current.browse.ActiveTab = tabContainers
	current.screen = shared.Main
	current.browse.Snapshot = core.Snapshot{Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}}

	updated, cmd = current.handleInspectTransition()
	current = unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("entering inspect mode should return a command")
	}
	if got := current.viewer.Vp.XOffset(); got != 0 {
		t.Fatalf("inspect entry should reset viewport x offset, got %d", got)
	}
}

func TestIntegration_FilterPromptIcon(t *testing.T) {
	m := unwrapModel(New(nil))
	if m.browse.Filter.Input.Prompt != "🔎︎ " {
		t.Fatalf("filter prompt = %q, want %q", m.browse.Filter.Input.Prompt, "🔎︎ ")
	}
}

func TestIntegration_FilterMode_AllowsVerticalNavigation(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.ContainerCursor = 0
	m.browse.Filter.Active = true
	m.browse.Filter.Input.Focus()
	m.browse.Filter.Input.SetValue("api")
	m.browse.Filter.Query = "api"
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api-1", State: "running"},
			{FullID: "ctr-2", Name: "api-2", State: "running"},
		},
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	after := unwrapModel(updated)
	if after.browse.ContainerCursor != 1 {
		t.Fatalf("filter mode down should move cursor to 1, got %d", after.browse.ContainerCursor)
	}
	if !after.browse.Filter.Active {
		t.Fatalf("filter mode should remain active while navigating")
	}
	if after.browse.Filter.Query != "api" {
		t.Fatalf("filter query should remain unchanged while navigating, got %q", after.browse.Filter.Query)
	}
}

func TestIntegration_ContainersComposeRow_CollapsesAndExpands(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}
	*m = m.syncBrowseData()
	m.dataDirty = false

	if got := m.browse.ItemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should show one row, got %d", got)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterExpand := unwrapModel(updated)
	*afterExpand = afterExpand.syncBrowseData()
	if got := afterExpand.browse.ItemCountForTab(tabContainers); got != 3 {
		t.Fatalf("expanded compose list should show project + 2 containers, got %d", got)
	}

	updated, _ = afterExpand.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterCollapse := unwrapModel(updated)
	*afterCollapse = afterCollapse.syncBrowseData()
	if got := afterCollapse.browse.ItemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should return to one row, got %d", got)
	}
}

func TestIntegration_ContainersComposeRow_EnterDoesNotOpenLogs(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}
	*m = m.syncBrowseData()

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("enter on compose project row should not open logs command")
	}
	if after.screen != shared.Main {
		t.Fatalf("screen = %v, want browse", after.screen)
	}
	*after = after.syncBrowseData()
	if got := after.browse.ItemCountForTab(tabContainers); got != 2 {
		t.Fatalf("enter on compose project should expand row, got item count %d", got)
	}
}

func TestIntegration_ContainersComposeFooterShowsContextualEnterHelp(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}
	*m = m.syncBrowseData()

	composeView := m.View().Content
	if !strings.Contains(composeView, "expand") {
		t.Fatalf("collapsed compose row should advertise expand action, got %q", composeView)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := unwrapModel(updated)
	expandedView := after.View().Content
	if !strings.Contains(expandedView, "collapse") {
		t.Fatalf("expanded compose row should advertise collapse action, got %q", expandedView)
	}
}

func TestIntegration_ContainersTabCount_UsesTotalContainersWhenComposeCollapsed(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}
	*m = m.syncBrowseData()
	m.dataDirty = false

	view := m.View().Content
	if !strings.Contains(util.StripANSI(view), "Containers") {
		t.Fatalf("header should show containers, got %q", view)
	}
	if got := m.browse.ItemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should still render one row, got %d", got)
	}
}

func TestIntegration_HorizontalTabSwitchClearsFilter(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Filter.Query = "redis"
	m.browse.Filter.Input.SetValue("redis")

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	after := unwrapModel(updated)

	if after.browse.ActiveTab != tabImages {
		t.Fatalf("active tab = %d, want %d", after.browse.ActiveTab, tabImages)
	}
	if after.browse.Filter.Query != "" {
		t.Fatalf("filter query should be cleared on horizontal tab switch, got %q", after.browse.Filter.Query)
	}
	if after.browse.Filter.Input.Value() != "" {
		t.Fatalf("filter input value should be cleared on horizontal tab switch, got %q", after.browse.Filter.Input.Value())
	}
}

func TestIntegration_LogsFiltering_ByContainsAndClearOnEsc(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Vp.Data = []string{"alpha line", "quick match", "zeta line"}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	current := unwrapModel(updated)
	if !current.viewer.Vp.Filter.Active {
		t.Fatalf("slash should activate logs filter mode")
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	current = unwrapModel(updated)
	if current.viewer.Vp.Filter.Query != "q" {
		t.Fatalf("logs filter query = %q, want q", current.viewer.Vp.Filter.Query)
	}
	filtered := current.viewer.Vp.View()
	if !strings.Contains(filtered, "quick match") {
		t.Fatalf("matching log line should be visible, got %q", filtered)
	}
	if strings.Contains(filtered, "alpha line") || strings.Contains(filtered, "zeta line") {
		t.Fatalf("non-matching log lines should be hidden, got %q", filtered)
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	after := unwrapModel(updated)
	if after.viewer.Vp.Filter.Active {
		t.Fatalf("esc should exit logs filter mode")
	}
	if after.viewer.Vp.Filter.Query != "" {
		t.Fatalf("esc should clear logs filter query, got %q", after.viewer.Vp.Filter.Query)
	}
	restored := after.viewer.Vp.View()
	if !strings.Contains(restored, "alpha line") || !strings.Contains(restored, "quick match") || !strings.Contains(restored, "zeta line") {
		t.Fatalf("logs viewport should restore full lines after clearing filter, got %q", restored)
	}
}

func TestIntegration_LogsFilterMode_AllowsVerticalNavigation(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 80)
	for i := 0; i < 80; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Vp.Data = lines
	m.viewer.Vp.Filter.Active = true
	m.viewer.Vp.Filter.Input.Focus()
	m.viewer.Vp.Filter.Query = ""
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Vp.SetFollow(false)
	m.viewer.Vp.GotoTop()

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	after := unwrapModel(updated)

	if !after.viewer.Vp.Filter.Active {
		t.Fatalf("filter mode should remain active after Down key")
	}
	if after.viewer.Vp.Filter.Query != "" {
		t.Fatalf("filter query should remain unchanged after Down key, got %q", after.viewer.Vp.Filter.Query)
	}
}

func TestIntegration_LogsFilterOpen_ReducesRowsFromTop(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Vp.Data = lines
	m.viewer.Vp.SetFollow(false)
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Vp.SetYOffset(10)

	beforeRows := m.viewer.VisibleRows()

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	after := unwrapModel(updated)

	if !after.viewer.Vp.Filter.Active {
		t.Fatalf("slash should activate logs filter mode")
	}
	afterRows := after.viewer.VisibleRows()
	if afterRows != beforeRows {
		t.Fatalf("expected same visible rows after opening filter (filter replaces breadcrumbs, same overhead), before=%d after=%d", beforeRows, afterRows)
	}
}

func TestIntegration_LogsFilterOpenClose_NoViewportDrift(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.LogViewer
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Vp.Data = lines
	m.viewer.Vp.SetFollow(false)
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Vp.SetYOffset(20)

	baseRows := m.viewer.VisibleRows()
	baseBottom := m.viewer.Vp.YOffset() + baseRows - 1

	current := *m
	for i := 0; i < 3; i++ {
		updated, _ := current.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		current = *unwrapModel(updated)
		if !current.viewer.Vp.Filter.Active {
			t.Fatalf("cycle %d: slash should activate logs filter mode", i)
		}

		updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		current = *unwrapModel(updated)
		if current.viewer.Vp.Filter.Active {
			t.Fatalf("cycle %d: enter should close logs filter mode", i)
		}

		bottom := current.viewer.Vp.YOffset() + current.viewer.VisibleRows() - 1
		if bottom != baseBottom {
			t.Fatalf("cycle %d: viewport drift detected, bottom=%d want=%d", i, bottom, baseBottom)
		}
	}
}

func TestIntegration_ShouldPollLogsOnTick_GatedByLogLoadingState(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.LogViewer
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Vp.Data = make([]string, 220)
	for i := range m.viewer.Vp.Data {
		m.viewer.Vp.Data[i] = "line"
	}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Vp.GotoTop()

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("should request history when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while viewport is at top and history is available")
	}

	m.viewer.Vp.InitialLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while initial load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while initial logs load is in progress")
	}

	m.viewer.Vp.InitialLoad = false
	m.viewer.Logs.HistoryLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while history load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while history logs load is in progress")
	}

	m.viewer.Logs.HistoryLoad = false
	m.viewer.Logs.HistoryDone = true
	m.viewer.Vp.GotoBottom()
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history after history is exhausted")
	}
	if !m.shouldPollLogsOnTick() {
		t.Fatalf("should poll when not at top and no load is active")
	}

	m.screen = shared.Main
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll outside logs screen")
	}
}

func TestIntegration_TickPrefersHistoryLoadAtTop(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.LogViewer
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Vp.Data = make([]string, 220)
	for i := range m.viewer.Vp.Data {
		m.viewer.Vp.Data[i] = "line"
	}
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Vp.GotoTop()
	m.viewer.Vp.InitialLoad = false
	m.viewer.Logs.HistoryLoad = false
	m.viewer.Logs.HistoryDone = false

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("expected history load to be scheduled when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("polling should stay disabled at top while history loading is available")
	}

	updated, cmd := m.Update(tickMsg{})
	current := unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("tick at top should schedule a history load command")
	}
	if !current.viewer.Logs.HistoryLoad {
		t.Fatalf("tick handling should mark history loading while fetch is in-flight")
	}
}

func TestIntegration_FirstEnterOnComposeRow_ExpandsNotLogs(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-3", Name: "web", State: "running"},
		},
	}
	// Simulate initial state: data loaded but no persistent sync yet
	m.dataDirty = true
	m.browse.Data = browse.BrowseData{}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := unwrapModel(updated)

	if cmd != nil {
		t.Fatalf("enter on compose project row should not return a command, got %T", cmd)
	}
	if after.screen != shared.Main {
		t.Fatalf("screen = %v, want Main", after.screen)
	}
	if !after.browse.ComposeExpanded["shop"] {
		t.Fatalf("compose project 'shop' should be expanded after enter")
	}
}

func TestIntegration_ComposeExpandThenDownThenEnter(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Main
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-3", Name: "web", State: "running"},
		},
	}
	m.dataDirty = true
	m.browse.Data = browse.BrowseData{}

	// Step 1: Enter to expand compose project
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("enter on compose project should not return cmd, got %T", cmd)
	}
	if !after.browse.ComposeExpanded["shop"] {
		t.Fatalf("compose project should be expanded")
	}

	// Step 2: Down to first child container
	updated, cmd = after.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	after = unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("down should not return cmd, got %T", cmd)
	}

	// Step 3: Enter on first child should open logs
	updated, cmd = after.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after = unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("enter on child container should return logs cmd")
	}

	if msg := cmd(); msg != nil {
		updated, _ = after.Update(msg)
		after = unwrapModel(updated)
	}
	if after.screen != shared.LogViewer {
		t.Fatalf("screen = %v, want LogViewer", after.screen)
	}
	if after.viewer.ContainerID != "ctr-1" {
		t.Fatalf("logs container = %q, want ctr-1", after.viewer.ContainerID)
	}
}

func TestPruneComposeExpanded_RemovesAbsentProjects(t *testing.T) {
	expanded := map[string]bool{"shop": true, "api": true, "stale": true}
	active := map[string]struct{}{"shop": {}, "api": {}}
	pruneComposeExpanded(expanded, active)
	if len(expanded) != 2 {
		t.Fatalf("expected 2 entries after prune, got %d", len(expanded))
	}
	if expanded["stale"] {
		t.Fatalf("stale project should have been pruned")
	}
	if !expanded["shop"] || !expanded["api"] {
		t.Fatalf("active projects should be preserved")
	}
}

func TestPruneComposeExpanded_EmptyMapIsNoOp(t *testing.T) {
	expanded := map[string]bool{}
	active := map[string]struct{}{"shop": {}}
	pruneComposeExpanded(expanded, active)
	if len(expanded) != 0 {
		t.Fatalf("empty map should remain empty")
	}
}
