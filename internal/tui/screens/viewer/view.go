package viewer

import (
	"fmt"
	"strings"

	"easydocker/internal/tui/ui/components"
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
	State            *State
	ContainerName    string
	Breadcrumb       string
	LineCount        *LineCountInfo
	LoadingMessage   string
	EmptyMessage     string
	LoadingIndicator string
	Width            int
	Height           int
	Styles           ViewStyles
	ContentType      ContentType
	ResourceType     ResourceType
}

const filterHeaderHeight = 2

func RenderContent(vm ViewModel) string {
	if vm.Width == 0 || vm.Height == 0 {
		return ""
	}
	layout := util.ComputeFrameLayout(vm.Width, vm.Height, vm.Styles.SubpageFrame)
	headerVM := vm
	headerVM.Width = layout.ContentWidth
	breadcrumb := vm.Breadcrumb
	if breadcrumb == "" {
		resourceLabel := getResourceLabel(vm.ResourceType)
		contentLabel := getContentLabel(vm.ContentType)
		breadcrumb = util.ClampSingleLine(
			fmt.Sprintf("%s / %s / %s", resourceLabel, vm.ContainerName, contentLabel),
			layout.ContentWidth,
		)
	} else {
		breadcrumb = util.ClampSingleLine(breadcrumb, layout.ContentWidth)
	}
	contentHeight := VisibleRowsForContent(layout.ContentHeight, vm.State.Filter.Active)

	if vm.State.Filter.Active {
		filterInput := vm.State.Filter.Input
		filterInput.SetWidth(components.DynamicInputWidth(filterInput.Prompt, layout.ContentWidth))
		filterHeader := components.RenderFilterHeader(filterInput.View(), layout.ContentWidth, vm.Styles.Divider)
		header := renderHeader(headerVM, breadcrumb)
		panel := renderPanel(vm, layout.ContentWidth, contentHeight)
		return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(header, filterHeader, panel))
	}

	headerDivider := components.RenderTitleDivider(vm.Styles.Divider, layout.ContentWidth)
	header := renderHeader(headerVM, breadcrumb)
	panel := renderPanel(vm, layout.ContentWidth, contentHeight)
	return util.RenderFramedContent(vm.Styles.SubpageFrame, layout, util.JoinSections(header, headerDivider, panel))
}

func VisibleRowsForContent(contentHeight int, filterActive bool) int {
	overhead := 2
	if filterActive {
		overhead += filterHeaderHeight
	}
	return max(1, contentHeight-overhead)
}

func renderHeader(vm ViewModel, breadcrumb string) string {
	wrap := "off"
	if vm.State.WrapLines {
		wrap = "on"
	}
	left := vm.Styles.Breadcrumb.Render(breadcrumb)
	wrapText := vm.Styles.FollowOff.Render(wrap)
	if vm.State.WrapLines {
		wrapText = vm.Styles.FollowOn.Render(wrap)
	}

	rightParts := []string{vm.Styles.Muted.Render("wrap:"), wrapText}

	if vm.ContentType == ContentTypeLogs {
		followLabel := "follow:off"
		followStyle := vm.Styles.FollowOff
		if vm.State.Follow {
			followLabel = "follow:on"
			followStyle = vm.Styles.FollowOn
		}
		followText := followStyle.Render(followLabel)
		rightParts = append(rightParts, vm.Styles.Muted.Render(" "), followText)
	}

	if vm.LineCount != nil {
		rightParts = append(rightParts, vm.Styles.Muted.Render(fmt.Sprintf(" lines:(%d-%d/%d)", vm.LineCount.Start, vm.LineCount.End, vm.LineCount.Total)))
	}

	right := lipgloss.JoinHorizontal(lipgloss.Left, rightParts...)
	return renderRightPriorityLine(left, right, vm.Width)
}

func renderPanel(vm ViewModel, width, height int) string {
	contentWidth := max(1, width)
	if vm.State.InitialLoad {
		loadingMsg := vm.LoadingMessage
		if loadingMsg == "" {
			loadingMsg = "Loading..."
		}
		return strings.Join(util.ClipAndPadLines([]string{renderLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator, loadingMsg)}, height, ""), "\n")
	}

	filtered := FilterLines(vm.State.Data, vm.State.Filter.Query)
	if len(filtered) == 0 {
		empty := vm.EmptyMessage
		if empty == "" {
			empty = "No content available."
		}
		if strings.TrimSpace(vm.State.Filter.Query) != "" {
			empty = "No lines match current filter."
		}
		return strings.Join(util.ClipAndPadLines([]string{util.ClampSingleLine(vm.Styles.Muted.Render(empty), contentWidth)}, height, ""), "\n")
	}

	lines := strings.Split(vm.State.Viewport.View(), "\n")
	lines = renderHorizontalScrollIndicators(vm.State, lines, filtered, max(1, vm.State.Viewport.Width()), vm.Styles.Muted.Reverse(true))
	if vm.State.HistoryLoad {
		lines = append([]string{renderHistoryLoadingLine(vm.Styles.Muted, contentWidth, vm.LoadingIndicator, vm.LoadingMessage)}, lines...)
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

func renderHistoryLoadingLine(style lipgloss.Style, width int, indicator string, message string) string {
	prefix := strings.TrimSpace(indicator)
	if prefix != "" {
		prefix += " "
	}
	if message == "" {
		message = "Loading more..."
	}
	return util.ClampSingleLine(style.Render(prefix+message), width)
}

func renderHorizontalScrollIndicators(state *State, lines, renderLines []string, width int, indicatorStyle lipgloss.Style) []string {
	if state.WrapLines || width <= 0 || len(lines) == 0 || len(renderLines) == 0 {
		return lines
	}

	start, end := viewportRange(state, len(renderLines))
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

func viewportRange(state *State, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	start := util.Clamp(state.Viewport.YOffset(), 0, max(0, total-1))
	visible := max(1, state.Viewport.VisibleLineCount())
	end := min(total, start+visible)
	return start, end
}

func getResourceLabel(rt ResourceType) string {
	switch rt {
	case ResourceTypeContainer:
		return "Containers"
	case ResourceTypeVolume:
		return "Volumes"
	case ResourceTypeNetwork:
		return "Networks"
	case ResourceTypeImage:
		return "Images"
	default:
		return "Containers"
	}
}

func GetResourceLabel(rt ResourceType) string {
	return getResourceLabel(rt)
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
