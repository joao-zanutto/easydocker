package tui

import (
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

func browseKeyMap() browse.KeyMap {
	return browse.NewKeyMap()
}

func viewerKeyMap() viewer.KeyMap {
	return viewer.NewKeyMap()
}

func (m model) footerKeyMap() help.KeyMap {
	if m.screen == shared.Logs || m.screen == shared.Inspect {
		viewerKeys := viewerKeyMap()
		contentType := viewer.ContentTypeLogs
		if m.screen == shared.Inspect {
			contentType = viewer.ContentTypeInspect
		}
		if m.logs.Filter.Active {
			bindings := []key.Binding{
				shared.EscBinding("clear/exit filter"),
				shared.EnterBinding("apply/close filter"),
			}
			return footerKeyMap{bindings: bindings}
		}
		containerState := ""
		if m.activeTab == tabContainers {
			if c, ok := m.selectedContainer(); ok {
				containerState = c.State
			}
		}
		return footerKeyMap{bindings: viewerKeys.ShortHelp(resourceTypeFromTab(m.activeTab), contentType, containerState)}
	}

	browseKeys := browseKeyMap()

	if m.browseFilter.Active {
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
	if m.activeTab == tabContainers {
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
	if m.activeTab == tabImages || m.activeTab == tabNetworks || m.activeTab == tabVolumes {
		bindings = append(bindings, browseKeys.OpenInspect)
	}
	return footerKeyMap{bindings: bindings}
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
