package tui

import (
	"easydocker/internal/tui/screens/browse"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) handleBrowseKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := browseKeyMap()

	// If filter mode is active, handle filter input first
	if m.browseFilter.Active {
		switch {
		case key.Matches(msg, keys.OpenMenu):
			// Esc exits filter mode and clears query
			m.browseFilter.Active = false
			m.browseFilter.Input.Blur()
			m.browseFilter.Query = ""
			m.browseFilter.Input.SetValue("")
			m.clampCursors()
			return m, nil
		case msg.String() == "enter":
			// Enter exits filter mode but keeps query
			m.browseFilter.Active = false
			m.browseFilter.Input.Blur()
			return m, nil
		case key.Matches(msg, keys.MoveUp):
			m.moveCursor(-1)
			return m, nil
		case key.Matches(msg, keys.MoveDown):
			m.moveCursor(1)
			return m, nil
		case key.Matches(msg, keys.PageUp):
			m.moveCursor(-browseCursorPageStep)
			return m, nil
		case key.Matches(msg, keys.PageDown):
			m.moveCursor(browseCursorPageStep)
			return m, nil
		default:
			// All other keys go to filter input
			var cmd tea.Cmd
			m.browseFilter.Input, cmd = m.browseFilter.Input.Update(msg)
			m.browseFilter.Query = m.browseFilter.Input.Value()
			// Recompute visible lists and clamp cursors to keep selection valid
			m.clampCursors()
			return m, cmd
		}
	}

	// Use controller for normal browse mode key handling
	browseState := browse.State{Filter: m.browseFilter}
	transition := browse.Controller{}.HandleKey(&browseState, msg, keys)
	m.browseFilter = browseState.Filter

	if transition.ChangeTab != 0 {
		m.moveActiveTab(transition.ChangeTab)
	}
	if transition.ActivateFilter {
		m.browseFilter.Active = true
		m.browseFilter.Input.Focus()
		m.browseFilter.Input.SetValue(m.browseFilter.Query)
	}
	if transition.ToggleScope {
		m.toggleContainerScope()
	}
	if transition.OpenResource {
		if m.toggleSelectedComposeProject() {
			return m, nil
		}
		if cmd := m.enterLogsModeIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if transition.OpenShell {
		if cmd := m.openShellIfContainerSelected(); cmd != nil {
			return m, cmd
		}
	}
	if transition.OpenInspect {
		return m.handleInspectTransition()
	}
	if transition.OpenMenu {
		m.menu.Active = true
		m.menu.Cursor = 0
	}
	if transition.CursorMove != 0 {
		m.moveCursor(transition.CursorMove)
	}

	return m, nil
}
