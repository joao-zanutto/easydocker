package theme

import "github.com/charmbracelet/lipgloss"

func Default() Set {
	return defaultForBackground(lipgloss.HasDarkBackground())
}

type palette struct {
	activeBG              lipgloss.Color
	headerTitleFg         lipgloss.Color
	headerTitleBg         lipgloss.Color
	headerTitleMetaFg     lipgloss.Color
	tabFg                 lipgloss.Color
	activeTabFg           lipgloss.Color
	badgeFg               lipgloss.Color
	badgeBg               lipgloss.Color
	footerFg              lipgloss.Color
	keyFg                 lipgloss.Color
	keyBg                 lipgloss.Color
	keyTextFg             lipgloss.Color
	errorTextFg           lipgloss.Color
	sectionFg             lipgloss.Color
	labelFg               lipgloss.Color
	valueFg               lipgloss.Color
	mutedFg               lipgloss.Color
	dividerFg             lipgloss.Color
	stateRunFg            lipgloss.Color
	stateWarnFg           lipgloss.Color
	stateStopFg           lipgloss.Color
	stateDeadFg           lipgloss.Color
	stateOtherFg          lipgloss.Color
	breadcrumbFg          lipgloss.Color
	followOnFg            lipgloss.Color
	followOffFg           lipgloss.Color
	monitorBorderFg       lipgloss.Color
	headerRowFg           lipgloss.Color
	rowFg                 lipgloss.Color
	mainFrameBorderFg     lipgloss.Color
	subpageFrameBorderFg  lipgloss.Color
}

func defaultForBackground(hasDarkBackground bool) Set {
	p := darkPalette()
	if !hasDarkBackground {
		p = lightPalette()
	}

	s := Set{Page: lipgloss.NewStyle(), ActiveBG: p.activeBG}
	applyChromeStyles(&s, p)
	applyBrowseStyles(&s, p)
	applyLogsStyles(&s, p)
	applyTableStyles(&s, p)
	applyFrameStyles(&s, p)
	return s
}

func darkPalette() palette {
	return palette{
		activeBG:             lipgloss.Color("236"),
		headerTitleFg:         lipgloss.Color("230"),
		headerTitleBg:         lipgloss.Color("24"),
		headerTitleMetaFg:     lipgloss.Color("252"),
		tabFg:                 lipgloss.Color("252"),
		activeTabFg:           lipgloss.Color("86"),
		badgeFg:               lipgloss.Color("229"),
		badgeBg:               lipgloss.Color("60"),
		footerFg:              lipgloss.Color("248"),
		keyFg:                 lipgloss.Color("230"),
		keyBg:                 lipgloss.Color("31"),
		keyTextFg:             lipgloss.Color("248"),
		errorTextFg:           lipgloss.Color("203"),
		sectionFg:             lipgloss.Color("186"),
		labelFg:               lipgloss.Color("109"),
		valueFg:               lipgloss.Color("252"),
		mutedFg:               lipgloss.Color("244"),
		dividerFg:             lipgloss.Color("60"),
		stateRunFg:            lipgloss.Color("42"),
		stateWarnFg:           lipgloss.Color("214"),
		stateStopFg:           lipgloss.Color("203"),
		stateDeadFg:           lipgloss.Color("199"),
		stateOtherFg:          lipgloss.Color("110"),
		breadcrumbFg:          lipgloss.Color("247"),
		followOnFg:            lipgloss.Color("252"),
		followOffFg:           lipgloss.Color("244"),
		monitorBorderFg:       lipgloss.Color("60"),
		headerRowFg:           lipgloss.Color("230"),
		rowFg:                 lipgloss.Color("252"),
		mainFrameBorderFg:     lipgloss.Color("67"),
		subpageFrameBorderFg:  lipgloss.Color("110"),
	}
}

func lightPalette() palette {
	return palette{
		activeBG:             lipgloss.Color("189"),
		headerTitleFg:         lipgloss.Color("231"),
		headerTitleBg:         lipgloss.Color("25"),
		headerTitleMetaFg:     lipgloss.Color("237"),
		tabFg:                 lipgloss.Color("237"),
		activeTabFg:           lipgloss.Color("25"),
		badgeFg:               lipgloss.Color("234"),
		badgeBg:               lipgloss.Color("153"),
		footerFg:              lipgloss.Color("239"),
		keyFg:                 lipgloss.Color("234"),
		keyBg:                 lipgloss.Color("153"),
		keyTextFg:             lipgloss.Color("239"),
		errorTextFg:           lipgloss.Color("160"),
		sectionFg:             lipgloss.Color("94"),
		labelFg:               lipgloss.Color("25"),
		valueFg:               lipgloss.Color("235"),
		mutedFg:               lipgloss.Color("245"),
		dividerFg:             lipgloss.Color("250"),
		stateRunFg:            lipgloss.Color("28"),
		stateWarnFg:           lipgloss.Color("166"),
		stateStopFg:           lipgloss.Color("160"),
		stateDeadFg:           lipgloss.Color("161"),
		stateOtherFg:          lipgloss.Color("25"),
		breadcrumbFg:          lipgloss.Color("237"),
		followOnFg:            lipgloss.Color("234"),
		followOffFg:           lipgloss.Color("245"),
		monitorBorderFg:       lipgloss.Color("250"),
		headerRowFg:           lipgloss.Color("235"),
		rowFg:                 lipgloss.Color("236"),
		mainFrameBorderFg:     lipgloss.Color("247"),
		subpageFrameBorderFg:  lipgloss.Color("250"),
	}
}
