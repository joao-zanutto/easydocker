package logs

import (
	"strings"
	"testing"

	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderPanel_InitialLoadPadsToHeight(t *testing.T) {
	state := NewState()
	state.InitialLoad = true

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 4)
	lines := strings.Split(panel, "\n")
	if len(lines) != 4 {
		t.Fatalf("lines len = %d, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "Loading logs") {
		t.Fatalf("expected loading line, got %q", lines[0])
	}
}

func TestRenderPanel_EmptyLogsPadsToHeight(t *testing.T) {
	state := NewState()
	state.InitialLoad = false
	state.Data.Logs = nil

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 3)
	lines := strings.Split(panel, "\n")
	if len(lines) != 3 {
		t.Fatalf("lines len = %d, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "No logs found") {
		t.Fatalf("expected empty logs hint, got %q", lines[0])
	}
}

func TestRenderPanel_HistoryLoadPrependsIndicator(t *testing.T) {
	state := NewState()
	state.Data.Logs = []string{"line-1", "line-2", "line-3"}
	state.SyncViewport([]string{"line-1", "line-2", "line-3"}, 78, 4)
	state.HistoryLoad = true

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 4)
	lines := strings.Split(panel, "\n")
	if len(lines) != 4 {
		t.Fatalf("lines len = %d, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "Loading older logs") {
		t.Fatalf("expected history loading line first, got %q", lines[0])
	}
}

func TestRenderContent_RendersFrameAndBreadcrumb(t *testing.T) {
	state := NewState()
	state.Data.Logs = []string{"line-1", "line-2"}
	state.SyncViewport([]string{"line-1", "line-2"}, 60, 4)

	view := RenderContent(ViewModel{
		State:         state,
		ContainerName: "api",
		Width:         80,
		Height:        8,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	})

	if view == "" {
		t.Fatalf("RenderContent returned empty string")
	}
	if !strings.Contains(view, "api") || !strings.Contains(view, "Logs") {
		t.Fatalf("expected breadcrumb in view, got %q", view)
	}
}

func TestRenderContent_KeepsLineCountIndicatorOnOneLine(t *testing.T) {
	state := NewState()
	logs := make([]string, 45)
	for i := range logs {
		logs[i] = strings.Repeat("x", 10)
	}
	state.Data.Logs = logs
	state.SyncViewport(logs, 60, 8)

	header := RenderHeader(ViewModel{
		State:         state,
		ContainerName: "server-postgres",
		Width:         76,
		Height:        12,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	}, "Containers / server-postgres / Logs", len(state.Data.Logs), 37, 45)

	view := RenderContent(ViewModel{
		State:         state,
		ContainerName: "server-postgres",
		Width:         80,
		Height:        12,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	})

	if !strings.Contains(view, "lines:(38-45/45)") {
		t.Fatalf("expected intact line-count indicator in view; header width=%d header=%q view=%q", util.DisplayWidth(header), header, view)
	}
	if strings.Contains(view, "lines:(38-45/…") {
		t.Fatalf("line-count indicator was unexpectedly ellipsized: %q", view)
	}
	if strings.Contains(view, "\n│ 45/45)") {
		t.Fatalf("line-count indicator was split across lines: %q", view)
	}
}
