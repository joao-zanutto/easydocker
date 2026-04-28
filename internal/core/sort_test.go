package core

import (
	"testing"
	"time"
)

func TestContainerStateRank(t *testing.T) {
	tests := []struct {
		name      string
		container ContainerRow
		want      int
	}{
		{name: "running healthy", container: ContainerRow{State: "running", Healthy: true}, want: 0},
		{name: "running", container: ContainerRow{State: "running", Healthy: false}, want: 1},
		{name: "created", container: ContainerRow{State: "created"}, want: 2},
		{name: "restarting", container: ContainerRow{State: "restarting"}, want: 3},
		{name: "paused", container: ContainerRow{State: "paused"}, want: 3},
		{name: "exited", container: ContainerRow{State: "exited"}, want: 4},
		{name: "stopped", container: ContainerRow{State: "stopped"}, want: 4},
		{name: "dead", container: ContainerRow{State: "dead"}, want: 5},
		{name: "unknown", container: ContainerRow{State: "whatever"}, want: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containerStateRank(tt.container); got != tt.want {
				t.Fatalf("containerStateRank(%+v) = %d, want %d", tt.container, got, tt.want)
			}
		})
	}
}

func TestSortContainers_StateRankThenCreatedThenName(t *testing.T) {
	rows := []ContainerRow{
		{Name: "zeta", State: "running", Healthy: false, CreatedUnix: 100},
		{Name: "alpha", State: "running", Healthy: false, CreatedUnix: 100},
		{Name: "newer", State: "running", Healthy: false, CreatedUnix: 200},
		{Name: "healthy", State: "running", Healthy: true, CreatedUnix: 50},
		{Name: "created", State: "created", CreatedUnix: 999},
	}

	SortContainers(rows)

	wantNames := []string{"healthy", "newer", "alpha", "zeta", "created"}
	for i, want := range wantNames {
		if rows[i].Name != want {
			t.Fatalf("rows[%d].Name = %q, want %q", i, rows[i].Name, want)
		}
	}
}

func TestSortImages_RepositoryThenTagAsc(t *testing.T) {
	rows := []ImageRow{
		{Tags: "redis:7", CreatedUnix: 20},
		{Tags: "alpine:3.20", CreatedUnix: 5},
		{Tags: "alpine:3.18", CreatedUnix: 50},
		{Tags: "ghcr.io:5000/app/backend:v1", CreatedUnix: 40},
		{Tags: "ghcr.io:5000/app/backend:v0", CreatedUnix: 10},
	}

	SortImages(rows)

	wantTags := []string{
		"alpine:3.18",
		"alpine:3.20",
		"ghcr.io:5000/app/backend:v0",
		"ghcr.io:5000/app/backend:v1",
		"redis:7",
	}
	for i, want := range wantTags {
		if rows[i].Tags != want {
			t.Fatalf("rows[%d].Tags = %q, want %q", i, rows[i].Tags, want)
		}
	}
}

func TestSortNetworks_CreatedDescThenNameAsc(t *testing.T) {
	older := time.Unix(1000, 0)
	newer := time.Unix(2000, 0)
	rows := []NetworkRow{
		{Name: "z", CreatedAt: older},
		{Name: "a", CreatedAt: older},
		{Name: "n", CreatedAt: newer},
	}

	SortNetworks(rows)

	wantNames := []string{"n", "a", "z"}
	for i, want := range wantNames {
		if rows[i].Name != want {
			t.Fatalf("rows[%d].Name = %q, want %q", i, rows[i].Name, want)
		}
	}
}

func TestSortVolumes_CreatedDescThenNameAsc(t *testing.T) {
	rows := []VolumeRow{
		{Name: "z", CreatedAt: "2024-01-01T00:00:00Z"},
		{Name: "a", CreatedAt: "2024-01-01T00:00:00Z"},
		{Name: "n", CreatedAt: "2025-01-01T00:00:00Z"},
	}

	SortVolumes(rows)

	wantNames := []string{"n", "a", "z"}
	for i, want := range wantNames {
		if rows[i].Name != want {
			t.Fatalf("rows[%d].Name = %q, want %q", i, rows[i].Name, want)
		}
	}
}
