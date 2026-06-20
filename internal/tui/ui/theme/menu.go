package theme

import "charm.land/lipgloss/v2"

func applyMenuStyles(s *Set) {
	s.Menu.Frame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Periwinkle).Background(SurfaceDeep).Padding(1, 2)
	s.Menu.Selector = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Menu.ItemNormal = lipgloss.NewStyle().Bold(true).Foreground(TextMuted)
	s.Menu.ItemDescription = lipgloss.NewStyle().Bold(true).Foreground(TextSilver)
	s.Menu.HelpFrame = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BlueGray).Background(SurfaceDeep).Padding(1, 2)
	s.Menu.HelpTitle = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	s.Menu.HelpSection = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	s.Menu.HelpCommand = lipgloss.NewStyle().Foreground(TextSecondary)
	s.Menu.HelpKey = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary)
	s.Menu.HelpFooter = lipgloss.NewStyle().Foreground(TextMuted)
	s.Menu.Scrollbar = lipgloss.NewStyle().Foreground(BlueGray)
	s.Menu.HelpContext = lipgloss.NewStyle().Foreground(BlueGray)
}
