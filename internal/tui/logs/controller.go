package logs

import (
	"easydocker/internal/core"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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

func (Controller) HandleKey(state *State, msg tea.KeyMsg, keys KeyMap, containersTab int) Transition {
	msgKey := msg.String()

	switch {
	case key.Matches(msg, keys.Right):
		state.SetFollow(false)
		state.Viewport.ScrollRight(8)
		return Transition{}
	case key.Matches(msg, keys.Left):
		state.SetFollow(false)
		state.Viewport.ScrollLeft(8)
		return Transition{}
	case key.Matches(msg, keys.Up):
		state.SetFollow(false)
		state.Viewport.LineUp(1)
		return historyTransitionIfNeeded(state)
	case key.Matches(msg, keys.Down):
		if state.Viewport.AtBottom() {
			state.SetFollow(true)
			return Transition{}
		}
		state.SetFollow(false)
		state.Viewport.LineDown(1)
		return Transition{}
	case key.Matches(msg, keys.PageUp):
		state.SetFollow(false)
		state.Viewport.PageUp()
		return historyTransitionIfNeeded(state)
	case key.Matches(msg, keys.PageDown):
		if state.Viewport.AtBottom() {
			state.SetFollow(true)
			return Transition{}
		}
		state.SetFollow(false)
		state.Viewport.PageDown()
		return Transition{}
	case key.Matches(msg, keys.Home):
		state.SetFollow(false)
		state.Viewport.SetXOffset(0)
		state.Viewport.GotoTop()
		return historyTransitionIfNeeded(state)
	case key.Matches(msg, keys.End):
		state.SetFollow(true)
		return Transition{}
	case key.Matches(msg, keys.ToggleFollow):
		state.SetFollow(!state.Follow)
		return Transition{}
	case key.Matches(msg, keys.Back):
		return Controller{}.Exit(state, containersTab)
	case msgKey == " " || msgKey == "b" || msgKey == "g" || msgKey == "G" || msgKey == "q" || msgKey == "tab":
		return Transition{}
	default:
		return Transition{}
	}
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
		state.ApplyHistory(msg.Data, state.Viewport.YOffset)
	case SourceInitial:
		state.ApplyInitial(msg.Data)
	default:
		state.ApplyPoll(msg.Data, state.Viewport.YOffset)
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
