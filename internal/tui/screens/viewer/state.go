package viewer

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

func (s *State) ResetForContainer(sessionID int, containerID string) {
	s.SessionID = sessionID
	s.ContainerID = containerID
	s.Data = nil
	s.Filter.Active = false
	s.Filter.Query = ""
	s.Filter.Input.SetValue("")
}

func (s *State) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	s.Viewport.SetWidth(visibleWidth)
	s.Viewport.SetHeight(visibleRows)
	s.Viewport.SetContent(joinLines(lines))
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

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
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
