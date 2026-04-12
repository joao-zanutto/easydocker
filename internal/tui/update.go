package tui

import (
	"strings"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/loading"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/mode"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var logsController = logs.Controller{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case tea.KeyMsg:
		return m.handleKey(msg.String())
	case containersResultMsg:
		return m.handleContainersResultMsg(msg)
	case resourcesResultMsg:
		return m.handleResourcesResultMsg(msg)
	case metricsResultMsg:
		return m.handleMetricsResultMsg(msg)
	case loadResultMsg:
		return m.handleLoadResultMsg(msg)
	case logs.ResultMsg:
		return m.handleLogsResultMsg(msg)
	case tickMsg:
		return m.handleTickMsg(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTickMsg(msg)
	}

	return m, nil
}

const browseCursorPageStep = 5

func (m model) handleBrowseKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "right", "l":
		m.moveActiveTab(1)
	case "left", "h":
		m.moveActiveTab(-1)
	case "a":
		m.toggleContainerScope()
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-browseCursorPageStep)
	case "pgdown":
		m.moveCursor(browseCursorPageStep)
	case "enter":
		if cmd := m.enterLogsModeIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	case "esc", "backspace":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	if m.screen == screenModeLogs {
		m.logs.SyncViewportFromData(m.logVisibleWidth(), m.logVisibleRows())
	}
	return m, nil
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	route := mode.RouteRootKey(key, toModeScreen(m.screen))
	switch route {
	case mode.RouteQuit:
		return m, tea.Quit
	case mode.RouteNoop:
		return m, nil
	case mode.RouteLogs:
		return m, m.handleLogsKey(key)
	case mode.RouteBrowse:
		return m.handleBrowseKey(key)
	}

	return m, nil
}

func toModeScreen(screen screenMode) mode.Screen {
	if screen == screenModeLogs {
		return mode.Logs
	}
	return mode.Browse
}

func fromModeScreen(screen mode.Screen) screenMode {
	if screen == mode.Logs {
		return screenModeLogs
	}
	return screenModeBrowse
}

func (m *model) handleLogsKey(key string) tea.Cmd {
	transition := logsController.HandleKey(&m.logs, key, tabContainers)
	return m.applyLogsTransition(transition)
}

func (m *model) enterLogsMode(container core.ContainerRow) tea.Cmd {
	transition := logsController.Enter(&m.logs, container.FullID)
	m.err = nil
	m.screen = fromModeScreen(mode.EnterLogsTransition())
	return m.applyLogsTransition(transition)
}

func (m *model) exitLogsMode() {
	transition := logsController.Exit(&m.logs, tabContainers)
	_ = m.applyLogsTransition(transition)
}

func (m *model) handleLogsResult(msg logs.ResultMsg) tea.Cmd {
	transition := logsController.HandleResult(&m.logs, msg, m.logVisibleWidth(), m.logVisibleRows())
	return m.applyLogsTransition(transition)
}

func (m *model) applyLogsTransition(transition logs.Transition) tea.Cmd {
	if transition.ExitToBrowse {
		targetScreen, _ := mode.ExitLogsTransition(transition.ForceTab)
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
	return m.loadLogsDataCmd(
		request.ContainerID,
		request.SessionID,
		request.PrevCPU,
		request.PrevMem,
		request.Tail,
		request.Src,
	)
}

func (m *model) applyLoadingTransition(transition loading.Transition) {
	m.loading = transition.Loading
	m.loadingStage = int(transition.Stage)
	m.err = transition.Err
}

func (m *model) setLoadError(err error) {
	m.applyLoadingTransition(loading.Fail(err))
}

func (m *model) beginLoadingStage(stage int) {
	m.applyLoadingTransition(loading.Begin(loading.Stage(stage)))
	m.snapshot.Timestamp = time.Now()
	m.clampCursors()
}

func (m *model) finishLoadingStage(err error) bool {
	transition, ok := loading.Finish(err)
	m.applyLoadingTransition(transition)
	return ok
}

func (m model) shouldReloadSnapshotOnTick() bool {
	return m.loadingStage == loadStageIdle
}

func (m model) shouldPollLogsOnTick() bool {
	return m.screen == screenModeLogs && m.logs.ContainerID != ""
}

func (m model) logsPollTail() int {
	if m.logs.TailLines <= 0 {
		return logs.InitialTail
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

	m.snapshot = msg.snapshot
	if err := m.reconcileLogsSelection(); err != nil {
		m.err = err
	}
	m.clampCursors()
	return m.respond(noSideEffect())
}

func (m model) handleLogsResultMsg(msg logs.ResultMsg) (tea.Model, tea.Cmd) {
	return m.respond(withSideEffect(m.handleLogsResult(msg)))
}

func (m model) handleTickMsg(_ tickMsg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{tickCmd()}
	if m.shouldReloadSnapshotOnTick() {
		cmds = append(cmds, m.loadDockerCmd())
	}
	if m.shouldPollLogsOnTick() {
		tail := m.logsPollTail()
		cmds = append(cmds, m.loadLogsDataCmd(m.logs.ContainerID, m.logs.SessionID, m.logs.Data.CPUHistory, m.logs.Data.MemHistory, tail, logs.SourcePoll))
	}
	return m.respond(withSideEffect(tea.Batch(cmds...)))
}

func (m model) handleSpinnerTickMsg(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if !m.shouldAnimateMetricsLoadingIndicator() {
		return m.respond(noSideEffect())
	}

	var cmd tea.Cmd
	m.metricsSpinner, cmd = m.metricsSpinner.Update(msg)
	return m.respond(withSideEffect(cmd))
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
		previous, ok := previousByID[row.FullID]
		if !ok || previous.MemoryUsage == "-" || previous.MemoryUsage == "loading" {
			continue
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
