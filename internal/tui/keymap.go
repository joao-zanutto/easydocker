package tui

import (
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/tables"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	Quit         key.Binding
	HelpNavigate key.Binding
	HelpSwitch   key.Binding
}

var (
	defaultBrowseKeyMap = newBrowseKeyMap()
	defaultLogsKeyMap   = logs.NewKeyMap()
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
	return " " + label + " "
}

func browseKeyMap() BrowseKeyMap {
	return defaultBrowseKeyMap
}

func logsKeyMap() logs.KeyMap {
	return defaultLogsKeyMap
}

func (m model) footerKeyMap() help.KeyMap {
	if m.screen == screenModeLogs {
		if m.logs.FilterActive {
			logsKeys := logsKeyMap()
			logsFilterVerticalNavigate := key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp(helpKeyLabel("↑/↓"), "navigate"),
			)
			bindings := []key.Binding{
				logsFilterVerticalNavigate,
				logsKeys.HelpPage,
				logsKeys.HelpHomeEnd,
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
		return logsKeyMap()
	}

	browseKeys := browseKeyMap()

	// If filter mode is active, show filter-specific controls
	if m.browseFilterActive {
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
		browseKeys.HelpNavigate,
		browseKeys.HelpSwitch,
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
			bindings = append(bindings, browseKeys.OpenLogs)
		}
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
