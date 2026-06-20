package core

import (
	"context"
	"errors"
	"testing"

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
	snapshot, err := svc.LoadSnapshot(context.Background())
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
	_, err := svc.LoadSnapshot(context.Background())
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
	snapshot, err := svc.LoadSnapshot(context.Background())
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
	snapshot, err := svc.LoadSnapshot(context.Background())
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
