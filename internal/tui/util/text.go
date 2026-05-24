package util

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func StripANSI(text string) string {
	return ansi.Strip(text)
}

func DisplayWidth(text string) int {
	return ansi.StringWidth(StripANSI(text))
}

func TruncateWithEllipsis(text string, width int) string {
	return ansi.Truncate(text, width, "…")
}

func ConstrainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	trimmed := strings.TrimRight(line, "\n")
	return TruncateWithEllipsis(trimmed, width)
}

func ClampSingleLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	flat := strings.ReplaceAll(line, "\n", " ")
	flat = strings.ReplaceAll(flat, "\r", " ")
	if DisplayWidth(flat) <= width {
		return flat
	}
	return ansi.Truncate(flat, width, "")
}
