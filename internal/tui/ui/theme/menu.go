package theme

import "charm.land/lipgloss/v2"

func applyMenuStyles(s *Set) {
	s.MenuFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Periwinkle).Background(SurfaceDeep).Padding(1, 2)
	s.MenuSelector = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.MenuItem = lipgloss.NewStyle().Bold(true).Foreground(TextMuted)
	s.MenuDesc = lipgloss.NewStyle().Bold(true).Foreground(TextSilver)
	s.HelpFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BlueGray).Background(SurfaceDeep).Padding(1, 2)
	s.HelpTitle = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.HelpSection = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	s.HelpCommand = lipgloss.NewStyle().Foreground(TextSecondary)
	s.HelpKey = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.HelpFooter = lipgloss.NewStyle().Foreground(TextMuted)
	s.Scrollbar = lipgloss.NewStyle().Foreground(BlueGray)
	s.HelpContext = lipgloss.NewStyle().Foreground(BlueGray)
}
