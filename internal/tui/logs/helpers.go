package logs

import (
	"strings"
	"unicode"

	"easydocker/internal/tui/util"
)

// MaxLiveLines controls retained live log history. 0 means unbounded.
const MaxLiveLines = 0

const (
	InitialTail = 200
	TailStep    = 200
)

// FilterLogLines returns only log lines containing query. Empty query keeps all lines.
func FilterLogLines(lines []string, query string) []string {
	if strings.TrimSpace(query) == "" {
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

// WrapLogLines splits each line into fixed-width visual rows.
func WrapLogLines(lines []string, width int) []string {
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
		runeWidth := util.DisplayWidth(string(r))
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

// ViewportRange returns the start and end indices of visible logs.
func ViewportRange(state State, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	start := util.Clamp(state.Viewport.YOffset(), 0, max(0, total-1))
	visible := max(1, state.Viewport.VisibleLineCount())
	end := min(total, start+visible)
	return start, end
}

// VisibleLogRange returns visible raw-log indices [start, end) from viewport state.
// In wrap mode, viewport offsets are row-based and need to be mapped back to raw lines.
func VisibleLogRange(state State, logLines []string) (int, int) {
	total := len(logLines)
	if total <= 0 {
		return 0, 0
	}
	if !state.WrapLines {
		return ViewportRange(state, total)
	}

	wrapWidth := max(1, state.Viewport.Width())
	totalRows := 0
	for _, line := range logLines {
		totalRows += wrappedRowCount(line, wrapWidth)
	}
	if totalRows <= 0 {
		return 0, 0
	}

	startRow := util.Clamp(state.Viewport.YOffset(), 0, totalRows-1)
	visibleRows := max(1, state.Viewport.VisibleLineCount())
	endRowExclusive := min(totalRows, startRow+visibleRows)

	startLine := rowToLineIndex(logLines, wrapWidth, startRow)
	endLine := rowToLineIndex(logLines, wrapWidth, max(startRow, endRowExclusive-1)) + 1
	return startLine, min(total, endLine)
}

// RawLineToViewportRowOffset maps a raw log line index to a wrapped viewport row offset.
// It returns the row offset for the first wrapped row of the target raw line.
func RawLineToViewportRowOffset(logLines []string, wrapWidth, lineIndex int) int {
	if len(logLines) == 0 || lineIndex <= 0 {
		return 0
	}
	if lineIndex > len(logLines) {
		lineIndex = len(logLines)
	}
	rows := 0
	for index := 0; index < lineIndex; index++ {
		rows += wrappedRowCount(logLines[index], wrapWidth)
	}
	return rows
}

func rowToLineIndex(logLines []string, wrapWidth, row int) int {
	if len(logLines) == 0 {
		return 0
	}
	cursor := 0
	for index, line := range logLines {
		rows := wrappedRowCount(line, wrapWidth)
		if row < cursor+rows {
			return index
		}
		cursor += rows
	}
	return len(logLines) - 1
}

func wrappedRowCount(line string, width int) int {
	if width <= 0 {
		return 1
	}
	lineWidth := util.DisplayWidth(line)
	if lineWidth <= 0 {
		return 1
	}
	return max(1, (lineWidth+width-1)/width)
}

// MergePolledLogs merges a fresh polled chunk into the previous log buffer.
// It returns the merged logs and whether an overlap was found.
func MergePolledLogs(previous, polled []string, maxLines int) ([]string, bool) {
	if len(previous) == 0 {
		return TrimLogs(polled, maxLines), true
	}
	if len(polled) == 0 {
		return previous, true
	}

	normalizedPrevious := make([]string, 0, len(previous))
	for _, line := range previous {
		normalizedPrevious = append(normalizedPrevious, strings.TrimRight(line, "\r"))
	}
	normalizedPolled := make([]string, 0, len(polled))
	for _, line := range polled {
		normalizedPolled = append(normalizedPolled, strings.TrimRight(line, "\r"))
	}

	maxOverlap := min(len(normalizedPrevious), len(normalizedPolled))
	for overlap := maxOverlap; overlap > 0; overlap-- {
		if !equalLogSlices(normalizedPrevious[len(normalizedPrevious)-overlap:], normalizedPolled[:overlap]) {
			continue
		}
		merged := append([]string{}, normalizedPrevious...)
		merged = append(merged, normalizedPolled[overlap:]...)
		return TrimLogs(merged, maxLines), true
	}

	if equalLogSlices(normalizedPrevious, normalizedPolled) {
		return TrimLogs(normalizedPrevious, maxLines), true
	}

	if len(normalizedPolled) < len(normalizedPrevious) && equalLogSlices(normalizedPrevious[len(normalizedPrevious)-len(normalizedPolled):], normalizedPolled) {
		return TrimLogs(normalizedPrevious, maxLines), true
	}

	return TrimLogs(normalizedPolled, maxLines), false
}

// TrimLogs keeps only the most recent maxLines items. maxLines <= 0 keeps all lines.
func TrimLogs(lines []string, maxLines int) []string {
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	return lines[len(lines)-maxLines:]
}

// SanitizeLogRenderLine normalizes ANSI-heavy log text for viewport rendering.
func SanitizeLogRenderLine(line string) string {
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
	line = strings.ReplaceAll(line, "\t", "    ")
	return line
}

func equalLogSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
