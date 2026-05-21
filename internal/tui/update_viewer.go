package tui

import (
	"easydocker/internal/tui/screens/viewer"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m *model) handleToggleWrap(visibleWidth, visibleRows int) {
	logList := viewer.FilterLines(m.logs.Data, m.logs.Filter.Query)
	startLine, _ := viewer.VisibleContentRange(&m.logs.State, logList)
	m.logs.SetWrapLines(!m.logs.WrapLines)
	m.logs.SyncFromData(visibleWidth, visibleRows)
	if !m.logs.Follow {
		targetYOffset := startLine
		if m.logs.WrapLines {
			targetYOffset = viewer.RawLineToViewportRowOffset(logList, visibleWidth, startLine)
		}
		m.logs.Viewport.SetYOffset(targetYOffset)
	}
}

type viewerKeyHandlers struct {
	visibleWidth   func() int
	visibleRows    func() int
	openFilter     func()
	closeFilter    func(bool)
	exitMode       func()
	historyLoad    func() *viewer.LoadRequest
	postTransition func()
}

func (m *model) handleViewerKey(msg tea.KeyPressMsg, h viewerKeyHandlers) tea.Cmd {
	keys := viewer.NewKeyMap()

	if m.logs.Filter.Active {
		switch {
		case key.Matches(msg, keys.Back):
			h.closeFilter(true)
			return nil
		case msg.String() == "enter":
			h.closeFilter(false)
			return nil
		case key.Matches(msg, keys.Up),
			key.Matches(msg, keys.Down),
			key.Matches(msg, keys.PageUp),
			key.Matches(msg, keys.PageDown),
			key.Matches(msg, keys.Home),
			key.Matches(msg, keys.End):
			transition := viewer.Controller{}.HandleKey(&m.logs.State, msg, keys)
			if h.historyLoad != nil {
				if historyReq := h.historyLoad(); historyReq != nil {
					transition.Load = historyReq
				}
			}
			return m.applyLogsTransition(viewer.Transition{Load: transition.Load})
		default:
			return m.updateViewerFilterInput(h.visibleWidth(), h.visibleRows(), msg)
		}
	}

	if key.Matches(msg, keys.OpenFilter) {
		h.openFilter()
		return nil
	}

	if key.Matches(msg, keys.Back) {
		h.exitMode()
		return nil
	}

	if key.Matches(msg, keys.ToggleWrap) {
		m.handleToggleWrap(h.visibleWidth(), h.visibleRows())
		return nil
	}

	transition := viewer.Controller{}.HandleKey(&m.logs.State, msg, keys)
	if h.postTransition != nil {
		h.postTransition()
	}
	if h.historyLoad != nil {
		if historyReq := h.historyLoad(); historyReq != nil {
			transition.Load = historyReq
		}
	}
	return m.applyLogsTransition(transition)
}

func (m *model) updateViewerFilterInput(visibleWidth, visibleRows int, msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	m.logs.Filter.Input, cmd = m.logs.Filter.Input.Update(msg)
	m.logs.Filter.Query = m.logs.Filter.Input.Value()
	m.logs.SyncFromData(visibleWidth, visibleRows)
	return cmd
}

func (m *model) openViewerFilter(visibleWidth func() int, visibleRows func() int) {
	previousRows := visibleRows()
	previousYOffset := m.logs.Viewport.YOffset()
	m.logs.OpenFilter()
	newRows := visibleRows()
	m.logs.SyncFromData(visibleWidth(), newRows)
	if !m.logs.Follow {
		targetYOffset := previousYOffset + (previousRows - newRows)
		m.logs.Viewport.SetYOffset(targetYOffset)
	}
}

func (m *model) closeViewerFilter(clear bool, visibleWidth func() int, visibleRows func() int) {
	previousRows := visibleRows()
	previousYOffset := m.logs.Viewport.YOffset()
	m.logs.CloseFilter(clear)
	newRows := visibleRows()
	m.logs.SyncFromData(visibleWidth(), newRows)
	if !m.logs.Follow && newRows > previousRows {
		m.logs.Viewport.SetYOffset(max(0, previousYOffset-(newRows-previousRows)))
	}
}

func (m *model) openLogsFilter() {
	m.openViewerFilter(func() int { return m.logVisibleWidth() }, func() int { return m.logVisibleRows() })
}

func (m *model) closeLogsFilter(clear bool) {
	m.closeViewerFilter(clear, func() int { return m.logVisibleWidth() }, func() int { return m.logVisibleRows() })
}

func (m *model) openInspectFilter() {
	m.openViewerFilter(func() int { return m.inspectVisibleWidth() }, func() int { return m.inspectVisibleRows() })
}

func (m *model) closeInspectFilter(clear bool) {
	m.closeViewerFilter(clear, func() int { return m.inspectVisibleWidth() }, func() int { return m.inspectVisibleRows() })
}
