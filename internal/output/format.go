package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Format controls output style.
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// Writer handles formatted output.
type Writer struct {
	Out    io.Writer
	Err    io.Writer
	Format Format
	Timing bool
	start  time.Time
}

// New creates a Writer with the given options.
func New(jsonOutput, timing bool) *Writer {
	w := &Writer{
		Out:    os.Stdout,
		Err:    os.Stderr,
		Timing: timing,
		start:  time.Now(),
	}
	if jsonOutput {
		w.Format = FormatJSON
	}
	return w
}

// JSON writes v as JSON to stdout.
func (w *Writer) JSON(v interface{}) error {
	enc := json.NewEncoder(w.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Text writes formatted text to stdout.
func (w *Writer) Text(format string, args ...interface{}) {
	fmt.Fprintf(w.Out, format, args...)
}

// Finish prints timing info if enabled.
func (w *Writer) Finish() {
	if w.Timing {
		elapsed := time.Since(w.start)
		fmt.Fprintf(w.Err, "%.1fms\n", float64(elapsed.Microseconds())/1000)
	}
}

// WordWrap wraps text to the given width.
func WordWrap(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}
	var lines []string
	words := strings.Fields(text)
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
