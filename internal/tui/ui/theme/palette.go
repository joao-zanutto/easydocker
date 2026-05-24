package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	// Surface colors
	Surface     = lipgloss.Color("236")
	SurfaceDeep = lipgloss.Color("234")
	SurfaceBG   = lipgloss.Color("24")

	// Text colors
	TextPrimary    = lipgloss.Color("230")
	TextSecondary  = lipgloss.Color("252")
	TextMuted      = lipgloss.Color("244")
	TextDim        = lipgloss.Color("248")
	TextSilver     = lipgloss.Color("250")
	TextBreadcrumb = lipgloss.Color("247")

	// Accent colors
	Gold        = lipgloss.Color("186")
	Teal        = lipgloss.Color("109")
	Cyan        = lipgloss.Color("86")
	BlueGray    = lipgloss.Color("110")
	Periwinkle  = lipgloss.Color("67")
	SteelBlue   = lipgloss.Color("31")
	LightYellow = lipgloss.Color("229")
	MutedPurple = lipgloss.Color("60")

	// Semantic/State colors
	Green  = lipgloss.Color("42")
	Orange = lipgloss.Color("214")
	Red    = lipgloss.Color("203")
	Pink   = lipgloss.Color("199")
)

func ContainerStateColor(state string) color.Color {
	switch strings.ToLower(state) {
	case "running":
		return Green
	case "paused", "restarting", "created":
		return Orange
	case "exited", "stopped":
		return Red
	case "dead":
		return Pink
	default:
		return BlueGray
	}
}
