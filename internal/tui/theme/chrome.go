package theme

import "github.com/charmbracelet/lipgloss"

func applyChromeStyles(s *Set, p palette) {
	s.Header = lipgloss.NewStyle().Padding(1, 1, 0, 1)
	s.Title = lipgloss.NewStyle().Bold(true).Foreground(p.headerTitleFg).Background(p.headerTitleBg).Padding(0, 1)
	s.TitleMeta = lipgloss.NewStyle().Foreground(p.headerTitleMetaFg).Background(p.headerTitleBg).Padding(0, 1)
	s.Tab = lipgloss.NewStyle().Foreground(p.tabFg).Padding(0, 1)
	s.ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(p.activeTabFg).Underline(true).Padding(0, 1)
	s.Badge = lipgloss.NewStyle().Foreground(p.badgeFg).Background(p.badgeBg).Padding(0, 1)
	s.Footer = lipgloss.NewStyle().Padding(0, 1, 1, 1).Foreground(p.footerFg)
	s.Key = lipgloss.NewStyle().Bold(true).Foreground(p.keyFg).Background(p.keyBg).Padding(0, 1)
	s.KeyText = lipgloss.NewStyle().Foreground(p.keyTextFg)
	s.ErrorText = lipgloss.NewStyle().Foreground(p.errorTextFg)
}
