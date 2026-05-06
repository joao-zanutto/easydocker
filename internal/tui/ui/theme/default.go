package theme

import "charm.land/lipgloss/v2"

func Default() Set {
	activeBG := lipgloss.Color("236")
	s := Set{Page: lipgloss.NewStyle(), ActiveBG: activeBG}
	applyChromeStyles(&s)
	applyBrowseStyles(&s)
	applyLogsStyles(&s)
	applyTableStyles(&s)
	applyFrameStyles(&s)
	return s
}
