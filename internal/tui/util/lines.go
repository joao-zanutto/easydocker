package util

// ConstrainLines applies ConstrainLine to every line in the slice.
func ConstrainLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, ConstrainLine(line, width))
	}
	return out
}

// ClipLines limits the number of lines to height.
func ClipLines(lines []string, height int) []string {
	if height <= 0 {
		return []string{}
	}
	if len(lines) <= height {
		return lines
	}
	return lines[:height]
}

// ClipAndPadLines clips to height and pads missing rows with fill.
func ClipAndPadLines(lines []string, height int, fill string) []string {
	if height <= 0 {
		return []string{}
	}
	clipped := ClipLines(lines, height)
	out := make([]string, 0, height)
	out = append(out, clipped...)
	for len(out) < height {
		out = append(out, fill)
	}
	return out
}
