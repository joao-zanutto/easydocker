package tui

import (
	"fmt"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"

	tea "charm.land/bubbletea/v2"
)

func (m *model) enterLogsModeIfContainerSelected() tea.Cmd {
	if m.browse.ActiveTab != shared.TabContainers {
		return nil
	}
	container, ok := m.selectedContainer()
	if !ok {
		return nil
	}
	return m.enterLogsMode(container)
}

func (m *model) openShellIfContainerSelected() tea.Cmd {
	if m.browse.ActiveTab != shared.TabContainers {
		return nil
	}
	container, ok := m.selectedContainer()
	if !ok {
		return nil
	}
	if !shared.CanOpenShell(container.State) {
		return nil
	}
	return shellCmd(m.service, container.FullID)
}

func (m *model) selectedContainer() (core.ContainerRow, bool) {
	return m.browse.SelectedContainer()
}

func (m *model) selectedImage() (core.ImageRow, bool) {
	return m.browse.SelectedImage()
}

func (m *model) selectedNetwork() (core.NetworkRow, bool) {
	return m.browse.SelectedNetwork()
}

func (m *model) selectedVolume() (core.VolumeRow, bool) {
	return m.browse.SelectedVolume()
}

func (m *model) selectedLogsContainer() (core.ContainerRow, bool) {
	return m.findContainerInSnapshot(m.viewer.ContainerID)
}

func (m *model) filteredContainers() []core.ContainerRow {
	scoped := core.FilterContainersRunningOnly(m.browse.Snapshot.Containers, m.browse.ShowAll)
	return core.FilterContainersByQuery(scoped, m.browse.Filter.Query)
}

func (m *model) findContainerInSnapshot(id string) (core.ContainerRow, bool) {
	for _, c := range m.browse.Snapshot.Containers {
		if c.FullID == id {
			return c, true
		}
	}
	return core.ContainerRow{}, false
}

func (m *model) findContainerByID(id string) (int, core.ContainerRow, bool) {
	for index, row := range m.browse.Data.ContainerListRows {
		if row.Kind != tables.RowContainer {
			continue
		}
		if row.Container.FullID == id {
			return index, row.Container, true
		}
	}
	return 0, core.ContainerRow{}, false
}

func (m *model) findContainerIndexByID(id string) (int, bool) {
	index, _, ok := m.findContainerByID(id)
	return index, ok
}

func (m *model) reconcileLogsSelection() error {
	if m.screen != shared.LogViewer {
		return nil
	}
	if _, ok := m.findContainerInSnapshot(m.viewer.ContainerID); !ok {
		m.screen = m.popScreen()
		return fmt.Errorf("selected container is no longer available")
	}
	return nil
}
