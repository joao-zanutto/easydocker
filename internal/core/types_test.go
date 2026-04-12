package core

import "testing"

func TestApplyMetricsToContainers_AppliesMatchingRowsOnly(t *testing.T) {
	rows := []ContainerRow{
		{FullID: "id-1", Name: "one", CPUPercent: 1.5, MemoryUsage: "-", MemoryLimit: "-"},
		{FullID: "id-2", Name: "two", CPUPercent: 2.5, MemoryUsage: "old", MemoryLimit: "old"},
	}
	originalCPU := rows[1].CPUPercent
	originalMem := rows[1].MemoryUsage

	metrics := map[string]ContainerMetrics{
		"id-1": {
			CPUPercent:       12.4,
			MemoryPercent:    43.2,
			MemoryUsage:      "512 MiB",
			MemoryLimit:      "2.0 GiB",
			MemoryUsageBytes: 536870912,
			MemoryLimitBytes: 2147483648,
		},
	}

	updated := ApplyMetricsToContainers(rows, metrics)

	if len(updated) != len(rows) {
		t.Fatalf("ApplyMetricsToContainers len = %d, want %d", len(updated), len(rows))
	}

	if updated[0].CPUPercent != metrics["id-1"].CPUPercent {
		t.Fatalf("row 0 CPUPercent = %v, want %v", updated[0].CPUPercent, metrics["id-1"].CPUPercent)
	}
	if updated[0].MemoryUsage != metrics["id-1"].MemoryUsage {
		t.Fatalf("row 0 MemoryUsage = %q, want %q", updated[0].MemoryUsage, metrics["id-1"].MemoryUsage)
	}

	if updated[1].CPUPercent != originalCPU || updated[1].MemoryUsage != originalMem {
		t.Fatalf("row 1 should be unchanged, got CPU=%v MemoryUsage=%q", updated[1].CPUPercent, updated[1].MemoryUsage)
	}

	if rows[0].CPUPercent != 1.5 {
		t.Fatalf("input slice mutated: rows[0].CPUPercent = %v, want 1.5", rows[0].CPUPercent)
	}
}
