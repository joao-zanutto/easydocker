package tui

import (
	"easydocker/internal/core"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func (m *model) handleLogsKey(msg tea.KeyPressMsg) tea.Cmd {
	return m.handleViewerKey(msg, viewerKeyHandlers{
		visibleWidth: m.logVisibleWidth,
		visibleRows:  m.logVisibleRows,
		openFilter:   m.openLogsFilter,
		closeFilter:  m.closeLogsFilter,
		exitMode:     m.exitLogsMode,
		historyLoad: func() *viewer.LoadRequest {
			return HistoryLoadRequest(&m.logs)
		},
	})
}

func (m *model) enterLogsMode(container core.ContainerRow) tea.Cmd {
	transition := EnterLogsState(&m.logs, container.FullID)
	m.err = nil
	m.screen = shared.EnterLogsTransition()
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

func (m model) handleLogsResultMsg(msg viewer.ContentMsg) (tea.Model, tea.Cmd) {
	return m, m.handleLogsResult(msg)
}

func (m *model) applyLogsTransition(transition viewer.Transition) tea.Cmd {
	if transition.LaunchShell {
		if container, ok := m.selectedLogsContainer(); ok {
			return shellCmd(m.service, container.FullID)
		}
		return nil
	}
	if transition.ExitToBrowse {
		targetScreen, _ := shared.ExitLogsTransition(transition.ForceTab)
		m.screen = targetScreen
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

func (m model) shouldLoadHistoryOnTick() bool {
	return m.screen == shared.Logs &&
		m.logs.ContainerID != "" &&
		m.logs.Viewport.AtTop() &&
		!m.logs.InitialLoad &&
		!m.logs.HistoryLoad &&
		!m.logs.HistoryDone
}

func (m model) shouldPollLogsOnTick() bool {
	return m.screen == shared.Logs &&
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
