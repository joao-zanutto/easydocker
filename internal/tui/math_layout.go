package tui

func allocateColumns(total int, desired []int) []int {
	if len(desired) == 0 {
		return []int{}
	}
	if total <= 0 {
		out := make([]int, len(desired))
		for i := range out {
			out[i] = 1
		}
		return out
	}

	out := make([]int, len(desired))
	sum := 0
	for i, width := range desired {
		if width < 1 {
			width = 1
		}
		out[i] = width
		sum += width
	}
	if sum == total {
		return out
	}
	if sum < total {
		out[len(out)-1] += total - sum
		return out
	}

	over := sum - total
	for over > 0 {
		changed := false
		for i := range out {
			if out[i] > 1 && over > 0 {
				out[i]--
				over--
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return out
}

func scrollWindow(total, cursor, height int) (int, int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}
	if height >= total {
		return 0, total
	}
	cursor = clamp(cursor, 0, total-1)
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
	}
	if start < 0 {
		start = 0
	}
	return start, end
}

func clamp(value, lower, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

func clampFloat(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
