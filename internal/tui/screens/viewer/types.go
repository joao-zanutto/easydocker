package viewer

import (
	"charm.land/lipgloss/v2"
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
	ContentTypeConfig
)

type Styles struct {
	Breadcrumb   lipgloss.Style
	FollowOn     lipgloss.Style
	FollowOff    lipgloss.Style
	Muted        lipgloss.Style
	Divider      lipgloss.Style
	SubpageFrame lipgloss.Style
	Key          lipgloss.Style
	KeyText      lipgloss.Style
}

type TransitionMsg struct {
	BackToBrowse bool
	LaunchShell  bool
}
