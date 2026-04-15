package logs

import (
	"strings"

	"easydocker/internal/core"
)

func (s *State) SetFollow(enabled bool) {
	s.Follow = enabled
	if enabled {
		s.Viewport.GotoBottom()
	}
}

func (s *State) ResetForContainer(sessionID int, containerID string, tail int) {
	*s = NewState()
	s.SessionID = sessionID
	s.ContainerID = containerID
	s.TailLines = tail
	s.InitialLoad = true
}

func (s *State) ResetForExit(sessionID int) {
	*s = NewState()
	s.SessionID = sessionID
}

func (s *State) CanLoadHistory() bool {
	return s.Viewport.AtTop() && !s.HistoryDone && !s.HistoryLoad
}

func (s *State) StartHistoryLoad(nextTail int) {
	s.HistoryLoad = true
	s.HistoryDone = false
	if nextTail > s.TailLines {
		s.TailLines = nextTail
	}
}

func (s *State) ApplyInitial(data core.ContainerLiveData) {
	s.InitialLoad = false
	s.HistoryLoad = false
	s.HistoryDone = false
	s.Data = data
}

func (s *State) ApplyHistory(data core.ContainerLiveData, previousYOffset int) {
	previousLen := len(s.Data.Logs)
	s.HistoryLoad = false
	if len(data.Logs) <= previousLen {
		s.HistoryDone = true
	}
	delta := len(data.Logs) - previousLen
	s.Data = data
	if !s.Follow {
		s.Viewport.SetYOffset(previousYOffset + max(0, delta))
	}
}

func (s *State) ApplyPoll(data core.ContainerLiveData, previousYOffset int) {
	mergedLogs, overlapFound := MergePolledLogs(s.Data.Logs, data.Logs, MaxLiveLines)
	if overlapFound || len(s.Data.Logs) == 0 {
		data.Logs = mergedLogs
	} else {
		data.Logs = TrimLogs(data.Logs, MaxLiveLines)
	}
	s.Data = data
	s.InitialLoad = false
	if !s.Follow {
		s.Viewport.SetYOffset(previousYOffset)
	}
}

func (s *State) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	s.Viewport.Width = visibleWidth
	s.Viewport.Height = visibleRows
	s.Viewport.SetContent(strings.Join(lines, "\n"))
	if s.Follow {
		s.Viewport.GotoBottom()
	}
}

func (s *State) SyncViewportFromData(visibleWidth, visibleRows int) {
	logLines := FilterLogLines(s.Data.Logs, s.FilterQuery)
	lines := make([]string, 0, len(logLines))
	for _, line := range logLines {
		lines = append(lines, SanitizeLogRenderLine(line))
	}
	s.SyncViewport(lines, visibleWidth, visibleRows)
}
