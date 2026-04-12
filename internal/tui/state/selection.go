package state

import "easydocker/internal/tui/util"

// Cursors holds per-tab cursor positions.
type Cursors struct {
	Container int
	Image     int
	Network   int
	Volume    int
}

// SelectionState groups browse tab/scope/cursor state.
type SelectionState struct {
	ActiveTab int
	ShowAll   bool
	Cursors   Cursors
}

// MoveActiveTab shifts active tab and clamps to [minTab, maxTab].
func MoveActiveTab(current, delta, minTab, maxTab int) int {
	return util.Clamp(current+delta, minTab, maxTab)
}

// ToggleContainerScope flips showAll only when the containers tab is active.
// Returns the possibly-updated showAll value and whether a toggle occurred.
func ToggleContainerScope(activeTab, containersTab int, showAll bool) (bool, bool) {
	if activeTab != containersTab {
		return showAll, false
	}
	return !showAll, true
}

// CursorForTab returns the cursor value for a tab.
func CursorForTab(c Cursors, tab int) (int, bool) {
	switch tab {
	case 0:
		return c.Container, true
	case 1:
		return c.Image, true
	case 2:
		return c.Network, true
	case 3:
		return c.Volume, true
	default:
		return 0, false
	}
}

// SetCursorForTab sets the cursor value for a tab and reports whether tab exists.
func SetCursorForTab(c *Cursors, tab, value int) bool {
	switch tab {
	case 0:
		c.Container = value
		return true
	case 1:
		c.Image = value
		return true
	case 2:
		c.Network = value
		return true
	case 3:
		c.Volume = value
		return true
	default:
		return false
	}
}

// MoveCursorForTab moves a tab cursor by delta and clamps to [0, max(0,itemCount-1)].
func MoveCursorForTab(c *Cursors, tab, delta, itemCount int) bool {
	cursor, ok := CursorForTab(*c, tab)
	if !ok {
		return false
	}
	upper := max(0, itemCount-1)
	return SetCursorForTab(c, tab, util.Clamp(cursor+delta, 0, upper))
}

// ClampCursorForTab clamps one tab cursor to [0, max(0,itemCount-1)].
func ClampCursorForTab(c *Cursors, tab, itemCount int) bool {
	cursor, ok := CursorForTab(*c, tab)
	if !ok {
		return false
	}
	upper := max(0, itemCount-1)
	return SetCursorForTab(c, tab, util.Clamp(cursor, 0, upper))
}

// ClampAllCursors clamps cursors for all tabs using a caller-provided item counter.
func ClampAllCursors(c *Cursors, tabs []int, itemCountForTab func(tab int) int) {
	for _, tab := range tabs {
		_ = ClampCursorForTab(c, tab, itemCountForTab(tab))
	}
}

// ReconcileCursorForTab updates a tab cursor to index when found is true.
func ReconcileCursorForTab(c *Cursors, tab, index int, found bool) bool {
	if !found {
		return false
	}
	return SetCursorForTab(c, tab, index)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
