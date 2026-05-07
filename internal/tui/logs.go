package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/viewer"

	tea "charm.land/bubbletea/v2"
)

const (
	InitialTail = 200
	TailStep    = 200
)

type LogsState = viewer.State

func NewLogsState() LogsState {
	st := viewer.NewState()
	st.Follow = true
	st.InitialLoad = true
	return st
}

func ResetLogsForContainer(s *LogsState, sessionID int, containerID string, tail int) {
	s.SessionID = sessionID
	s.ContainerID = containerID
	s.TailLines = tail
	s.InitialLoad = true
	s.Follow = true
	s.Data = nil
	s.HistoryLoad = false
	s.HistoryDone = false
	s.HistoryBaseLen = 0
	s.HistoryAppendedDuringLoad = 0
}

func ResetLogsForExit(s *LogsState, sessionID int) {
	s.SessionID = sessionID
}

func CanLoadHistory(s LogsState) bool {
	return s.Viewport.AtTop() && !s.HistoryDone && !s.HistoryLoad
}

func StartHistoryLoad(s *LogsState, nextTail int) {
	s.HistoryLoad = true
	s.HistoryDone = false
	s.HistoryBaseLen = len(s.Data)
	s.HistoryAppendedDuringLoad = 0
	if nextTail > s.TailLines {
		s.TailLines = nextTail
	}
}

func ApplyHistoryWithMerge(s *LogsState, data []string) (merged []string, done bool) {
	previousLen := len(s.Data)
	if s.HistoryBaseLen > 0 {
		previousLen = s.HistoryBaseLen
	}

	s.HistoryLoad = false
	prepended := len(data) - previousLen - s.HistoryAppendedDuringLoad
	if prepended < 0 {
		prepended = 0
	}

	if prepended == 0 {
		s.HistoryNoProgressCount++
	} else {
		s.HistoryNoProgressCount = 0
	}

	if s.HistoryNoProgressCount >= 3 {
		s.HistoryDone = true
	}

	merged = data
	s.Data = merged
	s.HistoryBaseLen = 0
	s.HistoryAppendedDuringLoad = 0

	return merged, s.HistoryDone
}

func ApplyPollWithMerge(s *LogsState, data []string) (merged []string, overlap bool) {
	previousLen := len(s.Data)
	merged, overlap = mergePolledLogs(s.Data, data, 0)
	if overlap || len(s.Data) == 0 {
		data = merged
	}
	if s.HistoryLoad {
		s.HistoryAppendedDuringLoad += max(0, len(data)-previousLen)
	}
	return data, overlap
}

func ApplyInitial(s *LogsState, data []string) {
	s.InitialLoad = false
	s.HistoryLoad = false
	s.HistoryDone = false
	s.Data = data
}

func EnterLogsState(s *LogsState, containerID string) viewer.Transition {
	nextSession := s.SessionID + 1
	ResetLogsForContainer(s, nextSession, containerID, InitialTail)
	return viewer.Transition{
		Load: &viewer.LoadRequest{
			ContainerID: containerID,
			SessionID:   s.SessionID,
			Tail:        s.TailLines,
			Src:         viewer.SourceInitial,
		},
	}
}

func ExitLogsState(s *LogsState, containersTab int) viewer.Transition {
	nextSession := s.SessionID + 1
	ResetLogsForExit(s, nextSession)
	return viewer.Transition{ExitToBrowse: true, ForceTab: containersTab}
}

func HistoryLoadRequest(s *LogsState) *viewer.LoadRequest {
	if !CanLoadHistory(*s) {
		return nil
	}
	nextTail := s.TailLines + TailStep
	StartHistoryLoad(s, nextTail)
	return &viewer.LoadRequest{
		ContainerID: s.ContainerID,
		SessionID:   s.SessionID,
		Tail:        nextTail,
		Src:         viewer.SourceHistory,
	}
}

func HandleLogsResult(s *LogsState, msg viewer.ContentMsg, w, h int) viewer.Transition {
	if msg.SessionID != s.SessionID || msg.ContainerID != s.ContainerID {
		return viewer.Transition{}
	}
	if msg.Err != nil {
		s.InitialLoad = false
		s.HistoryLoad = false
		return viewer.Transition{Err: msg.Err}
	}

	if msg.Tail > 0 && msg.Tail > s.TailLines {
		s.TailLines = msg.Tail
	}

	switch msg.Src {
	case viewer.SourceHistory:
		oldLen := len(s.Data)
		ApplyHistoryWithMerge(s, msg.Data)
		s.SyncFromData(w, h)
		newLen := len(s.Data)
		if newLen > oldLen {
			delta := newLen - oldLen
			s.Viewport.SetYOffset(s.Viewport.YOffset() + delta)
		}
	case viewer.SourceInitial:
		ApplyInitial(s, msg.Data)
		s.SyncFromData(w, h)
	default:
		merged, _ := ApplyPollWithMerge(s, msg.Data)
		s.ApplyContentForPoll(merged)
		s.SyncFromData(w, h)
	}
	return viewer.Transition{}
}

func SelectedLogsContainer(s LogsState, containers []core.ContainerRow) (core.ContainerRow, bool) {
	if s.ContainerID == "" {
		return core.ContainerRow{}, false
	}
	for _, c := range containers {
		if c.FullID == s.ContainerID {
			return c, true
		}
	}
	return core.ContainerRow{}, false
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

func mergePolledLogs(prev, polled []string, maxLines int) ([]string, bool) {
	if len(prev) == 0 {
		return trimLogs(polled, maxLines), true
	}
	if len(polled) == 0 {
		return prev, true
	}

	normPrev := make([]string, len(prev))
	for i, l := range prev {
		normPrev[i] = strings.TrimRight(l, "\r")
	}
	normPolled := make([]string, len(polled))
	for i, l := range polled {
		normPolled[i] = strings.TrimRight(l, "\r")
	}

	maxOverlap := min(len(normPrev), len(normPolled))
	for o := maxOverlap; o > 0; o-- {
		if equalLogSlices(normPrev[len(normPrev)-o:], normPolled[:o]) {
			merged := append([]string{}, normPrev...)
			merged = append(merged, normPolled[o:]...)
			return trimLogs(merged, maxLines), true
		}
	}

	if equalLogSlices(normPrev, normPolled) {
		return trimLogs(normPrev, maxLines), true
	}
	if len(normPolled) < len(normPrev) && equalLogSlices(normPrev[len(normPrev)-len(normPolled):], normPolled) {
		return trimLogs(normPrev, maxLines), true
	}

	return trimLogs(normPolled, maxLines), false
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
