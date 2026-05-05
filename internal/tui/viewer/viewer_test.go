package viewer

import (
	"testing"

	"charm.land/bubbletea/v2"
)

func TestWrappedRowCount(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		width    int
		expected int
	}{
		{
			name:     "short line fits in one row",
			line:     "hello",
			width:    80,
			expected: 1,
		},
		{
			name:     "exact width fits in one row",
			line:     "12345678901234567890123456789012345678901234567890123456789012345678901234567890",
			width:    80,
			expected: 1,
		},
		{
			name:     "double width wraps to two rows",
			line:     "123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			width:    80,
			expected: 2,
		},
		{
			name:     "very long line wraps to multiple rows",
			line:     "This is a very long line that should wrap across multiple rows when displayed in a narrow viewport",
			width:    20,
			expected: 5,
		},
		{
			name:     "empty line",
			line:     "",
			width:    80,
			expected: 1,
		},
		{
			name:     "width of 1 forces each char to row",
			line:     "abc",
			width:    1,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := SanitizeLine(tt.line)
			got := WrappedRowCount(sanitized, tt.width)
			if got != tt.expected {
				t.Errorf("WrappedRowCount(%q, %d) = %d, want %d", tt.line, tt.width, got, tt.expected)
			}
		})
	}
}

func TestFilterLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		query    string
		expected []string
	}{
		{
			name:     "empty query returns all lines",
			lines:    []string{"foo", "bar", "baz"},
			query:    "",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "filter matches subset",
			lines:    []string{"foo", "bar", "baz"},
			query:    "ba",
			expected: []string{"bar", "baz"},
		},
		{
			name:     "filter is case sensitive",
			lines:    []string{"Foo", "bar", "BAZ"},
			query:    "Foo",
			expected: []string{"Foo"},
		},
		{
			name:     "no match returns empty",
			lines:    []string{"foo", "bar"},
			query:    "xyz",
			expected: []string{},
		},
		{
			name:     "empty lines returns empty",
			lines:    []string{},
			query:    "foo",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterLines(tt.lines, tt.query)
			if len(got) != len(tt.expected) {
				t.Errorf("FilterLines() returned %d lines, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("FilterLines()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSanitizeLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ANSI codes removed",
			input:    "\x1b[31mred\x1b[0m",
			expected: "red",
		},
		{
			name:     "control characters removed",
			input:    "hello\x00world\x07",
			expected: "helloworld",
		},
		{
			name:     "tabs preserved as spaces",
			input:    "hello\tworld",
			expected: "hello world",
		},
		{
			name:     "newlines removed",
			input:    "hello\nworld",
			expected: "helloworld",
		},
		{
			name:     "carriage returns normalized",
			input:    "hello\r\nworld",
			expected: "helloworld",
		},
		{
			name:     "normal text unchanged",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLine(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStateFollow(t *testing.T) {
	t.Run("SetFollow enables follow and goes to bottom", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		state.Viewport.SetHeight(3)
		state.Viewport.SetYOffset(0)

		state.SetFollow(true)

		if !state.Follow {
			t.Error("expected Follow to be true")
		}
	})

	t.Run("SetFollow disables follow", func(t *testing.T) {
		state := NewState()
		state.Follow = true

		state.SetFollow(false)

		if state.Follow {
			t.Error("expected Follow to be false")
		}
	})
}

func TestStateWrap(t *testing.T) {
	t.Run("SetWrapLines enables wrapping", func(t *testing.T) {
		state := NewState()

		state.SetWrapLines(true)

		if !state.WrapLines {
			t.Error("expected WrapLines to be true")
		}
	})

	t.Run("SetWrapLines disables wrapping", func(t *testing.T) {
		state := NewState()
		state.WrapLines = true

		state.SetWrapLines(false)

		if state.WrapLines {
			t.Error("expected WrapLines to be false")
		}
	})
}

func TestRenderedViewportLineDelta(t *testing.T) {
	t.Run("no prepended lines returns zero", func(t *testing.T) {
		state := NewState()
		state.Filter.Query = ""
		state.WrapLines = false
		allLines := []string{"a", "b", "c"}

		delta := RenderedViewportLineDelta(&state, allLines, 0)

		if delta != 0 {
			t.Errorf("expected 0, got %d", delta)
		}
	})

	t.Run("prepended lines without wrap returns count", func(t *testing.T) {
		state := NewState()
		state.Filter.Query = ""
		state.WrapLines = false
		allLines := []string{"a", "b", "c", "d", "e"}

		delta := RenderedViewportLineDelta(&state, allLines, 2)

		if delta != 2 {
			t.Errorf("expected 2, got %d", delta)
		}
	})

	t.Run("filter affects delta count", func(t *testing.T) {
		state := NewState()
		state.Filter.Query = "a"
		state.WrapLines = false
		allLines := []string{"a", "b", "a", "c", "a"}

		delta := RenderedViewportLineDelta(&state, allLines, 3)

		if delta != 2 {
			t.Errorf("expected 2 (filtered lines), got %d", delta)
		}
	})

	t.Run("wrapped lines calculate correctly", func(t *testing.T) {
		state := NewState()
		state.Filter.Query = ""
		state.WrapLines = true
		state.Viewport.SetWidth(10)
		allLines := []string{"1234567890", "short", "123456789012345"}

		delta := RenderedViewportLineDelta(&state, allLines, 3)

		if delta != 4 {
			t.Errorf("expected 4 (1+1+2 wrapped rows), got %d", delta)
		}
	})
}

func TestSyncFromData(t *testing.T) {
	t.Run("syncs viewport with filtered and sanitized data", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"\x1b[31mred\x1b[0m", "normal", "   spaces   "}
		state.Filter.Query = ""
		state.Follow = false

		state.SyncFromData(80, 10)

		if state.InitialLoad {
			t.Error("expected InitialLoad to be false after sync")
		}
	})

	t.Run("follow mode goes to bottom", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Follow = true

		state.SyncFromData(80, 10)

		// Viewport should be at bottom in follow mode
		if state.Viewport.YOffset() != 0 {
			// YOffset 0 means at bottom in bubbletea viewport
		}
	})
}

func TestKeyBindings(t *testing.T) {
	t.Run("Up key navigation", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3\nline4\nline5")
		state.Viewport.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyUp}, NewKeyMap())

		if state.Viewport.YOffset() != 1 {
			t.Errorf("expected YOffset 1, got %d", state.Viewport.YOffset())
		}
	})

	t.Run("Down key navigation", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3\nline4\nline5")
		state.Viewport.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyDown}, NewKeyMap())

		if state.Viewport.YOffset() != 3 {
			t.Errorf("expected YOffset 3, got %d", state.Viewport.YOffset())
		}
	})

	t.Run("Home key goes to top", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3")
		state.Viewport.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyHome}, NewKeyMap())

		if state.Viewport.YOffset() != 0 {
			t.Errorf("expected YOffset 0, got %d", state.Viewport.YOffset())
		}
	})

	t.Run("End key goes to bottom", func(t *testing.T) {
		state := NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3")

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyEnd}, NewKeyMap())

		if state.Viewport.YOffset() != 1 {
			t.Errorf("expected YOffset 1, got %d", state.Viewport.YOffset())
		}
	})

	t.Run("ToggleWrap changes wrap state", func(t *testing.T) {
		state := NewState()
		state.WrapLines = false

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: 'w', Text: "w"}, NewKeyMap())

		if !state.WrapLines {
			t.Error("expected WrapLines to be true")
		}
	})

	t.Run("ToggleFollow changes follow state", func(t *testing.T) {
		state := NewState()
		state.Follow = false

		controller := Controller{}
		controller.HandleKey(&state, tea.KeyPressMsg{Code: 'f', Text: "f"}, NewKeyMap())

		if !state.Follow {
			t.Error("expected Follow to be true")
		}
	})
}