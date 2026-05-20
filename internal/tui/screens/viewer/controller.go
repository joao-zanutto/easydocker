package viewer

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

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
	} else {
		state.Viewport.ScrollLeft(step)
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
	state.Viewport.GotoTop()
	return Transition{}
}

func handleEnd(state *State) Transition {
	state.SetFollow(true)
	return Transition{}
}
