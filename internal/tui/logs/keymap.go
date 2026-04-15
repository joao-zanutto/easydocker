package logs

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines logs-mode bindings and corresponding help metadata.
type KeyMap struct {
	Right        key.Binding
	Left         key.Binding
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Home         key.Binding
	End          key.Binding
	ToggleFollow key.Binding
	Back         key.Binding
	HelpNavigate key.Binding
	HelpPage     key.Binding
	HelpHomeEnd  key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "scroll right"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "scroll left"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "line up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "line down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "bottom"),
		),
		ToggleFollow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp(helpKeyLabel("f"), "toggle follow"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp(helpKeyLabel("esc"), "back"),
		),
		HelpNavigate: key.NewBinding(
			key.WithKeys("left", "up", "down", "right"),
			key.WithHelp(helpKeyLabel("← ↑ ↓ →"), "navigate"),
		),
		HelpPage: key.NewBinding(
			key.WithKeys("pgup", "pgdown"),
			key.WithHelp(helpKeyLabel("pgup/dn"), "jump up/down"),
		),
		HelpHomeEnd: key.NewBinding(
			key.WithKeys("home", "end"),
			key.WithHelp(helpKeyLabel("home/end"), "go to top/bottom"),
		),
	}
}

func helpKeyLabel(label string) string {
	return " " + label + " "
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.HelpNavigate, k.HelpPage, k.HelpHomeEnd, k.ToggleFollow, k.Back}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.HelpNavigate, k.PageUp, k.PageDown},
		{k.Home, k.End},
		{k.ToggleFollow, k.Back},
	}
}
