package theme

import "charm.land/lipgloss/v2"

func Default() Set {
	s := Set{Chrome: ChromeStyles{Page: lipgloss.NewStyle()}, ActiveBG: Surface}
	applyChromeStyles(&s)
	applyBrowseStyles(&s)
	applyLogsStyles(&s)
	applyTableStyles(&s)
	applyFrameStyles(&s)
	applyMenuStyles(&s)
	return s
}
