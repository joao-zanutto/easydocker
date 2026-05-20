package viewer

import (
	"strings"
	"unicode"

	"easydocker/internal/tui/util"
)

func ViewportRange(state *State, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	start := util.Clamp(state.Viewport.YOffset(), 0, max(0, total-1))
	visible := max(1, state.Viewport.VisibleLineCount())
	end := min(total, start+visible)
	return start, end
}

func VisibleContentRange(state *State, lines []string) (int, int) {
	total := len(lines)
	if total <= 0 {
		return 0, 0
	}
	if !state.WrapLines {
		return ViewportRange(state, total)
	}

	wrapWidth := max(1, state.Viewport.Width())
	totalRows := 0
	for _, line := range lines {
		totalRows += WrappedRowCount(line, wrapWidth)
	}
	if totalRows <= 0 {
		return 0, 0
	}

	startRow := util.Clamp(state.Viewport.YOffset(), 0, totalRows-1)
	visibleRows := max(1, state.Viewport.VisibleLineCount())
	endRowExclusive := min(totalRows, startRow+visibleRows)

	startLine := rowToLineIndex(lines, wrapWidth, startRow)
	endLine := rowToLineIndex(lines, wrapWidth, max(startRow, endRowExclusive-1)) + 1
	return startLine, min(total, endLine)
}

func RawLineToViewportRowOffset(lines []string, wrapWidth, lineIndex int) int {
	if len(lines) == 0 || lineIndex <= 0 {
		return 0
	}
	if lineIndex > len(lines) {
		lineIndex = len(lines)
	}
	rows := 0
	for index := 0; index < lineIndex; index++ {
		rows += WrappedRowCount(lines[index], wrapWidth)
	}
	return rows
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
	lineWidth := util.DisplayWidth(line)
	if lineWidth <= 0 {
		return 1
	}
	return max(1, (lineWidth+width-1)/width)
}

func FilterLines(lines []string, query string) []string {
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
