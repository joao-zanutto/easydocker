package util

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

func StripANSI(text string) string {
	return ansi.Strip(text)
}

func DisplayWidth(text string) int {
	return ansi.StringWidth(StripANSI(text))
}

// TruncateWithEllipsis preserves ANSI sequences while constraining visible width.
func TruncateWithEllipsis(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if DisplayWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}
	prefix, tailANSI := truncateWithANSITail(text, width-1)
	return prefix + "…" + tailANSI
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
	return truncate(flat, width)
}

// truncate limits text to a visible display width while preserving ANSI sequences.
func truncate(text string, width int) string {
	prefix, tailANSI := truncateWithANSITail(text, width)
	return prefix + tailANSI
}

// truncateWithANSITail limits text to visible width and returns:
// - prefix: the visible segment plus any ANSI encountered before the cut
// - tailANSI: ANSI-only sequences from the cut tail (typically resets)
func truncateWithANSITail(text string, width int) (string, string) {
	if width <= 0 {
		return "", ""
	}

	var prefix strings.Builder
	w := 0
	cutIndex := len(text)
	for i := 0; i < len(text); {
		r := rune(text[i])
		if r == '\x1b' {
			seqLen := ansiSequenceLen(text[i:])
			if seqLen == 0 {
				i++
				continue
			}
			prefix.WriteString(text[i : i+seqLen])
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
			cutIndex = i
			break
		}

		prefix.WriteRune(r)
		w += rw
		i += size
	}

	if cutIndex >= len(text) {
		return prefix.String(), ""
	}

	var tail strings.Builder
	appendANSISequences(text[cutIndex:], &tail)
	return prefix.String(), tail.String()
}

// appendANSISequences keeps only ANSI escapes from the remaining tail.
// This preserves trailing reset codes so styles don't bleed into following text.
func appendANSISequences(text string, b *strings.Builder) {
	for i := 0; i < len(text); {
		if text[i] == '\x1b' {
			seqLen := ansiSequenceLen(text[i:])
			if seqLen > 0 {
				b.WriteString(text[i : i+seqLen])
				i += seqLen
				continue
			}
		}

		_, size := utf8.DecodeRuneInString(text[i:])
		if size <= 0 {
			i++
			continue
		}
		i += size
	}
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
