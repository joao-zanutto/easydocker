package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

func constrainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	trimmed := strings.TrimRight(line, "\n")
	if ansi.StringWidth(trimmed) <= width {
		return trimmed
	}
	if width <= 1 {
		return "…"
	}
	return truncateANSI(trimmed, width-1) + "…"
}

func constrainLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, constrainLine(line, width))
	}
	return out
}

func clipLines(lines []string, height int) []string {
	if height <= 0 {
		return []string{}
	}
	if len(lines) <= height {
		return lines
	}
	return lines[:height]
}

func logViewportLine(line string, offsetX, width int) string {
	if width <= 0 {
		return ""
	}
	clean := sanitizeLogViewportLine(line)
	if offsetX < 0 {
		offsetX = 0
	}
	totalWidth := ansi.StringWidth(clean)
	if offsetX >= totalWidth {
		return ""
	}
	start := offsetX
	limit := start + width
	if limit > totalWidth {
		limit = totalWidth
	}

	var b strings.Builder
	col := 0
	for _, r := range clean {
		rw := ansi.StringWidth(string(r))
		if rw < 0 {
			rw = 0
		}
		next := col + rw
		if next <= start {
			col = next
			continue
		}
		if col < start {
			col = next
			continue
		}
		if col >= limit {
			break
		}
		if rw == 0 {
			if col >= start && col < limit {
				b.WriteRune(r)
			}
			continue
		}
		if next > limit {
			break
		}
		b.WriteRune(r)
		col = next
	}
	return b.String()
}

// clipLogDisplayLine enforces a hard display-width cap without adding
// ellipsis. This keeps log rows from wrapping in the viewport.
func clipLogDisplayLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	clean := sanitizeLogViewportLine(line)
	if ansi.StringWidth(clean) <= width {
		return clean
	}
	return truncateANSI(clean, width)
}

// fit returns a string of exactly width display columns: truncated with "…" if
// the text is longer, or right-padded with spaces if shorter.
func fit(text string, width int) string {
	if width <= 0 {
		return ""
	}
	clean := strings.TrimSpace(normalizeLine(text))
	w := ansi.StringWidth(stripANSI(clean))
	switch {
	case w == width:
		return clean
	case w > width:
		if width == 1 {
			return "…"
		}
		return truncateANSI(clean, width-1) + "…"
	default:
		return clean + strings.Repeat(" ", width-w)
	}
}

func truncateANSI(text string, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	w := 0
	for i := 0; i < len(text); {
		r := rune(text[i])
		if r == '\x1b' {
			seqLen := ansiSequenceLen(text[i:])
			if seqLen == 0 {
				i++
				continue
			}
			b.WriteString(text[i : i+seqLen])
			i += seqLen
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		rw := ansi.StringWidth(string(r))
		if w+rw > width {
			break
		}
		b.WriteRune(r)
		w += rw
		i += size
	}
	return b.String()
}

func ansiSequenceLen(s string) int {
	if len(s) < 2 || s[0] != '\x1b' || s[1] != '[' {
		return 0
	}
	for i := 2; i < len(s); i++ {
		if s[i] >= '@' && s[i] <= '~' {
			return i + 1
		}
	}
	return 0
}

func stripANSI(text string) string {
	return ansi.Strip(text)
}

func normalizeLine(line string) string {
	line = strings.ReplaceAll(line, "\r", "")
	line = strings.ReplaceAll(line, "\t", "    ")
	return line
}

func logLineMaxOffset(line string, width int) int {
	if width <= 0 {
		return 0
	}
	clean := sanitizeLogViewportLine(line)
	totalWidth := ansi.StringWidth(clean)
	if totalWidth <= width {
		return 0
	}
	return totalWidth - width
}

func sanitizeLogViewportLine(line string) string {
	clean := normalizeLine(stripANSI(line))
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
