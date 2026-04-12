package tui

import (
	"fmt"

	"easydocker/internal/core"
	tuistate "easydocker/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) moveActiveTab(delta int) {
	m.activeTab = tuistate.MoveActiveTab(m.activeTab, delta, tabContainers, tabVolumes)
}

func (m *model) toggleContainerScope() {
	showAll, toggled := tuistate.ToggleContainerScope(m.activeTab, tabContainers, m.showAll)
	if !toggled {
		return
	}
	m.showAll = showAll
	m.clampCursors()
}

func (m *model) enterLogsModeIfContainerSelected() tea.Cmd {
	if m.activeTab != tabContainers {
		return nil
	}
	container, ok := m.selectedContainer()
	if !ok {
		return nil
	}
	return m.enterLogsMode(container)
}

func (m *model) moveCursor(delta int) {
	sel := m.selectionState()
	_ = tuistate.MoveCursorForTab(&sel.Cursors, sel.ActiveTab, delta, m.itemCountForTab(sel.ActiveTab))
	m.applySelectionState(sel)
}

func (m *model) clampCursors() {
	sel := m.selectionState()
	tuistate.ClampAllCursors(&sel.Cursors, []int{tabContainers, tabImages, tabNetworks, tabVolumes}, m.itemCountForTab)
	m.applySelectionState(sel)
}

func (m model) itemCountForTab(tab int) int {
	switch tab {
	case tabContainers:
		return len(m.filteredContainers())
	case tabImages:
		return len(m.snapshot.Images)
	case tabNetworks:
		return len(m.snapshot.Networks)
	case tabVolumes:
		return len(m.snapshot.Volumes)
	default:
		return 0
	}
}

func (m model) filteredContainers() []core.ContainerRow {
	return core.FilterContainersByScope(m.snapshot.Containers, m.showAll)
}

func (m model) findContainerIndexByID(id string) (int, bool) {
	for index, container := range m.filteredContainers() {
		if container.FullID == id {
			return index, true
		}
	}
	return 0, false
}

func (m model) selectedContainer() (core.ContainerRow, bool) {
	return selectedAt(m.filteredContainers(), m.containerCursor)
}

func (m model) selectedLogsContainer() (core.ContainerRow, bool) {
	return logsController.SelectedContainer(m.logs, m.snapshot.Containers)
}

func (m model) selectedImage() (core.ImageRow, bool) {
	return selectedAt(m.snapshot.Images, m.imageCursor)
}

func (m model) selectedNetwork() (core.NetworkRow, bool) {
	return selectedAt(m.snapshot.Networks, m.networkCursor)
}

func (m model) selectedVolume() (core.VolumeRow, bool) {
	return selectedAt(m.snapshot.Volumes, m.volumeCursor)
}

func selectedAt[T any](items []T, cursor int) (T, bool) {
	var zero T
	if len(items) == 0 || cursor < 0 || cursor >= len(items) {
		return zero, false
	}
	return items[cursor], true
}

func (m *model) reconcileLogsSelection() error {
	if m.screen != screenModeLogs {
		return nil
	}
	if index, ok := m.findContainerIndexByID(m.logs.ContainerID); ok {
		sel := m.selectionState()
		_ = tuistate.ReconcileCursorForTab(&sel.Cursors, tabContainers, index, ok)
		m.applySelectionState(sel)
		return nil
	}
	m.exitLogsMode()
	return fmt.Errorf("selected container is no longer available")
}

func (m model) selectionState() tuistate.SelectionState {
	return tuistate.SelectionState{
		ActiveTab: m.activeTab,
		ShowAll:   m.showAll,
		Cursors: tuistate.Cursors{
			Container: m.containerCursor,
			Image:     m.imageCursor,
			Network:   m.networkCursor,
			Volume:    m.volumeCursor,
		},
	}
}

func (m *model) applySelectionState(sel tuistate.SelectionState) {
	m.activeTab = sel.ActiveTab
	m.showAll = sel.ShowAll
	m.containerCursor = sel.Cursors.Container
	m.imageCursor = sel.Cursors.Image
	m.networkCursor = sel.Cursors.Network
	m.volumeCursor = sel.Cursors.Volume
}
