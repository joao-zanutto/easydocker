package menu

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Transition struct {
	Quit bool
	Back bool
}

type Controller struct{}

func (Controller) HandleKey(menu *MenuState, help *HelpState, msg tea.KeyPressMsg, keys MenuKeyMap) Transition {
	switch {
	case key.Matches(msg, keys.Up):
		if menu.Cursor > 0 {
			menu.Cursor--
		}
		return Transition{}
	case key.Matches(msg, keys.Down):
		if menu.Cursor < len(menu.Items)-1 {
			menu.Cursor++
		}
		return Transition{}
	case key.Matches(msg, keys.Select):
		if len(menu.Items) > 0 {
			item := menu.Items[menu.Cursor]
			switch item.Action {
			case MenuActionHelp:
				help.Active = true
				help.Cursor = 0
				menu.Active = false
			case MenuActionQuit:
				return Transition{Quit: true}
			}
		}
		return Transition{}
	case key.Matches(msg, keys.Back):
		menu.Active = false
	}

	return Transition{}
}

func (Controller) HandleHelpKey(help *HelpState, menu *MenuState, msg tea.KeyPressMsg, keys MenuKeyMap, contentHeight, visibleHeight int) Transition {
	maxScroll := contentHeight - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if help.Cursor > maxScroll {
		help.Cursor = maxScroll
	}
	if help.Cursor < 0 {
		help.Cursor = 0
	}

	switch {
	case key.Matches(msg, keys.Up):
		if help.Cursor > 0 {
			help.Cursor--
		}
	case key.Matches(msg, keys.Down):
		if help.Cursor < maxScroll {
			help.Cursor++
		}
	}
	return Transition{Back: key.Matches(msg, keys.Back)}
}
