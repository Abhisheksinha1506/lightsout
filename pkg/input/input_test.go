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
// 0 means none of them, for when that app is already kept open.
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
		{name: "none of these", apps: apps, input: "0\n", want: ""},
		{name: "retries after invalid input", apps: apps, input: "abc\n-1\n4\n\n1\n", want: "Mail"},
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

// The prompt lists every app with its number, starting from 1, plus 0 for
// none of them.
func TestPickAppListsNumberedApps(t *testing.T) {
	var out strings.Builder
	if _, err := pickApp(strings.NewReader("1\n"), &out, []string{"Mail", "Slack"}); err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"1) Mail", "2) Slack", "0) None of these"} {
		if !strings.Contains(out.String(), line) {
			t.Errorf("prompt output missing %q:\n%s", line, out.String())
		}
	}
}

// Without drops kept apps and sorts the rest.
func TestWithout(t *testing.T) {
	got := Without([]string{"Slack", "Terminal", "Mail", "Google Chrome"}, []string{"Google Chrome", "Terminal\n", "Zoom"})
	if want := []string{"Mail", "Slack"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Without() = %q, want %q", got, want)
	}
	if got := Without([]string{"A"}, []string{"A"}); len(got) != 0 {
		t.Errorf("Without(all kept) = %q, want none", got)
	}
}

// PickApp is pickApp for callers outside the package.
func TestPickAppExported(t *testing.T) {
	if got, err := PickApp(strings.NewReader("2\n"), io.Discard, []string{"Mail", "Slack"}); err != nil || got != "Slack" {
		t.Errorf("PickApp() = %q, %v; want \"Slack\", nil", got, err)
	}
}
