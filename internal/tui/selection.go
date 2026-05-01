package tui

import (
	"fmt"

	"easydocker/internal/core"
	tuistate "easydocker/internal/tui/state"
	"easydocker/internal/tui/tables"

	tea "charm.land/bubbletea/v2"
)

func (m *model) moveActiveTab(delta int) {
	previous := m.activeTab
	m.activeTab = tuistate.MoveActiveTab(m.activeTab, delta, tabContainers, tabVolumes)
	if m.activeTab != previous {
		m.clearBrowseFilter()
	}
}

func (m *model) clearBrowseFilter() {
	m.browseFilterQuery = ""
	m.browseFilterInput.SetValue("")
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

func (m *model) toggleSelectedComposeProject() bool {
	if m.activeTab != tabContainers {
		return false
	}
	row, ok := m.selectedContainerListRow()
	if !ok || row.Kind != tables.ContainerListRowComposeProject {
		return false
	}
	m.composeExpanded[row.ComposeProject.Name] = !m.composeExpanded[row.ComposeProject.Name]
	m.clampCursors()
	return true
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
		return len(m.containerListRows())
	case tabImages:
		return len(m.filteredImages())
	case tabNetworks:
		return len(m.filteredNetworks())
	case tabVolumes:
		return len(m.filteredVolumes())
	default:
		return 0
	}
}

func (m model) filteredContainers() []core.ContainerRow {
	scoped := core.FilterContainersByScope(m.snapshot.Containers, m.showAll)
	return core.FilterContainersByQuery(scoped, m.browseFilterQuery)
}

func (m model) filteredImages() []core.ImageRow {
	return core.FilterImagesByQuery(m.snapshot.Images, m.browseFilterQuery)
}

func (m model) filteredNetworks() []core.NetworkRow {
	return core.FilterNetworksByQuery(m.snapshot.Networks, m.browseFilterQuery)
}

func (m model) filteredVolumes() []core.VolumeRow {
	return core.FilterVolumesByQuery(m.snapshot.Volumes, m.browseFilterQuery)
}

func (m model) findContainerIndexByID(id string) (int, bool) {
	for index, row := range m.containerListRows() {
		if row.Kind != tables.ContainerListRowContainer {
			continue
		}
		if row.Container.FullID == id {
			return index, true
		}
	}
	return 0, false
}

func (m model) selectedContainer() (core.ContainerRow, bool) {
	row, ok := selectedAt(m.containerListRows(), m.containerCursor)
	if !ok || row.Kind != tables.ContainerListRowContainer {
		return core.ContainerRow{}, false
	}
	return row.Container, true
}

func (m model) selectedContainerListRow() (tables.ContainerListRow, bool) {
	return selectedAt(m.containerListRows(), m.containerCursor)
}

func (m model) selectedComposeProject() (core.ComposeProject, bool) {
	row, ok := m.selectedContainerListRow()
	if !ok || row.Kind != tables.ContainerListRowComposeProject {
		return core.ComposeProject{}, false
	}
	return row.ComposeProject, true
}

func (m model) containerListRows() []tables.ContainerListRow {
	return tables.BuildContainerListRows(m.filteredContainers(), m.composeExpanded)
}

func (m model) selectedLogsContainer() (core.ContainerRow, bool) {
	return logsController.SelectedContainer(m.logs, m.snapshot.Containers)
}

func (m model) selectedImage() (core.ImageRow, bool) {
	return selectedAt(m.filteredImages(), m.imageCursor)
}

func (m model) selectedNetwork() (core.NetworkRow, bool) {
	return selectedAt(m.filteredNetworks(), m.networkCursor)
}

func (m model) selectedVolume() (core.VolumeRow, bool) {
	return selectedAt(m.filteredVolumes(), m.volumeCursor)
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
