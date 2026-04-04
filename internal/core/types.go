package core

import "time"

type Snapshot struct {
	Containers []ContainerRow
	Images     []ImageRow
	Networks   []NetworkRow
	Volumes    []VolumeRow
	TotalCPU   float64
	TotalMem   uint64
	TotalLimit uint64
	Timestamp  time.Time
}

type ContainerRow struct {
	ID               string
	FullID           string
	Name             string
	Image            string
	State            string
	Status           string
	Ports            string
	Command          string
	CreatedUnix      int64
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsage      string
	MemoryLimit      string
	MemoryUsageBytes uint64
	MemoryLimitBytes uint64
	Healthy          bool
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

type ContainerLiveData struct {
	ContainerID   string
	Logs          []string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   string
	MemoryLimit   string
	MemoryBytes   uint64
	MemoryMax     uint64
	CPUHistory    []float64
	MemHistory    []float64
	State         string
	UpdatedAt     time.Time
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
