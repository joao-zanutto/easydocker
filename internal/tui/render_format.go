package tui

import (
	"fmt"
	"math"
	"strings"
)

func formatMemoryUsage(usage string, percent float64, limit string) string {
	if usage == "-" {
		return "-"
	}
	if limit != "" && limit != "-" {
		return fmt.Sprintf("%s / %s (%s)", usage, limit, renderPercent(percent))
	}
	return fmt.Sprintf("%s (%s)", usage, renderPercent(percent))
}

func renderPercent(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", value)
}

func joinSections(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, "\n")
}

func refCountText(ref int64) string {
	if ref <= 0 {
		return "0"
	}
	return fmt.Sprintf("%d", ref)
}
