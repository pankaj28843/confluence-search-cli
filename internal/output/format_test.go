package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWordWrap(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{"empty", "", 80, []string{""}},
		{"short", "hello", 80, []string{"hello"}},
		{"exact width", "hello world", 11, []string{"hello world"}},
		{"wraps at word boundary", "hello world foo bar", 11, []string{"hello world", "foo bar"}},
		{"single long word", "superlongword", 5, []string{"superlongword"}},
		{"multiple wraps", "a b c d e f g", 5, []string{"a b c", "d e f", "g"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WordWrap(tc.text, tc.width)
			if len(got) != len(tc.want) {
				t.Fatalf("WordWrap(%q, %d) = %v, want %v", tc.text, tc.width, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestWriterJSON(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Err: &bytes.Buffer{}, Format: FormatJSON}

	err := w.JSON(map[string]string{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, `"key": "value"`) {
		t.Errorf("JSON output = %q, want key/value pair", got)
	}
}

func TestWriterText(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Out: &buf, Err: &bytes.Buffer{}, Format: FormatText}

	w.Text("hello %s\n", "world")
	if buf.String() != "hello world\n" {
		t.Errorf("Text output = %q", buf.String())
	}
}

func TestWriterFinishWithTiming(t *testing.T) {
	var errBuf bytes.Buffer
	w := New(false, true)
	w.Err = &errBuf

	w.Finish()

	got := errBuf.String()
	if !strings.HasSuffix(got, "ms\n") {
		t.Errorf("timing output = %q, expected ms suffix", got)
	}
}

func TestWriterFinishWithoutTiming(t *testing.T) {
	var errBuf bytes.Buffer
	w := New(false, false)
	w.Err = &errBuf

	w.Finish()

	if errBuf.String() != "" {
		t.Errorf("expected no output without timing, got %q", errBuf.String())
	}
}
