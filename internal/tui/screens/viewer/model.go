package viewer

import (
	"strings"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	State
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

	HistoryLoad               bool
	HistoryDone               bool
	SessionID                 int
	TailLines                 int
	HistoryBaseLen            int
	HistoryAppendedDuringLoad int
	HistoryNoProgressCount    int

	LoadCmd func() tea.Msg
}

func NewModel() Model {
	st := NewState()
	return Model{
		State:      st,
		LoadingMsg: "Loading...",
		EmptyMsg:   "No content available.",
		Spinner:    spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinnerTickCmd()}
	if m.InitialLoad && m.LoadCmd != nil {
		cmds = append(cmds, m.LoadCmd)
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case ContentMsg:
		return m.handleContentMsg(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		if cmd != nil {
			return m, tea.Batch(cmd, m.spinnerTickCmd())
		}
		return m, m.spinnerTickCmd()
	}
	return m, nil
}

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	vm := ViewModel{
		State:            &m.State,
		ContainerName:    m.ContainerName,
		Breadcrumb:       m.Breadcrumb,
		LineCount:        m.lineCountInfo(),
		LoadingMessage:   m.LoadingMsg,
		EmptyMessage:     m.EmptyMsg,
		LoadingIndicator: m.loadingIndicator(),
		Width:            m.Width,
		Height:           m.Height,
		Styles:           m.Styles,
		ContentType:      m.ContentType,
		ResourceType:     m.ResourceType,
		HistoryLoad:      m.HistoryLoad,
	}

	return RenderContent(vm)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	keys := NewKeyMap()

	if m.Filter.Active {
		return m.handleFilterKey(msg, keys)
	}

	if key.Matches(msg, keys.Back) {
		return m, func() tea.Msg { return TransitionMsg{BackToBrowse: true} }
	}

	if key.Matches(msg, keys.ToggleWrap) {
		logList := FilterLines(m.Data, m.Filter.Query)
		currentStart, _ := VisibleContentRange(&m.State, logList)

		m.SetWrapLines(!m.WrapLines)
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())

		if m.WrapLines {
			w := m.VisibleWidth()
			row := 0
			for i := 0; i < currentStart && i < len(logList); i++ {
				row += WrappedRowCount(SanitizeLine(logList[i]), w)
			}
			m.Viewport.SetYOffset(row)
		} else {
			m.Viewport.SetYOffset(currentStart)
		}
		return m, nil
	}

	if key.Matches(msg, keys.ToggleFollow) {
		m.SetFollow(!m.Follow)
		return m, nil
	}

	if key.Matches(msg, keys.OpenFilter) {
		m.OpenFilter()
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	}

	if key.Matches(msg, keys.OpenShell) {
		return m, func() tea.Msg { return TransitionMsg{LaunchShell: true} }
	}

	transition := Controller{}.HandleKey(&m.State, msg, keys)
	m.SyncFromData(m.VisibleWidth(), m.VisibleRows())

	if transition.LaunchShell {
		return m, func() tea.Msg { return TransitionMsg{LaunchShell: true} }
	}
	return m, nil
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg, keys KeyMap) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.CloseFilter(true)
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	case msg.String() == "enter":
		m.CloseFilter(false)
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, nil
	default:
		var cmd tea.Cmd
		m.Filter.Input, cmd = m.Filter.Input.Update(msg)
		m.Filter.Query = m.Filter.Input.Value()
		m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
		return m, cmd
	}
}

func (m Model) handleContentMsg(msg ContentMsg) (Model, tea.Cmd) {
	if msg.SessionID != m.SessionID || msg.ContainerID != m.ContainerID {
		return m, nil
	}
	if msg.Err != nil {
		m.InitialLoad = false
		m.HistoryLoad = false
		return m, nil
	}

	if msg.Tail > 0 && msg.Tail > m.TailLines {
		m.TailLines = msg.Tail
	}

	switch msg.Src {
	case SourceHistory:
		m.applyHistoryWithMerge(msg.Data)
	case SourceInitial:
		m.ApplyInitial(msg.Data)
	default:
		m.applyPollWithMerge(msg.Data)
	}
	m.SyncFromData(m.VisibleWidth(), m.VisibleRows())
	return m, nil
}

func (m Model) loadingIndicator() string {
	if m.InitialLoad || m.HistoryLoad {
		return strings.TrimSpace(m.Spinner.View())
	}
	return ""
}

func (m Model) lineCountInfo() *LineCountInfo {
	logList := FilterLines(m.Data, m.Filter.Query)
	start, end := VisibleContentRange(&m.State, logList)
	return &LineCountInfo{Total: len(logList), Start: start + 1, End: max(start+1, end)}
}

func (m *Model) ApplyInitial(data []string) {
	m.InitialLoad = false
	m.HistoryLoad = false
	m.HistoryDone = false
	m.Data = data
}

func (m Model) VisibleWidth() int {
	return max(1, m.Width-m.Styles.SubpageFrame.GetHorizontalFrameSize())
}

func (m Model) VisibleRows() int {
	return VisibleRowsForContent(m.Height, m.Filter.Active)
}

func (m *Model) applyHistoryWithMerge(data []string) {
	previousLen := len(m.Data)
	if m.HistoryBaseLen > 0 {
		previousLen = m.HistoryBaseLen
	}

	m.HistoryLoad = false
	prepended := len(data) - previousLen - m.HistoryAppendedDuringLoad
	if prepended < 0 {
		prepended = 0
	}

	if prepended == 0 {
		m.HistoryNoProgressCount++
	} else {
		m.HistoryNoProgressCount = 0
	}

	if m.HistoryNoProgressCount >= 3 {
		m.HistoryDone = true
	}

	m.Data = data
	m.HistoryBaseLen = 0
	m.HistoryAppendedDuringLoad = 0
}

func (m *Model) applyPollWithMerge(data []string) {
	previousLen := len(m.Data)
	merged, _ := MergePolledLogs(m.Data, data)
	if m.HistoryLoad {
		m.HistoryAppendedDuringLoad += max(0, len(merged)-previousLen)
	}
	m.Data = merged
	m.InitialLoad = false
}

func (m Model) spinnerTickCmd() tea.Cmd {
	return tea.Tick(shared.SpinnerTickInterval, func(t time.Time) tea.Msg {
		return spinner.TickMsg{Time: t, ID: m.Spinner.ID()}
	})
}

type TransitionMsg struct {
	BackToBrowse bool
	LaunchShell  bool
}
