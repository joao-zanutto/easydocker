package docker

import (
	"strings"
	"testing"
	"time"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

func TestMapContainerRow(t *testing.T) {
	item := types.Container{
		ID:      "sha256:1234567890abcdef1234",
		Names:   []string{"/api"},
		Image:   "nginx:latest",
		State:   "running",
		Status:  "Up 3 minutes (healthy)",
		Command: "   /bin/sh -c \"sleep 100\"   ",
		Created: 42,
		Ports: []types.Port{
			{PrivatePort: 443, PublicPort: 8443, Type: "tcp"},
			{PrivatePort: 80, PublicPort: 0, Type: "tcp"},
			{PrivatePort: 53, PublicPort: 0, Type: "udp"},
		},
	}

	row := mapContainerRow(item)

	if row.ID != "1234567890ab" {
		t.Fatalf("ID = %q, want %q", row.ID, "1234567890ab")
	}
	if row.FullID != item.ID {
		t.Fatalf("FullID = %q, want %q", row.FullID, item.ID)
	}
	if row.Name != "api" {
		t.Fatalf("Name = %q, want %q", row.Name, "api")
	}
	if row.Ports != "53/udp, 80/tcp, 8443->443/tcp" {
		t.Fatalf("Ports = %q, want sorted formatted ports", row.Ports)
	}
	if row.Command != "/bin/sh -c \"sleep 100\"" {
		t.Fatalf("Command = %q, want trimmed command", row.Command)
	}
	if !row.Healthy {
		t.Fatalf("Healthy = false, want true")
	}
	if row.MemoryUsage != "-" || row.MemoryLimit != "-" {
		t.Fatalf("default memory fields should be '-' but got usage=%q limit=%q", row.MemoryUsage, row.MemoryLimit)
	}
}

func TestMapImageRow(t *testing.T) {
	item := image.Summary{
		ID:         "sha256:abcdef0123456789abcd",
		RepoTags:   nil,
		Size:       1024,
		Created:    time.Now().Add(-2 * time.Hour).Unix(),
		Containers: 3,
	}

	row := mapImageRow(item)

	if row.ID != "abcdef012345" {
		t.Fatalf("ID = %q, want %q", row.ID, "abcdef012345")
	}
	if row.Tags != "<none>:<none>" {
		t.Fatalf("Tags = %q, want <none>:<none>", row.Tags)
	}
	if row.Size != "1.0 KiB" {
		t.Fatalf("Size = %q, want %q", row.Size, "1.0 KiB")
	}
	if row.Containers != 3 {
		t.Fatalf("Containers = %d, want 3", row.Containers)
	}
	if row.CreatedUnix != item.Created {
		t.Fatalf("CreatedUnix = %d, want %d", row.CreatedUnix, item.Created)
	}
	if row.Created != "just now" && !strings.HasSuffix(row.Created, "ago") {
		t.Fatalf("Created = %q, want relative time", row.Created)
	}
}

func TestMapNetworkRow(t *testing.T) {
	created := time.Now().Add(-3 * time.Hour)
	item := network.Inspect{
		ID:         "sha256:1234567890abcdef1234",
		Name:       "bridge-x",
		Driver:     "bridge",
		Scope:      "local",
		Internal:   true,
		Attachable: false,
		Created:    created,
		Containers: map[string]network.EndpointResource{"1": {}, "2": {}},
	}

	row := mapNetworkRow(item)

	if row.ID != "1234567890ab" {
		t.Fatalf("ID = %q, want %q", row.ID, "1234567890ab")
	}
	if row.Internal != "yes" || row.Attachable != "no" {
		t.Fatalf("Internal/Attachable = %q/%q, want yes/no", row.Internal, row.Attachable)
	}
	if row.Endpoints != 2 {
		t.Fatalf("Endpoints = %d, want 2", row.Endpoints)
	}
	if !row.CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt = %v, want %v", row.CreatedAt, created)
	}
	if row.Created != "just now" && !strings.HasSuffix(row.Created, "ago") {
		t.Fatalf("Created = %q, want relative time", row.Created)
	}
}

func TestMapVolumeRow(t *testing.T) {
	created := time.Now().Add(-2*time.Hour - 5*time.Minute).UTC().Format(time.RFC3339Nano)
	item := &volume.Volume{
		Name:       "cache",
		Driver:     "local",
		Scope:      "local",
		Mountpoint: "/var/lib/docker/volumes/cache/_data",
		CreatedAt:  created,
		UsageData: &volume.UsageData{
			RefCount: 3,
			Size:     2048,
		},
	}

	row := mapVolumeRow(item)

	if row.Name != "cache" {
		t.Fatalf("Name = %q, want cache", row.Name)
	}
	if row.RefCount != 3 {
		t.Fatalf("RefCount = %d, want 3", row.RefCount)
	}
	if row.Size != "2.0 KiB" {
		t.Fatalf("Size = %q, want 2.0 KiB", row.Size)
	}
	if row.Created != "just now" && !strings.HasSuffix(row.Created, "ago") {
		t.Fatalf("Created = %q, want relative time", row.Created)
	}
}

func TestMapVolumeRow_UnknownUsageAndInvalidTimestamp(t *testing.T) {
	item := &volume.Volume{
		Name:       "cache",
		Driver:     "local",
		Scope:      "local",
		Mountpoint: "/var/lib/docker/volumes/cache/_data",
		CreatedAt:  "not-a-timestamp",
		UsageData:  nil,
	}

	row := mapVolumeRow(item)

	if row.RefCount != -1 {
		t.Fatalf("RefCount = %d, want -1", row.RefCount)
	}
	if row.Size != "-" {
		t.Fatalf("Size = %q, want '-'", row.Size)
	}
	if row.Created != "not-a-timestamp" {
		t.Fatalf("Created = %q, want original value", row.Created)
	}
}

func TestFormatPorts(t *testing.T) {
	if got := formatPorts(nil); got != "-" {
		t.Fatalf("formatPorts(nil) = %q, want '-'", got)
	}

	ports := []types.Port{
		{PrivatePort: 8080, PublicPort: 18080, Type: "tcp"},
		{PrivatePort: 53, PublicPort: 0, Type: "udp"},
		{PrivatePort: 8080, PublicPort: 0, Type: "tcp"},
	}

	got := formatPorts(ports)
	want := "53/udp, 8080/tcp, 18080->8080/tcp"
	if got != want {
		t.Fatalf("formatPorts(...) = %q, want %q", got, want)
	}
}

func TestCleanCommand(t *testing.T) {
	if got := cleanCommand("   "); got != "-" {
		t.Fatalf("cleanCommand(blank) = %q, want '-'", got)
	}
	if got := cleanCommand("  echo hello  "); got != "echo hello" {
		t.Fatalf("cleanCommand(trim) = %q, want 'echo hello'", got)
	}

	long := strings.Repeat("x", 70)
	got := cleanCommand(long)
	if len(got) != 64 {
		t.Fatalf("cleanCommand(long) len = %d, want 64", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("cleanCommand(long) = %q, want ellipsis suffix", got)
	}
}

func TestFormatTags(t *testing.T) {
	if got := formatTags(nil); got != "<none>:<none>" {
		t.Fatalf("formatTags(nil) = %q, want <none>:<none>", got)
	}
	if got := formatTags([]string{"a:1", "b:2"}); got != "a:1, b:2" {
		t.Fatalf("formatTags(...) = %q, want %q", got, "a:1, b:2")
	}
}

func TestShortID(t *testing.T) {
	if got := shortID("sha256:1234567890abcdef"); got != "1234567890ab" {
		t.Fatalf("shortID(sha) = %q, want %q", got, "1234567890ab")
	}
	if got := shortID("short"); got != "short" {
		t.Fatalf("shortID(short) = %q, want short", got)
	}
}
