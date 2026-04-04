package docker

import (
	"reflect"
	"testing"
)

func TestNormalizeLogs(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
		stderr string
		want   []string
	}{
		{
			name:   "joins stdout and stderr",
			stdout: "one\ntwo\n",
			stderr: "three\n",
			want:   []string{"one", "two", "three"},
		},
		{
			name:   "collapses carriage return progress lines",
			stdout: "step 1\rstep 2\rfinished\n",
			want:   []string{"finished"},
		},
		{
			name:   "applies backspaces",
			stdout: "abc\b\bdone\n",
			want:   []string{"adone"},
		},
		{
			name:   "treats terminal boundary restore as newline",
			stdout: "fetching\x1b7\x1b8done\n",
			want:   []string{"fetching", "done"},
		},
		{
			name:   "drops control only lines",
			stdout: "\x1b[31m\r\nreal line\n",
			want:   []string{"real line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLogs(tt.stdout, tt.stderr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("normalizeLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTailOption(t *testing.T) {
	if got := tailOption(0); got != "all" {
		t.Fatalf("tailOption(0) = %q, want %q", got, "all")
	}
	if got := tailOption(250); got != "250" {
		t.Fatalf("tailOption(250) = %q, want %q", got, "250")
	}
}
