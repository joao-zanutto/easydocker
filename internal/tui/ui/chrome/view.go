package chrome

import (
	"fmt"
	"math"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

func formatMemoryUsage(usage string, percent float64, limit string) string {
	if usage == "-" {
		return "-"
	}
	if limit != "" && limit != "-" {
		return fmt.Sprintf("%s / %s (%s)", usage, limit, renderPercent(percent))
	}
	return fmt.Sprintf("%s (%s)", usage, renderPercent(percent))
}

func renderPercent(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", value)
}

type TabSpec struct {
	Tab   shared.Tab
	Icon  string
	Name  string
	Count int
}

type HeaderStyles struct {
	Header    lipgloss.Style
	Title     lipgloss.Style
	TitleMeta lipgloss.Style
	Badge     lipgloss.Style
	ErrorText lipgloss.Style
	Key       lipgloss.Style
	KeyText   lipgloss.Style
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
}

type FooterStyles struct {
	Footer  lipgloss.Style
	Key     lipgloss.Style
	KeyText lipgloss.Style
}

type HeaderInput struct {
	Width            int
	Title            string
	TotalsText       string
	LoadingStageText string
	ActiveTab        shared.Tab
	ShowAll          bool
	HideScope        bool
	HideScopeKey     bool
	DimTabs          bool
	Err              error
	Tabs             []TabSpec
	Styles           HeaderStyles
	RenderTab        func(tab shared.Tab, label string) string
}

type FooterInput struct {
	Width  int
	KeyMap help.KeyMap
	Styles FooterStyles
}

type tabLabelVariant int

const (
	tabLabelFullWithParens tabLabelVariant = iota
	tabLabelFullCompact
	tabLabelIconWithCount
	tabLabelIconOnly
)

func RenderHeaderTabs(specs []TabSpec, maxWidth int, renderTab func(tab shared.Tab, label string) string) []string {
	for _, variant := range []tabLabelVariant{tabLabelFullWithParens, tabLabelFullCompact, tabLabelIconWithCount} {
		tabs := renderHeaderTabsVariant(specs, variant, renderTab)
		if joinedDisplayWidth(tabs) <= maxWidth {
			return tabs
		}
	}
	return renderHeaderTabsVariant(specs, tabLabelIconOnly, renderTab)
}

func ScopeLabel(showAll bool) string {
	if showAll {
		return "all"
	}
	return "running"
}

func RenderHeader(input HeaderInput) string {
	innerWidth := max(1, input.Width-input.Styles.Header.GetHorizontalFrameSize())

	tabs := RenderHeaderTabs(input.Tabs, max(1, innerWidth-6), input.RenderTab)
	if !input.HideScope && input.ActiveTab == shared.TabContainers {
		scope := ScopeLabel(input.ShowAll)
		scopeRendered := input.Styles.Badge.Render(fmt.Sprintf("scope:%s", scope))
		var prefix string
		if input.HideScopeKey {
			prefix = scopeRendered
		} else {
			keyRendered := input.Styles.Key.Render("a")
			prefix = keyRendered + scopeRendered
		}
		for i, tab := range input.Tabs {
			if tab.Tab == shared.TabContainers {
				tabs[i] = prefix + " " + tabs[i]
				break
			}
		}
	}
	tabsText := strings.Join(tabs, " │ ")
	tabsWidth := util.DisplayWidth(tabsText)

	leftAvail := max(1, innerWidth-tabsWidth-1)
	titleRendered := input.Styles.Title.Render(input.Title)
	titleWidth := util.DisplayWidth(titleRendered)

	var left string
	if input.TotalsText != "" || input.LoadingStageText != "" {
		metaAvail := max(1, leftAvail-titleWidth)
		metaContentWidth := max(1, metaAvail-input.Styles.TitleMeta.GetHorizontalFrameSize())

		metaText := " " + input.TotalsText
		if input.LoadingStageText != "" {
			stageText := " " + input.LoadingStageText
			combined := metaText + stageText
			if util.DisplayWidth(metaText)+util.DisplayWidth(stageText) > metaContentWidth+5 {
				metaText = stageText
			} else {
				metaText = combined
			}
		}
		metaRendered := input.Styles.TitleMeta.Render(util.ConstrainLine(metaText, metaContentWidth))
		left = titleRendered + metaRendered
	} else {
		titleContentWidth := max(1, leftAvail-input.Styles.Title.GetHorizontalFrameSize())
		left = input.Styles.Title.Render(util.ConstrainLine(input.Title, titleContentWidth))
	}
	line := renderEdgeAlignedLine(left, tabsText, innerWidth)

	if input.DimTabs && tabsWidth > 0 {
		dimStart := max(0, innerWidth-tabsWidth)
		line = dimLineRight(line, dimStart, innerWidth)
	}

	if input.Err != nil {
		line = lipgloss.JoinVertical(lipgloss.Left, line, util.ConstrainLine(input.Styles.ErrorText.Render(input.Err.Error()), innerWidth))
	}
	return input.Styles.Header.Render(line)
}

func dimLineRight(line string, dimStart, fullWidth int) string {
	if fullWidth <= 0 || dimStart >= fullWidth {
		return line
	}
	bgLayer := lipgloss.NewLayer(line)
	comp := lipgloss.NewCompositor(bgLayer)
	bounds := comp.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return line
	}
	canvas := lipgloss.NewCanvas(bounds.Dx(), bounds.Dy()).Compose(comp)
	fg := lipgloss.Color("244")
	bg := lipgloss.Color("233")
	for y := 0; y < bounds.Dy(); y++ {
		for x := dimStart; x < bounds.Dx(); x++ {
			cell := canvas.CellAt(x, y)
			if cell == nil || cell.IsZero() {
				continue
			}
			cell = cell.Clone()
			cell.Style.Fg = fg
			cell.Style.Bg = bg
			canvas.SetCell(x, y, cell)
		}
	}
	return canvas.Render()
}

func RenderFooter(input FooterInput) string {
	innerWidth := max(1, input.Width-input.Styles.Footer.GetHorizontalFrameSize())
	helpModel := help.New()
	helpModel.SetWidth(innerWidth)
	helpModel.ShortSeparator = "   "
	helpModel.Ellipsis = "…"
	helpModel.Styles = help.Styles{
		ShortKey:       input.Styles.Key,
		ShortDesc:      input.Styles.KeyText,
		ShortSeparator: lipgloss.NewStyle(),
		FullKey:        input.Styles.Key,
		FullDesc:       input.Styles.KeyText,
		FullSeparator:  lipgloss.NewStyle(),
		Ellipsis:       lipgloss.NewStyle(),
	}
	line := util.ConstrainLine(helpModel.View(input.KeyMap), innerWidth)
	return input.Styles.Footer.Render(lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, line))
}

func RenderTotalsLabel(snapshot core.Snapshot, loadingStage shared.Stage, metricsLoaded bool, loadingIndicator string) string {
	if !metricsLoaded && (loadingStage == shared.StageMetrics || (loadingStage != shared.StageIdle && snapshot.TotalCPU == 0 && snapshot.TotalMem == 0)) {
		indicator := loadingIndicator
		if strings.TrimSpace(indicator) == "" {
			indicator = "-"
		}
		return fmt.Sprintf("\x1b[1mCPU\x1b[22m %6s  \x1b[1mMEM\x1b[22m %6s", indicator, indicator)
	}
	return fmt.Sprintf("\x1b[1mCPU\x1b[22m %6s  \x1b[1mMEM\x1b[22m %s", renderPercent(snapshot.TotalCPU), memTotal(snapshot))
}

func memTotal(snapshot core.Snapshot) string {
	mem := core.HumanBytes(snapshot.TotalMem)
	if snapshot.TotalLimit > 0 {
		return formatMemoryUsage(mem, (float64(snapshot.TotalMem)/float64(snapshot.TotalLimit))*100, core.HumanBytes(snapshot.TotalLimit))
	}
	return mem
}

func RenderLoadingStageLabel(loadingStage shared.Stage, metricsLoaded bool) string {
	switch loadingStage {
	case shared.StageContainers:
		return "loading containers"
	case shared.StageResources:
		return "loading resources"
	case shared.StageMetrics:
		if metricsLoaded {
			return ""
		}
		return "loading metrics"
	default:
		return ""
	}
}

func renderEdgeAlignedLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.TrimSpace(right) == "" {
		return util.ClampSingleLine(util.ConstrainLine(left, width), width)
	}
	leftWidth := util.DisplayWidth(left)
	rightWidth := util.DisplayWidth(right)
	if leftWidth+rightWidth+1 > width {
		return util.ClampSingleLine(util.ConstrainLine(left+" "+right, width), width)
	}
	return util.ClampSingleLine(left+strings.Repeat(" ", width-leftWidth-rightWidth)+right, width)
}

func renderHeaderTabsVariant(specs []TabSpec, variant tabLabelVariant, renderTab func(tab shared.Tab, label string) string) []string {
	tabs := make([]string, 0, len(specs))
	for _, spec := range specs {
		tabs = append(tabs, renderTab(spec.Tab, headerTabLabel(spec, variant)))
	}
	return tabs
}

func headerTabLabel(spec TabSpec, variant tabLabelVariant) string {
	switch variant {
	case tabLabelFullWithParens:
		return fmt.Sprintf("%s %s (%d)", spec.Icon, spec.Name, spec.Count)
	case tabLabelFullCompact:
		return fmt.Sprintf("%s %s %d", spec.Icon, spec.Name, spec.Count)
	case tabLabelIconWithCount:
		return fmt.Sprintf("%s %d", spec.Icon, spec.Count)
	default:
		return spec.Icon
	}
}

func joinedDisplayWidth(parts []string) int {
	total := 0
	for i, part := range parts {
		total += util.DisplayWidth(part)
		if i > 0 {
			total += 3
		}
	}
	return total
}
