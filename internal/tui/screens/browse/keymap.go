package browse

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	TabRight    key.Binding
	TabLeft     key.Binding
	MoveUp      key.Binding
	MoveDown    key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	ToggleScope key.Binding
	OpenLogs    key.Binding
	OpenFilter  key.Binding
	OpenShell   key.Binding
	OpenInspect key.Binding
	Quit        key.Binding
	OpenMenu    key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		TabRight: key.NewBinding(
			key.WithKeys("right"),
		),
		TabLeft: key.NewBinding(
			key.WithKeys("left"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("down"),
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
			key.WithKeys("s"),
		),
		OpenInspect: key.NewBinding(
			key.WithKeys("i"),
		),
		OpenMenu: key.NewBinding(
			key.WithKeys("esc"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
		),
	}
}
