package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultUsesDetectedBackground(t *testing.T) {
	previous := lipgloss.HasDarkBackground()
	t.Cleanup(func() {
		lipgloss.SetHasDarkBackground(previous)
	})

	tests := []struct {
		name        string
		dark        bool
		activeBG    lipgloss.Color
		labelFG     lipgloss.Color
		headerRowFG  lipgloss.Color
	}{
		{name: "dark", dark: true, activeBG: lipgloss.Color("236"), labelFG: lipgloss.Color("109"), headerRowFG: lipgloss.Color("230")},
		{name: "light", dark: false, activeBG: lipgloss.Color("189"), labelFG: lipgloss.Color("25"), headerRowFG: lipgloss.Color("235")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lipgloss.SetHasDarkBackground(tc.dark)

			s := Default()
			if s.ActiveBG != tc.activeBG {
				t.Fatalf("ActiveBG = %q, want %q", s.ActiveBG, tc.activeBG)
			}
			if got := s.Label.GetForeground(); got != tc.labelFG {
				t.Fatalf("Label foreground = %v, want %q", got, tc.labelFG)
			}
			if got := s.HeaderRow.GetForeground(); got != tc.headerRowFG {
				t.Fatalf("HeaderRow foreground = %v, want %q", got, tc.headerRowFG)
			}
		})
	}
}