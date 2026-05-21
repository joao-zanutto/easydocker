package tui

import (
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

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

func (m *model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	if m.screen == shared.Logs {
		m.logs.SyncFromData(m.logVisibleWidth(), m.logVisibleRows())
	}
	if m.screen == shared.Inspect {
		m.logs.SyncFromData(m.inspectVisibleWidth(), m.inspectVisibleRows())
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.help.Active {
		return m.handleHelpKey(msg)
	}

	if m.menu.Active {
		return m.handleMenuKey(msg)
	}

	if m.screen == shared.Browse && m.browseFilter.Active {
		return m.handleBrowseKey(msg)
	}

	if m.screen == shared.Logs && m.logs.Filter.Active {
		return m, m.handleLogsKey(msg)
	}

	route := shared.RouteRootKey(msg.String(), m.screen)
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

func (m model) handleMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := menu.NewKeyMap()
	transition := menu.Controller{}.HandleKey(&m.menu, &m.help, msg, keys)
	if transition.Quit {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleHelpKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := menu.NewKeyMap()
	helpHeight := m.height * 9 / 10
	bodyHeight := menu.HelpBodyHeight(helpHeight, m.styles.HelpFrame)
	contentHeight := menu.HelpBodyLineCount(m.help.Commands)
	transition := menu.Controller{}.HandleHelpKey(&m.help, &m.menu, msg, keys, contentHeight, bodyHeight)
	if transition.Back {
		m.help.Active = false
		m.menu.Active = true
	}
	return m, nil
}

func (m model) shouldAnimateLogsLoadingIndicator() bool {
	return m.screen == shared.Logs && (m.logs.InitialLoad || m.logs.HistoryLoad)
}

func (m model) shouldAnimateMetricsLoadingIndicator() bool {
	return !m.metricsLoaded && m.loadingStage != loadStageIdle
}

func (m model) handleTickMsg(_ tickMsg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{tickCmd()}
	if m.shouldReloadSnapshotOnTick() {
		cmds = append(cmds, loadDockerCmd(m.service))
	}
	if m.shouldLoadHistoryOnTick() {
		tail := len(m.logs.Data) + TailStep
		cmds = append(cmds, LoadLogsCmd(m.service, m.logs.ContainerID, m.logs.SessionID, tail, viewer.SourceHistory))
	} else if m.shouldPollLogsOnTick() {
		tail := m.logsPollTail()
		cmds = append(cmds, LoadLogsCmd(m.service, m.logs.ContainerID, m.logs.SessionID, tail, viewer.SourcePoll))
	}
	return m, tea.Batch(cmds...)
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
		return m, nil
	}

	return m, tea.Batch(cmds...)
}
