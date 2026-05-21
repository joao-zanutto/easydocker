package shared

import "easydocker/internal/tui/util"

// Cursors holds per-tab cursor positions.
type Cursors struct {
	Container int
	Image     int
	Network   int
	Volume    int
}

func (c Cursors) byTab(tab Tab) (int, bool) {
	switch tab {
	case TabContainers:
		return c.Container, true
	case TabImages:
		return c.Image, true
	case TabNetworks:
		return c.Network, true
	case TabVolumes:
		return c.Volume, true
	default:
		return 0, false
	}
}

func (c *Cursors) setByTab(tab Tab, value int) bool {
	switch tab {
	case TabContainers:
		c.Container = value
		return true
	case TabImages:
		c.Image = value
		return true
	case TabNetworks:
		c.Network = value
		return true
	case TabVolumes:
		c.Volume = value
		return true
	default:
		return false
	}
}

// SelectionState groups browse tab/scope/cursor state.
type SelectionState struct {
	ActiveTab Tab
	ShowAll   bool
	Cursors   Cursors
}

// MoveActiveTab shifts active tab and clamps to [minTab, maxTab].
func MoveActiveTab(current Tab, delta int, minTab, maxTab Tab) Tab {
	return Tab(util.Clamp(int(current)+delta, int(minTab), int(maxTab)))
}

// ToggleContainerScope flips showAll only when the containers tab is active.
// Returns the possibly-updated showAll value and whether a toggle occurred.
func ToggleContainerScope(activeTab Tab, showAll bool) (bool, bool) {
	if activeTab != TabContainers {
		return showAll, false
	}
	return !showAll, true
}

// CursorForTab returns the cursor value for a tab.
func CursorForTab(c Cursors, tab Tab) (int, bool) {
	return c.byTab(tab)
}

// SetCursorForTab sets the cursor value for a tab and reports whether tab exists.
func SetCursorForTab(c *Cursors, tab Tab, value int) bool {
	return c.setByTab(tab, value)
}

// MoveCursorForTab moves a tab cursor by delta and clamps to [0, max(0,itemCount-1)].
func MoveCursorForTab(c *Cursors, tab Tab, delta, itemCount int) bool {
	cursor, ok := CursorForTab(*c, tab)
	if !ok {
		return false
	}
	upper := max(0, itemCount-1)
	return SetCursorForTab(c, tab, util.Clamp(cursor+delta, 0, upper))
}

// ClampCursorForTab clamps one tab cursor to [0, max(0,itemCount-1)].
func ClampCursorForTab(c *Cursors, tab Tab, itemCount int) bool {
	cursor, ok := CursorForTab(*c, tab)
	if !ok {
		return false
	}
	upper := max(0, itemCount-1)
	return SetCursorForTab(c, tab, util.Clamp(cursor, 0, upper))
}

// ClampAllCursors clamps cursors for all tabs using a caller-provided item counter.
func ClampAllCursors(c *Cursors, tabs []Tab, itemCountForTab func(tab Tab) int) {
	for _, tab := range tabs {
		_ = ClampCursorForTab(c, tab, itemCountForTab(tab))
	}
}

// ReconcileCursorForTab updates a tab cursor to index when found is true.
func ReconcileCursorForTab(c *Cursors, tab Tab, index int, found bool) bool {
	if !found {
		return false
	}
	return SetCursorForTab(c, tab, index)
}
