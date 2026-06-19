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

func TestFilterJSONLines(t *testing.T) {
	inspectJSON := []string{
		`{`,
		`  "Id": "abc123",`,
		`  "Config": {`,
		`    "Env": [`,
		`      "PATH=/usr/bin:/bin",`,
		`      "HOME=/root"`,
		`    ],`,
		`    "Cmd": null`,
		`  },`,
		`  "Name": "/foo"`,
		`}`,
	}

	tests := []struct {
		name     string
		lines    []string
		query    string
		expected []string
	}{
		{
			name:     "empty query returns all lines",
			lines:    inspectJSON,
			query:    "",
			expected: inspectJSON,
		},
		{
			name:  "key match shows full subtree",
			lines: inspectJSON,
			query: "Env",
			expected: []string{
				`{`,
				`  "Config": {`,
				`    "Env": [`,
				`      "PATH=/usr/bin:/bin",`,
				`      "HOME=/root"`,
				`    ],`,
				`  },`,
				`}`,
			},
		},
		{
			name:  "value match shows leaf line",
			lines: inspectJSON,
			query: "/root",
			expected: []string{
				`{`,
				`  "Config": {`,
				`    "Env": [`,
				`      "HOME=/root"`,
				`    ],`,
				`  },`,
				`}`,
			},
		},
		{
			name:  "inner key match shows parent subtree",
			lines: inspectJSON,
			query: "Cmd",
			expected: []string{
				`{`,
				`  "Config": {`,
				`    "Cmd": null`,
				`  },`,
				`}`,
			},
		},
		{
			name:  "top-level key shows full subtree through close",
			lines: inspectJSON,
			query: "Config",
			expected: []string{
				`{`,
				`  "Config": {`,
				`    "Env": [`,
				`      "PATH=/usr/bin:/bin",`,
				`      "HOME=/root"`,
				`    ],`,
				`    "Cmd": null`,
				`  },`,
				`}`,
			},
		},
		{
			name:     "no match returns empty",
			lines:    inspectJSON,
			query:    "NOSUCHKEY",
			expected: []string{},
		},
		{
			name:     "empty lines returns empty",
			lines:    []string{},
			query:    "foo",
			expected: []string{},
		},
		{
			name:  "multiple matches merge overlapping blocks",
			lines: inspectJSON,
			query: "PATH",
			expected: []string{
				`{`,
				`  "Config": {`,
				`    "Env": [`,
				`      "PATH=/usr/bin:/bin",`,
				`    ],`,
				`  },`,
				`}`,
			},
		},
		{
			name: "nested match includes closing bracket at same level",
			lines: []string{
				`{`,
				`  "A": {`,
				`    "B": 1`,
				`  },`,
				`  "C": 2`,
				`}`,
			},
			query: "A",
			expected: []string{
				`{`,
				`  "A": {`,
				`    "B": 1`,
				`  },`,
				`}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterJSONLines(tt.lines, tt.query)
			if len(got) != len(tt.expected) {
				t.Errorf("FilterJSONLines() returned %d lines, want %d", len(got), len(tt.expected))
				for i, l := range got {
					t.Logf("  got[%d] = %q", i, l)
				}
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("FilterJSONLines()[%d] = %q, want %q", i, got[i], tt.expected[i])
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

func TestVpFollow(t *testing.T) {
	t.Run("SetFollow enables follow and goes to bottom", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		vp.SetHeight(3)
		vp.SetYOffset(0)

		vp.SetFollow(true)

		if !vp.Follow {
			t.Error("expected Follow to be true")
		}
	})

	t.Run("SetFollow disables follow", func(t *testing.T) {
		vp := NewViewport()
		vp.Follow = true

		vp.SetFollow(false)

		if vp.Follow {
			t.Error("expected Follow to be false")
		}
	})
}

func TestVpWrap(t *testing.T) {
	t.Run("SetWrapLines enables wrapping", func(t *testing.T) {
		vp := NewViewport()

		vp.SetWrapLines(true)

		if !vp.WrapLines {
			t.Error("expected WrapLines to be true")
		}
	})

	t.Run("SetWrapLines disables wrapping", func(t *testing.T) {
		vp := NewViewport()
		vp.WrapLines = true

		vp.SetWrapLines(false)

		if vp.WrapLines {
			t.Error("expected WrapLines to be false")
		}
	})
}

func TestSyncFromData(t *testing.T) {
	t.Run("syncs viewport with filtered and sanitized data", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"\x1b[31mred\x1b[0m", "normal", "   spaces   "}
		vp.Filter.Query = ""
		vp.Follow = false

		vp.SyncFromData(80, 10)

		if vp.InitialLoad {
			t.Error("expected InitialLoad to be false after sync")
		}
	})

	t.Run("follow mode goes to bottom", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3"}
		vp.Follow = true

		vp.SyncFromData(80, 10)

		// Viewport should be at bottom in follow mode
		if got := vp.YOffset(); got != 0 {
			t.Errorf("viewport YOffset = %d, want 0 (meaning bottom)", got)
		}
	})
}

func TestKeyBindings(t *testing.T) {
	t.Run("Up key navigation", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		vp.SetHeight(2)
		vp.SetContent("line1\nline2\nline3\nline4\nline5")
		vp.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(vp, tea.KeyPressMsg{Code: tea.KeyUp}, NewKeyMap())

		if vp.YOffset() != 1 {
			t.Errorf("expected YOffset 1, got %d", vp.YOffset())
		}
	})

	t.Run("Down key navigation", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3", "line4", "line5"}
		vp.SetHeight(2)
		vp.SetContent("line1\nline2\nline3\nline4\nline5")
		vp.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(vp, tea.KeyPressMsg{Code: tea.KeyDown}, NewKeyMap())

		if vp.YOffset() != 3 {
			t.Errorf("expected YOffset 3, got %d", vp.YOffset())
		}
	})

	t.Run("Home key goes to top", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3"}
		vp.SetHeight(2)
		vp.SetContent("line1\nline2\nline3")
		vp.SetYOffset(2)

		controller := Controller{}
		controller.HandleKey(vp, tea.KeyPressMsg{Code: tea.KeyHome}, NewKeyMap())

		if vp.YOffset() != 0 {
			t.Errorf("expected YOffset 0, got %d", vp.YOffset())
		}
	})

	t.Run("End key goes to bottom", func(t *testing.T) {
		vp := NewViewport()
		vp.Data = []string{"line1", "line2", "line3"}
		vp.SetHeight(2)
		vp.SetContent("line1\nline2\nline3")

		controller := Controller{}
		controller.HandleKey(vp, tea.KeyPressMsg{Code: tea.KeyEnd}, NewKeyMap())

		if vp.YOffset() != 1 {
			t.Errorf("expected YOffset 1, got %d", vp.YOffset())
		}
	})

}
