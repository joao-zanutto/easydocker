package viewer

import (
	"easydocker/internal/tui/ui/components"

	"charm.land/bubbles/v2/viewport"
)

type Transition struct {
	ExitToBrowse bool
	ForceTab     int
	Load         *LoadRequest
	Err          error
	LaunchShell  bool
}

type Source string

const (
	SourceInitial Source = "initial"
	SourceHistory Source = "history"
	SourcePoll    Source = "poll"
)

type LoadRequest struct {
	ContainerID string
	SessionID   int
	Tail        int
	Src         Source
}

type ResultMsg struct {
	ContainerID string
	SessionID   int
	Data        any
	Err         error
	Tail        int
	Src         Source
}

type ContentMsg struct {
	SessionID   int
	ContainerID string
	Data        []string
	Err         error
	Tail        int
	Src         Source
}

type LineCountInfo struct {
	Total int
	Start int
	End   int
}

type ContentType int

const (
	ContentTypeLogs ContentType = iota
	ContentTypeInspect
)

type ResourceType int

const (
	ResourceTypeContainer ResourceType = iota
	ResourceTypeVolume
	ResourceTypeNetwork
	ResourceTypeImage
)

type State struct {
	ContainerID               string
	SessionID                 int
	Data                      []string
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
