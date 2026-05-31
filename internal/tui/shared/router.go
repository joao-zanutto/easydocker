package shared

// Screen identifies a top-level TUI mode.
type Screen int

const (
	Main Screen = iota
	LogViewer
	InspectViewer
)
