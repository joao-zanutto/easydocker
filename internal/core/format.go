package core

import (
	"fmt"
	"strings"
)

func HumanBytes(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := uint64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func ContainerCPUValue(container ContainerRow, loadingIndicator string) string {
	if container.CPUPercent < 0 {
		if container.State == StateRunning {
			return metricsLoadingValue(loadingIndicator)
		}
		return MetricsNA
	}
	if container.CPUPercent < 0.05 {
		if container.State == StateRunning {
			return "0.0%"
		}
		return MetricsNA
	}
	return fmt.Sprintf("%.1f%%", container.CPUPercent)
}

func ContainerMemoryTableValue(container ContainerRow, loadingIndicator string) string {
	if container.MemoryUsage == MetricsNA || container.MemoryUsage == MetricsLoading {
		if container.State == StateRunning {
			return metricsLoadingValue(loadingIndicator)
		}
		return MetricsNA
	}
	return container.MemoryUsage
}

func metricsLoadingValue(loadingIndicator string) string {
	if strings.TrimSpace(loadingIndicator) == "" {
		return MetricsNA
	}
	return loadingIndicator
}

func ResourceLabel(rt ResourceType) string {
	return rt.String()
}

func ContainerStateText(container ContainerRow) string {
	if container.Healthy && container.State == StateRunning {
		return "● healthy"
	}
	return "● " + string(container.State)
}
