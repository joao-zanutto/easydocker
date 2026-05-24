package core

import "testing"

func TestContainerCPUValue_RunningNeverDash(t *testing.T) {
	loading := ContainerCPUValue(ContainerRow{State: "running", CPUPercent: -1}, "⠋")
	if loading != "⠋" {
		t.Fatalf("running loading cpu = %q, want spinner icon", loading)
	}

	idle := ContainerCPUValue(ContainerRow{State: "running", CPUPercent: 0}, "⠋")
	if idle != "0.0%" {
		t.Fatalf("running zero cpu = %q, want 0.0%%", idle)
	}

	afterInitial := ContainerCPUValue(ContainerRow{State: "running", CPUPercent: -1}, "")
	if afterInitial != "-" {
		t.Fatalf("running loading cpu with no indicator = %q, want -", afterInitial)
	}
}

func TestContainerMemoryTableValue_OmitsLimit(t *testing.T) {
	running := ContainerRow{State: "running", MemoryUsage: "128 MiB", MemoryLimit: "2 GiB", MemoryPercent: 6.25}
	got := ContainerMemoryTableValue(running, "⠋")
	if got != "128 MiB" {
		t.Fatalf("table memory value = %q, want %q", got, "128 MiB (6.2%)")
	}

	loading := ContainerMemoryTableValue(ContainerRow{State: "running", MemoryUsage: "-"}, "⠋")
	if loading != "⠋" {
		t.Fatalf("running placeholder memory value = %q, want spinner icon", loading)
	}

	afterInitial := ContainerMemoryTableValue(ContainerRow{State: "running", MemoryUsage: "-"}, "")
	if afterInitial != "-" {
		t.Fatalf("running placeholder memory value with no indicator = %q, want -", afterInitial)
	}
}
