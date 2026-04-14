package chrome

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/util"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type TabSpec struct {
	Tab   int
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
	ActiveTab        int
	ShowAll          bool
	Err              error
	Tabs             []TabSpec
	Styles           HeaderStyles
	RenderTab        func(tab int, label string) string
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

var (
	browseFooterBaseHelp = []key.Binding{
		newFooterBinding([]string{"up", "down"}, helpKeyLabel("↑/↓"), "navigate"),
		newFooterBinding([]string{"left", "right"}, helpKeyLabel("←/→"), "switch tabs"),
		newFooterBinding([]string{"esc"}, helpKeyLabel("esc"), "quit"),
	}
	browseContainerFooterHelp = []key.Binding{
		newFooterBinding([]string{"a"}, helpKeyLabel("a"), "toggle running/all"),
		newFooterBinding([]string{"enter"}, helpKeyLabel("enter"), "logs"),
	}
	logsFooterHelp = []key.Binding{
		newFooterBinding([]string{"left", "up", "down", "right"}, helpKeyLabel("← ↑ ↓ →"), "navigate"),
		newFooterBinding([]string{"pgup", "pgdn"}, helpKeyLabel("pgup/dn"), "jump up/down"),
		newFooterBinding([]string{"home", "end"}, helpKeyLabel("home/end"), "go to top/bottom"),
		newFooterBinding([]string{"f"}, helpKeyLabel("f"), "toggle follow"),
		newFooterBinding([]string{"esc"}, helpKeyLabel("esc"), "back"),
	}
)

func RenderHeaderTabs(specs []TabSpec, maxWidth int, renderTab func(tab int, label string) string) []string {
	for _, variant := range []tabLabelVariant{tabLabelFullWithParens, tabLabelFullCompact, tabLabelIconWithCount} {
		tabs := renderHeaderTabsVariant(specs, variant, renderTab)
		if joinedDisplayWidth(tabs) <= maxWidth {
			return tabs
		}
	}
	return renderHeaderTabsVariant(specs, tabLabelIconOnly, renderTab)
}

func RenderScopeBadge(showAll bool, maxWidth int, renderBadge func(string) string) string {
	scope := "running"
	if showAll {
		scope = "all"
	}
	labels := []string{"container scope: " + scope, "scope: " + scope, scope}
	for _, label := range labels {
		badge := renderBadge(label)
		if util.DisplayWidth(badge) <= max(1, maxWidth) {
			return badge
		}
	}
	return renderBadge(scope)
}

func NewFooterKeyMap(isLogsScreen bool, isContainersTab bool) help.KeyMap {
	if isLogsScreen {
		return footerKeyMap{bindings: logsFooterHelp}
	}
	bindings := append([]key.Binding{}, browseFooterBaseHelp...)
	if isContainersTab {
		bindings = append(bindings, browseContainerFooterHelp...)
	}
	return footerKeyMap{bindings: bindings}
}

func RenderHeader(input HeaderInput) string {
	innerWidth := max(1, input.Width-input.Styles.Header.GetHorizontalFrameSize())
	totalsText := input.TotalsText
	if input.LoadingStageText != "" {
		totalsText += " " + input.LoadingStageText
	}
	tabs := RenderHeaderTabs(input.Tabs, max(1, innerWidth-6), input.RenderTab)
	tabsText := strings.Join(tabs, " ")
	tabsWidth := util.DisplayWidth(tabsText)
	leftAvail := max(1, innerWidth-tabsWidth-1)
	firstLeft := input.Styles.Title.Render(util.ConstrainLine(input.Title, max(1, leftAvail-input.Styles.Title.GetHorizontalFrameSize())))
	firstLine := renderEdgeAlignedLine(firstLeft, tabsText, innerWidth)

	secondRight := ""
	if input.ActiveTab == 0 {
		secondRight = RenderScopeBadge(input.ShowAll, max(1, innerWidth/3), func(label string) string {
			return input.Styles.Badge.Render(label)
		})
	}
	secondLine := renderPinnedHeaderLine(input.Styles.TitleMeta, totalsText, secondRight, innerWidth)
	line := lipgloss.JoinVertical(lipgloss.Left, firstLine, secondLine)
	if input.Err != nil {
		line = lipgloss.JoinVertical(lipgloss.Left, line, util.ConstrainLine(input.Styles.ErrorText.Render(input.Err.Error()), innerWidth))
	}
	return input.Styles.Header.Render(line)
}

func RenderFooter(input FooterInput) string {
	innerWidth := max(1, input.Width-input.Styles.Footer.GetHorizontalFrameSize())
	helpModel := help.New()
	helpModel.Width = innerWidth
	helpModel.ShortSeparator = "   "
	helpModel.Ellipsis = "…"
	helpModel.Styles = help.Styles{
		ShortKey:       input.Styles.Key.Padding(0),
		ShortDesc:      input.Styles.KeyText,
		ShortSeparator: lipgloss.NewStyle(),
		FullKey:        input.Styles.Key.Padding(0),
		FullDesc:       input.Styles.KeyText,
		FullSeparator:  lipgloss.NewStyle(),
		Ellipsis:       lipgloss.NewStyle(),
	}
	line := util.ConstrainLine(helpModel.View(input.KeyMap), innerWidth)
	return input.Styles.Footer.Render(lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, line))
}

type footerKeyMap struct {
	bindings []key.Binding
}

func (m footerKeyMap) ShortHelp() []key.Binding {
	return m.bindings
}

func (m footerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.bindings}
}

func newFooterBinding(keys []string, helpKey, description string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(helpKey, description))
}

func helpKeyLabel(label string) string {
	return " " + label + " "
}

func RenderTotalsLabel(snapshot core.Snapshot, loadingStage, loadStageIdle, loadStageMetrics int, metricsLoaded bool, loadingIndicator string) string {
	if !metricsLoaded && (loadingStage == loadStageMetrics || (loadingStage != loadStageIdle && snapshot.TotalCPU == 0 && snapshot.TotalMem == 0)) {
		indicator := loadingIndicator
		if strings.TrimSpace(indicator) == "" {
			indicator = "-"
		}
		return fmt.Sprintf("CPU %s  MEM %s", indicator, indicator)
	}
	mem := core.HumanBytes(int64(snapshot.TotalMem))
	if snapshot.TotalLimit > 0 {
		return fmt.Sprintf("CPU %s  MEM %s", util.RenderPercent(snapshot.TotalCPU), util.FormatMemoryUsage(mem, (float64(snapshot.TotalMem)/float64(snapshot.TotalLimit))*100, core.HumanBytes(int64(snapshot.TotalLimit))))
	}
	return fmt.Sprintf("CPU %s  MEM %s", util.RenderPercent(snapshot.TotalCPU), mem)
}

func RenderLoadingStageLabel(loadingStage, loadStageContainers, loadStageResources, loadStageMetrics int, metricsLoaded bool) string {
	switch loadingStage {
	case loadStageContainers:
		return "loading containers"
	case loadStageResources:
		return "loading resources"
	case loadStageMetrics:
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

func renderCenteredLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = util.ConstrainLine(line, width)
	lineWidth := util.DisplayWidth(line)
	if lineWidth >= width {
		return util.ClampSingleLine(line, width)
	}
	leftPad := (width - lineWidth) / 2
	return util.ClampSingleLine(strings.Repeat(" ", leftPad)+line, width)
}

func renderPinnedHeaderLine(leftStyle lipgloss.Style, leftText, right string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.TrimSpace(right) == "" {
		contentWidth := max(1, width-leftStyle.GetHorizontalFrameSize())
		return util.ClampSingleLine(leftStyle.Render(util.ConstrainLine(leftText, contentWidth)), width)
	}
	rightWidth := util.DisplayWidth(right)
	leftTotalWidth := max(1, width-rightWidth-1)
	leftContentWidth := max(1, leftTotalWidth-leftStyle.GetHorizontalFrameSize())
	left := leftStyle.Render(util.ConstrainLine(leftText, leftContentWidth))
	leftWidth := util.DisplayWidth(left)
	spacing := max(1, width-leftWidth-rightWidth)
	return util.ClampSingleLine(left+strings.Repeat(" ", spacing)+right, width)
}

func renderHeaderTabsVariant(specs []TabSpec, variant tabLabelVariant, renderTab func(tab int, label string) string) []string {
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
			total++
		}
	}
	return total
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
