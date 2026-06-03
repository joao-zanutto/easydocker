package viewer

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/ui/components"
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

type ViewModel struct {
	Vp               *Viewport
	ContainerName    string
	Breadcrumb       string
	LineCount        *LineCountInfo
	LoadingMessage   string
	EmptyMessage     string
	LoadingIndicator string
	Width            int
	Height           int
	Styles           Styles
	ContentType      ContentType
	ResourceType     core.ResourceType
	Logs             LogsViewer
}

func RenderContent(vm ViewModel) string {
	if vm.Width == 0 || vm.Height == 0 {
		return ""
	}
	layout := util.ComputeFrameLayout(vm.Width, vm.Height, vm.Styles.SubpageFrame)
	headerVM := vm
	headerVM.Width = layout.ContentWidth
	breadcrumb := vm.Breadcrumb
	if breadcrumb == "" {
		resourceLabel := util.ResourceLabel(vm.ResourceType)
		contentLabel := getContentLabel(vm.ContentType)
		breadcrumb = util.ClampSingleLine(
			fmt.Sprintf("%s > %s > %s |", resourceLabel, vm.ContainerName, contentLabel),
			layout.ContentWidth,
		)
	} else {
		breadcrumb = util.ClampSingleLine(breadcrumb+" |", layout.ContentWidth)
	}
	contentHeight := VisibleRowsForContent(layout.ContentHeight)

	header := renderHeader(headerVM, breadcrumb, vm.Vp.Filter)
	headerDivider := components.RenderTitleDivider(vm.Styles.Divider, layout.ContentWidth)
	panel := renderPanel(vm, layout.ContentWidth, contentHeight)

	return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(header, headerDivider, panel))
}

func VisibleRowsForContent(contentHeight int) int {
	return max(1, contentHeight-2)
}

func renderHeader(vm ViewModel, breadcrumb string, filterState components.FilterState) string {
	var left string
	if filterState.Active {
		input := filterState.Input
		input.SetWidth(components.DynamicInputWidth(input.Prompt, vm.Width))
		left = input.View()
	} else {
		left = vm.Styles.Breadcrumb.Render(breadcrumb)
	}

	var mid string
	if !filterState.Active {
		if filterState.Query != "" {
			mid = vm.Styles.Muted.Render(" ") + vm.Styles.KeyText.Render("🔍 "+filterState.Query)
		} else {
			mid = vm.Styles.Muted.Render(" ") + vm.Styles.Key.Inline(true).Render(" / ") + vm.Styles.KeyText.Render(" filter")
		}
	}

	wrapVal := fmt.Sprintf("%-3s", "on")
	if !vm.Vp.WrapLines {
		wrapVal = "off"
	}
	wrapKey := vm.Styles.Key.Inline(true).Render(" w ")
	if filterState.Active {
		wrapKey = vm.Styles.Muted.Render("   ")
	}
	wrapLabel := vm.Styles.Muted.Render(" wrap:")
	wrapText := vm.Styles.FollowOff.Render(wrapVal)
	if vm.Vp.WrapLines {
		wrapText = vm.Styles.FollowOn.Render(wrapVal)
	}
	rightParts := []string{wrapKey, wrapLabel, wrapText}

	if vm.ContentType == ContentTypeLogs {
		onVal := fmt.Sprintf("%-3s", "on")
		if !vm.Vp.Follow {
			onVal = "off"
		}
		followKey := vm.Styles.Key.Inline(true).Render(" f ")
		if filterState.Active {
			followKey = vm.Styles.Muted.Render("   ")
		}
		followLabel := vm.Styles.Muted.Render(" follow:")
		followText := vm.Styles.FollowOff.Render(onVal)
		if vm.Vp.Follow {
			followText = vm.Styles.FollowOn.Render(onVal)
		}
		rightParts = append(rightParts, vm.Styles.Muted.Render(" "), followKey, followLabel, followText)
	}

	if vm.LineCount != nil {
		rightParts = append(rightParts, vm.Styles.Muted.Render(fmt.Sprintf(" lines:(%d-%d/%d)", vm.LineCount.Start, vm.LineCount.End, vm.LineCount.Total)))
	}

	right := lipgloss.JoinHorizontal(lipgloss.Left, rightParts...)
	return renderThreePartLine(left, mid, right, vm.Width)
}

func renderPanel(vm ViewModel, width, height int) string {
	contentWidth := max(1, width)
	if vm.Vp.InitialLoad {
		loadingMsg := vm.LoadingMessage
		if loadingMsg == "" {
			loadingMsg = "Loading..."
		}
		return strings.Join(util.ClipAndPadLines([]string{renderLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator, loadingMsg)}, height, ""), "\n")
	}

	filtered := FilterLines(vm.Vp.Data, vm.Vp.Filter.Query)
	if len(filtered) == 0 {
		empty := vm.EmptyMessage
		if empty == "" {
			empty = "No content available."
		}
		if strings.TrimSpace(vm.Vp.Filter.Query) != "" {
			empty = "No lines match current filter."
		}
		return strings.Join(util.ClipAndPadLines([]string{util.ClampSingleLine(vm.Styles.Muted.Render(empty), contentWidth)}, height, ""), "\n")
	}

	lines := strings.Split(vm.Vp.View(), "\n")
	lines = renderHorizontalScrollIndicators(vm.Vp, lines, filtered, contentWidth, vm.Styles.Muted.Reverse(true))
	if vm.ContentType == ContentTypeLogs && vm.Logs.HistoryLoad {
		msg := vm.LoadingMessage
		if msg == "" {
			msg = "Loading more..."
		}
		lines = append([]string{renderLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator, msg)}, lines...)
	}
	lines = util.ClipAndPadLines(lines, height, "")
	return strings.Join(lines, "\n")
}

func renderLoadingLine(style lipgloss.Style, width int, indicator string, message string) string {
	prefix := strings.TrimSpace(indicator)
	if prefix != "" {
		prefix += " "
	}
	if message == "" {
		message = "Loading..."
	}
	return util.ClampSingleLine(style.Render(prefix+message), width)
}

func renderHorizontalScrollIndicators(vp *Viewport, lines, renderLines []string, width int, indicatorStyle lipgloss.Style) []string {
	if vp.WrapLines || width <= 0 || len(lines) == 0 || len(renderLines) == 0 {
		return lines
	}

	start, end := ViewportRange(vp, len(renderLines))
	visible := renderLines[start:end]
	if len(visible) == 0 {
		return lines
	}

	xOffset := max(0, vp.XOffset())
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
		middle := components.PadVisibleWidth(util.ClampSingleLine(line, max(0, width-2)), max(0, width-2))
		return left + middle + right
	}
	if canScrollLeft {
		rest := components.PadVisibleWidth(util.ClampSingleLine(line, max(0, width-1)), max(0, width-1))
		return left + rest
	}
	prefix := components.PadVisibleWidth(util.ClampSingleLine(line, max(0, width-1)), max(0, width-1))
	return prefix + right
}

func renderThreePartLine(left, mid, right string, width int) string {
	if width <= 0 {
		return ""
	}
	right = util.ClampSingleLine(right, width)
	rightWidth := util.DisplayWidth(right)
	if rightWidth >= width {
		return right
	}
	leftMid := left + mid
	leftMax := max(1, width-rightWidth-1)
	if util.DisplayWidth(leftMid) > leftMax {
		midWidth := util.DisplayWidth(mid)
		left = util.ClampSingleLine(left, max(0, leftMax-midWidth))
		leftRendered := util.DisplayWidth(left)
		mid = util.ClampSingleLine(mid, max(0, leftMax-leftRendered))
		leftMid = left + mid
	}
	leftMidWidth := util.DisplayWidth(leftMid)
	spacing := max(0, width-leftMidWidth-rightWidth)
	return leftMid + strings.Repeat(" ", spacing) + right
}

func getContentLabel(ct ContentType) string {
	switch ct {
	case ContentTypeLogs:
		return "Logs"
	case ContentTypeInspect:
		return "Inspect"
	default:
		return "Logs"
	}
}
