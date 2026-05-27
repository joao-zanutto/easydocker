package docker

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"time"

	"easydocker/internal/core"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (r *Repository) LoadContainerMetrics(ctx context.Context, rows []core.ContainerRow) (map[string]core.ContainerMetrics, float64, uint64, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return nil, 0, 0, err
	}

	runningRows := make([]core.ContainerRow, 0, len(rows))
	for _, row := range rows {
		if strings.EqualFold(row.State, "running") {
			runningRows = append(runningRows, row)
		}
	}
	if len(runningRows) == 0 {
		return map[string]core.ContainerMetrics{}, 0, 0, nil
	}

	metricsByID := make(map[string]core.ContainerMetrics, len(runningRows))
	workerCount := min(len(runningRows), max(2, min(runtime.NumCPU(), 6)))
	jobs := make(chan core.ContainerRow)
	var mu sync.Mutex
	var totalCPU float64
	var totalMem uint64
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for row := range jobs {
			metrics, err := r.loadSingleContainerMetrics(ctx, cli, row.FullID)
			if err != nil {
				continue
			}

			mu.Lock()
			metricsByID[row.FullID] = metrics
			totalCPU += metrics.CPUPercent
			totalMem += metrics.MemoryUsageBytes
			mu.Unlock()
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
	for _, row := range runningRows {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return metricsByID, totalCPU, totalMem, ctx.Err()
		case jobs <- row:
		}
	}
	close(jobs)
	wg.Wait()

	return metricsByID, totalCPU, totalMem, nil
}

func (r *Repository) loadSingleContainerMetrics(ctx context.Context, cli *client.Client, containerID string) (core.ContainerMetrics, error) {
	return retryWithBackoff(ctx, 2, 1500*time.Millisecond, func(ctx context.Context) (core.ContainerMetrics, error) {
		statsReader, err := cli.ContainerStats(ctx, containerID, false)
		if err != nil {
			return core.ContainerMetrics{}, err
		}
		defer func() { _ = statsReader.Body.Close() }()

		var stats container.StatsResponse
		if err := json.NewDecoder(statsReader.Body).Decode(&stats); err != nil {
			return core.ContainerMetrics{}, err
		}

		cpuPercent := computeCPUPercent(stats)
		memoryPercent, memoryUsage, memoryLimit, memoryBytes, memoryMax := computeMemoryUsage(stats)

		return core.ContainerMetrics{
			CPUPercent:       cpuPercent,
			MemoryPercent:    memoryPercent,
			MemoryUsage:      memoryUsage,
			MemoryLimit:      memoryLimit,
			MemoryUsageBytes: memoryBytes,
			MemoryLimitBytes: memoryMax,
		}, nil
	})
}

func retryWithBackoff[T any](ctx context.Context, maxAttempts int, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		opCtx, cancel := context.WithTimeout(ctx, timeout)
		result, err := fn(opCtx)
		cancel()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	var zero T
	return zero, lastErr
}

func computeCPUPercent(stats container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if onlineCPUs == 0 {
		onlineCPUs = 1
	}
	return (cpuDelta / systemDelta) * onlineCPUs * 100
}

func computeMemoryUsage(stats container.StatsResponse) (float64, string, string, uint64, uint64) {
	used := effectiveMemoryUsage(stats)
	limit := stats.MemoryStats.Limit
	if limit == 0 {
		return 0, core.HumanBytes(used), "-", used, 0
	}
	percent := (float64(used) / float64(limit)) * 100
	return percent, core.HumanBytes(used), core.HumanBytes(limit), used, limit
}

func effectiveMemoryUsage(stats container.StatsResponse) uint64 {
	used := stats.MemoryStats.Usage
	for _, key := range []string{"total_inactive_file", "inactive_file", "cache"} {
		if cached, ok := stats.MemoryStats.Stats[key]; ok && used >= cached {
			used -= cached
			break
		}
	}
	return used
}
