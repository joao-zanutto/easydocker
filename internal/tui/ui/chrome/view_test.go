package chrome

import (
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type testFooterKeyMap struct {
	bindings []key.Binding
}

func (m testFooterKeyMap) ShortHelp() []key.Binding {
	return m.bindings
}

func (m testFooterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.bindings}
}

func TestRenderHeaderTabs_WidthFallback(t *testing.T) {
	specs := []TabSpec{
		{Tab: 0, Icon: "🐳", Name: "Containers", Count: 2},
		{Tab: 1, Icon: "💿", Name: "Images", Count: 2},
		{Tab: 2, Icon: "🔌", Name: "Networks", Count: 1},
		{Tab: 3, Icon: "📂", Name: "Volumes", Count: 2},
	}
	renderTab := func(tab int, label string) string { return label }

	wide := util.StripANSI(strings.Join(RenderHeaderTabs(specs, 200, renderTab), " "))
	for _, token := range []string{"Containers (2)", "Images (2)", "Networks (1)", "Volumes (2)"} {
		if !strings.Contains(wide, token) {
			t.Fatalf("expected wide header tabs to contain %q, got %q", token, wide)
		}
	}

	tiny := RenderHeaderTabs(specs, 1, renderTab)
	if len(tiny) != 4 {
		t.Fatalf("expected 4 tabs, got %d", len(tiny))
	}
	icons := []string{"🐳", "💿", "🔌", "📂"}
	for i, want := range icons {
		got := strings.TrimSpace(util.StripANSI(tiny[i]))
		if got != want {
			t.Fatalf("tab %d icon fallback = %q, want %q", i, got, want)
		}
	}
}

func TestRenderScopeBadge_FallbackByWidth(t *testing.T) {
	tests := []struct {
		name       string
		showAll    bool
		wideToken  string
		narrowWant string
	}{
		{name: "all", showAll: true, wideToken: "container scope: all", narrowWant: "all"},
		{name: "running", showAll: false, wideToken: "container scope: running", narrowWant: "running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wide := util.StripANSI(RenderScopeBadge(tt.showAll, 40, func(label string) string { return label }))
			if !strings.Contains(wide, tt.wideToken) {
				t.Fatalf("expected wide scope badge to contain %q, got %q", tt.wideToken, wide)
			}
			narrow := strings.TrimSpace(util.StripANSI(RenderScopeBadge(tt.showAll, 2, func(label string) string { return label })))
			if narrow != tt.narrowWant {
				t.Fatalf("narrow scope badge = %q, want %q", narrow, tt.narrowWant)
			}
		})
	}
}

func TestRenderHeaderAndFooter(t *testing.T) {
	header := RenderHeader(HeaderInput{
		Width:            220,
		Title:            "EasyDocker",
		TotalsText:       "CPU 1.0%  MEM 2.0%",
		LoadingStageText: "loading resources",
		ActiveTab:        0,
		ShowAll:          true,
		Err:              nil,
		Tabs: []TabSpec{
			{Tab: 0, Icon: "🐳", Name: "Containers", Count: 2},
			{Tab: 1, Icon: "💿", Name: "Images", Count: 2},
		},
		Styles: HeaderStyles{
			Header:    lipgloss.NewStyle(),
			Title:     lipgloss.NewStyle(),
			TitleMeta: lipgloss.NewStyle(),
			Badge:     lipgloss.NewStyle(),
			ErrorText: lipgloss.NewStyle(),
		},
		RenderTab: func(tab int, label string) string { return label },
	})
	if !strings.Contains(util.StripANSI(header), "EasyDocker") {
		t.Fatalf("expected header to contain title, got %q", header)
	}

	footer := RenderFooter(FooterInput{
		Width: 220,
		KeyMap: testFooterKeyMap{bindings: []key.Binding{
			key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
			key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "switch tabs")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle running/all")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "logs")),
		}},
		Styles: FooterStyles{
			Footer:  lipgloss.NewStyle(),
			Key:     lipgloss.NewStyle(),
			KeyText: lipgloss.NewStyle(),
		},
	})
	for _, token := range []string{"navigate", "switch tabs", "toggle running/all", "logs"} {
		if !strings.Contains(util.StripANSI(footer), token) {
			t.Fatalf("expected footer to contain %q, got %q", token, footer)
		}
	}
}

func TestRenderTotalsLabel(t *testing.T) {
	snapshot := core.Snapshot{TotalCPU: 12.3, TotalMem: 1024, TotalLimit: 2048}
	if got := RenderTotalsLabel(snapshot, 0, 1, 3, true, ""); !strings.Contains(got, "CPU") || !strings.Contains(got, "MEM") {
		t.Fatalf("RenderTotalsLabel() = %q, want CPU/MEM text", got)
	}
}

func TestRenderTotalsLabel_UsesIndicatorOnlyBeforeFirstMetricsLoad(t *testing.T) {
	snapshot := core.Snapshot{TotalCPU: 12.3, TotalMem: 1024, TotalLimit: 2048}

	loading := RenderTotalsLabel(snapshot, 3, 0, 3, false, "⠋")
	if !strings.Contains(loading, "CPU ⠋") || !strings.Contains(loading, "MEM ⠋") {
		t.Fatalf("pre-metrics totals should show spinner indicator, got %q", loading)
	}

	stale := RenderTotalsLabel(snapshot, 3, 0, 3, true, "⠋")
	if strings.Contains(stale, "⠋") {
		t.Fatalf("post-metrics totals should keep stale numbers, got %q", stale)
	}
}
