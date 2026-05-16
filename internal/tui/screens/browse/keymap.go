package browse

import (
	"easydocker/internal/tui/shared"

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
	OpenMenu    key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		TabRight: shared.RightBinding("next tab"),
		TabLeft:  shared.LeftBinding("prev tab"),
		MoveUp:   shared.UpBinding("move up"),
		MoveDown: shared.DownBinding("move down"),
		PageUp:   shared.PageUpBinding("page up"),
		PageDown: shared.PageDownBinding("page down"),
		ToggleScope: shared.ActionBinding("a", "toggle running/all"),
		OpenLogs:    shared.EnterBinding("logs"),
		OpenFilter:  shared.SlashBinding("filter"),
		OpenShell:   shared.ActionBinding("s", "shell"),
		OpenInspect: shared.ActionBinding("i", "inspect"),
		OpenMenu:    shared.EscBinding("menu"),
	}
}
