package tui

import (
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
	m.previousScreen = m.screen
	m.screen = shared.Logs
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.viewer.ContainerID = container.FullID
	m.viewer.SessionID++
	m.viewer.TailLines = InitialTail
	m.viewer.InitialLoad = true
	m.viewer.Follow = true
	m.viewer.Data = nil
	m.viewer.HistoryLoad = false
	m.viewer.HistoryDone = false
	m.viewer.HistoryBaseLen = 0
	m.viewer.HistoryAppendedDuringLoad = 0
	m.viewer.HistoryNoProgressCount = 0
	m.viewer.Viewport.SetXOffset(0)
	m.viewer.Filter.Active = false
	m.viewer.Filter.Query = ""
	m.viewer.Filter.Input.SetValue("")
	m.viewer.ContentType = viewer.ContentTypeLogs
	m.viewer.ResourceType = core.ResourceContainer
	m.viewer.LoadingMsg = "Loading logs..."
	m.viewer.EmptyMsg = "No logs found for this container."
	m.viewer.Breadcrumb = ""
	m.viewer.ContainerName = container.Name
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Breadcrumb,
		FollowOn:     m.styles.FollowOn,
		FollowOff:    m.styles.FollowOff,
		Muted:        m.styles.Muted,
		Divider:      m.styles.Divider,
		SubpageFrame: m.styles.SubpageFrame,
	}
	m.err = nil
	return LoadLogsCmd(m.service, m.viewer.ContainerID, m.viewer.SessionID, m.viewer.TailLines, viewer.SourceInitial)
}

func LoadLogsCmd(service *core.Service, containerID string, sessionID, tail int, src viewer.Source) tea.Cmd {
	return func() tea.Msg {
		logs, err := service.LoadContainerLogs(containerID, tail)
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
