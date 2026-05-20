package viewer

import "strings"

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

func (s *State) OpenFilter() {
	s.Filter.Active = true
	s.Filter.Input.Focus()
	s.Filter.Input.SetValue(s.Filter.Query)
}

func (s *State) CloseFilter(clear bool) {
	s.Filter.Active = false
	s.Filter.Input.Blur()
	if clear {
		s.Filter.Query = ""
		s.Filter.Input.SetValue("")
	}
}

func (s *State) ResetForContainer(sessionID int, containerID string, tail int) {
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
	s.HistoryNoProgressCount = 0
	s.HorizontalOffset = 0
	s.WrapXOffset = 0
	s.Viewport.SetXOffset(0)
	s.Filter.Active = false
	s.Filter.Query = ""
	s.Filter.Input.SetValue("")
}

func (s *State) ResetForExit(sessionID int) {
	s.SessionID = sessionID
}

func (s *State) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	s.Viewport.SetWidth(visibleWidth)
	s.Viewport.SetHeight(visibleRows)
	s.Viewport.SetContent(strings.Join(lines, "\n"))
	if s.Follow {
		s.Viewport.GotoBottom()
	}
}

func (s *State) SyncFromData(visibleWidth, visibleRows int) {
	s.InitialLoad = false

	if s.Data == nil {
		s.SyncViewport(nil, visibleWidth, visibleRows)
		return
	}

	lines := FilterLines(s.Data, s.Filter.Query)

	sanitized := make([]string, 0, len(lines))
	for _, line := range lines {
		sanitized = append(sanitized, SanitizeLine(line))
	}

	if s.WrapLines {
		sanitized = WrapLines(sanitized, visibleWidth)
	}

	s.SyncViewport(sanitized, visibleWidth, visibleRows)
}

func (s *State) ApplyContentForPoll(data []string) {
	s.Data = data
	s.InitialLoad = false
}

func (s *State) ApplyContentInitial(data []string) {
	s.Data = data
	s.InitialLoad = false
	s.HistoryLoad = false
	s.HistoryDone = false
}
