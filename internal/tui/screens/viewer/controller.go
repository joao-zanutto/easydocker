package viewer

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Controller struct{}

func (Controller) HandleKey(vp *Viewport, msg tea.KeyPressMsg, keys KeyMap) Transition {
	switch {
	case key.Matches(msg, keys.Right):
		return handleHorizontalScroll(vp, true)
	case key.Matches(msg, keys.Left):
		return handleHorizontalScroll(vp, false)
	case key.Matches(msg, keys.Up):
		return handleVerticalScroll(vp, -1, false)
	case key.Matches(msg, keys.Down):
		return handleVerticalScroll(vp, 1, false)
	case key.Matches(msg, keys.PageUp):
		return handleVerticalScroll(vp, -1, true)
	case key.Matches(msg, keys.PageDown):
		return handleVerticalScroll(vp, 1, true)
	case key.Matches(msg, keys.Home):
		return handleHome(vp)
	case key.Matches(msg, keys.End):
		return handleEnd(vp)
	default:
		return Transition{}
	}
}

func handleHorizontalScroll(vp *Viewport, right bool) Transition {
	if vp.WrapLines {
		return Transition{}
	}
	vp.SetFollow(false)
	step := 8
	if right {
		vp.ScrollRight(step)
	} else {
		vp.ScrollLeft(step)
	}
	return Transition{}
}

func handleVerticalScroll(vp *Viewport, direction int, isPage bool) Transition {
	vp.SetFollow(false)
	if isPage {
		if direction > 0 {
			vp.PageDown()
		} else {
			vp.PageUp()
		}
	} else {
		if direction > 0 {
			vp.ScrollDown(1)
		} else {
			vp.ScrollUp(1)
		}
	}
	if direction > 0 && vp.AtBottom() {
		vp.SetFollow(true)
	}
	return Transition{}
}

func handleHome(vp *Viewport) Transition {
	vp.SetFollow(false)
	vp.SetXOffset(0)
	vp.GotoTop()
	return Transition{}
}

func handleEnd(vp *Viewport) Transition {
	vp.SetFollow(true)
	return Transition{}
}
