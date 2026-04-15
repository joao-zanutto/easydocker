package core

import "testing"

func TestFilterContainersByScope_ShowAllReturnsOriginalSlice(t *testing.T) {
	containers := []ContainerRow{{Name: "one"}, {Name: "two"}}

	got := FilterContainersByScope(containers, true)
	if len(got) != len(containers) {
		t.Fatalf("FilterContainersByScope(..., true) len = %d, want %d", len(got), len(containers))
	}
	if &got[0] != &containers[0] {
		t.Fatalf("FilterContainersByScope(..., true) should return original slice")
	}
}

func TestFilterContainersByScope_RunningOnlyCaseInsensitive(t *testing.T) {
	containers := []ContainerRow{
		{Name: "run-1", State: "running"},
		{Name: "run-2", State: "RUNNING"},
		{Name: "stopped", State: "exited"},
		{Name: "other", State: "created"},
	}

	got := FilterContainersByScope(containers, false)
	if len(got) != 2 {
		t.Fatalf("FilterContainersByScope(..., false) len = %d, want 2", len(got))
	}
	if got[0].Name != "run-1" || got[1].Name != "run-2" {
		t.Fatalf("FilterContainersByScope(..., false) = %#v, want running rows only", got)
	}
}

// Tests for FilterContainersByQuery
func TestFilterContainersByQuery_EmptyQueryReturnsAll(t *testing.T) {
	containers := []ContainerRow{
		{Name: "web", Image: "nginx:latest"},
		{Name: "db", Image: "postgres:13"},
	}

	got := FilterContainersByQuery(containers, "")
	if len(got) != 2 {
		t.Fatalf("FilterContainersByQuery with empty query should return all, got %d", len(got))
	}

	got = FilterContainersByQuery(containers, "   ")
	if len(got) != 2 {
		t.Fatalf("FilterContainersByQuery with whitespace query should return all, got %d", len(got))
	}
}

func TestFilterContainersByQuery_MatchesByNameCaseInsensitive(t *testing.T) {
	containers := []ContainerRow{
		{Name: "web-server", Image: "nginx:latest"},
		{Name: "db-master", Image: "postgres:13"},
		{Name: "WEB-cache", Image: "redis:latest"},
	}

	got := FilterContainersByQuery(containers, "web")
	if len(got) != 2 {
		t.Fatalf("FilterContainersByQuery('web') len = %d, want 2", len(got))
	}
	if got[0].Name != "web-server" || got[1].Name != "WEB-cache" {
		t.Fatalf("FilterContainersByQuery('web') returned wrong containers")
	}
}

func TestFilterContainersByQuery_MatchesByImageCaseInsensitive(t *testing.T) {
	containers := []ContainerRow{
		{Name: "container1", Image: "nginx:latest"},
		{Name: "container2", Image: "postgres:13"},
		{Name: "container3", Image: "NGINX:alpine"},
	}

	got := FilterContainersByQuery(containers, "nginx")
	if len(got) != 2 {
		t.Fatalf("FilterContainersByQuery('nginx') len = %d, want 2", len(got))
	}
}

// Tests for FilterImagesByQuery
func TestFilterImagesByQuery_EmptyQueryReturnsAll(t *testing.T) {
	images := []ImageRow{
		{Tags: "nginx:latest"},
		{Tags: "postgres:13"},
	}

	got := FilterImagesByQuery(images, "")
	if len(got) != 2 {
		t.Fatalf("FilterImagesByQuery with empty query should return all, got %d", len(got))
	}
}

func TestFilterImagesByQuery_MatchesByNameCaseInsensitive(t *testing.T) {
	images := []ImageRow{
		{Tags: "nginx:latest"},
		{Tags: "postgres:13"},
		{Tags: "NGINX:alpine"},
	}

	got := FilterImagesByQuery(images, "nginx")
	if len(got) != 2 {
		t.Fatalf("FilterImagesByQuery('nginx') len = %d, want 2", len(got))
	}
}

// Tests for FilterNetworksByQuery
func TestFilterNetworksByQuery_EmptyQueryReturnsAll(t *testing.T) {
	networks := []NetworkRow{
		{Name: "bridge"},
		{Name: "overlay"},
	}

	got := FilterNetworksByQuery(networks, "")
	if len(got) != 2 {
		t.Fatalf("FilterNetworksByQuery with empty query should return all, got %d", len(got))
	}
}

func TestFilterNetworksByQuery_MatchesByNameCaseInsensitive(t *testing.T) {
	networks := []NetworkRow{
		{Name: "bridge-net"},
		{Name: "overlay-net"},
		{Name: "BRIDGE-cache"},
	}

	got := FilterNetworksByQuery(networks, "bridge")
	if len(got) != 2 {
		t.Fatalf("FilterNetworksByQuery('bridge') len = %d, want 2", len(got))
	}
}

// Tests for FilterVolumesByQuery
func TestFilterVolumesByQuery_EmptyQueryReturnsAll(t *testing.T) {
	volumes := []VolumeRow{
		{Name: "data"},
		{Name: "cache"},
	}

	got := FilterVolumesByQuery(volumes, "")
	if len(got) != 2 {
		t.Fatalf("FilterVolumesByQuery with empty query should return all, got %d", len(got))
	}
}

func TestFilterVolumesByQuery_MatchesByNameCaseInsensitive(t *testing.T) {
	volumes := []VolumeRow{
		{Name: "data-vol"},
		{Name: "cache-vol"},
		{Name: "DATA-backup"},
	}

	got := FilterVolumesByQuery(volumes, "data")
	if len(got) != 2 {
		t.Fatalf("FilterVolumesByQuery('data') len = %d, want 2", len(got))
	}
}

