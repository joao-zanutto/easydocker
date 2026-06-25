package viewer

import (
	"log/slog"
	"strings"

	"easydocker/internal/core"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
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
	Spinner       spinner.Model
}

func NewModel() Model {
	vp := NewViewport()
	return Model{
		Vp:         vp,
		Logs:       NewLogsViewer(),
		Inspect:    NewInspectViewer(),
		Spinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
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

	filtered := m.Vp.FilteredLines()

	vm := ViewModel{
		Vp:               m.Vp,
		ContainerName:    m.ContainerName,
		Breadcrumb:       m.Breadcrumb,
		LineCount:        m.lineCountInfo(filtered),
		LoadingMessage:   m.LoadingMsg,
		EmptyMessage:     m.EmptyMsg,
		LoadingIndicator: m.loadingIndicator(),
		Width:            m.Width,
		Height:           m.Height,
		Styles:           m.Styles,
		ContentType:      m.Vp.ContentType,
		ResourceType:     m.ResourceType,
		Logs:             m.Logs,
		FilteredLines:    filtered,
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
		logList := m.Vp.FilteredLines()
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

	Controller{}.HandleKey(m.Vp, msg, keys)
	return m, nil
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg, keys KeyMap) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.Vp.CloseFilter(true)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		m.Vp.SetYOffset(m.Vp.savedYOffset)
		return m, nil
	case msg.String() == "enter":
		m.Vp.CloseFilter(false)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down),
		key.Matches(msg, keys.PageUp), key.Matches(msg, keys.PageDown),
		key.Matches(msg, keys.Home), key.Matches(msg, keys.End):
		Controller{}.HandleKey(m.Vp, msg, keys)
		return m, nil
	default:
		var cmd tea.Cmd
		m.Vp.Filter.Input, cmd = m.Vp.Filter.Input.Update(msg)
		prevQuery := m.Vp.Filter.Query
		m.Vp.Filter.Query = m.Vp.Filter.Input.Value()
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		if prevQuery != "" && m.Vp.Filter.Query == "" {
			m.Vp.SetYOffset(m.Vp.savedYOffset)
		}
		return m, cmd
	}
}

func (m Model) handleContentMsg(msg ContentMsg) (Model, tea.Cmd) {
	if msg.SessionID != m.Logs.SessionID || msg.ContainerID != m.ContainerID {
		return m, nil
	}
	if msg.Err != nil {
		slog.Error("log load failed", "sessionID", msg.SessionID, "error", msg.Err)
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
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
	case SourceInitial:
		m.ApplyInitial(msg.Data)
		m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
	default:
		if m.applyPollWithMerge(msg.Data) {
			m.Vp.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		}
	}
	return m, nil
}

func (m Model) loadingIndicator() string {
	if m.Vp.InitialLoad || m.Logs.HistoryLoad {
		return strings.TrimSpace(m.Spinner.View())
	}
	return ""
}

func (m Model) lineCountInfo(logList []string) *LineCountInfo {
	start, end := VisibleContentRange(m.Vp, logList)
	return &LineCountInfo{Total: len(logList), Start: start + 1, End: max(start+1, end)}
}

func (m *Model) ApplyInitial(data []string) {
	m.Vp.InitialLoad = false
	m.Logs.HistoryLoad = false
	m.Logs.HistoryDone = false
	m.Vp.Data = data
	m.Vp.InvalidateSanitizeCache()
}

func (m Model) VisibleWidth() int {
	return max(1, m.Width-m.Styles.SubpageFrame.GetHorizontalFrameSize())
}

func (m Model) VisibleRows() int {
	frameHeight := m.Styles.SubpageFrame.GetVerticalFrameSize()
	contentHeight := max(1, max(1, m.Height)-frameHeight)
	return VisibleRowsForContent(contentHeight)
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

	currentYOffset := m.Vp.YOffset()
	m.Vp.Data = data
	if prepended > 0 {
		m.Vp.PrependToCache(data, prepended)
	} else {
		m.Vp.InvalidateSanitizeCache()
	}
	m.Logs.HistoryBaseLen = 0
	m.Logs.HistoryAppendedDuringLoad = 0

	if prepended > 0 {
		if !m.Vp.WrapLines {
			m.Vp.SetYOffset(currentYOffset + prepended)
		} else {
			wrapWidth := max(1, m.Vp.Width())
			extraRows := 0
			for i := 0; i < prepended && i < len(data); i++ {
				extraRows += WrappedRowCount(SanitizeLine(data[i]), wrapWidth)
			}
			m.Vp.SetYOffset(currentYOffset + extraRows)
		}
	}
}

func (m *Model) applyPollWithMerge(data []string) bool {
	previous := m.Vp.Data
	previousLen := len(previous)
	merged, ok := MergePolledLogs(previous, data)
	if m.Logs.HistoryLoad {
		m.Logs.HistoryAppendedDuringLoad += max(0, len(merged)-previousLen)
	}
	if !ok {
		m.Vp.InvalidateSanitizeCache()
	}
	same := len(previous) == len(merged) &&
		(len(previous) == 0 || &previous[0] == &merged[0])
	m.Vp.wrapCanAppend = ok && !same && len(merged) > len(previous)
	m.Vp.Data = merged
	m.Vp.InitialLoad = false
	return !same
}
