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

// ViewportRange returns the start and end indices of visible logs.
func ViewportRange(state State, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	start := util.Clamp(state.Viewport.YOffset, 0, max(0, total-1))
	visible := max(1, state.Viewport.VisibleLineCount())
	end := min(total, start+visible)
	return start, end
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
