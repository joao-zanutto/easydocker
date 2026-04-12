package mode

import "testing"

func TestRouteRootKey(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		screen Screen
		want   RootKeyRoute
	}{
		{name: "global quit", key: "ctrl+c", screen: Browse, want: RouteQuit},
		{name: "global noop q", key: "q", screen: Browse, want: RouteNoop},
		{name: "global noop tab", key: "tab", screen: Logs, want: RouteNoop},
		{name: "logs delegated", key: "up", screen: Logs, want: RouteLogs},
		{name: "browse delegated", key: "up", screen: Browse, want: RouteBrowse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RouteRootKey(tt.key, tt.screen); got != tt.want {
				t.Fatalf("RouteRootKey(%q, %v) = %v, want %v", tt.key, tt.screen, got, tt.want)
			}
		})
	}
}

func TestTransitions(t *testing.T) {
	if got := EnterLogsTransition(); got != Logs {
		t.Fatalf("EnterLogsTransition() = %v, want Logs", got)
	}

	screen, tab := ExitLogsTransition(0)
	if screen != Browse || tab != 0 {
		t.Fatalf("ExitLogsTransition() = (%v, %d), want (Browse, 0)", screen, tab)
	}
}
