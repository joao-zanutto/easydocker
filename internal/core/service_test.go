package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type mockRepository struct {
	loadContainerRowsFn       func(ctx context.Context) ([]ContainerRow, error)
	loadSupportingResourcesFn func(ctx context.Context) (Snapshot, error)
	loadContainerMetricsFn    func(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error)
	loadContainerLiveDataFn   func(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error)

	calls []string
}

func (m *mockRepository) LoadContainerRows(ctx context.Context) ([]ContainerRow, error) {
	m.calls = append(m.calls, "rows")
	if m.loadContainerRowsFn != nil {
		return m.loadContainerRowsFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) LoadSupportingResources(ctx context.Context) (Snapshot, error) {
	m.calls = append(m.calls, "resources")
	if m.loadSupportingResourcesFn != nil {
		return m.loadSupportingResourcesFn(ctx)
	}
	return Snapshot{}, nil
}

func (m *mockRepository) LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
	m.calls = append(m.calls, "metrics")
	if m.loadContainerMetricsFn != nil {
		return m.loadContainerMetricsFn(ctx, rows)
	}
	return nil, 0, 0, nil
}

func (m *mockRepository) LoadContainerLiveData(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
	m.calls = append(m.calls, "live")
	if m.loadContainerLiveDataFn != nil {
		return m.loadContainerLiveDataFn(ctx, containerID, previousCPU, previousMem, tail)
	}
	return ContainerLiveData{}, nil
}

func TestServiceLoadSnapshot_ComposesDataAndMetrics(t *testing.T) {
	rows := []ContainerRow{{FullID: "id-1", Name: "one"}, {FullID: "id-2", Name: "two"}}
	metrics := map[string]ContainerMetrics{
		"id-1": {
			CPUPercent:       10.5,
			MemoryPercent:    33.0,
			MemoryUsage:      "512 MiB",
			MemoryLimit:      "2.0 GiB",
			MemoryUsageBytes: 512,
			MemoryLimitBytes: 2048,
		},
	}
	resources := Snapshot{
		Images:   []ImageRow{{ID: "img"}},
		Networks: []NetworkRow{{Name: "net"}},
		Volumes:  []VolumeRow{{Name: "vol"}},
	}

	repo := &mockRepository{}
	repo.loadContainerRowsFn = func(ctx context.Context) ([]ContainerRow, error) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatalf("LoadContainerRows context should have a deadline")
		}
		return rows, nil
	}
	repo.loadSupportingResourcesFn = func(ctx context.Context) (Snapshot, error) {
		return resources, nil
	}
	repo.loadContainerMetricsFn = func(ctx context.Context, gotRows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
		if !reflect.DeepEqual(gotRows, rows) {
			t.Fatalf("LoadContainerMetrics rows = %#v, want %#v", gotRows, rows)
		}
		return metrics, 99.9, 12345, nil
	}

	svc := NewService(repo)
	snapshot, err := svc.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v, want nil", err)
	}

	if !reflect.DeepEqual(repo.calls, []string{"rows", "resources", "metrics"}) {
		t.Fatalf("repository call order = %#v, want [rows resources metrics]", repo.calls)
	}

	if len(snapshot.Containers) != 2 {
		t.Fatalf("snapshot.Containers len = %d, want 2", len(snapshot.Containers))
	}
	if snapshot.Containers[0].CPUPercent != 10.5 {
		t.Fatalf("snapshot container CPU = %v, want 10.5", snapshot.Containers[0].CPUPercent)
	}
	if snapshot.Containers[1].CPUPercent != 0 {
		t.Fatalf("snapshot container without metrics should remain unchanged")
	}
	if snapshot.TotalCPU != 99.9 || snapshot.TotalMem != 12345 {
		t.Fatalf("snapshot totals = (%v, %v), want (99.9, 12345)", snapshot.TotalCPU, snapshot.TotalMem)
	}
	if snapshot.Timestamp.IsZero() {
		t.Fatalf("snapshot timestamp should be populated")
	}
}

func TestServiceLoadSnapshot_StopsOnResourceError(t *testing.T) {
	repo := &mockRepository{}
	repo.loadContainerRowsFn = func(ctx context.Context) ([]ContainerRow, error) {
		return []ContainerRow{{FullID: "id-1"}}, nil
	}
	repo.loadSupportingResourcesFn = func(ctx context.Context) (Snapshot, error) {
		return Snapshot{}, errors.New("boom")
	}
	repo.loadContainerMetricsFn = func(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
		t.Fatalf("LoadContainerMetrics should not be called after resource failure")
		return nil, 0, 0, nil
	}

	svc := NewService(repo)
	_, err := svc.LoadSnapshot()
	if err == nil {
		t.Fatalf("LoadSnapshot() error = nil, want non-nil")
	}
}

func TestServiceLoadContainerLiveData_UsesTailDependentTimeout(t *testing.T) {
	tests := []struct {
		name string
		tail int
		want time.Duration
	}{
		{name: "default timeout", tail: 100, want: 5 * time.Second},
		{name: "medium tail timeout", tail: 600, want: 20 * time.Second},
		{name: "all logs timeout", tail: 0, want: 60 * time.Second},
		{name: "large tail timeout", tail: 5000, want: 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotDuration time.Duration
			repo := &mockRepository{}
			repo.loadContainerLiveDataFn = func(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatalf("LoadContainerLiveData context should have a deadline")
				}
				gotDuration = time.Until(deadline)
				return ContainerLiveData{ContainerID: containerID}, nil
			}

			svc := NewService(repo)
			_, err := svc.LoadContainerLiveData("id-1", nil, nil, tt.tail)
			if err != nil {
				t.Fatalf("LoadContainerLiveData() error = %v, want nil", err)
			}

			assertDurationApprox(t, gotDuration, tt.want, 2*time.Second)
		})
	}
}

func TestServiceLoadContainerLiveData_UsesConfiguredTimeouts(t *testing.T) {
	config := ServiceConfig{
		RequestTimeout:              3 * time.Second,
		LiveDataMediumTailThreshold: 50,
		LiveDataMediumTailTimeout:   7 * time.Second,
		LiveDataLargeTailThreshold:  100,
		LiveDataLargeTailTimeout:    11 * time.Second,
	}

	tests := []struct {
		name string
		tail int
		want time.Duration
	}{
		{name: "uses configured default timeout", tail: 10, want: 3 * time.Second},
		{name: "uses configured medium timeout", tail: 60, want: 7 * time.Second},
		{name: "uses configured large timeout for tail all", tail: 0, want: 11 * time.Second},
		{name: "uses configured large timeout over large threshold", tail: 200, want: 11 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotDuration time.Duration
			repo := &mockRepository{}
			repo.loadContainerLiveDataFn = func(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatalf("LoadContainerLiveData context should have a deadline")
				}
				gotDuration = time.Until(deadline)
				return ContainerLiveData{ContainerID: containerID}, nil
			}

			svc := NewServiceWithConfig(repo, config)
			_, err := svc.LoadContainerLiveData("id-1", nil, nil, tt.tail)
			if err != nil {
				t.Fatalf("LoadContainerLiveData() error = %v, want nil", err)
			}

			assertDurationApprox(t, gotDuration, tt.want, 2*time.Second)
		})
	}
}

func TestNewServiceWithConfig_ZeroValuesUseDefaults(t *testing.T) {
	var gotDuration time.Duration
	repo := &mockRepository{}
	repo.loadContainerLiveDataFn = func(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("LoadContainerLiveData context should have a deadline")
		}
		gotDuration = time.Until(deadline)
		return ContainerLiveData{ContainerID: containerID}, nil
	}

	svc := NewServiceWithConfig(repo, ServiceConfig{})
	_, err := svc.LoadContainerLiveData("id-1", nil, nil, 100)
	if err != nil {
		t.Fatalf("LoadContainerLiveData() error = %v, want nil", err)
	}

	assertDurationApprox(t, gotDuration, 5*time.Second, 2*time.Second)
}

func assertDurationApprox(t *testing.T, got, want, tolerance time.Duration) {
	t.Helper()
	min := want - tolerance
	max := want + tolerance
	if got < min || got > max {
		t.Fatalf("duration = %v, want within [%v, %v]", got, min, max)
	}
}
