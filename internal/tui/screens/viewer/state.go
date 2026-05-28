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
		s.Viewport.SetXOffset(0)
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

func (s *State) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	s.Viewport.SetWidth(visibleWidth)
	s.Viewport.SetHeight(visibleRows)
	s.Viewport.SetContent(strings.Join(lines, "\n"))
	if s.Follow {
		s.Viewport.GotoBottom()
	}
}

func PrepareContentLines(data []string, query string, wrapWidth int, wrapEnabled bool) []string {
	lines := FilterLines(data, query)

	sanitized := make([]string, 0, len(lines))
	for _, line := range lines {
		sanitized = append(sanitized, SanitizeLine(line))
	}

	if wrapEnabled && wrapWidth > 0 {
		sanitized = WrapLines(sanitized, wrapWidth)
	}

	return sanitized
}

func (s *State) SyncFromData(visibleWidth, visibleRows int) {
	s.InitialLoad = false

	if s.Data == nil {
		s.SyncViewport(nil, visibleWidth, visibleRows)
		return
	}

	lines := PrepareContentLines(s.Data, s.Filter.Query, visibleWidth, s.WrapLines)
	s.SyncViewport(lines, visibleWidth, visibleRows)
}
