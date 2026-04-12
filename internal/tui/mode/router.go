package mode

// Screen identifies a top-level TUI mode.
type Screen int

const (
	Browse Screen = iota
	Logs
)

// RootKeyRoute classifies top-level key routing decisions.
type RootKeyRoute int

const (
	RouteBrowse RootKeyRoute = iota
	RouteLogs
	RouteNoop
	RouteQuit
)

// RouteRootKey decides whether a key is global, ignored, or delegated by screen.
func RouteRootKey(key string, screen Screen) RootKeyRoute {
	switch key {
	case "ctrl+c":
		return RouteQuit
	case "q", "tab":
		return RouteNoop
	}

	if screen == Logs {
		return RouteLogs
	}
	return RouteBrowse
}

// EnterLogsTransition returns the target screen for entering logs mode.
func EnterLogsTransition() Screen {
	return Logs
}

// ExitLogsTransition returns the target screen and tab when leaving logs mode.
func ExitLogsTransition(containersTab int) (Screen, int) {
	return Browse, containersTab
}