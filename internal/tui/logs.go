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
	m.pushScreen(m.screen)
	m.screen = shared.LogViewer
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.viewer.ContainerID = container.FullID
	m.viewer.Logs.SessionID++
	m.viewer.Logs.TailLines = InitialTail
	m.viewer.Vp.InitialLoad = true
	m.viewer.Vp.Follow = true
	m.viewer.Vp.Data = nil
	m.viewer.Logs.HistoryLoad = false
	m.viewer.Logs.HistoryDone = false
	m.viewer.Logs.HistoryBaseLen = 0
	m.viewer.Logs.HistoryAppendedDuringLoad = 0
	m.viewer.Logs.HistoryNoProgressCount = 0
	m.viewer.Vp.SetXOffset(0)
	m.viewer.Vp.Filter.Active = false
	m.viewer.Vp.Filter.Query = ""
	m.viewer.Vp.Filter.Input.SetValue("")
	m.viewer.Vp.ContentType = viewer.ContentTypeLogs
	m.viewer.ResourceType = core.ResourceContainer
	m.viewer.LoadingMsg = "Loading logs..."
	m.viewer.EmptyMsg = "No logs found for this container."
	m.viewer.Breadcrumb = ""
	m.viewer.ContainerName = container.Name
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Viewer.Breadcrumb,
		FollowOn:     m.styles.Viewer.FollowOn,
		FollowOff:    m.styles.Viewer.FollowOff,
		Muted:        m.styles.Browse.Muted,
		Divider:      m.styles.Browse.Divider,
		SubpageFrame: m.styles.Viewer.SubpageFrame,
		Key:          m.styles.Chrome.Key,
		KeyText:      m.styles.Chrome.KeyText,
	}
	m.err = nil
	return LoadLogsCmd(m.service, m.viewer.ContainerID, m.viewer.Logs.SessionID, m.viewer.Logs.TailLines, viewer.SourceInitial)
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
