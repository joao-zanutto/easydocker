package viewer

import (
	"strings"

	"easydocker/internal/tui/ui/components"

	bubblesview "charm.land/bubbles/v2/viewport"
)

type Viewport struct {
	*bubblesview.Model
	WrapLines      bool
	WrapAnchored   bool
	Filter         components.FilterState
	Data           []string
	Follow         bool
	ContentType    ContentType
	InitialLoad    bool
	rawLines       []string
	sanitizedLines []string
	dataGen        int
}

func NewViewport() *Viewport {
	vp := bubblesview.New(bubblesview.WithWidth(1), bubblesview.WithHeight(1))
	vp.SetHorizontalStep(8)
	vp.SetContent("")
	filterState := components.NewFilterState()
	return &Viewport{
		Model:  &vp,
		Filter: filterState,
		Follow: true,
	}
}

func (vp *Viewport) SetFollow(enabled bool) {
	vp.Follow = enabled
	if enabled {
		vp.GotoBottom()
	}
}

func (vp *Viewport) SetWrapLines(enabled bool) {
	if vp.WrapLines == enabled {
		return
	}
	if enabled {
		vp.SetXOffset(0)
	}
	vp.WrapLines = enabled
}

func (vp *Viewport) OpenFilter() {
	vp.Filter.Active = true
	vp.Filter.Input.Focus()
	vp.Filter.Input.SetValue(vp.Filter.Query)
}

func (vp *Viewport) CloseFilter(clear bool) {
	vp.Filter.Active = false
	vp.Filter.Input.Blur()
	if clear {
		vp.Filter.Query = ""
		vp.Filter.Input.SetValue("")
	}
}

func (vp *Viewport) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	vp.SetWidth(visibleWidth)
	vp.SetHeight(visibleRows)
	vp.SetContent(strings.Join(lines, "\n"))
	if vp.Follow {
		vp.GotoBottom()
	}
}

func (vp *Viewport) getSanitizedLines(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	if vp.dataGen > 0 && len(vp.rawLines) > 0 && &raw[0] == &vp.rawLines[0] {
		return vp.sanitizedLines
	}
	vp.rawLines = raw
	vp.sanitizedLines = make([]string, len(raw))
	for i, line := range raw {
		vp.sanitizedLines[i] = SanitizeLine(line)
	}
	vp.dataGen++
	return vp.sanitizedLines
}

func (vp *Viewport) FilteredLines() []string {
	sanitized := vp.getSanitizedLines(vp.Data)
	if vp.ContentType == ContentTypeInspect {
		return FilterJSONLines(sanitized, vp.Filter.Query)
	}
	return FilterLines(sanitized, vp.Filter.Query)
}

func (vp *Viewport) ClearCache() {
	vp.rawLines = nil
	vp.sanitizedLines = nil
	vp.dataGen = 0
}

func (vp *Viewport) SyncFromData(visibleWidth, visibleRows int) {
	vp.InitialLoad = false
	if vp.Data == nil {
		vp.SyncViewport(nil, visibleWidth, visibleRows)
		return
	}
	lines := vp.PrepareContentLines(visibleWidth, vp.WrapLines)
	vp.SyncViewport(lines, visibleWidth, visibleRows)
}
