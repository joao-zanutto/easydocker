package logs

import (
	"easydocker/internal/core"
	"easydocker/internal/tui/components"

	"charm.land/bubbles/v2/viewport"
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
	ExitToBrowse   bool
	ForceTab       int
	Load           *LoadRequest
	Err            error
	LaunchTerminal bool
}

type State struct {
	ContainerID               string
	SessionID                 int
	Data                      core.ContainerLiveData
	TailLines                 int
	HistoryBaseLen            int
	HistoryAppendedDuringLoad int
	HistoryNoProgressCount    int
	Filter                    components.FilterState
	HorizontalOffset          int
	WrapLines                 bool
	WrapXOffset               int
	InitialLoad               bool
	HistoryDone               bool
	HistoryLoad               bool
	Follow                    bool
	Viewport                  viewport.Model
}

func NewState() State {
	vp := viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))
	vp.SetHorizontalStep(8)
	vp.SetContent("")

	filterState := components.NewFilterState()

	return State{Follow: true, Viewport: vp, Filter: filterState}
}
