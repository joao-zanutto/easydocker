package core

import (
	"testing"
	"time"
)

func TestAggregateComposeProjects_GroupsSortsAndRollsUpMetrics(t *testing.T) {
	now := time.Now()
	containers := []ContainerRow{
		{
			FullID:             "c-3",
			Name:               "worker",
			ComposeProject:     "shop",
			ComposeService:     "worker",
			ComposeWorkingDir:  "/srv/shop",
			ComposeConfigFiles: "compose.yaml",
			State:              "exited",
			CreatedUnix:        now.Add(-3 * time.Minute).Unix(),
			CPUPercent:         0,
			MemoryUsageBytes:   0,
			MemoryLimitBytes:   0,
		},
		{
			FullID:             "c-1",
			Name:               "api",
			ComposeProject:     "shop",
			ComposeService:     "api",
			ComposeWorkingDir:  "/srv/shop",
			ComposeConfigFiles: "compose.yaml",
			State:              "running",
			Healthy:            true,
			CreatedUnix:        now.Add(-1 * time.Minute).Unix(),
			CPUPercent:         10.5,
			MemoryPercent:      50,
			MemoryUsageBytes:   100,
			MemoryLimitBytes:   200,
		},
		{
			FullID:             "c-2",
			Name:               "cache",
			ComposeProject:     "shop",
			ComposeService:     "cache",
			ComposeWorkingDir:  "/srv/shop",
			ComposeConfigFiles: "compose.yaml",
			State:              "running",
			CreatedUnix:        now.Add(-2 * time.Minute).Unix(),
			CPUPercent:         4.5,
			MemoryPercent:      50,
			MemoryUsageBytes:   50,
			MemoryLimitBytes:   100,
		},
	}

	projects := AggregateComposeProjects(containers)
	if len(projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(projects))
	}

	project := projects[0]
	if project.Name != "shop" {
		t.Fatalf("project name = %q, want shop", project.Name)
	}
	if project.ContainerCount != 3 || project.RunningCount != 2 || project.HealthyCount != 1 {
		t.Fatalf("project counts = %#v, want 3/2/1", project)
	}
	if project.Network != "shop_default" {
		t.Fatalf("project network = %q, want shop_default", project.Network)
	}
	if project.WorkingDir != "/srv/shop" || project.ConfigFiles != "compose.yaml" {
		t.Fatalf("project metadata = workingDir %q configFiles %q", project.WorkingDir, project.ConfigFiles)
	}
	if len(project.Services) != 3 {
		t.Fatalf("project services len = %d, want 3", len(project.Services))
	}
	if project.Created == "" || project.Created == "-" {
		t.Fatalf("project created time should be populated, got %q", project.Created)
	}
	if project.CPUPercent != 15.0 {
		t.Fatalf("project CPUPercent = %v, want 15.0", project.CPUPercent)
	}
	if project.MemoryUsage != "150 B" {
		t.Fatalf("project MemoryUsage = %q, want 150 B", project.MemoryUsage)
	}
	if project.MemoryLimit != "300 B" {
		t.Fatalf("project MemoryLimit = %q, want 300 B", project.MemoryLimit)
	}
	if project.MemoryPercent != 50.0 {
		t.Fatalf("project MemoryPercent = %v, want 50", project.MemoryPercent)
	}
	if len(project.Containers) != 3 {
		t.Fatalf("project containers len = %d, want 3", len(project.Containers))
	}
	if project.Containers[0].Name != "api" || project.Containers[1].Name != "cache" || project.Containers[2].Name != "worker" {
		t.Fatalf("project containers were not sorted by core sort order: %#v", project.Containers)
	}
}

func TestAggregateComposeProjects_MemoryPercentDoesNotUseSummedLimits(t *testing.T) {
	now := time.Now()
	containers := []ContainerRow{
		{
			FullID:           "c-1",
			Name:             "api",
			ComposeProject:   "shop",
			State:            "running",
			CreatedUnix:      now.Add(-1 * time.Minute).Unix(),
			MemoryPercent:    50,
			MemoryUsageBytes: 10,
			MemoryLimitBytes: 20,
		},
		{
			FullID:           "c-2",
			Name:             "worker",
			ComposeProject:   "shop",
			State:            "running",
			CreatedUnix:      now.Add(-2 * time.Minute).Unix(),
			MemoryPercent:    10,
			MemoryUsageBytes: 90,
			MemoryLimitBytes: 900,
		},
	}

	projects := AggregateComposeProjects(containers)
	if len(projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(projects))
	}

	if projects[0].MemoryPercent != 30 {
		t.Fatalf("project MemoryPercent = %v, want 30 (average of 50 and 10)", projects[0].MemoryPercent)
	}
}

func TestHumanAge_FormatsRecentTimes(t *testing.T) {
	if got := HumanAge(time.Now().Add(-2 * time.Minute)); got != "2m ago" && got != "1m ago" {
		t.Fatalf("HumanAge(...) = %q, want a minutes-ago string", got)
	}
}
