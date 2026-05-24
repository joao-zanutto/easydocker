package theme

import "charm.land/lipgloss/v2"

func applyChromeStyles(s *Set) {
	s.Header = lipgloss.NewStyle().Padding(0, 1, 0, 1)
	s.Title = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SurfaceBG).Padding(0, 1)
	s.TitleMeta = lipgloss.NewStyle().Foreground(TextSecondary).Background(SurfaceBG).Padding(0, 1)
	s.Tab = lipgloss.NewStyle().Foreground(TextSecondary).Padding(0, 1)
	s.ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(Cyan).Underline(true).Padding(0, 1)
	s.Badge = lipgloss.NewStyle().Foreground(LightYellow).Background(MutedPurple).Padding(0, 1)
	s.Footer = lipgloss.NewStyle().Padding(0, 0, 0, 0).Foreground(TextDim)
	s.Key = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SteelBlue).Padding(0, 1)
	s.KeyText = lipgloss.NewStyle().Foreground(TextDim)
	s.ErrorText = lipgloss.NewStyle().Foreground(Red)
}
