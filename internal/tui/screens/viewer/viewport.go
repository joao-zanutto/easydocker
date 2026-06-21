package viewer

import (
	"reflect"
	"unsafe"

	"easydocker/internal/tui/ui/components"
	"easydocker/internal/tui/util"

	bubblesview "charm.land/bubbles/v2/viewport"
)

var (
	bvLinesOffset          = uintptr(0)
	bvLongestLineWidthOffset = uintptr(0)
)

func init() {
	t := reflect.TypeOf(bubblesview.Model{})
	if f, ok := t.FieldByName("lines"); ok {
		bvLinesOffset = f.Offset
	} else {
		panic("bubbleview viewport missing field: lines")
	}
	if f, ok := t.FieldByName("longestLineWidth"); ok {
		bvLongestLineWidthOffset = f.Offset
	} else {
		panic("bubbleview viewport missing field: longestLineWidth")
	}
}

type Viewport struct {
	*bubblesview.Model
	WrapLines      bool
	WrapAnchored   bool
	Filter         components.FilterState
	Data           []string
	Follow         bool
	ContentType    ContentType
	InitialLoad    bool
	sanitizedLines    []string
	needFullSan       bool
	dataGen           int
	longestLineWidth  int
	savedYOffset      int
	wrappedLines      []string
	wrappedWidth      int
	wrappedSourceCount int
	wrapCanAppend     bool
	wrapTotalRows     int
	wrapCacheWidth    int
	wrapCacheGen      int
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
	vp.wrapCacheGen++
	vp.wrappedLines = nil
}

func (vp *Viewport) OpenFilter() {
	vp.savedYOffset = vp.YOffset()
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

//nolint:gosec
func setBvContent(m *bubblesview.Model, lines []string, longestLineWidth int) {
	base := unsafe.Pointer(m)
	*(*[]string)(unsafe.Pointer(uintptr(base) + bvLinesOffset)) = lines
	*(*int)(unsafe.Pointer(uintptr(base) + bvLongestLineWidthOffset)) = longestLineWidth
}

func (vp *Viewport) SyncViewport(lines []string, visibleWidth, visibleRows int) {
	vp.SetWidth(visibleWidth)
	vp.SetHeight(visibleRows)

	if len(lines) == 1 && util.DisplayWidthPure(lines[0]) == 0 {
		setBvContent(vp.Model, nil, vp.longestLineWidth)
	} else {
		setBvContent(vp.Model, lines, vp.longestLineWidth)
	}

	if !vp.SoftWrap {
		maxOffset := max(0, len(lines)-visibleRows)
		if vp.YOffset() > maxOffset {
			vp.SetYOffset(maxOffset)
		}
	}
	if vp.Follow {
		vp.GotoBottom()
	}
}

func (vp *Viewport) getSanitizedLines(raw []string) []string {
	if len(raw) == 0 {
		vp.longestLineWidth = 0
		return nil
	}
	if !vp.needFullSan && len(raw) == len(vp.sanitizedLines) && vp.dataGen > 0 {
		return vp.sanitizedLines
	}
	if !vp.needFullSan && len(raw) > len(vp.sanitizedLines) && vp.dataGen > 0 {
		oldLen := len(vp.sanitizedLines)
		vp.sanitizedLines = append(vp.sanitizedLines, make([]string, len(raw)-oldLen)...)
		for i := oldLen; i < len(raw); i++ {
			vp.sanitizedLines[i] = SanitizeLine(raw[i])
			if w := util.DisplayWidthPure(vp.sanitizedLines[i]); w > vp.longestLineWidth {
				vp.longestLineWidth = w
			}
		}
		vp.needFullSan = false
		vp.dataGen++
		return vp.sanitizedLines
	}
	vp.sanitizedLines = make([]string, len(raw))
	maxW := 0
	for i, line := range raw {
		vp.sanitizedLines[i] = SanitizeLine(line)
		if w := util.DisplayWidthPure(vp.sanitizedLines[i]); w > maxW {
			maxW = w
		}
	}
	vp.longestLineWidth = maxW
	vp.needFullSan = false
	vp.dataGen++
	return vp.sanitizedLines
}

func (vp *Viewport) InvalidateSanitizeCache() {
	vp.needFullSan = true
}

func (vp *Viewport) PrependToCache(data []string, prependedCount int) {
	if prependedCount <= 0 {
		return
	}
	newSan := make([]string, prependedCount)
	for i := 0; i < prependedCount; i++ {
		newSan[i] = SanitizeLine(data[i])
		if w := util.DisplayWidthPure(newSan[i]); w > vp.longestLineWidth {
			vp.longestLineWidth = w
		}
	}
	vp.sanitizedLines = append(newSan, vp.sanitizedLines...)
	vp.dataGen++
	vp.wrappedLines = nil
	vp.needFullSan = false
}

func (vp *Viewport) FilteredLines() []string {
	sanitized := vp.getSanitizedLines(vp.Data)
	if vp.ContentType == ContentTypeInspect {
		return FilterJSONLines(sanitized, vp.Filter.Query)
	}
	return FilterLines(sanitized, vp.Filter.Query)
}

func (vp *Viewport) ClearCache() {
	vp.sanitizedLines = nil
	vp.needFullSan = false
	vp.dataGen = 0
	vp.longestLineWidth = 0
	vp.wrappedLines = nil
	vp.wrapTotalRows = 0
	vp.wrapCacheWidth = 0
	vp.wrapCacheGen = 0
}

func (vp *Viewport) SyncFromData(visibleWidth, visibleRows int) {
	if vp.Data != nil {
		vp.InitialLoad = false
	}
	if vp.Data == nil {
		vp.SyncViewport(nil, visibleWidth, visibleRows)
		vp.wrapTotalRows = 0
		vp.wrapCanAppend = false
		return
	}
	if !vp.wrapCanAppend {
		vp.wrappedLines = nil
	}
	lines := vp.PrepareContentLines(visibleWidth, vp.WrapLines)
	vp.SyncViewport(lines, visibleWidth, visibleRows)
	vp.wrapTotalRows = len(lines)
	vp.wrapCacheWidth = visibleWidth
	vp.wrapCacheGen = vp.dataGen
	vp.wrapCanAppend = false
}
