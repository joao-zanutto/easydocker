package util

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
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

func ContainerCPUValue(container core.ContainerRow, loadingIndicator string) string {
	if container.CPUPercent < 0 {
		if container.State == core.StateRunning {
			return metricsLoadingValue(loadingIndicator)
		}
		return core.MetricsNA
	}
	if container.CPUPercent < 0.05 {
		if container.State == core.StateRunning {
			return "0.0%"
		}
		return core.MetricsNA
	}
	return fmt.Sprintf("%.1f%%", container.CPUPercent)
}

func ContainerMemoryTableValue(container core.ContainerRow, loadingIndicator string) string {
	if container.MemoryUsage == core.MetricsNA || container.MemoryUsage == core.MetricsLoading {
		if container.State == core.StateRunning {
			return metricsLoadingValue(loadingIndicator)
		}
		return core.MetricsNA
	}
	return container.MemoryUsage
}

func metricsLoadingValue(loadingIndicator string) string {
	if strings.TrimSpace(loadingIndicator) == "" {
		return core.MetricsNA
	}
	return loadingIndicator
}

func ResourceLabel(rt core.ResourceType) string {
	return rt.String()
}

func ContainerStateText(container core.ContainerRow) string {
	if container.Healthy && container.State == core.StateRunning {
		return "● healthy"
	}
	return "● " + string(container.State)
}
