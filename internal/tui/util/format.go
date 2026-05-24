package util

import (
	"fmt"
	"strings"
)

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
