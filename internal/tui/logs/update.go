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

func (s *State) SetWrapLines(enabled bool) {
	if s.WrapLines == enabled {
		return
	}
	if enabled {
		s.WrapXOffset = s.HorizontalOffset
		s.Viewport.SetXOffset(0)
	} else {
		s.Viewport.SetXOffset(s.WrapXOffset)
	}
	s.WrapLines = enabled
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
	s.HistoryBaseLen = len(s.Data.Logs)
	s.HistoryAppendedDuringLoad = 0
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
	if s.HistoryBaseLen > 0 {
		previousLen = s.HistoryBaseLen
	}
	s.HistoryLoad = false
	prepended := len(data.Logs) - previousLen - s.HistoryAppendedDuringLoad
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
	prependViewportDelta := renderedViewportLineDelta(s, data.Logs, prepended)
	s.Data = data
	s.HistoryBaseLen = 0
	s.HistoryAppendedDuringLoad = 0
	if !s.Follow {
		s.Viewport.SetYOffset(previousYOffset + prependViewportDelta)
	}
}

func (s *State) ApplyPoll(data core.ContainerLiveData, previousYOffset int) {
	previousLen := len(s.Data.Logs)
	mergedLogs, overlapFound := MergePolledLogs(s.Data.Logs, data.Logs, MaxLiveLines)
	if overlapFound || len(s.Data.Logs) == 0 {
		data.Logs = mergedLogs
	} else {
		data.Logs = TrimLogs(data.Logs, MaxLiveLines)
	}
	if s.HistoryLoad {
		s.HistoryAppendedDuringLoad += max(0, len(data.Logs)-previousLen)
	}
	s.Data = data
	s.InitialLoad = false
	if !s.Follow {
		s.Viewport.SetYOffset(previousYOffset)
	}
}

func (s *State) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	s.Viewport.SetWidth(visibleWidth)
	s.Viewport.SetHeight(visibleRows)
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
	if s.WrapLines {
		lines = WrapLogLines(lines, visibleWidth)
	}
	s.SyncViewport(lines, visibleWidth, visibleRows)
}

func renderedViewportLineDelta(state *State, allLines []string, prepended int) int {
	if prepended <= 0 || len(allLines) == 0 {
		return 0
	}
	if prepended > len(allLines) {
		prepended = len(allLines)
	}

	added := FilterLogLines(allLines[:prepended], state.FilterQuery)
	if !state.WrapLines {
		return len(added)
	}

	width := max(1, state.Viewport.Width())
	delta := 0
	for _, line := range added {
		delta += wrappedRowCount(SanitizeLogRenderLine(line), width)
	}
	return delta
}
