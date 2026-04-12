package core

import "testing"

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{name: "negative", size: -1, want: "-1 B"},
		{name: "bytes", size: 0, want: "0 B"},
		{name: "under kibibyte", size: 1023, want: "1023 B"},
		{name: "one kibibyte", size: 1024, want: "1.0 KiB"},
		{name: "fractional kibibyte", size: 1536, want: "1.5 KiB"},
		{name: "one mebibyte", size: 1024 * 1024, want: "1.0 MiB"},
		{name: "one gibibyte", size: 1024 * 1024 * 1024, want: "1.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanBytes(tt.size); got != tt.want {
				t.Fatalf("HumanBytes(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}
