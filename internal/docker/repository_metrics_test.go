package docker

import (
	"math"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestComputeCPUPercent(t *testing.T) {
	tests := []struct {
		name  string
		stats container.StatsResponse
		want  float64
	}{
		{
			name: "normal with online cpus",
			stats: container.StatsResponse{
				Stats: container.Stats{
					CPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 200},
						SystemUsage: 1100,
						OnlineCPUs:  2,
					},
					PreCPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 100},
						SystemUsage: 100,
					},
				},
			},
			want: 20,
		},
		{
			name: "fallback to percpu usage length",
			stats: container.StatsResponse{
				Stats: container.Stats{
					CPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 200, PercpuUsage: []uint64{1, 1, 1, 1}},
						SystemUsage: 1100,
						OnlineCPUs:  0,
					},
					PreCPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 100},
						SystemUsage: 100,
					},
				},
			},
			want: 40,
		},
		{
			name: "fallback to single cpu",
			stats: container.StatsResponse{
				Stats: container.Stats{
					CPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 200},
						SystemUsage: 1100,
						OnlineCPUs:  0,
					},
					PreCPUStats: container.CPUStats{
						CPUUsage:    container.CPUUsage{TotalUsage: 100},
						SystemUsage: 100,
					},
				},
			},
			want: 10,
		},
		{
			name: "zero cpu delta",
			stats: container.StatsResponse{
				Stats: container.Stats{
					CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 1100, OnlineCPUs: 2},
					PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 100},
				},
			},
			want: 0,
		},
		{
			name: "zero system delta",
			stats: container.StatsResponse{
				Stats: container.Stats{
					CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 200}, SystemUsage: 100, OnlineCPUs: 2},
					PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 100}, SystemUsage: 100},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCPUPercent(tt.stats)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Fatalf("computeCPUPercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEffectiveMemoryUsage(t *testing.T) {
	tests := []struct {
		name  string
		stats container.StatsResponse
		want  uint64
	}{
		{
			name: "subtracts total_inactive_file first",
			stats: container.StatsResponse{
				Stats: container.Stats{
					MemoryStats: container.MemoryStats{
						Usage: 1000,
						Stats: map[string]uint64{"total_inactive_file": 100, "cache": 300},
					},
				},
			},
			want: 900,
		},
		{
			name: "subtracts cache when others absent",
			stats: container.StatsResponse{
				Stats: container.Stats{
					MemoryStats: container.MemoryStats{
						Usage: 1000,
						Stats: map[string]uint64{"cache": 250},
					},
				},
			},
			want: 750,
		},
		{
			name: "does not underflow when cached exceeds usage",
			stats: container.StatsResponse{
				Stats: container.Stats{
					MemoryStats: container.MemoryStats{
						Usage: 200,
						Stats: map[string]uint64{"cache": 250},
					},
				},
			},
			want: 200,
		},
		{
			name: "no cache keys",
			stats: container.StatsResponse{
				Stats: container.Stats{MemoryStats: container.MemoryStats{Usage: 1234, Stats: map[string]uint64{}}},
			},
			want: 1234,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveMemoryUsage(tt.stats)
			if got != tt.want {
				t.Fatalf("effectiveMemoryUsage() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestComputeMemoryUsage(t *testing.T) {
	t.Run("with limit", func(t *testing.T) {
		stats := container.StatsResponse{
			Stats: container.Stats{
				MemoryStats: container.MemoryStats{
					Usage: 2048,
					Limit: 2048,
					Stats: map[string]uint64{"cache": 512},
				},
			},
		}

		percent, usage, limit, usedBytes, maxBytes := computeMemoryUsage(stats)

		if math.Abs(percent-75.0) > 0.0001 {
			t.Fatalf("percent = %v, want 75", percent)
		}
		if usage != "1.5 KiB" {
			t.Fatalf("usage = %q, want %q", usage, "1.5 KiB")
		}
		if limit != "2.0 KiB" {
			t.Fatalf("limit = %q, want %q", limit, "2.0 KiB")
		}
		if usedBytes != 1536 || maxBytes != 2048 {
			t.Fatalf("bytes = (%d, %d), want (1536, 2048)", usedBytes, maxBytes)
		}
	})

	t.Run("without limit", func(t *testing.T) {
		stats := container.StatsResponse{
			Stats: container.Stats{MemoryStats: container.MemoryStats{Usage: 1536, Limit: 0, Stats: map[string]uint64{}}},
		}

		percent, usage, limit, usedBytes, maxBytes := computeMemoryUsage(stats)
		if percent != 0 {
			t.Fatalf("percent = %v, want 0", percent)
		}
		if usage != "1.5 KiB" {
			t.Fatalf("usage = %q, want %q", usage, "1.5 KiB")
		}
		if limit != "-" {
			t.Fatalf("limit = %q, want '-'", limit)
		}
		if usedBytes != 1536 || maxBytes != 0 {
			t.Fatalf("bytes = (%d, %d), want (1536, 0)", usedBytes, maxBytes)
		}
	})
}
