package tui

import (
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

func (m model) footerKeyMap() help.KeyMap {
	if m.screen == shared.Logs || m.screen == shared.Inspect {
		viewerKeys := viewer.NewKeyMap()
		contentType := viewer.ContentTypeLogs
		if m.screen == shared.Inspect {
			contentType = viewer.ContentTypeInspect
		}
		if m.viewer.Filter.Active {
			bindings := []key.Binding{
				shared.EscBinding("clear/exit filter"),
				shared.EnterBinding("apply/close filter"),
			}
			return footerKeyMap{bindings: bindings}
		}
		containerState := ""
		if m.browse.ActiveTab == tabContainers {
			if c, ok := m.selectedContainer(); ok {
				containerState = c.State
			}
		}
		return footerKeyMap{bindings: viewerKeys.ShortHelp(resourceTypeFromTab(m.browse.ActiveTab), contentType, containerState)}
	}

	browseKeys := browse.NewKeyMap()

	if m.browse.Filter.Active {
		bindings := []key.Binding{
			shared.EscBinding("clear/exit filter"),
			shared.EnterBinding("apply/close filter"),
		}
		return footerKeyMap{bindings: bindings}
	}

	bindings := []key.Binding{
		browseKeys.OpenFilter,
		browseKeys.OpenMenu,
	}
	if m.browse.ActiveTab == tabContainers {
		bindings = append(bindings, browseKeys.ToggleScope)
		if row, ok := m.selectedContainerListRow(); ok && row.Kind == tables.ContainerListRowComposeProject {
			action := "expand"
			if row.ComposeExpanded {
				action = "collapse"
			}
			bindings = append(bindings, shared.EnterBinding(action))
		} else {
			bindings = append(bindings, browseKeys.OpenLogs, browseKeys.OpenInspect)
			if container, ok := m.selectedContainer(); ok && shared.CanOpenShell(container.State) {
				bindings = append(bindings, browseKeys.OpenShell)
			}
		}
	}
	if m.browse.ActiveTab == tabImages || m.browse.ActiveTab == tabNetworks || m.browse.ActiveTab == tabVolumes {
		bindings = append(bindings, browseKeys.OpenInspect)
	}
	return footerKeyMap{bindings: bindings}
}

func (m model) selectedContainerListRow() (tables.ContainerListRow, bool) {
	rows := m.containerListRows()
	var zero tables.ContainerListRow
	if len(rows) == 0 || m.browse.ContainerCursor < 0 || m.browse.ContainerCursor >= len(rows) {
		return zero, false
	}
	return rows[m.browse.ContainerCursor], true
}

type footerKeyMap struct {
	bindings []key.Binding
}

func (m footerKeyMap) ShortHelp() []key.Binding {
	return m.bindings
}

func (m footerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.bindings}
}

func resourceTypeFromTab(tab shared.Tab) viewer.ResourceType {
	switch tab {
	case tabContainers:
		return viewer.ResourceTypeContainer
	case tabImages:
		return viewer.ResourceTypeImage
	case tabNetworks:
		return viewer.ResourceTypeNetwork
	case tabVolumes:
		return viewer.ResourceTypeVolume
	default:
		return viewer.ResourceTypeContainer
	}
}
