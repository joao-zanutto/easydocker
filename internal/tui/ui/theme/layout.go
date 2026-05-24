package theme

import "charm.land/lipgloss/v2"

func applyFrameStyles(s *Set) {
	s.MainFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Periwinkle).Padding(0, 1)
	s.SubpageFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BlueGray).Padding(0, 1)
}
