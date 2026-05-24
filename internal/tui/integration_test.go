package tui

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/spinner"
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
	m := model{
		width:  120,
		height: 30,
		styles: defaultStyles(),
		browse: browse.Model{
			ActiveTab: tabContainers,
			ShowAll:   true,
			Filter:    browse.NewFilterState(),
			Snapshot: core.Snapshot{
				Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
			},
		},
		viewer: viewer.Model{
			State:   viewer.NewState(),
			Spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		},
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}

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
	if current.screen != shared.Logs {
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
	if current.screen != shared.Browse {
		t.Fatalf("screen = %v, want browse", current.screen)
	}
	if current.browse.ActiveTab != tabContainers {
		t.Fatalf("ActiveTab = %d, want %d", current.browse.ActiveTab, tabContainers)
	}
}

func TestIntegration_ViewRendersBrowseAndLogsModes(t *testing.T) {
	m := model{
		width:  100,
		height: 28,
		screen: shared.Browse,
		styles: defaultStyles(),
		browse: browse.Model{
			ActiveTab: tabContainers,
			ShowAll:   true,
			Filter:    browse.NewFilterState(),
			Snapshot: core.Snapshot{
				Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", Image: "nginx", Status: "Up"}},
			},
		},
		viewer: viewer.Model{
			State:   viewer.NewState(),
			Spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		},
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}

	browseView := m.View().Content
	if !strings.Contains(browseView, "EasyDocker") {
		t.Fatalf("browse view missing header")
	}

	m.screen = shared.Logs
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Data = []string{"line-1", "line-2"}
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	logsView := m.View().Content
	if !strings.Contains(logsView, "Logs") || !strings.Contains(logsView, "api") {
		t.Fatalf("logs view missing logs breadcrumb context")
	}
}

func TestIntegration_UpdateResultFlow(t *testing.T) {
	m := model{
		loading:      true,
		loadingStage: loadStageContainers,
		styles:       defaultStyles(),
		browse: browse.Model{
			Filter: browse.NewFilterState(),
		},
		viewer: viewer.Model{
			State:   viewer.NewState(),
			Spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		},
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}

	updated, cmd := m.Update(containersResultMsg{containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}})
	current := unwrapModel(updated)
	if cmd == nil || current.loadingStage != loadStageResources {
		t.Fatalf("expected transition to resources stage")
	}

	updated, cmd = current.Update(resourcesResultMsg{snapshot: core.Snapshot{Images: []core.ImageRow{{ID: "img-1"}}, Networks: []core.NetworkRow{{ID: "net-1"}}, Volumes: []core.VolumeRow{{Name: "vol-1"}}, TotalLimit: 1000}})
	current = unwrapModel(updated)
	if cmd == nil || current.loadingStage != loadStageMetrics {
		t.Fatalf("expected transition to metrics stage")
	}

	updated, cmd = current.Update(metricsResultMsg{metricsByID: map[string]core.ContainerMetrics{"ctr-1": {CPUPercent: 12.5, MemoryUsage: "10 MiB", MemoryLimit: "100 MiB", MemoryUsageBytes: 10, MemoryLimitBytes: 100, MemoryPercent: 10}}, totalCPU: 12.5, totalMem: 10})
	current = unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("metrics stage should not schedule command")
	}
	if current.loading || current.loadingStage != loadStageIdle {
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
	m := model{
		width:        120,
		height:       30,
		loading:      false,
		loadingStage: loadStageIdle,
		styles:       defaultStyles(),
		browse: browse.Model{
			ActiveTab: tabContainers,
			ShowAll:   true,
			Filter:    browse.NewFilterState(),
			Snapshot: core.Snapshot{
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
		},
		viewer: viewer.Model{
			State:   viewer.NewState(),
			Spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		},
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}

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
	m := model{
		width:        120,
		height:       30,
		loading:      true,
		loadingStage: loadStageMetrics,
		styles:       defaultStyles(),
		browse: browse.Model{
			ActiveTab: tabContainers,
			ShowAll:   true,
			Filter:    browse.NewFilterState(),
			Snapshot: core.Snapshot{
				Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"}},
			},
		},
		viewer: viewer.Model{
			State:   viewer.NewState(),
			Spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		},
		metricsSpinner:   spinner.New(spinner.WithSpinner(spinner.Dot)),
		containerSpinner: spinner.New(spinner.WithSpinner(spinner.Line)),
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
	m := unwrapModel(New(nil))
	m.width = 100
	m.height = 30
	m.screen = shared.Browse

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
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Data = []string{"abcdefghijklmnopqrstuvwxyz"}
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	current := unwrapModel(updated)
	if !current.viewer.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}

	wrappedView := current.viewer.Viewport.View()
	if !strings.Contains(wrappedView, "\n") {
		t.Fatalf("wrapped viewport should render on multiple lines, got %q", wrappedView)
	}

	if !strings.Contains(current.View().Content, "lines:") {
		t.Fatalf("wrapped line rows should not inflate total log count, view=%q", current.View().Content)
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	after := unwrapModel(updated)
	if after.viewer.Viewport.XOffset() != current.viewer.Viewport.XOffset() {
		t.Fatalf("horizontal scroll should be ignored while wrapped, got %d want %d", after.viewer.Viewport.XOffset(), current.viewer.Viewport.XOffset())
	}
}

func TestIntegration_LogsWrapTogglePreservesRawLineAnchorWhenNotFollowing(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 80
	m.height = 24
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	logsData := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		logsData = append(logsData, strconv.Itoa(i)+" "+strings.Repeat("x", 48))
	}
	m.viewer.Data = logsData
	m.viewer.SetFollow(false)
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	nearBottom := max(0, len(logsData)-m.viewer.VisibleRows()-1)
	m.viewer.Viewport.SetYOffset(nearBottom)

	beforeList := viewer.FilterLines(m.viewer.Data, m.viewer.Filter.Query)
	beforeStart, _ := viewer.VisibleContentRange(&m.viewer.State, beforeList)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	after := unwrapModel(updated)
	if !after.viewer.WrapLines {
		t.Fatalf("wrap should be enabled after pressing w")
	}

	afterList := viewer.FilterLines(after.viewer.Data, after.viewer.Filter.Query)
	afterStart, _ := viewer.VisibleContentRange(&after.viewer.State, afterList)
	if afterStart != beforeStart {
		t.Fatalf("visible raw log anchor changed across wrap toggle, before=%d after=%d", beforeStart, afterStart)
	}
}

func TestIntegration_ViewerEntryResetsHorizontalPosition(t *testing.T) {
	m := unwrapModel(New(nil))
	m.viewer.Viewport.SetXOffset(24)
	m.screen = shared.Browse
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
	if got := current.viewer.Viewport.XOffset(); got != 0 {
		t.Fatalf("logs entry should reset viewport x offset, got %d", got)
	}

	current.viewer.Viewport.SetXOffset(32)
	current.browse.ActiveTab = tabContainers
	current.screen = shared.Browse
	current.browse.Snapshot = core.Snapshot{Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}}}

	updated, cmd = current.handleInspectTransition()
	current = unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("entering inspect mode should return a command")
	}
	if got := current.viewer.Viewport.XOffset(); got != 0 {
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
	m.screen = shared.Browse
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
	m.screen = shared.Browse
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}
	*m = m.syncBrowseData()

	if got := m.browse.ItemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should show one row, got %d", got)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterExpand := unwrapModel(updated)
	if got := afterExpand.browse.ItemCountForTab(tabContainers); got != 3 {
		t.Fatalf("expanded compose list should show project + 2 containers, got %d", got)
	}

	updated, _ = afterExpand.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	afterCollapse := unwrapModel(updated)
	if got := afterCollapse.browse.ItemCountForTab(tabContainers); got != 1 {
		t.Fatalf("collapsed compose list should return to one row, got %d", got)
	}
}

func TestIntegration_ContainersComposeRow_EnterDoesNotOpenLogs(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Browse
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	after := unwrapModel(updated)
	if cmd != nil {
		t.Fatalf("enter on compose project row should not open logs command")
	}
	if after.screen != shared.Browse {
		t.Fatalf("screen = %v, want browse", after.screen)
	}
	if got := after.browse.ItemCountForTab(tabContainers); got != 2 {
		t.Fatalf("enter on compose project should expand row, got item count %d", got)
	}
}

func TestIntegration_ContainersComposeFooterShowsContextualEnterHelp(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.screen = shared.Browse
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"}},
	}

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
	m.screen = shared.Browse
	m.browse.ActiveTab = tabContainers
	m.browse.ShowAll = true
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{
			{FullID: "ctr-1", Name: "api", ComposeProject: "shop", State: "running"},
			{FullID: "ctr-2", Name: "worker", ComposeProject: "shop", State: "running"},
		},
	}
	*m = m.syncBrowseData()

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
	m.screen = shared.Browse
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
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Data = []string{"alpha line", "quick match", "zeta line"}
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	current := unwrapModel(updated)
	if !current.viewer.Filter.Active {
		t.Fatalf("slash should activate logs filter mode")
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	current = unwrapModel(updated)
	if current.viewer.Filter.Query != "q" {
		t.Fatalf("logs filter query = %q, want q", current.viewer.Filter.Query)
	}
	filtered := current.viewer.Viewport.View()
	if !strings.Contains(filtered, "quick match") {
		t.Fatalf("matching log line should be visible, got %q", filtered)
	}
	if strings.Contains(filtered, "alpha line") || strings.Contains(filtered, "zeta line") {
		t.Fatalf("non-matching log lines should be hidden, got %q", filtered)
	}

	updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	after := unwrapModel(updated)
	if after.viewer.Filter.Active {
		t.Fatalf("esc should exit logs filter mode")
	}
	if after.viewer.Filter.Query != "" {
		t.Fatalf("esc should clear logs filter query, got %q", after.viewer.Filter.Query)
	}
	restored := after.viewer.Viewport.View()
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
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 80)
	for i := 0; i < 80; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Data = lines
	m.viewer.Filter.Active = true
	m.viewer.Filter.Input.Focus()
	m.viewer.Filter.Query = ""
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.SetFollow(false)
	m.viewer.Viewport.GotoTop()

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	after := unwrapModel(updated)

	if !after.viewer.Filter.Active {
		t.Fatalf("filter mode should remain active after Down key")
	}
	if after.viewer.Filter.Query != "" {
		t.Fatalf("filter query should remain unchanged after Down key, got %q", after.viewer.Filter.Query)
	}
}

func TestIntegration_LogsFilterOpen_ReducesRowsFromTop(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Data = lines
	m.viewer.SetFollow(false)
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Viewport.SetYOffset(10)

	beforeRows := m.viewer.VisibleRows()

	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	after := unwrapModel(updated)

	if !after.viewer.Filter.Active {
		t.Fatalf("slash should activate logs filter mode")
	}
	afterRows := after.viewer.VisibleRows()
	if afterRows >= beforeRows {
		t.Fatalf("expected fewer visible rows after opening filter, before=%d after=%d", beforeRows, afterRows)
	}
}

func TestIntegration_LogsFilterOpenClose_NoViewportDrift(t *testing.T) {
	m := unwrapModel(New(nil))
	m.width = 120
	m.height = 34
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.screen = shared.Logs
	m.browse.ActiveTab = tabContainers
	m.browse.Snapshot = core.Snapshot{
		Containers: []core.ContainerRow{{FullID: "ctr-1", Name: "api", State: "running"}},
	}
	m.viewer.ContainerID = "ctr-1"

	lines := make([]string, 0, 300)
	for i := 0; i < 300; i++ {
		lines = append(lines, "line-"+strconv.Itoa(i))
	}
	m.viewer.Data = lines
	m.viewer.SetFollow(false)
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Viewport.SetYOffset(20)

	baseRows := m.viewer.VisibleRows()
	baseBottom := m.viewer.Viewport.YOffset() + baseRows - 1

	current := *m
	for i := 0; i < 3; i++ {
		updated, _ := current.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		current = *unwrapModel(updated)
		if !current.viewer.Filter.Active {
			t.Fatalf("cycle %d: slash should activate logs filter mode", i)
		}

		updated, _ = current.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		current = *unwrapModel(updated)
		if current.viewer.Filter.Active {
			t.Fatalf("cycle %d: enter should close logs filter mode", i)
		}

		bottom := current.viewer.Viewport.YOffset() + current.viewer.VisibleRows() - 1
		if bottom != baseBottom {
			t.Fatalf("cycle %d: viewport drift detected, bottom=%d want=%d", i, bottom, baseBottom)
		}
	}
}

func TestIntegration_ShouldPollLogsOnTick_GatedByLogLoadingState(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Logs
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Data = make([]string, 220)
	for i := range m.viewer.Data {
		m.viewer.Data[i] = "line"
	}
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Viewport.GotoTop()

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("should request history when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while viewport is at top and history is available")
	}

	m.viewer.InitialLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while initial load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while initial logs load is in progress")
	}

	m.viewer.InitialLoad = false
	m.viewer.HistoryLoad = true
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history while history load is in progress")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll while history logs load is in progress")
	}

	m.viewer.HistoryLoad = false
	m.viewer.HistoryDone = true
	m.viewer.Viewport.GotoBottom()
	if m.shouldLoadHistoryOnTick() {
		t.Fatalf("should not load history after history is exhausted")
	}
	if !m.shouldPollLogsOnTick() {
		t.Fatalf("should poll when not at top and no load is active")
	}

	m.screen = shared.Browse
	if m.shouldPollLogsOnTick() {
		t.Fatalf("should not poll outside logs screen")
	}
}

func TestIntegration_TickPrefersHistoryLoadAtTop(t *testing.T) {
	m := unwrapModel(New(nil))
	m.screen = shared.Logs
	m.viewer.ContainerID = "ctr-1"
	m.viewer.Data = make([]string, 220)
	for i := range m.viewer.Data {
		m.viewer.Data[i] = "line"
	}
	m.viewer.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	m.viewer.Viewport.GotoTop()
	m.viewer.InitialLoad = false
	m.viewer.HistoryLoad = false
	m.viewer.HistoryDone = false

	if !m.shouldLoadHistoryOnTick() {
		t.Fatalf("expected history load to be scheduled when viewport is at top")
	}
	if m.shouldPollLogsOnTick() {
		t.Fatalf("polling should stay disabled at top while history loading is available")
	}

	updated, cmd := m.Update(tickMsg(time.Now()))
	current := unwrapModel(updated)
	if cmd == nil {
		t.Fatalf("tick at top should schedule a history load command")
	}
	if current.viewer.HistoryLoad {
		t.Fatalf("tick handling should not mark history loading without result handling")
	}
}
