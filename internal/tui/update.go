package tui

import (
	"strings"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

var browseController = browse.Controller{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case containersResultMsg:
		return m.handleContainersResultMsg(msg)
	case resourcesResultMsg:
		return m.handleResourcesResultMsg(msg)
	case metricsResultMsg:
		return m.handleMetricsResultMsg(msg)
	case loadResultMsg:
		return m.handleLoadResultMsg(msg)
	case viewer.ContentMsg:
		return m.handleLogsResultMsg(msg)
	case inspectResultMsg:
		return m.handleInspectResultMsg(msg)
	case shellDoneMsg:
		return m, nil
	case tickMsg:
		return m.handleTickMsg(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTickMsg(msg)
	}

	return m, nil
}

const browseCursorPageStep = 5

func (m model) handleBrowseKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := browseKeyMap()

	// If filter mode is active, handle filter input first
	if m.browseFilter.Active {
		switch {
		case key.Matches(msg, keys.Quit):
			// Esc exits filter mode and clears query
			m.browseFilter.Active = false
			m.browseFilter.Input.Blur()
			m.browseFilter.Query = ""
			m.browseFilter.Input.SetValue("")
			m.clampCursors()
			return m, nil
		case msg.String() == "enter":
			// Enter exits filter mode but keeps query
			m.browseFilter.Active = false
			m.browseFilter.Input.Blur()
			return m, nil
		case key.Matches(msg, keys.MoveUp):
			m.moveCursor(-1)
			return m, nil
		case key.Matches(msg, keys.MoveDown):
			m.moveCursor(1)
			return m, nil
		case key.Matches(msg, keys.PageUp):
			m.moveCursor(-browseCursorPageStep)
			return m, nil
		case key.Matches(msg, keys.PageDown):
			m.moveCursor(browseCursorPageStep)
			return m, nil
		default:
			// All other keys go to filter input
			var cmd tea.Cmd
			m.browseFilter.Input, cmd = m.browseFilter.Input.Update(msg)
			m.browseFilter.Query = m.browseFilter.Input.Value()
			// Recompute visible lists and clamp cursors to keep selection valid
			m.clampCursors()
			return m, cmd
		}
	}

	// Use controller for normal browse mode key handling
	browseState := browse.State{Filter: m.browseFilter}
	transition := browseController.HandleKey(&browseState, msg, browse.NewKeyMap())
	m.browseFilter = browseState.Filter

	if transition.ChangeTab != 0 {
		m.moveActiveTab(transition.ChangeTab)
	}
	if transition.ActivateFilter {
		m.browseFilter.Active = true
		m.browseFilter.Input.Focus()
		m.browseFilter.Input.SetValue(m.browseFilter.Query)
	}
	if transition.ToggleScope {
		m.toggleContainerScope()
	}
	if transition.OpenResource {
		if m.toggleSelectedComposeProject() {
			return m, nil
		}
		if cmd := m.enterLogsModeIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if transition.OpenShell {
		if cmd := m.openShellIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if transition.OpenInspect {
		return m.handleInspectTransition()
	}
	if transition.Quit {
		return m, tea.Quit
	}
	if transition.CursorMove != 0 {
		m.moveCursor(transition.CursorMove)
	}

	return m, nil
}

func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	if m.screen == screenModeLogs {
		m.logs.SyncFromData(m.logVisibleWidth(), m.logVisibleRows())
	}
	if m.screen == screenModeInspect {
		m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.screen == screenModeBrowse && m.browseFilter.Active {
		return m.handleBrowseKey(msg)
	}

	if m.screen == screenModeLogs && m.logs.Filter.Active {
		return m, m.handleLogsKey(msg)
	}

	route := shared.RouteRootKey(msg.String(), toModeScreen(m.screen))
	switch route {
	case shared.RouteQuit:
		return m, tea.Quit
	case shared.RouteNoop:
		return m, nil
	case shared.RouteLogs:
		return m, m.handleLogsKey(msg)
	case shared.RouteInspect:
		return m, m.handleInspectKey(msg)
	case shared.RouteBrowse:
		return m.handleBrowseKey(msg)
	}

	return m, nil
}

func toModeScreen(screen screenMode) shared.Screen {
	switch screen {
	case screenModeLogs:
		return shared.Logs
	case screenModeInspect:
		return shared.Inspect
	}
	return shared.Browse
}

func fromModeScreen(screen shared.Screen) screenMode {
	switch screen {
	case shared.Logs:
		return screenModeLogs
	case shared.Inspect:
		return screenModeInspect
	}
	return screenModeBrowse
}

func (m *model) handleLogsKey(msg tea.KeyPressMsg) tea.Cmd {
	keys := viewer.NewKeyMap()

	if m.logs.Filter.Active {
		switch {
		case key.Matches(msg, keys.Back):
			previousRows := m.logVisibleRows()
			previousYOffset := m.logs.Viewport.YOffset()
			m.logs.Filter.Active = false
			m.logs.Filter.Input.Blur()
			m.logs.Filter.Query = ""
			m.logs.Filter.Input.SetValue("")
			newRows := m.logVisibleRows()
			m.logs.SyncFromData(m.logVisibleWidth(), newRows)
			if !m.logs.Follow && newRows > previousRows {
				m.logs.Viewport.SetYOffset(max(0, previousYOffset-(newRows-previousRows)))
			}
			return nil
		case msg.String() == "enter":
			previousRows := m.logVisibleRows()
			previousYOffset := m.logs.Viewport.YOffset()
			m.logs.Filter.Active = false
			m.logs.Filter.Input.Blur()
			newRows := m.logVisibleRows()
			m.logs.SyncFromData(m.logVisibleWidth(), newRows)
			if !m.logs.Follow && newRows > previousRows {
				m.logs.Viewport.SetYOffset(max(0, previousYOffset-(newRows-previousRows)))
			}
			return nil
		case key.Matches(msg, keys.Up),
			key.Matches(msg, keys.Down),
			key.Matches(msg, keys.PageUp),
			key.Matches(msg, keys.PageDown),
			key.Matches(msg, keys.Home),
			key.Matches(msg, keys.End):
			transition := viewer.Controller{}.HandleKey(&m.logs, msg, keys)
			if historyReq := HistoryLoadRequest(&m.logs); historyReq != nil {
				transition.Load = historyReq
			}
			return m.applyLogsTransition(viewer.Transition{Load: transition.Load})
		default:
			var cmd tea.Cmd
			m.logs.Filter.Input, cmd = m.logs.Filter.Input.Update(msg)
			m.logs.Filter.Query = m.logs.Filter.Input.Value()
			m.logs.SyncFromData(m.logVisibleWidth(), m.logVisibleRows())
			return cmd
		}
	}

	if key.Matches(msg, keys.OpenFilter) {
		previousRows := m.logVisibleRows()
		previousYOffset := m.logs.Viewport.YOffset()
		m.logs.Filter.Active = true
		m.logs.Filter.Input.Focus()
		m.logs.Filter.Input.SetValue(m.logs.Filter.Query)
		newRows := m.logVisibleRows()
		m.logs.SyncFromData(m.logVisibleWidth(), newRows)
		if !m.logs.Follow && newRows < previousRows {
			m.logs.Viewport.SetYOffset(previousYOffset + (previousRows - newRows))
		}
		return nil
	}

	if key.Matches(msg, keys.Back) {
		m.exitLogsMode()
		return nil
	}

	if key.Matches(msg, keys.ToggleWrap) {
		logList := viewer.FilterLines(m.logs.Data, m.logs.Filter.Query)
		startLine, _ := viewer.VisibleContentRange(&m.logs, logList)
		visibleWidth := m.logVisibleWidth()
		visibleRows := m.logVisibleRows()
		m.logs.SetWrapLines(!m.logs.WrapLines)
		m.logs.SyncFromData(visibleWidth, visibleRows)
		if !m.logs.Follow {
			targetYOffset := startLine
			if m.logs.WrapLines {
				targetYOffset = viewer.RawLineToViewportRowOffset(logList, visibleWidth, startLine)
			}
			m.logs.Viewport.SetYOffset(targetYOffset)
		}
		return nil
	}

	transition := viewer.Controller{}.HandleKey(&m.logs, msg, viewer.NewKeyMap())
	if historyReq := HistoryLoadRequest(&m.logs); historyReq != nil {
		transition.Load = historyReq
	}
	return m.applyLogsTransition(transition)
}

func (m *model) enterLogsMode(container core.ContainerRow) tea.Cmd {
	transition := EnterLogsState(&m.logs, container.FullID)
	m.err = nil
	m.screen = fromModeScreen(shared.EnterLogsTransition())
	return m.applyLogsTransition(transition)
}

func (m *model) exitLogsMode() {
	transition := ExitLogsState(&m.logs, tabContainers)
	_ = m.applyLogsTransition(transition)
}

func (m *model) handleLogsResult(msg viewer.ContentMsg) tea.Cmd {
	transition := HandleLogsResult(&m.logs, msg, m.logVisibleWidth(), m.logVisibleRows())
	return m.applyLogsTransition(transition)
}

func (m *model) applyLogsTransition(transition viewer.Transition) tea.Cmd {
	if transition.LaunchShell {
		if container, ok := m.selectedLogsContainer(); ok {
			return m.shellCmd(container.FullID)
		}
		return nil
	}
	if transition.ExitToBrowse {
		targetScreen, _ := shared.ExitLogsTransition(transition.ForceTab)
		m.screen = fromModeScreen(targetScreen)
		m.activeTab = transition.ForceTab
	}
	if transition.Err != nil {
		m.err = transition.Err
	}
	if transition.Load == nil {
		return nil
	}
	request := transition.Load
	loadCmd := LoadLogsCmd(
		m.service,
		request.ContainerID,
		request.SessionID,
		request.Tail,
		request.Src,
	)
	if m.shouldAnimateLogsLoadingIndicator() {
		return tea.Batch(loadCmd, m.logsSpinner.Tick)
	}
	return loadCmd
}

func (m *model) applyLoadingTransition(transition shared.Transition) {
	m.loading = transition.Loading
	m.loadingStage = int(transition.Stage)
	m.err = transition.Err
}

func (m *model) setLoadError(err error) {
	m.applyLoadingTransition(shared.Fail(err))
}

func (m *model) beginLoadingStage(stage int) {
	m.applyLoadingTransition(shared.Begin(shared.Stage(stage)))
	m.snapshot.Timestamp = time.Now()
	m.clampCursors()
}

func (m *model) finishLoadingStage(err error) bool {
	transition, ok := shared.Finish(err)
	m.applyLoadingTransition(transition)
	return ok
}

func (m model) shouldReloadSnapshotOnTick() bool {
	return m.loadingStage == loadStageIdle
}

func (m model) shouldLoadHistoryOnTick() bool {
	return m.screen == screenModeLogs &&
		m.logs.ContainerID != "" &&
		m.logs.Viewport.AtTop() &&
		!m.logs.InitialLoad &&
		!m.logs.HistoryLoad &&
		!m.logs.HistoryDone
}

func (m model) shouldPollLogsOnTick() bool {
	return m.screen == screenModeLogs &&
		m.logs.ContainerID != "" &&
		!m.logs.Viewport.AtTop() &&
		!m.logs.InitialLoad &&
		!m.logs.HistoryLoad
}

func (m model) logsPollTail() int {
	if m.logs.TailLines <= 0 {
		return InitialTail
	}
	return m.logs.TailLines
}

type handlerResult struct {
	cmd tea.Cmd
}

func noSideEffect() handlerResult {
	return handlerResult{}
}

func withSideEffect(cmd tea.Cmd) handlerResult {
	return handlerResult{cmd: cmd}
}

func (m model) respond(result handlerResult) (tea.Model, tea.Cmd) {
	return m, result.cmd
}

func (m model) handleContainersResultMsg(msg containersResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m.respond(noSideEffect())
	}

	m.snapshot.Containers = preserveRunningContainerMetrics(msg.containers, m.snapshot.Containers)
	m.beginLoadingStage(loadStageResources)
	return m.respond(withSideEffect(m.loadResourcesCmd()))
}

func (m model) handleResourcesResultMsg(msg resourcesResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m.respond(noSideEffect())
	}

	m.snapshot.Images = msg.snapshot.Images
	m.snapshot.Networks = msg.snapshot.Networks
	m.snapshot.Volumes = msg.snapshot.Volumes
	m.snapshot.TotalLimit = msg.snapshot.TotalLimit
	m.beginLoadingStage(loadStageMetrics)
	return m.respond(withSideEffect(m.loadMetricsCmd(m.snapshot.Containers)))
}

func (m model) handleMetricsResultMsg(msg metricsResultMsg) (tea.Model, tea.Cmd) {
	if !m.finishLoadingStage(msg.err) {
		return m.respond(noSideEffect())
	}

	m.snapshot.Containers = core.ApplyMetricsToContainers(m.snapshot.Containers, msg.metricsByID)
	m.snapshot.TotalCPU = msg.totalCPU
	m.snapshot.TotalMem = msg.totalMem
	m.snapshot.Timestamp = time.Now()
	m.metricsLoaded = true
	m.clampCursors()
	return m.respond(noSideEffect())
}

func (m model) handleLoadResultMsg(msg loadResultMsg) (tea.Model, tea.Cmd) {
	if !m.finishLoadingStage(msg.err) {
		return m.respond(noSideEffect())
	}

	previousContainers := m.snapshot.Containers
	m.snapshot = msg.snapshot
	m.snapshot.Containers = preserveRunningContainerMetrics(m.snapshot.Containers, previousContainers)
	if err := m.reconcileLogsSelection(); err != nil {
		m.err = err
	}
	m.clampCursors()
	return m.respond(noSideEffect())
}

func (m model) handleLogsResultMsg(msg viewer.ContentMsg) (tea.Model, tea.Cmd) {
	return m.respond(withSideEffect(m.handleLogsResult(msg)))
}

func (m model) handleTickMsg(_ tickMsg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{tickCmd()}
	if m.shouldReloadSnapshotOnTick() {
		cmds = append(cmds, m.loadDockerCmd())
	}
	if m.shouldLoadHistoryOnTick() {
		tail := len(m.logs.Data) + TailStep
		cmds = append(cmds, LoadLogsCmd(m.service, m.logs.ContainerID, m.logs.SessionID, tail, viewer.SourceHistory))
	} else if m.shouldPollLogsOnTick() {
		tail := m.logsPollTail()
		cmds = append(cmds, LoadLogsCmd(m.service, m.logs.ContainerID, m.logs.SessionID, tail, viewer.SourcePoll))
	}
	return m.respond(withSideEffect(tea.Batch(cmds...)))
}

func (m model) handleSpinnerTickMsg(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 3)

	if m.shouldAnimateMetricsLoadingIndicator() {
		var cmd tea.Cmd
		m.metricsSpinner, cmd = m.metricsSpinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		m.containerSpinner, cmd = m.containerSpinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if m.shouldAnimateLogsLoadingIndicator() {
		var cmd tea.Cmd
		m.logsSpinner, cmd = m.logsSpinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) == 0 {
		return m.respond(noSideEffect())
	}

	return m.respond(withSideEffect(tea.Batch(cmds...)))
}

func (m model) shouldAnimateLogsLoadingIndicator() bool {
	return m.screen == screenModeLogs && (m.logs.InitialLoad || m.logs.HistoryLoad)
}

func (m *model) handleInspectTransition() (tea.Model, tea.Cmd) {
	resourceType, resourceID, resourceName, ok := m.selectedInspectResource()
	if !ok {
		return m, nil
	}
	m.previousScreen = m.screen
	m.screen = screenModeInspect
	m.logs.InitialLoad = true
	m.logs.Data = nil
	m.logs.ResourceType = viewer.ResourceType(resourceType)
	m.logs.ContainerID = resourceID
	m.logs.ResourceName = resourceName
	return m, m.loadInspectCmd(resourceType, resourceID, resourceName)
}

func (m *model) loadInspectCmd(resourceType int, resourceID, resourceName string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		var data []string
		var err error
		switch resourceType {
		case tabContainers:
			data, err = svc.InspectContainer(resourceID)
		case tabImages:
			data, err = svc.InspectImage(resourceID)
		case tabNetworks:
			data, err = svc.InspectNetwork(resourceID)
		case tabVolumes:
			data, err = svc.InspectVolume(resourceID)
		}
		return inspectResultMsg{
			resourceType: resourceType,
			resourceID:   resourceID,
			resourceName: resourceName,
			data:         data,
			err:          err,
		}
	}
}

func (m model) handleInspectResultMsg(msg inspectResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.logs.InitialLoad = false
		return m, nil
	}
	m.logs.ApplyContentInitial(msg.data)
	m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
	return m, nil
}

func (m *model) handleInspectKey(msg tea.KeyPressMsg) tea.Cmd {
	keys := viewer.NewKeyMap()

	if m.logs.Filter.Active {
		switch {
		case key.Matches(msg, keys.Back):
			previousYOffset := m.logs.Viewport.YOffset()
			m.logs.Filter.Active = false
			m.logs.Filter.Input.Blur()
			m.logs.Filter.Query = ""
			m.logs.Filter.Input.SetValue("")
			newRows := m.inspectVisibleRows()
			m.logs.SyncFromData(m.inspectVisibleWidth(), newRows)
			m.logs.Viewport.SetYOffset(max(0, previousYOffset))
			return nil
		case msg.String() == "enter":
			m.logs.Filter.Active = false
			m.logs.Filter.Input.Blur()
			return nil
		default:
			var cmd tea.Cmd
			m.logs.Filter.Input, cmd = m.logs.Filter.Input.Update(msg)
			m.logs.Filter.Query = m.logs.Filter.Input.Value()
			m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
			return cmd
		}
	}

	if key.Matches(msg, keys.OpenFilter) {
		m.logs.Filter.Active = true
		m.logs.Filter.Input.Focus()
		m.logs.Filter.Input.SetValue(m.logs.Filter.Query)
		return nil
	}

	if key.Matches(msg, keys.Back) {
		m.exitInspectMode()
		return nil
	}

	if key.Matches(msg, keys.ToggleWrap) {
		logList := viewer.FilterLines(m.logs.Data, m.logs.Filter.Query)
		startLine, _ := viewer.VisibleContentRange(&m.logs, logList)
		visibleWidth := m.inspectVisibleWidth()
		visibleRows := m.inspectVisibleRows()
		m.logs.SetWrapLines(!m.logs.WrapLines)
		m.logs.SyncFromData(visibleWidth, visibleRows)
		if !m.logs.WrapLines {
			targetYOffset := viewer.RawLineToViewportRowOffset(logList, visibleWidth, startLine)
			m.logs.Viewport.SetYOffset(targetYOffset)
		}
		return nil
	}

	viewer.Controller{}.HandleKey(&m.logs, msg, viewer.NewKeyMap())
	m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
	return nil
}

func (m *model) exitInspectMode() {
	m.screen = m.previousScreen
}

func (m model) selectedInspectResource() (int, string, string, bool) {
	switch m.activeTab {
	case tabContainers:
		c, ok := m.selectedContainer()
		if !ok {
			return 0, "", "", false
		}
		return tabContainers, c.FullID, c.Name, true
	case tabImages:
		img, ok := m.selectedImage()
		if !ok {
			return 0, "", "", false
		}
		return tabImages, img.ID, img.Tags, true
	case tabNetworks:
		net, ok := m.selectedNetwork()
		if !ok {
			return 0, "", "", false
		}
		return tabNetworks, net.ID, net.Name, true
	case tabVolumes:
		vol, ok := m.selectedVolume()
		if !ok {
			return 0, "", "", false
		}
		return tabVolumes, vol.Name, vol.Name, true
	default:
		return 0, "", "", false
	}
}

func (m model) inspectVisibleWidth() int {
	totalWidth := max(1, m.width)
	return m.logsPageContentWidth(totalWidth)
}

func (m model) inspectVisibleRows() int {
	mainHeight := util.MainAreaHeight(m.height, m.renderHeader(), m.renderFooter())
	contentHeight := m.logsPageContentHeight(mainHeight)
	return viewer.VisibleRowsForContent(contentHeight, m.logs.Filter.Active)
}

func (m model) shouldAnimateMetricsLoadingIndicator() bool {
	return !m.metricsLoaded && m.loadingStage != loadStageIdle
}

func preserveRunningContainerMetrics(currentRows, previousRows []core.ContainerRow) []core.ContainerRow {
	if len(currentRows) == 0 || len(previousRows) == 0 {
		return currentRows
	}

	previousByID := make(map[string]core.ContainerRow, len(previousRows))
	for _, row := range previousRows {
		previousByID[row.FullID] = row
	}

	merged := make([]core.ContainerRow, len(currentRows))
	copy(merged, currentRows)
	for index, row := range merged {
		if !strings.EqualFold(row.State, "running") {
			continue
		}
		// Only preserve old metrics if current metrics are stale/missing
		if row.CPUPercent >= 0 && row.MemoryUsage != "-" && row.MemoryUsage != "loading" {
			continue // Current has real metrics, don't overwrite
		}
		previous, ok := previousByID[row.FullID]
		if !ok || previous.MemoryUsage == "-" || previous.MemoryUsage == "loading" {
			continue // Previous doesn't have good metrics either
		}
		merged[index].CPUPercent = previous.CPUPercent
		merged[index].MemoryPercent = previous.MemoryPercent
		merged[index].MemoryUsage = previous.MemoryUsage
		merged[index].MemoryLimit = previous.MemoryLimit
		merged[index].MemoryUsageBytes = previous.MemoryUsageBytes
		merged[index].MemoryLimitBytes = previous.MemoryLimitBytes
	}

	return merged
}
