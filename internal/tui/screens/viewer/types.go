package viewer

import (
	"easydocker/internal/tui/ui/components"

	"charm.land/bubbles/v2/viewport"
)

type Transition struct {
	LaunchShell bool
}

type ContentMsg struct {
	SessionID   int
	ContainerID string
	Data        []string
	Err         error
	Tail        int
	Src         Source
}

type Source string

const (
	SourceInitial Source = "initial"
	SourceHistory Source = "history"
	SourcePoll    Source = "poll"
)

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

type State struct {
	ContainerID  string
	Data         []string
	Filter       components.FilterState
	WrapLines    bool
	InitialLoad  bool
	Follow       bool
	Viewport     viewport.Model
	ContentType  ContentType
	ResourceName string
}

func NewState() State {
	vp := viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))
	vp.SetHorizontalStep(8)
	vp.SetContent("")

	filterState := components.NewFilterState()

	return State{Follow: true, Viewport: vp, Filter: filterState}
}
