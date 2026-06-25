package tui

import (
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

func (m model) footerKeyMap() help.KeyMap {
	if m.screen == shared.LogViewer || m.screen == shared.InspectViewer {
		if m.viewer.Vp.ContentType == viewer.ContentTypeConfig {
			return footerKeyMap{bindings: []key.Binding{shared.EscBinding("back")}}
		}
		viewerKeys := viewer.NewKeyMap()
		if m.viewer.Vp.Filter.Active {
			bindings := []key.Binding{
				shared.EscBinding("clear filter"),
				shared.EnterBinding("apply filter"),
			}
			return footerKeyMap{bindings: bindings}
		}
		var containerState core.ContainerState
		if m.browse.ActiveTab == shared.TabContainers {
			if c, ok := m.selectedContainer(); ok {
				containerState = c.State
			}
		}
		return footerKeyMap{bindings: viewerKeys.ShortHelp(shared.TabToResourceType(m.browse.ActiveTab), containerState)}
	}

	browseKeys := browse.NewKeyMap()

	if m.browse.Filter.Active {
		bindings := []key.Binding{
			shared.EscBinding("clear filter"),
			shared.EnterBinding("apply filter"),
		}
		return footerKeyMap{bindings: bindings}
	}

	bindings := []key.Binding{}
	if m.browse.ActiveTab == shared.TabContainers {
		rows := m.browse.Data.ContainerListRows
		var row tables.ContainerListRow
		if len(rows) > 0 && m.browse.Cursors.Container >= 0 && m.browse.Cursors.Container < len(rows) {
			row = rows[m.browse.Cursors.Container]
		}
		if row.Kind == tables.RowComposeProject {
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
	if m.browse.ActiveTab == shared.TabImages || m.browse.ActiveTab == shared.TabNetworks || m.browse.ActiveTab == shared.TabVolumes {
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

func keyLabel(binding key.Binding) string {
	return strings.TrimSpace(binding.Help().Key)
}

func joinKeyLabels(sep string, bindings ...key.Binding) string {
	labels := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		label := keyLabel(binding)
		if label == "" {
			continue
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, sep)
}

func (m *model) buildHelpCommands() []menu.HelpCommand {
	browseKeys := browse.NewKeyMap()
	viewerKeys := viewer.NewKeyMap()
	menuKeys := menu.NewKeyMap()

	return menu.BuildHelpCommands(menu.HelpKeyLabels{
		MoveUpDown:   joinKeyLabels("/", browseKeys.MoveUp, browseKeys.MoveDown),
		PageUpDown:   joinKeyLabels("/", viewerKeys.PageUp, viewerKeys.PageDown),
		HomeEnd:      joinKeyLabels("/", viewerKeys.Home, viewerKeys.End),
		TabLeftRight: joinKeyLabels("/", browseKeys.TabLeft, browseKeys.TabRight),
		LeftRight:    joinKeyLabels("/", viewerKeys.Left, viewerKeys.Right),
		OpenLogs:     keyLabel(browseKeys.OpenLogs),
		OpenFilter:   keyLabel(browseKeys.OpenFilter),
		ToggleScope:  keyLabel(browseKeys.ToggleScope),
		OpenInspect:  keyLabel(browseKeys.OpenInspect),
		OpenShell:    keyLabel(browseKeys.OpenShell),
		ToggleWrap:   keyLabel(viewerKeys.ToggleWrap),
		ToggleFollow: keyLabel(viewerKeys.ToggleFollow),
		BackClose:    keyLabel(menuKeys.Back),
	})
}
