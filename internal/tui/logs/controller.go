package logs

import (
	"easydocker/internal/core"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Controller struct{}

func (Controller) Enter(state *State, containerID string) Transition {
	nextSession := state.SessionID + 1
	state.ResetForContainer(nextSession, containerID, InitialTail)
	return Transition{
		Load: &LoadRequest{
			ContainerID: containerID,
			SessionID:   state.SessionID,
			Tail:        state.TailLines,
			Src:         SourceInitial,
		},
	}
}

func (Controller) Exit(state *State, containersTab int) Transition {
	nextSession := state.SessionID + 1
	state.ResetForExit(nextSession)
	return Transition{ExitToBrowse: true, ForceTab: containersTab}
}

func (Controller) HandleKey(state *State, msg tea.KeyPressMsg, keys KeyMap, containersTab int) Transition {
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

	case key.Matches(msg, keys.ToggleFollow):
		state.SetFollow(!state.Follow)
		return Transition{}
	case "s":
		return Transition{LaunchTerminal: true}
	case key.Matches(msg, keys.Back):
		return Controller{}.Exit(state, containersTab)

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

	// When scrolling down and already at bottom, re-enable follow
	if direction > 0 && state.Viewport.AtBottom() {
		state.SetFollow(true)
	}

	return historyTransitionIfNeeded(state)
}

func handleHome(state *State) Transition {
	state.SetFollow(false)
	state.Viewport.SetXOffset(0)
	state.HorizontalOffset = 0
	state.Viewport.GotoTop()
	return historyTransitionIfNeeded(state)
}

func handleEnd(state *State) Transition {
	state.SetFollow(true)
	return Transition{}
}

func historyTransitionIfNeeded(state *State) Transition {
	if !state.CanLoadHistory() {
		return Transition{}
	}
	nextTail := len(state.Data.Logs) + TailStep
	state.StartHistoryLoad(nextTail)
	return Transition{
		Load: &LoadRequest{
			ContainerID: state.ContainerID,
			SessionID:   state.SessionID,
			PrevCPU:     state.Data.CPUHistory,
			PrevMem:     state.Data.MemHistory,
			Tail:        nextTail,
			Src:         SourceHistory,
		},
	}
}

func (Controller) HandleResult(state *State, msg ResultMsg, visibleWidth, visibleRows int) Transition {
	if msg.SessionID != state.SessionID || msg.ContainerID != state.ContainerID {
		return Transition{}
	}
	if msg.Err != nil {
		state.InitialLoad = false
		state.HistoryLoad = false
		return Transition{Err: msg.Err}
	}

	if msg.Tail > 0 && msg.Tail > state.TailLines {
		state.TailLines = msg.Tail
	}

	switch msg.Src {
	case SourceHistory:
		state.ApplyHistory(msg.Data, state.Viewport.YOffset())
	case SourceInitial:
		state.ApplyInitial(msg.Data)
	default:
		state.ApplyPoll(msg.Data, state.Viewport.YOffset())
	}
	state.SyncViewportFromData(visibleWidth, visibleRows)
	return Transition{}
}

func (Controller) SelectedContainer(state State, containers []core.ContainerRow) (core.ContainerRow, bool) {
	if state.ContainerID == "" {
		return core.ContainerRow{}, false
	}
	for _, container := range containers {
		if container.FullID == state.ContainerID {
			return container, true
		}
	}
	return core.ContainerRow{}, false
}
