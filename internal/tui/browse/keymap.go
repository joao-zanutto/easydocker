package browse

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
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
	Quit        key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		TabRight: key.NewBinding(
			key.WithKeys("right", "l"),
		),
		TabLeft: key.NewBinding(
			key.WithKeys("left", "h"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
		),
		ToggleScope: key.NewBinding(
			key.WithKeys("a"),
		),
		OpenLogs: key.NewBinding(
			key.WithKeys("enter"),
		),
		OpenFilter: key.NewBinding(
			key.WithKeys("/"),
		),
		OpenShell: key.NewBinding(
			key.WithKeys("t"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
		),
	}
}