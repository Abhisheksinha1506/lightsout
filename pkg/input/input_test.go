package input

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// pickApp asks the user to type the number of the app they are running
// lightsout from, and keeps asking until the answer is a listed number.
func TestPickApp(t *testing.T) {
	apps := []string{"Mail", "Slack", "Terminal"}

	tests := []struct {
		name    string
		apps    []string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid number", apps: apps, input: "2\n", want: "Slack"},
		{name: "spaces around number", apps: apps, input: "  3 \n", want: "Terminal"},
		{name: "retries after invalid input", apps: apps, input: "abc\n0\n4\n\n1\n", want: "Mail"},
		{name: "last line without newline", apps: apps, input: "3", want: "Terminal"},
		{name: "input ends without a choice", apps: apps, input: "abc\n", wantErr: true},
		{name: "no input at all", apps: apps, input: "", wantErr: true},
		{name: "no apps to choose from", apps: nil, input: "1\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pickApp(strings.NewReader(tt.input), io.Discard, tt.apps)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("pickApp() = %q, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("pickApp() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("pickApp() = %q, want %q", got, tt.want)
			}
		})
	}
}

// failingWriter rejects every write, like a closed terminal.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("terminal closed") }

// If the prompt can't be shown, pickApp must stop instead of waiting for an
// answer to a question the user never saw.
func TestPickAppStopsWhenPromptCannotBeShown(t *testing.T) {
	if got, err := pickApp(strings.NewReader("1\n"), failingWriter{}, []string{"Mail"}); err == nil {
		t.Fatalf("pickApp() = %q, want an error", got)
	}
}

// The prompt lists every app with its number, starting from 1.
func TestPickAppListsNumberedApps(t *testing.T) {
	var out strings.Builder
	if _, err := pickApp(strings.NewReader("1\n"), &out, []string{"Mail", "Slack"}); err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"1) Mail", "2) Slack"} {
		if !strings.Contains(out.String(), line) {
			t.Errorf("prompt output missing %q:\n%s", line, out.String())
		}
	}
}

// When run from kitty, kitty is auto-detected and kept open, whether or not
// it is also in the ignore list. An ignored kitty used to fail with
// "kitty is not open" even though it was running.
func TestExecutingTermKittyAutoDetected(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	running := []string{"Slack", "kitty", "Mail"}

	tests := []struct {
		name    string
		ignored []string
	}{
		{name: "kitty not ignored", ignored: nil},
		{name: "kitty ignored", ignored: []string{"kitty"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExecutingTerm(running, tt.ignored)
			if want := []string{"Mail", "Slack"}; !reflect.DeepEqual(got, want) {
				t.Errorf("ExecutingTerm() = %q, want %q", got, want)
			}
		})
	}
}
