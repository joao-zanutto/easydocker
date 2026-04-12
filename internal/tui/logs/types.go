package logs

import (
	"easydocker/internal/core"

	"github.com/charmbracelet/bubbles/viewport"
)

type Source string

const (
	SourceInitial Source = "initial"
	SourceHistory Source = "history"
	SourcePoll    Source = "poll"
)

type ResultMsg struct {
	ContainerID string
	SessionID   int
	Data        core.ContainerLiveData
	Err         error
	Tail        int
	Src         Source
}

type LoadRequest struct {
	ContainerID string
	SessionID   int
	PrevCPU     []float64
	PrevMem     []float64
	Tail        int
	Src         Source
}

type Transition struct {
	ExitToBrowse bool
	ForceTab     int
	Load         *LoadRequest
	Err          error
}

type State struct {
	ContainerID string
	SessionID   int
	Data        core.ContainerLiveData
	TailLines   int
	InitialLoad bool
	HistoryDone bool
	HistoryLoad bool
	Follow      bool
	Viewport    viewport.Model
}

func NewState() State {
	vp := viewport.New(1, 1)
	vp.SetHorizontalStep(8)
	vp.SetContent("")
	return State{Follow: true, Viewport: vp}
}
