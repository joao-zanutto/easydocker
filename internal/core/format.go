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
		if strings.EqualFold(container.State, "running") {
			return metricsLoadingValue(loadingIndicator)
		}
		return "-"
	}
	if container.CPUPercent < 0.05 {
		if strings.EqualFold(container.State, "running") {
			return "0.0%"
		}
		return "-"
	}
	return fmt.Sprintf("%.1f%%", container.CPUPercent)
}

func ContainerMemoryTableValue(container ContainerRow, loadingIndicator string) string {
	if container.MemoryUsage == "-" || strings.EqualFold(container.MemoryUsage, "loading") {
		if strings.EqualFold(container.State, "running") {
			return metricsLoadingValue(loadingIndicator)
		}
		return "-"
	}
	return container.MemoryUsage
}

func metricsLoadingValue(loadingIndicator string) string {
	if strings.TrimSpace(loadingIndicator) == "" {
		return "-"
	}
	return loadingIndicator
}

func ResourceLabel(rt ResourceType) string {
	switch rt {
	case ResourceContainer:
		return "Containers"
	case ResourceImage:
		return "Images"
	case ResourceNetwork:
		return "Networks"
	case ResourceVolume:
		return "Volumes"
	default:
		return "Containers"
	}
}

func ContainerStateText(container ContainerRow) string {
	if container.Healthy && container.State == "running" {
		return "● healthy"
	}
	return "● " + container.State
}
