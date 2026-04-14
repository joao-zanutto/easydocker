package tui

import (
	"easydocker/internal/tui/logs"

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
		Quit: key.NewBinding(
			key.WithKeys("esc", "backspace"),
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
		return logsKeyMap()
	}

	browseKeys := browseKeyMap()
	bindings := []key.Binding{
		browseKeys.HelpNavigate,
		browseKeys.HelpSwitch,
		browseKeys.Quit,
	}
	if m.activeTab == tabContainers {
		bindings = append(bindings, browseKeys.ToggleScope, browseKeys.OpenLogs)
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
