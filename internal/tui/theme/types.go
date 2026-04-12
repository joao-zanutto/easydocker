package theme

import "github.com/charmbracelet/lipgloss"

// Set contains all style primitives used by the root TUI orchestrator.
type Set struct {
	Page         lipgloss.Style
	Header       lipgloss.Style
	Title        lipgloss.Style
	TitleMeta    lipgloss.Style
	Breadcrumb   lipgloss.Style
	Tab          lipgloss.Style
	ActiveTab    lipgloss.Style
	Badge        lipgloss.Style
	Muted        lipgloss.Style
	MainFrame    lipgloss.Style
	SubpageFrame lipgloss.Style
	Divider      lipgloss.Style
	HeaderRow    lipgloss.Style
	Row          lipgloss.Style
	ActiveRow    lipgloss.Style
	Section      lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	ErrorText    lipgloss.Style
	Footer       lipgloss.Style
	Key          lipgloss.Style
	KeyText      lipgloss.Style
	FollowOn     lipgloss.Style
	FollowOff    lipgloss.Style
	MonitorBox   lipgloss.Style
	StateRun     lipgloss.Style
	StateWarn    lipgloss.Style
	StateStop    lipgloss.Style
	StateDead    lipgloss.Style
	StateOther   lipgloss.Style
	ActiveBG     lipgloss.Color
}
