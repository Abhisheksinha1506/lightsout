package output

import (
	"io"
	"os"
	"strings"
	"testing"
)

// The banner is six lines of block letters.
func TestTitle(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	Title()
	os.Stdout = stdout
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if lines := strings.Count(string(out), "\n"); lines != 6 {
		t.Errorf("Title() printed %d lines, want 6:\n%s", lines, out)
	}
	if !strings.Contains(string(out), "██╗") {
		t.Errorf("Title() output has no block letters:\n%s", out)
	}
}
