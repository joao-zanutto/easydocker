package viewer

import (
	"easydocker/internal/core"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Vp            *Viewport
	Logs          LogsViewer
	Inspect       InspectViewer
	ContainerID   string
	Loading       bool
	LoadingMsg    string
	EmptyMsg      string
	Breadcrumb    string
	ContainerName string
	ResourceType  core.ResourceType
	Width         int
	Height        int
	Styles        Styles
}

func NewModel() Model {
	vp := NewViewport()
	return Model{
		Vp:         vp,
		Logs:       NewLogsViewer(),
		Inspect:    NewInspectViewer(),
		LoadingMsg: "Loading...",
		EmptyMsg:   "No content available.",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case ContentMsg:
		return m.handleContentMsg(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	vm := ViewModel{
		Vp:               m.Vp,
		ContainerName:    m.ContainerName,
		Breadcrumb:       m.Breadcrumb,
		LineCount:        m.lineCountInfo(),
		LoadingMessage:   m.LoadingMsg,
		EmptyMessage:     m.EmptyMsg,
		LoadingIndicator: m.loadingIndicator(),
		Width:            m.Width,
		Height:           m.Height,
		Styles:           m.Styles,
		ContentType:      m.Vp.ContentType,
		ResourceType:     m.ResourceType,
		Logs:             m.Logs,
	}

	return RenderContent(vm)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	keys := NewKeyMap()

	if m.Vp.Filter.Active {
		return m.handleFilterKey(msg, keys)
	}

	if key.Matches(msg, keys.Back) {
		return m, func() tea.Msg { return TransitionMsg{BackToBrowse: true} }
	}

	if key.Matches(msg, keys.ToggleWrap) {
		logList := FilterLines(m.Vp.Data, m.Vp.Filter.Query)
		currentStart, _ := VisibleContentRange(m.Vp, logList)
		m.Vp.SetWrapLines(!m.Vp.WrapLines)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		if m.Vp.WrapLines {
			w := m.VisibleWidth()
			row := 0
			for i := 0; i < currentStart && i < len(logList); i++ {
				row += WrappedRowCount(SanitizeLine(logList[i]), w)
			}
			m.Vp.SetYOffset(row)
		} else {
			m.Vp.SetYOffset(currentStart)
		}
		return m, nil
	}

	if key.Matches(msg, keys.ToggleFollow) {
		m.Vp.SetFollow(!m.Vp.Follow)
		return m, nil
	}

	if key.Matches(msg, keys.OpenFilter) {
		m.Vp.OpenFilter()
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	}

	if key.Matches(msg, keys.OpenShell) {
		return m, func() tea.Msg { return TransitionMsg{LaunchShell: true} }
	}

	transition := Controller{}.HandleKey(m.Vp, msg, keys)
	m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())

	if transition.LaunchShell {
		return m, func() tea.Msg { return TransitionMsg{LaunchShell: true} }
	}
	return m, nil
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg, keys KeyMap) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.Vp.CloseFilter(true)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	case msg.String() == "enter":
		m.Vp.CloseFilter(false)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	default:
		var cmd tea.Cmd
		m.Vp.Filter.Input, cmd = m.Vp.Filter.Input.Update(msg)
		m.Vp.Filter.Query = m.Vp.Filter.Input.Value()
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, cmd
	}
}

func (m Model) handleContentMsg(msg ContentMsg) (Model, tea.Cmd) {
	if msg.SessionID != m.Logs.SessionID || msg.ContainerID != m.ContainerID {
		return m, nil
	}
	if msg.Err != nil {
		m.Vp.InitialLoad = false
		m.Logs.HistoryLoad = false
		return m, nil
	}

	if msg.Tail > 0 && msg.Tail > m.Logs.TailLines {
		m.Logs.TailLines = msg.Tail
	}

	switch msg.Src {
	case SourceHistory:
		m.applyHistoryWithMerge(msg.Data)
	case SourceInitial:
		m.ApplyInitial(msg.Data)
	default:
		m.applyPollWithMerge(msg.Data)
	}
	m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
	return m, nil
}

func (m Model) loadingIndicator() string {
	if m.Vp.InitialLoad || m.Logs.HistoryLoad {
		return ""
	}
	return ""
}

func (m Model) lineCountInfo() *LineCountInfo {
	logList := FilterLines(m.Vp.Data, m.Vp.Filter.Query)
	start, end := VisibleContentRange(m.Vp, logList)
	return &LineCountInfo{Total: len(logList), Start: start + 1, End: max(start+1, end)}
}

func (m *Model) ApplyInitial(data []string) {
	m.Vp.InitialLoad = false
	m.Logs.HistoryLoad = false
	m.Logs.HistoryDone = false
	m.Vp.Data = data
}

func (m Model) VisibleWidth() int {
	return max(1, m.Width-m.Styles.SubpageFrame.GetHorizontalFrameSize())
}

func (m Model) VisibleRows() int {
	return VisibleRowsForContent(m.Height)
}

func (m *Model) applyHistoryWithMerge(data []string) {
	previousLen := len(m.Vp.Data)
	if m.Logs.HistoryBaseLen > 0 {
		previousLen = m.Logs.HistoryBaseLen
	}

	m.Logs.HistoryLoad = false
	prepended := len(data) - previousLen - m.Logs.HistoryAppendedDuringLoad
	if prepended < 0 {
		prepended = 0
	}

	if prepended == 0 {
		m.Logs.HistoryNoProgressCount++
	} else {
		m.Logs.HistoryNoProgressCount = 0
	}

	if m.Logs.HistoryNoProgressCount >= 3 {
		m.Logs.HistoryDone = true
	}

	m.Vp.Data = data
	m.Logs.HistoryBaseLen = 0
	m.Logs.HistoryAppendedDuringLoad = 0
}

func (m *Model) applyPollWithMerge(data []string) {
	previousLen := len(m.Vp.Data)
	merged, _ := MergePolledLogs(m.Vp.Data, data)
	if m.Logs.HistoryLoad {
		m.Logs.HistoryAppendedDuringLoad += max(0, len(merged)-previousLen)
	}
	m.Vp.Data = merged
	m.Vp.InitialLoad = false
}
