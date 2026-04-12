package logs

import (
	"errors"
	"testing"

	"easydocker/internal/core"
)

func TestControllerEnterReturnsInitialLoadTransition(t *testing.T) {
	controller := Controller{}
	state := NewState()
	state.SessionID = 10

	transition := controller.Enter(&state, "ctr-1")
	if transition.Load == nil {
		t.Fatalf("enter transition should include load request")
	}
	if transition.Load.ContainerID != "ctr-1" {
		t.Fatalf("containerID = %q, want ctr-1", transition.Load.ContainerID)
	}
	if transition.Load.SessionID != 11 {
		t.Fatalf("sessionID = %d, want 11", transition.Load.SessionID)
	}
	if transition.Load.Tail != InitialTail {
		t.Fatalf("tail = %d, want %d", transition.Load.Tail, InitialTail)
	}
	if transition.Load.Src != SourceInitial {
		t.Fatalf("src = %q, want %q", transition.Load.Src, SourceInitial)
	}
}

func TestControllerHandleKeyHistoryProducesLoadRequest(t *testing.T) {
	controller := Controller{}
	state := NewState()
	state.ContainerID = "ctr-1"
	state.SessionID = 4
	state.Data = core.ContainerLiveData{Logs: []string{"a", "b", "c"}, CPUHistory: []float64{1.2}, MemHistory: []float64{3.4}}
	state.Viewport.SetContent("a\nb\nc")
	state.Viewport.GotoTop()

	transition := controller.HandleKey(&state, "home", 0)
	if transition.Load == nil {
		t.Fatalf("expected history load request")
	}
	if transition.Load.Src != SourceHistory {
		t.Fatalf("src = %q, want %q", transition.Load.Src, SourceHistory)
	}
	if transition.Load.Tail != len(state.Data.Logs)+TailStep {
		t.Fatalf("tail = %d, want %d", transition.Load.Tail, len(state.Data.Logs)+TailStep)
	}
	if !state.HistoryLoad {
		t.Fatalf("history load flag should be enabled")
	}
}

func TestControllerHandleKeyExitRequestsBrowse(t *testing.T) {
	controller := Controller{}
	state := NewState()
	state.SessionID = 3
	state.ContainerID = "ctr-1"

	transition := controller.HandleKey(&state, "esc", 2)
	if !transition.ExitToBrowse {
		t.Fatalf("exit transition expected")
	}
	if transition.ForceTab != 2 {
		t.Fatalf("force tab = %d, want 2", transition.ForceTab)
	}
	if state.SessionID != 4 {
		t.Fatalf("sessionID = %d, want 4", state.SessionID)
	}
	if state.ContainerID != "" {
		t.Fatalf("containerID should reset on exit")
	}
}

func TestControllerHandleResultMapsErrorsAndSyncsViewport(t *testing.T) {
	controller := Controller{}
	state := NewState()
	state.ContainerID = "ctr-1"
	state.SessionID = 9
	state.InitialLoad = true

	errTransition := controller.HandleResult(&state, ResultMsg{
		ContainerID: "ctr-1",
		SessionID:   9,
		Err:         errors.New("boom"),
	}, 40, 4)
	if errTransition.Err == nil {
		t.Fatalf("error should be propagated in transition")
	}

	state.InitialLoad = true
	state.HistoryLoad = true
	okTransition := controller.HandleResult(&state, ResultMsg{
		ContainerID: "ctr-1",
		SessionID:   9,
		Data:        core.ContainerLiveData{Logs: []string{"l1", "l2", "l3"}},
		Src:         SourceInitial,
	}, 30, 2)
	if okTransition.Err != nil {
		t.Fatalf("unexpected error transition: %v", okTransition.Err)
	}
	if state.Viewport.Width != 30 || state.Viewport.Height != 2 {
		t.Fatalf("viewport size = (%d,%d), want (30,2)", state.Viewport.Width, state.Viewport.Height)
	}
}
