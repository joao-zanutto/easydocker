package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultUsesDetectedBackground(t *testing.T) {
	s := Default()
	if s.ActiveBG != lipgloss.Color("236") {
		t.Fatalf("ActiveBG = %q, want %q", s.ActiveBG, lipgloss.Color("236"))
	}
	if got := s.Label.GetForeground(); got != lipgloss.Color("109") {
		t.Fatalf("Label foreground = %v, want %q", got, lipgloss.Color("109"))
	}
	if got := s.HeaderRow.GetForeground(); got != lipgloss.Color("230") {
		t.Fatalf("HeaderRow foreground = %v, want %q", got, lipgloss.Color("230"))
	}
}
