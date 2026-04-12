package util

import (
	"fmt"
	"math"
	"strings"
)

// FormatMemoryUsage formats memory usage with optional limit and percentage.
func FormatMemoryUsage(usage string, percent float64, limit string) string {
	if usage == "-" {
		return "-"
	}
	if limit != "" && limit != "-" {
		return fmt.Sprintf("%s / %s (%s)", usage, limit, RenderPercent(percent))
	}
	return fmt.Sprintf("%s (%s)", usage, RenderPercent(percent))
}

// RenderPercent formats a float as a percentage, handling NaN and Inf.
func RenderPercent(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", value)
}

// JoinSections joins non-empty strings with newlines, filtering out blank sections.
func JoinSections(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, "\n")
}

// RefCountText formats a reference count, returning "0" for zero or negative values.
func RefCountText(ref int64) string {
	if ref <= 0 {
		return "0"
	}
	return fmt.Sprintf("%d", ref)
}
