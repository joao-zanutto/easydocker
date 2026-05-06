package browse

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Controller struct{}

type Transition struct {
	ChangeTab        int
	CursorMove       int
	ActivateFilter   bool
	DeactivateFilter bool
	OpenResource     bool
	OpenShell        bool
	Quit             bool
	ToggleScope      bool
}

func (Controller) HandleKey(state *State, msg tea.KeyPressMsg, keys KeyMap) Transition {
	switch {
	case key.Matches(msg, keys.TabRight):
		return Transition{ChangeTab: 1}
	case key.Matches(msg, keys.TabLeft):
		return Transition{ChangeTab: -1}
	case key.Matches(msg, keys.MoveUp):
		return Transition{CursorMove: -1}
	case key.Matches(msg, keys.MoveDown):
		return Transition{CursorMove: 1}
	case key.Matches(msg, keys.PageUp):
		return Transition{CursorMove: -5}
	case key.Matches(msg, keys.PageDown):
		return Transition{CursorMove: 5}
	case key.Matches(msg, keys.ToggleScope):
		return Transition{ToggleScope: true}
	case key.Matches(msg, keys.OpenLogs):
		return Transition{OpenResource: true}
	case key.Matches(msg, keys.OpenShell):
		return Transition{OpenShell: true}
	case key.Matches(msg, keys.OpenFilter):
		return Transition{ActivateFilter: true}
	case key.Matches(msg, keys.Quit):
		return Transition{Quit: true}
	default:
		return Transition{}
	}
}
