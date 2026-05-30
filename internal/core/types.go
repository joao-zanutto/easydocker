package core

import (
	"strings"
	"time"
)

type ResourceType int

const (
	ResourceContainer ResourceType = iota
	ResourceImage
	ResourceNetwork
	ResourceVolume
)

type Snapshot struct {
	Containers      []ContainerRow
	ComposeProjects []ComposeProject
	Images          []ImageRow
	Networks        []NetworkRow
	Volumes         []VolumeRow
	TotalCPU        float64
	TotalMem        uint64
	TotalLimit      uint64
	Timestamp       time.Time
}

type ContainerRow struct {
	ID                     string
	FullID                 string
	Name                   string
	ComposeProject         string
	ComposeService         string
	ComposeWorkingDir      string
	ComposeConfigFiles     string
	Image                  string
	State                  string
	Status                 string
	Ports                  string
	Command                string
	CreatedUnix            int64
	CPUPercent             float64
	MemoryPercent          float64
	MemoryUsage            string
	MemoryLimit            string
	MemoryUsageBytes       uint64
	MemoryLimitBytes       uint64
	Healthy                bool
}

type ComposeProject struct {
	Name             string
	Containers       []ContainerRow
	ContainerCount   int
	RunningCount     int
	HealthyCount     int
	CreatedUnix      int64
	Created          string
	Network          string
	WorkingDir       string
	ConfigFiles      string
	Services         []string
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsage      string
	MemoryLimit      string
	MemoryUsageBytes uint64
	MemoryLimitBytes uint64
}

type ImageRow struct {
	ID          string
	Tags        string
	Size        string
	Created     string
	CreatedUnix int64
	Containers  int64
}

type NetworkRow struct {
	ID         string
	Name       string
	Driver     string
	Scope      string
	Internal   string
	Attachable string
	Endpoints  int
	Created    string
	CreatedAt  time.Time
}

type VolumeRow struct {
	Name       string
	Driver     string
	Scope      string
	Mountpoint string
	Size       string
	RefCount   int64
	Created    string
	CreatedAt  string
}

type ContainerMetrics struct {
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsage      string
	MemoryLimit      string
	MemoryUsageBytes uint64
	MemoryLimitBytes uint64
}

func ApplyMetricsToContainers(rows []ContainerRow, metricsByID map[string]ContainerMetrics) []ContainerRow {
	updated := make([]ContainerRow, len(rows))
	copy(updated, rows)
	for index, row := range updated {
		metrics, ok := metricsByID[row.FullID]
		if !ok {
			continue
		}
		updated[index].CPUPercent = metrics.CPUPercent
		updated[index].MemoryPercent = metrics.MemoryPercent
		updated[index].MemoryUsage = metrics.MemoryUsage
		updated[index].MemoryLimit = metrics.MemoryLimit
		updated[index].MemoryUsageBytes = metrics.MemoryUsageBytes
		updated[index].MemoryLimitBytes = metrics.MemoryLimitBytes
	}
	return updated
}

func PreserveRunningContainerMetrics(currentRows, previousRows []ContainerRow) []ContainerRow {
	if len(currentRows) == 0 || len(previousRows) == 0 {
		return currentRows
	}

	previousByID := make(map[string]ContainerRow, len(previousRows))
	for _, row := range previousRows {
		previousByID[row.FullID] = row
	}

	merged := make([]ContainerRow, len(currentRows))
	copy(merged, currentRows)
	for index, row := range merged {
		if !strings.EqualFold(row.State, "running") {
			continue
		}
		if row.CPUPercent >= 0 && row.MemoryUsage != "-" && row.MemoryUsage != "loading" {
			continue
		}
		previous, ok := previousByID[row.FullID]
		if !ok || previous.MemoryUsage == "-" || previous.MemoryUsage == "loading" {
			continue
		}
		merged[index].CPUPercent = previous.CPUPercent
		merged[index].MemoryPercent = previous.MemoryPercent
		merged[index].MemoryUsage = previous.MemoryUsage
		merged[index].MemoryLimit = previous.MemoryLimit
		merged[index].MemoryUsageBytes = previous.MemoryUsageBytes
		merged[index].MemoryLimitBytes = previous.MemoryLimitBytes
	}

	return merged
}
