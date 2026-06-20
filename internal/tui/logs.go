package tui

import (
	"context"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

const (
	InitialTail = 200
	TailStep    = 200
)

func (m *model) enterLogsMode(container core.ContainerRow) tea.Cmd {
	m.initViewer(shared.LogViewer, viewer.ContentTypeLogs, core.ResourceContainer,
		container.FullID, container.Name, "Loading logs...", "No logs found for this container.")
	m.viewer.Vp.Follow = true
	m.viewer.Logs = viewer.NewLogsViewer()
	return LoadLogsCmd(m.service, m.viewer.ContainerID, m.viewer.Logs.SessionID,
		m.viewer.Logs.TailLines, viewer.SourceInitial)
}

func LoadLogsCmd(service core.ServiceInterface, containerID string, sessionID, tail int, src viewer.Source) tea.Cmd {
	return func() tea.Msg {
		logs, err := service.LoadContainerLogs(context.Background(), containerID, tail)
		return viewer.ContentMsg{
			ContainerID: containerID,
			SessionID:   sessionID,
			Data:        logs,
			Err:         err,
			Tail:        tail,
			Src:         src,
		}
	}
}
