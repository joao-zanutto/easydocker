package shared

// Screen identifies a top-level TUI mode.
type Screen int

const (
	Browse Screen = iota
	Logs
	Inspect
)



