package util

import (
	"strings"
	"testing"
)

func TestDisplayWidth_StripsANSI(t *testing.T) {
	if got := DisplayWidth("\x1b[31mabc\x1b[39m"); got != 3 {
		t.Fatalf("displayWidth(ansi text) = %d, want 3", got)
	}
}

func TestClipAndPadLines(t *testing.T) {
	t.Run("non-positive height", func(t *testing.T) {
		if got := ClipAndPadLines([]string{"a"}, 0, "-"); len(got) != 0 {
			t.Fatalf("clipAndPadLines(..., 0, ...) len = %d, want 0", len(got))
		}
	})

	t.Run("clips long input", func(t *testing.T) {
		got := ClipAndPadLines([]string{"a", "b", "c"}, 2, "-")
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("clipAndPadLines clip result = %#v, want [\"a\", \"b\"]", got)
		}
	})

	t.Run("pads short input", func(t *testing.T) {
		got := ClipAndPadLines([]string{"a"}, 3, "-")
		if len(got) != 3 || got[0] != "a" || got[1] != "-" || got[2] != "-" {
			t.Fatalf("clipAndPadLines pad result = %#v, want [\"a\", \"-\", \"-\"]", got)
		}
	})
}

func TestConstrainLine_WithANSIAndTightWidth(t *testing.T) {
	line := "\x1b[32mabcdef\x1b[39m"
	got := ConstrainLine(line, 4)
	if DisplayWidth(got) != 4 {
		t.Fatalf("constrainLine width = %d, want 4 (got %q)", DisplayWidth(got), got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("constrainLine should include ellipsis when truncated, got %q", got)
	}
	if got1 := ConstrainLine(line, 1); got1 != "…" {
		t.Fatalf("constrainLine(..., 1) = %q, want ellipsis", got1)
	}
}
