package logs

import (
	"reflect"
	"testing"
)

func TestFilterLogLines(t *testing.T) {
	lines := []string{"alpha one", "beta two", "alpha three"}

	if got := FilterLogLines(lines, ""); !reflect.DeepEqual(got, lines) {
		t.Fatalf("FilterLogLines(..., empty) = %v, want %v", got, lines)
	}

	if got := FilterLogLines(lines, "alpha"); !reflect.DeepEqual(got, []string{"alpha one", "alpha three"}) {
		t.Fatalf("FilterLogLines(..., alpha) = %v, want only matching lines", got)
	}

	if got := FilterLogLines(lines, "zzz"); len(got) != 0 {
		t.Fatalf("FilterLogLines(..., zzz) len = %d, want 0", len(got))
	}
}

func TestMergePolledLogs(t *testing.T) {
	tests := []struct {
		name        string
		previous    []string
		polled      []string
		maxLines    int
		want        []string
		wantOverlap bool
	}{
		{
			name:        "empty previous returns polled",
			previous:    nil,
			polled:      []string{"a", "b"},
			maxLines:    0,
			want:        []string{"a", "b"},
			wantOverlap: true,
		},
		{
			name:        "empty polled keeps previous",
			previous:    []string{"a", "b"},
			polled:      nil,
			maxLines:    0,
			want:        []string{"a", "b"},
			wantOverlap: true,
		},
		{
			name:        "overlap appends only new suffix",
			previous:    []string{"1", "2", "3"},
			polled:      []string{"3", "4", "5"},
			maxLines:    0,
			want:        []string{"1", "2", "3", "4", "5"},
			wantOverlap: true,
		},
		{
			name:        "identical slices stay stable",
			previous:    []string{"1", "2"},
			polled:      []string{"1", "2"},
			maxLines:    0,
			want:        []string{"1", "2"},
			wantOverlap: true,
		},
		{
			name:        "smaller polled suffix keeps previous",
			previous:    []string{"1", "2", "3", "4"},
			polled:      []string{"3", "4"},
			maxLines:    0,
			want:        []string{"1", "2", "3", "4"},
			wantOverlap: true,
		},
		{
			name:        "carriage returns are normalized for overlap",
			previous:    []string{"1\r", "2\r", "3\r"},
			polled:      []string{"3", "4"},
			maxLines:    0,
			want:        []string{"1", "2", "3", "4"},
			wantOverlap: true,
		},
		{
			name:        "disjoint poll replaces log buffer",
			previous:    []string{"1", "2"},
			polled:      []string{"8", "9"},
			maxLines:    0,
			want:        []string{"8", "9"},
			wantOverlap: false,
		},
		{
			name:        "max lines trims merged result",
			previous:    []string{"1", "2", "3"},
			polled:      []string{"3", "4", "5"},
			maxLines:    3,
			want:        []string{"3", "4", "5"},
			wantOverlap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOverlap := MergePolledLogs(tt.previous, tt.polled, tt.maxLines)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("mergePolledLogs() logs = %v, want %v", got, tt.want)
			}
			if gotOverlap != tt.wantOverlap {
				t.Fatalf("mergePolledLogs() overlap = %v, want %v", gotOverlap, tt.wantOverlap)
			}
		})
	}
}
