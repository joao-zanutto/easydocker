package tui

import (
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.syncBrowseData()

	var updated tea.Model
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		updated, cmd = m.handleWindowSizeMsg(msg)
	case tea.KeyPressMsg:
		updated, cmd = m.handleKey(msg)
	case containersResultMsg:
		updated, cmd = m.handleContainersResultMsg(msg)
	case resourcesResultMsg:
		updated, cmd = m.handleResourcesResultMsg(msg)
	case metricsResultMsg:
		updated, cmd = m.handleMetricsResultMsg(msg)
	case loadResultMsg:
		updated, cmd = m.handleLoadResultMsg(msg)
	case viewer.ContentMsg:
		updated, cmd = m.handleViewerContentMsg(msg)
	case inspectResultMsg:
		updated, cmd = m.handleInspectResultMsg(msg)
	case shellDoneMsg:
		m.err = msg.err
		updated, cmd = m, nil
	case tickMsg:
		updated, cmd = m.handleTickMsg(msg)
	case spinner.TickMsg:
		updated, cmd = m.handleSpinnerTickMsg(msg)
	case browse.TransitionMsg:
		updated, cmd = m.handleBrowseTransition(msg)
	case viewer.TransitionMsg:
		updated, cmd = m.handleViewerTransition(msg)
	}

	switch v := updated.(type) {
	case model:
		m = v
	case *model:
		m = *v
	}
	m = m.syncBrowseData()
	return m, cmd
}

func (m *model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	var cmd tea.Cmd
	m.browse, cmd = m.browse.Update(msg)
	if m.screen == shared.LogViewer || m.screen == shared.InspectViewer {
		var vcmd tea.Cmd
		m.viewer, vcmd = m.viewer.Update(msg)
		if vcmd != nil {
			cmd = tea.Batch(cmd, vcmd)
		}
	}
	return m, cmd
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

	if m.screen == shared.Main {
		var cmd tea.Cmd
		m.browse, cmd = m.browse.Update(msg)
		return m, cmd
	}

	if m.screen == shared.LogViewer || m.screen == shared.InspectViewer {
		var cmd tea.Cmd
		m.viewer, cmd = m.viewer.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleBrowseTransition(msg browse.TransitionMsg) (tea.Model, tea.Cmd) {
	if msg.OpenMenu {
		m.menu.Active = true
		m.menu.Cursor = 0
		return m, nil
	}
	if msg.OpenResource {
		if cmd := m.enterLogsModeIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if msg.OpenShell {
		if cmd := m.openShellIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if msg.OpenInspect {
		return m.handleInspectTransition()
	}
	return m, nil
}

func (m model) handleViewerTransition(msg viewer.TransitionMsg) (tea.Model, tea.Cmd) {
	if msg.BackToBrowse {
		m.screen = m.previousScreen
		return m, nil
	}
	if msg.LaunchShell {
		if container, ok := m.selectedLogsContainer(); ok {
			return m, shellCmd(m.service, container.FullID)
		}
	}
	return m, nil
}

func (m model) handleMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := menu.NewKeyMap()
	m.help.Commands = m.buildHelpCommands()
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

func (m model) handleTickMsg(_ tickMsg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{tickCmd()}
	if m.shouldReloadSnapshotOnTick() {
		cmds = append(cmds, loadDockerCmd(m.service))
	}
	if m.shouldLoadHistoryOnTick() {
		tail := len(m.viewer.Data) + TailStep
		cmds = append(cmds, LoadLogsCmd(m.service, m.viewer.ContainerID, m.viewer.SessionID, tail, viewer.SourceHistory))
	} else if m.shouldPollLogsOnTick() {
		tail := m.logsPollTail()
		cmds = append(cmds, LoadLogsCmd(m.service, m.viewer.ContainerID, m.viewer.SessionID, tail, viewer.SourcePoll))
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
		m.viewer.Spinner, cmd = m.viewer.Spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) == 0 {
		return m, nil
	}

	return m, tea.Batch(cmds...)
}
