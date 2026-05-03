package logs

import (
	"fmt"
	"strings"

	"easydocker/internal/tui/components"
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

type ViewStyles struct {
	Breadcrumb   lipgloss.Style
	FollowOn     lipgloss.Style
	FollowOff    lipgloss.Style
	Muted        lipgloss.Style
	Divider      lipgloss.Style
	SubpageFrame lipgloss.Style
}

type ViewModel struct {
	State            State
	ContainerName    string
	LoadingIndicator string
	Width            int
	Height           int
	Styles           ViewStyles
}

const filterHeaderHeight = 2

func RenderContent(vm ViewModel) string {
	if vm.Width == 0 || vm.Height == 0 {
		return ""
	}
	layout := util.ComputeFrameLayout(vm.Width, vm.Height, vm.Styles.SubpageFrame)
	headerVM := vm
	headerVM.Width = layout.ContentWidth
	breadcrumb := util.ClampSingleLine("Containers / "+vm.ContainerName+" / Logs", layout.ContentWidth)
	logsHeight := VisibleRowsForContent(layout.ContentHeight, vm.State.Filter.Active)
	logList := FilterLogLines(vm.State.Data.Logs, vm.State.Filter.Query)
	start, end := VisibleLogRange(vm.State, logList)
	headline := RenderHeader(headerVM, breadcrumb, len(logList), start, end)
	logsPanel := RenderPanel(vm, layout.ContentWidth, logsHeight)

	if vm.State.Filter.Active {
		filterInput := vm.State.Filter.Input
		filterInput.SetWidth(components.DynamicInputWidth(filterInput.Prompt, layout.ContentWidth))
		filterHeader := renderFilterHeader(filterInput.View(), layout.ContentWidth, vm.Styles.Divider)
		return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(headline, filterHeader, logsPanel))
	}

	headerDivider := components.RenderTitleDivider(vm.Styles.Divider, layout.ContentWidth)
	return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(headline, headerDivider, logsPanel))
}

func VisibleRowsForContent(contentHeight int, filterActive bool) int {
	overhead := 2 // header + divider + frame spacing behavior
	if filterActive {
		overhead += filterHeaderHeight
	}
	return max(1, contentHeight-overhead)
}

func RenderHeader(vm ViewModel, breadcrumb string, total, start, end int) string {
	follow := "off"
	if vm.State.Follow {
		follow = "on"
	}
	wrap := "off"
	if vm.State.WrapLines {
		wrap = "on"
	}
	first := 0
	last := 0
	if total > 0 {
		first = start + 1
		last = max(first, end)
	}
	left := vm.Styles.Breadcrumb.Render(breadcrumb)
	followText := vm.Styles.FollowOff.Render(follow)
	if vm.State.Follow {
		followText = vm.Styles.FollowOn.Render(follow)
	}
	wrapText := vm.Styles.FollowOff.Render(wrap)
	if vm.State.WrapLines {
		wrapText = vm.Styles.FollowOn.Render(wrap)
	}
	right := lipgloss.JoinHorizontal(
		lipgloss.Left,
		vm.Styles.Muted.Render("wrap:"),
		wrapText,
		vm.Styles.Muted.Render("  "),
		vm.Styles.Muted.Render("follow:"),
		followText,
		vm.Styles.Muted.Render(fmt.Sprintf("  lines:(%d-%d/%d)", first, last, total)),
	)
	return renderRightPriorityLine(left, right, vm.Width)
}

func RenderPanel(vm ViewModel, width, height int) string {
	contentWidth := max(1, width-2)
	if vm.State.InitialLoad {
		return strings.Join(util.ClipAndPadLines([]string{renderLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator)}, height, ""), "\n")
	}

	logList := FilterLogLines(vm.State.Data.Logs, vm.State.Filter.Query)
	if len(logList) == 0 {
		empty := "No logs found for this container."
		if strings.TrimSpace(vm.State.Filter.Query) != "" {
			empty = "No log lines match current filter."
		}
		return strings.Join(util.ClipAndPadLines([]string{util.ClampSingleLine(vm.Styles.Muted.Render(empty), contentWidth)}, height, ""), "\n")
	}

	renderLines := make([]string, 0, len(logList))
	for _, line := range logList {
		renderLines = append(renderLines, SanitizeLogRenderLine(line))
	}

	lines := strings.Split(vm.State.Viewport.View(), "\n")
	lines = renderHorizontalScrollIndicators(vm.State, lines, renderLines, max(1, vm.State.Viewport.Width()), vm.Styles.Muted.Reverse(true))
	if vm.State.HistoryLoad {
		lines = append([]string{renderHistoryLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator)}, lines...)
	}
	lines = util.ClipAndPadLines(lines, height, "")
	return strings.Join(lines, "\n")
}

func renderFilterHeader(input string, width int, dividerStyle lipgloss.Style) string {
	return components.RenderFilterHeader(input, width, dividerStyle)
}

func padVisibleWidth(line string, width int) string {
	return components.PadVisibleWidth(line, width)
}

func renderLoadingLine(style lipgloss.Style, width int, indicator string) string {
	prefix := strings.TrimSpace(indicator)
	if prefix != "" {
		prefix += " "
	}
	return util.ClampSingleLine(style.Render(prefix+"Loading logs..."), width)
}

func renderHistoryLoadingLine(style lipgloss.Style, width int, indicator string) string {
	prefix := strings.TrimSpace(indicator)
	if prefix != "" {
		prefix += " "
	}
	return util.ClampSingleLine(style.Render(prefix+"Loading older logs..."), width)
}

func renderHorizontalScrollIndicators(state State, lines, renderLines []string, width int, indicatorStyle lipgloss.Style) []string {
	if state.WrapLines || width <= 0 || len(lines) == 0 || len(renderLines) == 0 {
		return lines
	}

	start, end := ViewportRange(state, len(renderLines))
	visible := renderLines[start:end]
	if len(visible) == 0 {
		return lines
	}

	xOffset := max(0, state.Viewport.XOffset())
	anyCanScrollLeft := xOffset > 0
	out := append([]string(nil), lines...)
	for i := 0; i < len(out) && i < len(visible); i++ {
		lineWidth := util.DisplayWidth(visible[i])
		maxOffset := max(0, lineWidth-width)
		canScrollRight := xOffset < maxOffset
		if !anyCanScrollLeft && !canScrollRight {
			continue
		}
		out[i] = applyScrollIndicator(out[i], width, anyCanScrollLeft, canScrollRight, indicatorStyle)
	}

	return out
}

func applyScrollIndicator(line string, width int, canScrollLeft, canScrollRight bool, style lipgloss.Style) string {
	if width <= 0 {
		return ""
	}

	left := style.Render("<")
	right := style.Render(">")
	if canScrollLeft && canScrollRight {
		middle := padVisibleWidth(util.ClampSingleLine(line, max(0, width-2)), max(0, width-2))
		return left + middle + right
	}
	if canScrollLeft {
		rest := padVisibleWidth(util.ClampSingleLine(line, max(0, width-1)), max(0, width-1))
		return left + rest
	}
	prefix := padVisibleWidth(util.ClampSingleLine(line, max(0, width-1)), max(0, width-1))
	return prefix + right
}

func renderRightPriorityLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	right = util.ClampSingleLine(right, width)
	rightWidth := util.DisplayWidth(right)
	if rightWidth >= width {
		return right
	}
	leftWidth := max(1, width-rightWidth-1)
	left = util.ClampSingleLine(left, leftWidth)
	leftRenderedWidth := util.DisplayWidth(left)
	spacing := max(0, width-leftRenderedWidth-rightWidth)
	return left + strings.Repeat(" ", spacing) + right
}
