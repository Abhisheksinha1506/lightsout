package macos

import (
	"errors"
	"reflect"
	"testing"
)

// Only Dock apps are listed: system agents such as loginwindow (policy 1) and
// background services (policy 2) must never be quit. Names are kept whole,
// even with commas, and same-named apps keep all their process IDs.
func TestParseApps(t *testing.T) {
	out := []byte(`[
		{"name": "loginwindow", "pid": 422, "policy": 1, "bundle": "com.apple.loginwindow"},
		{"name": "Finder", "pid": 642, "policy": 0, "bundle": "com.apple.finder"},
		{"name": "Google Chrome", "pid": 954, "policy": 0, "bundle": "com.google.Chrome"},
		{"name": "AutoFill (Google Chrome)", "pid": 1021, "policy": 2, "bundle": "com.apple.SafariPlatformSupport.Helper"},
		{"name": "Notes, Reminders & More", "pid": 1100, "policy": 0, "bundle": "com.example.notes"},
		{"name": "Code", "pid": 1200, "policy": 0, "bundle": "com.microsoft.VSCode"},
		{"name": "Code", "pid": 1201, "policy": 0, "bundle": "com.microsoft.VSCode"},
		{"name": "", "pid": 1300, "policy": 0, "bundle": "com.example.unnamed"},
		{"name": "Ghost", "pid": -1, "policy": 0, "bundle": "com.example.ghost"}
	]`)

	got, err := parseApps(out)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]int{
		"Google Chrome":           {954},
		"Notes, Reminders & More": {1100},
		"Code":                    {1200, 1201},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseApps() = %v, want %v", got, want)
	}
}

func TestParseAppsRejectsBadOutput(t *testing.T) {
	if _, err := parseApps([]byte("execution error: not JSON")); err == nil {
		t.Fatal("parseApps() error = nil, want an error")
	}
}

func TestQuit(t *testing.T) {
	tests := []struct {
		name          string
		pids          []int
		requestErr    error
		runningPolls  int // processes report running for this many polls; -1 means forever
		wantErr       bool
		wantRequested []int
	}{
		{name: "quits at once", pids: []int{10}, runningPolls: 0, wantRequested: []int{10}},
		{name: "quits after a moment", pids: []int{10, 11}, runningPolls: 3, wantRequested: []int{10, 11}},
		{name: "stays open (save dialog)", pids: []int{10}, runningPolls: -1, wantErr: true, wantRequested: []int{10}},
		{name: "quit request fails", pids: []int{10}, requestErr: errors.New("osascript failed"), wantErr: true, wantRequested: []int{10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requested []int
			request := func(pid int) error {
				requested = append(requested, pid)
				return tt.requestErr
			}
			polls := 0
			running := func(int) bool {
				polls++
				return tt.runningPolls < 0 || polls <= tt.runningPolls
			}

			// Long enough for several polls at pollInterval.
			err := quit(tt.pids, request, running, 5*pollInterval)

			if (err != nil) != tt.wantErr {
				t.Errorf("quit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(requested, tt.wantRequested) {
				t.Errorf("requested quit for %v, want %v", requested, tt.wantRequested)
			}
		})
	}
}
