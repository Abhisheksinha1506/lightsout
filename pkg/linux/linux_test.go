package linux

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseWindowIDs(t *testing.T) {
	out := "0x01e00003 -1 host Desktop\n" +
		"0x03c00003  0 host Terminal - user@host\n" +
		"\n" +
		"0x04200007  0 host Inbox - Mail\n"

	got := parseWindowIDs(out)

	want := []string{"0x01e00003", "0x03c00003", "0x04200007"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseWindowIDs() = %q, want %q", got, want)
	}
}

// Windows are grouped by WM_CLASS class name. Desktop and panel windows, and
// windows with no WM_CLASS at all, must be left alone.
func TestParseXprop(t *testing.T) {
	tests := []struct {
		name      string
		out       string
		wantClass string
		wantOK    bool
	}{
		{
			name:      "normal window",
			out:       "WM_CLASS(STRING) = \"gnome-terminal-server\", \"Gnome-terminal\"\n_NET_WM_WINDOW_TYPE(ATOM) = _NET_WM_WINDOW_TYPE_NORMAL\n",
			wantClass: "Gnome-terminal", wantOK: true,
		},
		{
			name:      "no window type set",
			out:       "WM_CLASS(STRING) = \"firefox\", \"firefox\"\n_NET_WM_WINDOW_TYPE:  not found.\n",
			wantClass: "firefox", wantOK: true,
		},
		{
			name:      "comma inside class name",
			out:       "WM_CLASS(STRING) = \"app\", \"My, App\"\n",
			wantClass: "My, App", wantOK: true,
		},
		{
			name:   "desktop window",
			out:    "WM_CLASS(STRING) = \"desktop_window\", \"Nautilus\"\n_NET_WM_WINDOW_TYPE(ATOM) = _NET_WM_WINDOW_TYPE_DESKTOP\n",
			wantOK: false,
		},
		{
			name:   "panel (dock) window",
			out:    "WM_CLASS(STRING) = \"xfce4-panel\", \"Xfce4-panel\"\n_NET_WM_WINDOW_TYPE(ATOM) = _NET_WM_WINDOW_TYPE_DOCK, _NET_WM_WINDOW_TYPE_NORMAL\n",
			wantOK: false,
		},
		{
			name:   "no WM_CLASS",
			out:    "WM_CLASS:  not found.\n_NET_WM_WINDOW_TYPE(ATOM) = _NET_WM_WINDOW_TYPE_NORMAL\n",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class, ok := parseXprop(tt.out)
			if ok != tt.wantOK || class != tt.wantClass {
				t.Errorf("parseXprop() = (%q, %v), want (%q, %v)", class, ok, tt.wantClass, tt.wantOK)
			}
		})
	}
}

func TestCloseWindows(t *testing.T) {
	tests := []struct {
		name          string
		ids           []string
		requestErr    error
		openAfter     [][]string // window IDs open on each check; the last entry repeats
		openErr       error
		wantErr       bool
		wantRequested []string
	}{
		{
			name: "all windows close", ids: []string{"0x1", "0x2"},
			openAfter:     [][]string{{"0x1", "0x2", "0x9"}, {"0x2", "0x9"}, {"0x9"}},
			wantRequested: []string{"0x1", "0x2"},
		},
		{
			name: "a window stays open (save dialog)", ids: []string{"0x1"},
			openAfter: [][]string{{"0x1"}},
			wantErr:   true, wantRequested: []string{"0x1"},
		},
		{
			name: "close request fails", ids: []string{"0x1"},
			requestErr: errors.New("wmctrl failed"),
			wantErr:    true, wantRequested: []string{"0x1"},
		},
		{
			name: "window list keeps failing", ids: []string{"0x1"},
			openErr: errors.New("wmctrl failed"),
			wantErr: true, wantRequested: []string{"0x1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requested []string
			request := func(id string) error {
				requested = append(requested, id)
				return tt.requestErr
			}
			checks := 0
			open := func() ([]string, error) {
				if tt.openErr != nil {
					return nil, tt.openErr
				}
				i := checks
				if i >= len(tt.openAfter) {
					i = len(tt.openAfter) - 1
				}
				checks++
				return tt.openAfter[i], nil
			}

			// Long enough for several polls at pollInterval.
			err := closeWindows(tt.ids, request, open, 5*pollInterval)

			if (err != nil) != tt.wantErr {
				t.Errorf("closeWindows() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(requested, tt.wantRequested) {
				t.Errorf("requested close for %q, want %q", requested, tt.wantRequested)
			}
		})
	}
}
