package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	page         lipgloss.Style
	header       lipgloss.Style
	title        lipgloss.Style
	titleMeta    lipgloss.Style
	breadcrumb   lipgloss.Style
	tab          lipgloss.Style
	activeTab    lipgloss.Style
	badge        lipgloss.Style
	muted        lipgloss.Style
	mainFrame    lipgloss.Style
	subpageFrame lipgloss.Style
	divider      lipgloss.Style
	headerRow    lipgloss.Style
	row          lipgloss.Style
	activeRow    lipgloss.Style
	section      lipgloss.Style
	label        lipgloss.Style
	value        lipgloss.Style
	errorText    lipgloss.Style
	footer       lipgloss.Style
	key          lipgloss.Style
	keyText      lipgloss.Style
	followOn     lipgloss.Style
	followOff    lipgloss.Style
	monitorBox   lipgloss.Style
	stateRun     lipgloss.Style
	stateWarn    lipgloss.Style
	stateStop    lipgloss.Style
	stateDead    lipgloss.Style
	stateOther   lipgloss.Style
	activeBG     lipgloss.Color
}

func defaultStyles() styles {
	activeBG := lipgloss.Color("236")
	return styles{
		page:         lipgloss.NewStyle(),
		header:       lipgloss.NewStyle().Padding(1, 1, 0, 1),
		title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1),
		titleMeta:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("24")).Padding(0, 1),
		breadcrumb:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("247")),
		tab:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1),
		activeTab:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Underline(true).Padding(0, 1),
		badge:        lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("60")).Padding(0, 1),
		muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		mainFrame:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("67")).Padding(0, 1),
		subpageFrame: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("110")).Padding(0, 1),
		divider:      lipgloss.NewStyle().Foreground(lipgloss.Color("60")),
		headerRow:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")),
		row:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		activeRow:    lipgloss.NewStyle().Bold(true).Background(activeBG),
		section:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186")),
		label:        lipgloss.NewStyle().Foreground(lipgloss.Color("109")),
		value:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		errorText:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		footer:       lipgloss.NewStyle().Padding(0, 1, 1, 1).Foreground(lipgloss.Color("248")),
		key:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("31")).Padding(0, 1),
		keyText:      lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		followOn:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		followOff:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		monitorBox:   lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("60")).Padding(0, 1),
		stateRun:     lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		stateWarn:    lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		stateStop:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		stateDead:    lipgloss.NewStyle().Foreground(lipgloss.Color("199")),
		stateOther:   lipgloss.NewStyle().Foreground(lipgloss.Color("110")),
		activeBG:     activeBG,
	}
}
