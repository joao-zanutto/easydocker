package viewer

import (
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "line up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "line down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "scroll left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "scroll right"),
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
		ToggleWrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp(helpKeyLabel("w"), "toggle wrap"),
		),
		ToggleFollow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp(helpKeyLabel("f"), "toggle follow"),
		),
		OpenFilter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp(helpKeyLabel("/"), "filter"),
		),
		OpenShell: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp(helpKeyLabel("s"), "shell"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp(helpKeyLabel("esc"), "back"),
		),
	}
}

func helpKeyLabel(label string) string {
	return util.HelpKeyLabel(label)
}

type Controller struct{}

func (Controller) HandleKey(state *State, msg tea.KeyPressMsg, keys KeyMap) Transition {
	switch {
	case key.Matches(msg, keys.Right):
		return handleHorizontalScroll(state, true)
	case key.Matches(msg, keys.Left):
		return handleHorizontalScroll(state, false)
	case key.Matches(msg, keys.Up):
		return handleVerticalScroll(state, -1, false)
	case key.Matches(msg, keys.Down):
		return handleVerticalScroll(state, 1, false)
	case key.Matches(msg, keys.PageUp):
		return handleVerticalScroll(state, -1, true)
	case key.Matches(msg, keys.PageDown):
		return handleVerticalScroll(state, 1, true)
	case key.Matches(msg, keys.Home):
		return handleHome(state)
	case key.Matches(msg, keys.End):
		return handleEnd(state)
	case key.Matches(msg, keys.ToggleWrap):
		state.SetWrapLines(!state.WrapLines)
		return Transition{}
	case key.Matches(msg, keys.ToggleFollow):
		state.SetFollow(!state.Follow)
		return Transition{}
	case key.Matches(msg, keys.OpenFilter):
		state.Filter.Active = !state.Filter.Active
		if state.Filter.Active {
			state.Filter.Input.Focus()
		} else {
			state.Filter.Input.Blur()
			state.Filter.Query = ""
			state.Filter.Input.SetValue("")
		}
		return Transition{}
	case key.Matches(msg, keys.OpenShell):
		return Transition{LaunchShell: true}
	case key.Matches(msg, keys.Back):
		return Transition{ExitToBrowse: true}
	default:
		return Transition{}
	}
}

func handleHorizontalScroll(state *State, right bool) Transition {
	if state.WrapLines {
		return Transition{}
	}
	state.SetFollow(false)
	step := 8
	if right {
		state.Viewport.ScrollRight(step)
		state.HorizontalOffset += step
	} else {
		state.Viewport.ScrollLeft(step)
		state.HorizontalOffset = max(0, state.HorizontalOffset-step)
	}
	return Transition{}
}

func handleVerticalScroll(state *State, direction int, isPage bool) Transition {
	state.SetFollow(false)
	if isPage {
		if direction > 0 {
			state.Viewport.PageDown()
		} else {
			state.Viewport.PageUp()
		}
	} else {
		if direction > 0 {
			state.Viewport.ScrollDown(1)
		} else {
			state.Viewport.ScrollUp(1)
		}
	}
	if direction > 0 && state.Viewport.AtBottom() {
		state.SetFollow(true)
	}
	return Transition{}
}

func handleHome(state *State) Transition {
	state.SetFollow(false)
	state.Viewport.SetXOffset(0)
	state.HorizontalOffset = 0
	state.Viewport.GotoTop()
	return Transition{}
}

func handleEnd(state *State) Transition {
	state.SetFollow(true)
	return Transition{}
}

func (k KeyMap) ShortHelp(resourceType ResourceType) []key.Binding {
	bindings := []key.Binding{
		k.HelpNavigate(), k.HelpPage(), k.HelpHomeEnd(), k.ToggleFollow, k.ToggleWrap, k.Back,
	}

	bindings = append(bindings, k.OpenFilter)

	if resourceType == ResourceTypeContainer {
		bindings = append(bindings, k.OpenShell)
	}

	return bindings
}

func (k KeyMap) HelpNavigate() key.Binding {
	return key.NewBinding(
		key.WithKeys("left", "up", "down", "right"),
		key.WithHelp(helpKeyLabel("← ↑ ↓ →"), "navigate"),
	)
}

func (k KeyMap) HelpPage() key.Binding {
	return key.NewBinding(
		key.WithKeys("pgup", "pgdown"),
		key.WithHelp(helpKeyLabel("pgup/dn"), "jump up/down"),
	)
}

func (k KeyMap) HelpHomeEnd() key.Binding {
	return key.NewBinding(
		key.WithKeys("home", "end"),
		key.WithHelp(helpKeyLabel("home/end"), "go to top/bottom"),
	)
}
