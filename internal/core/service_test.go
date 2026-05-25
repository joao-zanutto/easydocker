package core

import (
	"context"
	"errors"
	"testing"
	"time"

	gomock "go.uber.org/mock/gomock"
)

func TestServiceLoadSnapshot_ComposesDataAndMetrics(t *testing.T) {
	rows := []ContainerRow{{FullID: "id-1", Name: "one", ComposeProject: "shop"}, {FullID: "id-2", Name: "two", ComposeProject: "shop"}}
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

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().LoadContainerRows(gomock.Any()).Return(rows, nil)
	repo.EXPECT().LoadSupportingResources(gomock.Any()).Return(resources, nil)
	repo.EXPECT().LoadContainerMetrics(gomock.Any(), rows).Return(metrics, float64(99.9), uint64(12345), nil)

	svc := NewService(repo)
	snapshot, err := svc.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v, want nil", err)
	}

	if len(snapshot.Containers) != 2 {
		t.Fatalf("snapshot.Containers len = %d, want 2", len(snapshot.Containers))
	}
	if len(snapshot.ComposeProjects) != 1 {
		t.Fatalf("snapshot.ComposeProjects len = %d, want 1", len(snapshot.ComposeProjects))
	}
	if snapshot.ComposeProjects[0].Name != "shop" || snapshot.ComposeProjects[0].ContainerCount != 2 {
		t.Fatalf("compose project summary = %#v, want shop with 2 containers", snapshot.ComposeProjects[0])
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

func TestServiceLoadSnapshot_FailsOnContainerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().LoadContainerRows(gomock.Any()).Return(nil, errors.New("docker not available"))
	// LoadSupportingResources runs in parallel — it may or may not be called.
	repo.EXPECT().LoadSupportingResources(gomock.Any()).Return(Snapshot{}, nil).AnyTimes()

	svc := NewService(repo)
	_, err := svc.LoadSnapshot()
	if err == nil {
		t.Fatalf("LoadSnapshot() error = nil, want non-nil")
	}
}

func TestServiceLoadSnapshot_ReturnsPartialWhenResourcesFail(t *testing.T) {
	rows := []ContainerRow{{FullID: "id-1", Name: "one"}, {FullID: "id-2", Name: "two"}}
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().LoadContainerRows(gomock.Any()).Return(rows, nil)
	repo.EXPECT().LoadSupportingResources(gomock.Any()).Return(Snapshot{}, errors.New("resource error"))

	svc := NewService(repo)
	snapshot, err := svc.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v, want nil (partial tolerance)", err)
	}

	if len(snapshot.Containers) != 2 {
		t.Fatalf("snapshot.Containers len = %d, want 2", len(snapshot.Containers))
	}
	if snapshot.ComposeProjects == nil {
		t.Fatalf("ComposeProjects should be computed even when resources fail")
	}
	if snapshot.Timestamp.IsZero() {
		t.Fatalf("snapshot timestamp should be populated")
	}
}

func TestServiceLoadSnapshot_ReturnsPartialWhenMetricsFail(t *testing.T) {
	rows := []ContainerRow{{FullID: "id-1", Name: "one", ComposeProject: "shop"}}
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().LoadContainerRows(gomock.Any()).Return(rows, nil)
	repo.EXPECT().LoadSupportingResources(gomock.Any()).Return(Snapshot{Images: []ImageRow{{ID: "img"}}}, nil)
	repo.EXPECT().LoadContainerMetrics(gomock.Any(), rows).Return(nil, float64(0), uint64(0), errors.New("metrics error"))

	svc := NewService(repo)
	snapshot, err := svc.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v, want nil (partial tolerance)", err)
	}

	if len(snapshot.Containers) != 1 {
		t.Fatalf("snapshot.Containers len = %d, want 1", len(snapshot.Containers))
	}
	if len(snapshot.Images) != 1 {
		t.Fatalf("snapshot.Images len = %d, want 1 (resources should still be present)", len(snapshot.Images))
	}
	if snapshot.Containers[0].CPUPercent != 0 {
		t.Fatalf("container CPU should be 0 when metrics fail, got %v", snapshot.Containers[0].CPUPercent)
	}
	if snapshot.Timestamp.IsZero() {
		t.Fatalf("snapshot timestamp should be populated")
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
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().LoadContainerLiveData(gomock.Any(), "id-1", nil, nil, tt.tail).DoAndReturn(
				func(ctx context.Context, _ string, _, _ []float64, _ int) (ContainerLiveData, error) {
					deadline, ok := ctx.Deadline()
					if !ok {
						t.Fatalf("LoadContainerLiveData context should have a deadline")
					}
					gotDuration = time.Until(deadline)
					return ContainerLiveData{ContainerID: "id-1"}, nil
				},
			)

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
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().LoadContainerLiveData(gomock.Any(), "id-1", nil, nil, tt.tail).DoAndReturn(
				func(ctx context.Context, _ string, _, _ []float64, _ int) (ContainerLiveData, error) {
					deadline, ok := ctx.Deadline()
					if !ok {
						t.Fatalf("LoadContainerLiveData context should have a deadline")
					}
					gotDuration = time.Until(deadline)
					return ContainerLiveData{ContainerID: "id-1"}, nil
				},
			)

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
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().LoadContainerLiveData(gomock.Any(), "id-1", nil, nil, 100).DoAndReturn(
		func(ctx context.Context, _ string, _, _ []float64, _ int) (ContainerLiveData, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatalf("LoadContainerLiveData context should have a deadline")
			}
			gotDuration = time.Until(deadline)
			return ContainerLiveData{ContainerID: "id-1"}, nil
		},
	)

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
