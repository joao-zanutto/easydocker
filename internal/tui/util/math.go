package util

// Clamp constrains v to the inclusive range [low, high].
func Clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
