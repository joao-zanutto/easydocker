package theme

import "charm.land/lipgloss/v2"

func applyChromeStyles(s *Set) {
	s.Chrome.Header = lipgloss.NewStyle().Padding(0, 1, 0, 1)
	s.Chrome.Title = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SurfaceBG).Padding(0, 1)
	s.Chrome.TitleMeta = lipgloss.NewStyle().Foreground(TextSecondary).Padding(0, 1)
	s.Chrome.Badge = lipgloss.NewStyle().Foreground(LightYellow).Padding(0, 1)
	s.Chrome.ErrorText = lipgloss.NewStyle().Foreground(Red)
	s.Chrome.Footer = lipgloss.NewStyle().Padding(0, 0, 0, 0).Foreground(TextDim)
	s.Chrome.Key = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SteelBlue).Padding(0, 1)
	s.Chrome.KeyText = lipgloss.NewStyle().Foreground(TextDim)
	s.Tabs.Tab = lipgloss.NewStyle().Foreground(TextSecondary).Padding(0, 1)
	s.Tabs.ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(TextPrimary).Background(SelectorPurple).Padding(0, 1)
}
