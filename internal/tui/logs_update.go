package tui

import (
	"strings"

	"easydocker/internal/core"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 0 means unbounded retention so users can scroll the full log history.
const maxLiveLogLines = 0

const (
	logsInitialTail = 200
	logsTailStep    = 200
)

func (m *model) handleLogsKey(key string) tea.Cmd {
	switch key {
	case "right":
		m.logs.follow = false
		m.logs.scrollX += 8
		m.clampLogScroll()
	case "left":
		m.logs.follow = false
		m.logs.scrollX -= 8
		m.clampLogScroll()
	case "up", "k":
		m.logs.follow = false
		m.logs.scrollY--
		m.clampLogScroll()
		return m.loadMoreLogsCmdIfNeeded()
	case "down", "j":
		if m.logViewportEnd() >= len(m.logs.data.Logs) {
			m.logs.follow = true
			m.clampLogScroll()
			return nil
		}
		m.logs.follow = false
		m.logs.scrollY++
		m.clampLogScroll()
	case "pgup":
		m.logs.follow = false
		m.logs.scrollY -= m.logPageStep()
		m.clampLogScroll()
		return m.loadMoreLogsCmdIfNeeded()
	case "pgdown":
		if m.logViewportEnd() >= len(m.logs.data.Logs) {
			m.logs.follow = true
			m.clampLogScroll()
			return nil
		}
		m.logs.follow = false
		m.logs.scrollY += m.logPageStep()
		m.clampLogScroll()
	case "home":
		m.logs.follow = false
		m.logs.scrollX = 0
		m.logs.scrollY = 0
		m.clampLogScroll()
		return m.loadMoreLogsCmdIfNeeded()
	case "end":
		m.logs.follow = true
		m.clampLogScroll()
	case "f":
		m.logs.follow = !m.logs.follow
		m.clampLogScroll()
	case "esc", "backspace":
		m.exitLogsMode()
	case " ", "b", "g", "G", "q", "tab":
		return nil
	}
	return nil
}

func (m *model) enterLogsMode(container core.ContainerRow) tea.Cmd {
	m.err = nil
	m.screen = screenModeLogs
	m.logs.sessionID++
	m.logs.containerID = container.FullID
	m.logs.data = core.ContainerLiveData{}
	m.logs.scrollX = 0
	m.logs.scrollY = 0
	m.logs.follow = true
	m.logs.tailLines = logsInitialTail
	m.logs.initialLoad = true
	m.logs.historyDone = false
	m.logs.historyLoad = false
	return m.loadLogsDataCmd(container.FullID, m.logs.sessionID, nil, nil, m.logs.tailLines, "initial")
}

func (m *model) exitLogsMode() {
	m.screen = screenModeBrowse
	m.logs.sessionID++
	m.activeTab = tabContainers
	m.logs = logsState{follow: true}
}

func (m *model) loadMoreLogsCmdIfNeeded() tea.Cmd {
	if m.logs.scrollY != 0 || m.logs.historyDone || m.logs.historyLoad {
		return nil
	}
	m.logs.historyLoad = true
	m.logs.historyDone = false
	nextTail := len(m.logs.data.Logs) + logsTailStep
	return m.loadLogsDataCmd(m.logs.containerID, m.logs.sessionID, m.logs.data.CPUHistory, m.logs.data.MemHistory, nextTail, "history")
}

func (m *model) handleLogsResult(msg logsResultMsg) tea.Cmd {
	if msg.sessionID != m.logs.sessionID || msg.containerID != m.logs.containerID {
		return nil
	}
	if msg.err != nil {
		m.err = msg.err
		m.logs.initialLoad = false
		m.logs.historyLoad = false
		return nil
	}

	previousLen := len(m.logs.data.Logs)
	previousScrollY := m.logs.scrollY
	if msg.tail > 0 && msg.tail > m.logs.tailLines {
		m.logs.tailLines = msg.tail
	}

	if msg.src == "history" {
		m.logs.historyLoad = false
		if len(msg.data.Logs) <= previousLen {
			m.logs.historyDone = true
		}
		delta := len(msg.data.Logs) - previousLen
		m.logs.data = msg.data
		if !m.logs.follow && delta > 0 {
			m.logs.scrollY = previousScrollY + delta
		}
		m.clampLogScroll()
		return nil
	}

	if msg.src == "initial" {
		m.logs.initialLoad = false
		m.logs.historyLoad = false
		m.logs.historyDone = false
		m.logs.data = msg.data
		m.clampLogScroll()
		return nil
	}

	mergedLogs, overlapFound := mergePolledLogs(m.logs.data.Logs, msg.data.Logs, maxLiveLogLines)
	if overlapFound || len(m.logs.data.Logs) == 0 {
		msg.data.Logs = mergedLogs
	} else {
		msg.data.Logs = trimLogs(msg.data.Logs, maxLiveLogLines)
	}
	m.logs.data = msg.data
	m.logs.initialLoad = false
	if !m.logs.follow {
		m.logs.scrollY = previousScrollY
	}
	m.clampLogScroll()
	return nil
}

func (m *model) clampLogScroll() {
	visibleRows := m.logVisibleRows()
	maxY := max(0, len(m.logs.data.Logs)-visibleRows)
	if m.logs.follow {
		m.logs.scrollY = maxY
	} else {
		m.logs.scrollY = clamp(m.logs.scrollY, 0, maxY)
	}
	visibleWidth := m.logVisibleWidth()
	start := clamp(m.logs.scrollY, 0, maxY)
	end := min(len(m.logs.data.Logs), start+visibleRows)
	maxX := 0
	for _, line := range m.logs.data.Logs[start:end] {
		maxX = max(maxX, logLineMaxOffset(line, visibleWidth))
	}
	m.logs.scrollX = clamp(m.logs.scrollX, 0, maxX)
}

func (m model) logPageStep() int {
	return max(3, m.logVisibleRows()-1)
}

func (m model) logVisibleRows() int {
	return max(1, m.logSectionHeight())
}

func (m model) logVisibleWidth() int {
	totalWidth := max(1, m.width)
	if m.screen == screenModeLogs {
		pageContentWidth := m.logsPageContentWidth(totalWidth)
		return max(1, pageContentWidth-4)
	}
	innerWidth := totalWidth - m.styles.mainFrame.GetHorizontalFrameSize()
	return max(1, innerWidth-2)
}

func (m model) logSectionHeight() int {
	mainHeight := m.height - lipgloss.Height(m.renderHeader()) - lipgloss.Height(m.renderFooter())
	if mainHeight < 1 {
		mainHeight = 1
	}
	if m.screen == screenModeLogs {
		return max(1, m.logsPageContentHeight(mainHeight)-3)
	}
	innerHeight := max(1, mainHeight-m.styles.mainFrame.GetVerticalFrameSize())
	return max(1, innerHeight-2)
}

func (m model) logViewportEnd() int {
	visibleRows := m.logVisibleRows()
	maxY := max(0, len(m.logs.data.Logs)-visibleRows)
	start := clamp(m.logs.scrollY, 0, maxY)
	return min(len(m.logs.data.Logs), start+visibleRows)
}

func (m model) loadLogsDataCmd(containerID string, sessionID int, previousCPU, previousMem []float64, tail int, src string) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		data, err := svc.LoadContainerLiveData(containerID, previousCPU, previousMem, tail)
		return logsResultMsg{containerID: containerID, sessionID: sessionID, data: data, err: err, tail: tail, src: src}
	}
}

func (m model) selectedLogsContainer() (core.ContainerRow, bool) {
	if m.logs.containerID == "" {
		return core.ContainerRow{}, false
	}
	for _, container := range m.snapshot.Containers {
		if container.FullID == m.logs.containerID {
			return container, true
		}
	}
	return core.ContainerRow{}, false
}

func mergePolledLogs(previous, polled []string, maxLines int) ([]string, bool) {
	if len(previous) == 0 {
		return trimLogs(polled, maxLines), true
	}
	if len(polled) == 0 {
		return previous, true
	}

	normalizedPrevious := make([]string, 0, len(previous))
	for _, line := range previous {
		normalizedPrevious = append(normalizedPrevious, strings.TrimRight(line, "\r"))
	}
	normalizedPolled := make([]string, 0, len(polled))
	for _, line := range polled {
		normalizedPolled = append(normalizedPolled, strings.TrimRight(line, "\r"))
	}

	maxOverlap := min(len(normalizedPrevious), len(normalizedPolled))
	for overlap := maxOverlap; overlap > 0; overlap-- {
		if !equalLogSlices(normalizedPrevious[len(normalizedPrevious)-overlap:], normalizedPolled[:overlap]) {
			continue
		}
		merged := append([]string{}, normalizedPrevious...)
		merged = append(merged, normalizedPolled[overlap:]...)
		return trimLogs(merged, maxLines), true
	}

	if equalLogSlices(normalizedPrevious, normalizedPolled) {
		return trimLogs(normalizedPrevious, maxLines), true
	}

	if len(normalizedPolled) < len(normalizedPrevious) && equalLogSlices(normalizedPrevious[len(normalizedPrevious)-len(normalizedPolled):], normalizedPolled) {
		return trimLogs(normalizedPrevious, maxLines), true
	}

	return trimLogs(normalizedPolled, maxLines), false
}

func trimLogs(lines []string, maxLines int) []string {
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	return lines[len(lines)-maxLines:]
}

func equalLogSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
