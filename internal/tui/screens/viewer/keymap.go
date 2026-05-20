package viewer

import (
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Home         key.Binding
	End          key.Binding
	ToggleWrap   key.Binding
	ToggleFollow key.Binding
	OpenFilter   key.Binding
	OpenShell    key.Binding
	Back         key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Up:           shared.UpBinding("line up"),
		Down:         shared.DownBinding("line down"),
		Left:         shared.LeftBinding("scroll left"),
		Right:        shared.RightBinding("scroll right"),
		PageUp:       shared.PageUpBinding("page up"),
		PageDown:     shared.PageDownBinding("page down"),
		Home:         shared.HomeBinding("top"),
		End:          shared.EndBinding("bottom"),
		ToggleWrap:   shared.ActionBinding("w", "toggle wrap"),
		ToggleFollow: shared.ActionBinding("f", "toggle follow"),
		OpenFilter:   shared.SlashBinding("filter"),
		OpenShell:    shared.ActionBinding("s", "shell"),
		Back:         shared.EscBinding("back"),
	}
}

func (k KeyMap) ShortHelp(resourceType ResourceType, contentType ContentType, containerState string) []key.Binding {
	bindings := []key.Binding{
		k.ToggleWrap, k.Back,
	}

	bindings = append(bindings, k.OpenFilter)

	if contentType == ContentTypeLogs {
		bindings = append(bindings, k.ToggleFollow)
	}

	if resourceType == ResourceTypeContainer && shared.CanOpenShell(containerState) {
		bindings = append(bindings, k.OpenShell)
	}

	return bindings
}
