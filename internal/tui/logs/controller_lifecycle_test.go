package logs

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"easydocker/internal/core"

	tea "github.com/charmbracelet/bubbletea"
)

func TestControllerHandleResult_IgnoresMismatchedMessage(t *testing.T) {
	controller := Controller{}
	state := newControllerState(logLines("base", 5))
	before := append([]string(nil), state.Data.Logs...)

	transition := controller.HandleResult(&state, ResultMsg{
		ContainerID: "other-container",
		SessionID:   state.SessionID + 1,
		Data:        core.ContainerLiveData{Logs: logLines("new", 5)},
		Src:         SourcePoll,
	}, 80, 8)

	if transition != (Transition{}) {
		t.Fatalf("mismatched message should yield empty transition, got %#v", transition)
	}
	if !reflect.DeepEqual(state.Data.Logs, before) {
		t.Fatalf("state logs changed on mismatched message")
	}
}

func TestControllerHandleResult_ErrorClearsLoadFlags(t *testing.T) {
	controller := Controller{}
	state := newControllerState(logLines("base", 5))
	state.InitialLoad = true
	state.HistoryLoad = true

	transition := controller.HandleResult(&state, ResultMsg{
		ContainerID: state.ContainerID,
		SessionID:   state.SessionID,
		Err:         errors.New("boom"),
		Src:         SourcePoll,
	}, 80, 8)

	if transition.Err == nil {
		t.Fatalf("expected error transition")
	}
	if state.InitialLoad {
		t.Fatalf("initialLoad should be cleared on error")
	}
	if state.HistoryLoad {
		t.Fatalf("historyLoad should be cleared on error")
	}
}

func TestControllerHandleResult_InitialSourceResetsFlagsAndData(t *testing.T) {
	controller := Controller{}
	state := newControllerState(logLines("base", 5))
	state.InitialLoad = true
	state.HistoryLoad = true
	state.HistoryDone = true

	data := core.ContainerLiveData{Logs: []string{"a", "b", "c"}}
	_ = controller.HandleResult(&state, ResultMsg{
		ContainerID: state.ContainerID,
		SessionID:   state.SessionID,
		Data:        data,
		Src:         SourceInitial,
	}, 40, 4)

	if state.InitialLoad {
		t.Fatalf("initialLoad should be false after initial result")
	}
	if state.HistoryLoad {
		t.Fatalf("historyLoad should be false after initial result")
	}
	if state.HistoryDone {
		t.Fatalf("historyDone should be reset after initial result")
	}
	if !reflect.DeepEqual(state.Data.Logs, data.Logs) {
		t.Fatalf("logs mismatch: got %v want %v", state.Data.Logs, data.Logs)
	}
}

func TestControllerHandleResult_HistorySource(t *testing.T) {
	controller := Controller{}

	t.Run("advances offset when prepending history", func(t *testing.T) {
		state := newControllerState(logLines("base", 30))
		state.HistoryLoad = true
		state.Follow = false
		state.Viewport.SetYOffset(2)
		previousYOffset := state.Viewport.YOffset

		history := append(logLines("older", 5), state.Data.Logs...)
		_ = controller.HandleResult(&state, ResultMsg{
			ContainerID: state.ContainerID,
			SessionID:   state.SessionID,
			Data:        core.ContainerLiveData{Logs: history},
			Src:         SourceHistory,
		}, 80, 8)

		if state.HistoryLoad {
			t.Fatalf("historyLoad should be false")
		}
		if state.HistoryDone {
			t.Fatalf("historyDone should remain false when history grows")
		}
		if got, want := state.Viewport.YOffset, previousYOffset+5; got != want {
			t.Fatalf("YOffset = %d, want %d", got, want)
		}
	})

	t.Run("marks done when history does not grow", func(t *testing.T) {
		state := newControllerState(logLines("base", 20))
		state.HistoryLoad = true

		_ = controller.HandleResult(&state, ResultMsg{
			ContainerID: state.ContainerID,
			SessionID:   state.SessionID,
			Data:        core.ContainerLiveData{Logs: append([]string(nil), state.Data.Logs...)},
			Src:         SourceHistory,
		}, 80, 8)

		if !state.HistoryDone {
			t.Fatalf("historyDone should be true when history does not grow")
		}
	})
}

func TestControllerHandleResult_PollSourcePreservesOffsetWhenNotFollowing(t *testing.T) {
	controller := Controller{}
	state := newControllerState(logLines("base", 30))
	state.InitialLoad = true
	state.Follow = false
	state.Viewport.SetYOffset(4)
	previousYOffset := state.Viewport.YOffset

	polled := core.ContainerLiveData{Logs: logLines("poll", 30)}
	_ = controller.HandleResult(&state, ResultMsg{
		ContainerID: state.ContainerID,
		SessionID:   state.SessionID,
		Data:        polled,
		Src:         SourcePoll,
	}, 80, 8)

	if state.InitialLoad {
		t.Fatalf("initialLoad should be false after poll")
	}
	if got, want := state.Viewport.YOffset, previousYOffset; got != want {
		t.Fatalf("YOffset = %d, want %d", got, want)
	}
	if !reflect.DeepEqual(state.Data.Logs, polled.Logs) {
		t.Fatalf("poll logs mismatch")
	}
}

func TestControllerHandleKey_Behavior(t *testing.T) {
	controller := Controller{}

	t.Run("f toggles follow", func(t *testing.T) {
		state := newControllerState(logLines("base", 100))
		_ = controller.HandleKey(&state, keyMsg(tea.KeyRunes, 'f'), NewKeyMap(), 0)
		if state.Follow {
			t.Fatalf("follow should be disabled")
		}

		state.Viewport.GotoTop()
		_ = controller.HandleKey(&state, keyMsg(tea.KeyRunes, 'f'), NewKeyMap(), 0)
		if !state.Follow {
			t.Fatalf("follow should be enabled")
		}
		if !state.Viewport.AtBottom() {
			t.Fatalf("viewport should jump to bottom when follow is enabled")
		}
	})

	t.Run("home requests history and end enables follow", func(t *testing.T) {
		state := newControllerState(logLines("base", 100))
		state.Follow = true

		tr := controller.HandleKey(&state, keyMsg(tea.KeyHome), NewKeyMap(), 0)
		if tr.Load == nil || tr.Load.Src != SourceHistory {
			t.Fatalf("home should request history load")
		}
		if state.Follow {
			t.Fatalf("follow should be disabled after home")
		}

		state.Follow = false
		state.Viewport.GotoTop()
		tr = controller.HandleKey(&state, keyMsg(tea.KeyEnd), NewKeyMap(), 0)
		if tr != (Transition{}) {
			t.Fatalf("end should not emit transition, got %#v", tr)
		}
		if !state.Follow {
			t.Fatalf("follow should be enabled after end")
		}
		if !state.Viewport.AtBottom() {
			t.Fatalf("viewport should be at bottom after end")
		}
	})

	t.Run("pgup and pgdown behavior", func(t *testing.T) {
		state := newControllerState(logLines("base", 100))
		state.Follow = true
		state.Viewport.GotoTop()

		tr := controller.HandleKey(&state, keyMsg(tea.KeyPgUp), NewKeyMap(), 0)
		if tr.Load == nil {
			t.Fatalf("pgup at top should request history")
		}
		if state.Follow {
			t.Fatalf("follow should be disabled after pgup")
		}

		state.Follow = false
		state.Viewport.GotoBottom()
		_ = controller.HandleKey(&state, keyMsg(tea.KeyPgDown), NewKeyMap(), 0)
		if !state.Follow {
			t.Fatalf("pgdown at bottom should re-enable follow")
		}
	})

	t.Run("down moves when not at bottom", func(t *testing.T) {
		state := newControllerState(logLines("base", 100))
		state.Follow = true
		state.Viewport.GotoTop()
		before := state.Viewport.YOffset

		_ = controller.HandleKey(&state, keyMsg(tea.KeyDown), NewKeyMap(), 0)
		if state.Follow {
			t.Fatalf("follow should be disabled")
		}
		if state.Viewport.YOffset <= before {
			t.Fatalf("down should move viewport")
		}
	})

	t.Run("esc exits with forced tab", func(t *testing.T) {
		state := newControllerState(logLines("base", 20))
		tr := controller.HandleKey(&state, keyMsg(tea.KeyEsc), NewKeyMap(), 3)
		if !tr.ExitToBrowse || tr.ForceTab != 3 {
			t.Fatalf("esc should request exit to browse with tab 3")
		}
	})
}

func newControllerState(lines []string) State {
	state := NewState()
	state.SessionID = 7
	state.ContainerID = "container-7"
	state.Data = core.ContainerLiveData{Logs: append([]string(nil), lines...)}
	state.SyncViewportFromData(80, 8)
	return state
}

func logLines(prefix string, count int) []string {
	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lines = append(lines, fmt.Sprintf("%s-%02d", prefix, i))
	}
	return lines
}
