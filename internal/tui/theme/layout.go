package theme

import "github.com/charmbracelet/lipgloss"

func applyFrameStyles(s *Set, p palette) {
	s.MainFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.mainFrameBorderFg).Padding(0, 1)
	s.SubpageFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.subpageFrameBorderFg).Padding(0, 1)
}
