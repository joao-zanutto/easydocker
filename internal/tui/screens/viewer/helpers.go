package viewer

import (
	"slices"
	"sort"
	"strings"
	"unicode"

	"easydocker/internal/tui/util"
)

func ViewportRange(vp *Viewport, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	start := util.Clamp(vp.YOffset(), 0, max(0, total-1))
	visible := max(1, vp.VisibleLineCount())
	end := min(total, start+visible)
	return start, end
}

func VisibleContentRange(vp *Viewport, lines []string) (int, int) {
	total := len(lines)
	if total <= 0 {
		return 0, 0
	}
	if !vp.WrapLines {
		return ViewportRange(vp, total)
	}

	wrapWidth := max(1, vp.Width())
	var totalRows int
	if vp.wrapCacheWidth == wrapWidth && vp.wrapCacheGen == vp.dataGen {
		totalRows = vp.wrapTotalRows
	} else {
		for _, line := range lines {
			totalRows += WrappedRowCount(line, wrapWidth)
		}
	}
	if totalRows <= 0 {
		return 0, 0
	}

	startRow := util.Clamp(vp.YOffset(), 0, totalRows-1)
	visibleRows := max(1, vp.VisibleLineCount())
	endRowExclusive := min(totalRows, startRow+visibleRows)

	startLine := rowToLineIndex(lines, wrapWidth, startRow)
	endLine := rowToLineIndex(lines, wrapWidth, max(startRow, endRowExclusive-1)) + 1
	return startLine, min(total, endLine)
}

func rowToLineIndex(lines []string, wrapWidth, row int) int {
	if len(lines) == 0 {
		return 0
	}
	cursor := 0
	for index, line := range lines {
		rows := WrappedRowCount(line, wrapWidth)
		if row < cursor+rows {
			return index
		}
		cursor += rows
	}
	return len(lines) - 1
}

func WrappedRowCount(line string, width int) int {
	if width <= 0 {
		return 1
	}
	lineWidth := util.DisplayWidthPure(line)
	if lineWidth <= 0 {
		return 1
	}
	return max(1, (lineWidth+width-1)/width)
}

func FilterLines(lines []string, query string) []string {
	if query == "" || strings.TrimSpace(query) == "" {
		return lines
	}
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, query) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func jsonIndentLevel(line string) int {
	trimmed := strings.TrimLeft(line, " ")
	return (len(line) - len(trimmed)) / 2
}

func isJSONCloser(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "}" || trimmed == "}," || trimmed == "]" || trimmed == "],"
}

func isJSONKey(line string) bool {
	return !isJSONCloser(line) && strings.Contains(line, "\": ")
}

func isBlockOpener(line string) bool {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimSuffix(trimmed, ",")
	return strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[")
}

func findJSONParent(lines []string, idx int) int {
	level := jsonIndentLevel(lines[idx])
	if level <= 0 {
		return -1
	}
	if level == 1 {
		return 0
	}
	for k := idx - 1; k >= 0; k-- {
		l := jsonIndentLevel(lines[k])
		if l < level && (l == 0 || isJSONKey(lines[k])) {
			return k
		}
	}
	return 0
}

func findJSONCloser(lines []string, idx int) int {
	if !isBlockOpener(lines[idx]) {
		return -1
	}
	level := jsonIndentLevel(lines[idx])
	for j := idx + 1; j < len(lines); j++ {
		if jsonIndentLevel(lines[j]) == level && isJSONCloser(lines[j]) {
			return j
		}
	}
	return -1
}

func buildAncestryPath(lines []string, matchIdx int) []int {
	path := []int{matchIdx}
	for {
		parent := findJSONParent(lines, path[0])
		if parent < 0 || parent == path[0] {
			break
		}
		path = append([]int{parent}, path...)
		if jsonIndentLevel(lines[parent]) == 0 {
			break
		}
	}
	return path
}

func FilterJSONLines(lines []string, query string) []string {
	if query == "" || strings.TrimSpace(query) == "" {
		return lines
	}

	lineSet := make(map[int]bool)

	for i, line := range lines {
		if !strings.Contains(line, query) {
			continue
		}

		path := buildAncestryPath(lines, i)

		closers := make([]int, len(path))
		for p, idx := range path {
			closers[p] = findJSONCloser(lines, idx)
		}

		for _, idx := range path {
			lineSet[idx] = true
		}

		matchIdx := path[len(path)-1]
		matchCloser := closers[len(closers)-1]
		if matchCloser > matchIdx+1 {
			for j := matchIdx + 1; j < matchCloser; j++ {
				lineSet[j] = true
			}
		}

		for p := len(path) - 1; p >= 0; p-- {
			if closers[p] >= 0 {
				lineSet[closers[p]] = true
			}
		}
	}

	if len(lineSet) == 0 {
		return nil
	}

	indices := make([]int, 0, len(lineSet))
	for idx := range lineSet {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = lines[idx]
	}
	return result
}

func WrapLines(lines []string, width int) []string {
	if width <= 0 {
		return nil
	}
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapLogLine(line, width)...)
	}
	return wrapped
}

func wrapLogLine(line string, width int) []string {
	if line == "" {
		return []string{""}
	}

	chunks := make([]string, 0, max(1, len(line)/max(1, width)+1))
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, current.String())
		current.Reset()
		currentWidth = 0
	}

	for _, r := range line {
		runeWidth := util.DisplayWidthPure(string(r))
		if runeWidth <= 0 {
			current.WriteRune(r)
			continue
		}
		if currentWidth > 0 && currentWidth+runeWidth > width {
			flush()
		}
		current.WriteRune(r)
		currentWidth += runeWidth
		if currentWidth >= width {
			flush()
		}
	}

	flush()
	if len(chunks) == 0 {
		return []string{""}
	}
	return chunks
}

func SanitizeLine(line string) string {
	clean := normalizeLine(util.StripANSI(line))
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, clean)
}

func normalizeLine(line string) string {
	line = strings.ReplaceAll(line, "\r", "")
	line = strings.ReplaceAll(line, "\t", " ")
	return line
}

func MergePolledLogs(prev, polled []string) ([]string, bool) {
	if len(prev) == 0 {
		return polled, true
	}
	if len(polled) == 0 {
		return prev, true
	}

	maxOverlap := min(len(prev), len(polled))
	for o := maxOverlap; o > 0; o-- {
		match := true
		for i := 0; i < o; i++ {
			if strings.TrimRight(prev[len(prev)-o+i], "\r") != strings.TrimRight(polled[i], "\r") {
				match = false
				break
			}
		}
		if match {
			if o == len(polled) {
				return prev, true
			}
			result := make([]string, len(prev)+len(polled)-o)
			for i, l := range prev {
				result[i] = strings.TrimRight(l, "\r")
			}
			for i := o; i < len(polled); i++ {
				result[len(prev)+(i-o)] = strings.TrimRight(polled[i], "\r")
			}
			return result, true
		}
	}

	normPrev := make([]string, len(prev))
	for i, l := range prev {
		normPrev[i] = strings.TrimRight(l, "\r")
	}
	normPolled := make([]string, len(polled))
	for i, l := range polled {
		normPolled[i] = strings.TrimRight(l, "\r")
	}

	if slices.Equal(normPrev, normPolled) {
		return normPrev, true
	}
	if len(normPolled) < len(normPrev) && slices.Equal(normPrev[len(normPrev)-len(normPolled):], normPolled) {
		return normPrev, true
	}

	return normPolled, false
}

func (vp *Viewport) PrepareContentLines(wrapWidth int, wrapEnabled bool) []string {
	lines := vp.FilteredLines()

	if !wrapEnabled || wrapWidth <= 0 {
		vp.wrappedLines = nil
		return lines
	}

	if len(vp.wrappedLines) > 0 && vp.wrappedWidth == wrapWidth &&
		len(lines) == vp.wrappedSourceCount {
		return vp.wrappedLines
	}

	if vp.wrapCanAppend && len(vp.wrappedLines) > 0 && vp.wrappedWidth == wrapWidth &&
		len(lines) > vp.wrappedSourceCount {
		newLines := lines[vp.wrappedSourceCount:]
		newWrapped := WrapLines(newLines, wrapWidth)
		vp.wrappedLines = append(vp.wrappedLines, newWrapped...)
		vp.wrappedSourceCount = len(lines)
		return vp.wrappedLines
	}

	vp.wrappedLines = WrapLines(lines, wrapWidth)
	vp.wrappedWidth = wrapWidth
	vp.wrappedSourceCount = len(lines)
	return vp.wrappedLines
}
