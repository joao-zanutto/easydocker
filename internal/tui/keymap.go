package tui

import (
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// BrowseKeyMap defines browse-mode key bindings and help metadata.
type BrowseKeyMap struct {
	TabRight     key.Binding
	TabLeft      key.Binding
	MoveUp       key.Binding
	MoveDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	ToggleScope  key.Binding
	OpenLogs     key.Binding
	OpenFilter   key.Binding
	OpenShell    key.Binding
	OpenInspect  key.Binding
	Quit         key.Binding
	HelpNavigate key.Binding
	HelpSwitch   key.Binding
}

var (
	defaultBrowseKeyMap = newBrowseKeyMap()
	defaultViewerKeyMap = viewer.NewKeyMap()
)

func newBrowseKeyMap() BrowseKeyMap {
	return BrowseKeyMap{
		TabRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next tab"),
		),
		TabLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev tab"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		ToggleScope: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp(helpKeyLabel("a"), "toggle running/all"),
		),
		OpenLogs: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp(helpKeyLabel("enter"), "logs"),
		),
		OpenFilter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp(helpKeyLabel("/"), "filter"),
		),
		OpenShell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp(helpKeyLabel("s"), "shell"),
		),
		OpenInspect: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp(helpKeyLabel("i"), "inspect"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp(helpKeyLabel("esc"), "quit"),
		),
		HelpNavigate: key.NewBinding(
			key.WithKeys("up", "down"),
			key.WithHelp(helpKeyLabel("↑/↓"), "navigate"),
		),
		HelpSwitch: key.NewBinding(
			key.WithKeys("left", "right"),
			key.WithHelp(helpKeyLabel("←/→"), "switch tabs"),
		),
	}
}

func helpKeyLabel(label string) string {
	return util.HelpKeyLabel(label)
}

func browseKeyMap() BrowseKeyMap {
	return defaultBrowseKeyMap
}

func viewerKeyMap() viewer.KeyMap {
	return defaultViewerKeyMap
}

func canOpenShell(state string) bool {
	return shared.CanOpenShell(state)
}

func (m model) footerKeyMap() help.KeyMap {
	if m.screen == screenModeLogs || m.screen == screenModeInspect {
		viewerKeys := viewerKeyMap()
		contentType := viewer.ContentTypeLogs
		if m.screen == screenModeInspect {
			contentType = viewer.ContentTypeInspect
		}
		if m.logs.Filter.Active {
			logsFilterVerticalNavigate := key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp(helpKeyLabel("↑/↓"), "navigate"),
			)
			bindings := []key.Binding{
				logsFilterVerticalNavigate,
				viewerKeys.HelpPage(),
				viewerKeys.HelpHomeEnd(),
				key.NewBinding(
					key.WithKeys("esc"),
					key.WithHelp(helpKeyLabel("esc"), "clear/exit filter"),
				),
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp(helpKeyLabel("enter"), "apply/close filter"),
				),
			}
			return footerKeyMap{bindings: bindings}
		}
		return footerKeyMap{bindings: viewerKeys.ShortHelp(viewer.ResourceType(m.activeTab), contentType)}
	}

	browseKeys := browseKeyMap()

	// If filter mode is active, show filter-specific controls
	if m.browseFilter.Active {
		bindings := []key.Binding{
			browseKeys.HelpNavigate,
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp(helpKeyLabel("esc"), "clear/exit filter"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp(helpKeyLabel("enter"), "apply/close filter"),
			),
		}
		return footerKeyMap{bindings: bindings}
	}

	bindings := []key.Binding{
		browseKeys.OpenFilter,
		browseKeys.Quit,
	}
	if m.activeTab == tabContainers {
		bindings = append(bindings, browseKeys.ToggleScope)
		if row, ok := m.selectedContainerListRow(); ok && row.Kind == tables.ContainerListRowComposeProject {
			action := "expand"
			if row.ComposeExpanded {
				action = "collapse"
			}
			bindings = append(bindings, key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp(helpKeyLabel("enter"), action),
			))
		} else {
			bindings = append(bindings, browseKeys.OpenLogs, browseKeys.OpenInspect)
			if container, ok := m.selectedContainer(); ok && canOpenShell(container.State) {
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
