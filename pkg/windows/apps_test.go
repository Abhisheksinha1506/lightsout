package windows

import (
	"errors"
	"reflect"
	"testing"
)

// appWindow is a normal, visible, titled, unowned app window; tests change
// one field at a time.
func appWindow() window {
	return window{
		Handle:  1,
		Title:   "Inbox - Mail",
		Class:   "MailWindowClass",
		Exe:     `C:\Program Files\Mail\mail.exe`,
		Visible: true,
	}
}

// Only windows that would show on the taskbar count as apps: never the
// desktop, the taskbar itself, dialogs, tool palettes, hidden or cloaked
// windows, or lightsout's own console.
func TestOnTaskbar(t *testing.T) {
	tests := []struct {
		name   string
		change func(w *window)
		want   bool
	}{
		{name: "normal app window", change: func(w *window) {}, want: true},
		{name: "hidden", change: func(w *window) { w.Visible = false }, want: false},
		{name: "untitled", change: func(w *window) { w.Title = "" }, want: false},
		{name: "cloaked (e.g. suspended Store app)", change: func(w *window) { w.Cloaked = true }, want: false},
		{name: "owned dialog", change: func(w *window) { w.Owned = true }, want: false},
		{name: "tool window", change: func(w *window) { w.ToolWindow = true }, want: false},
		{name: "tool window marked as app window", change: func(w *window) { w.ToolWindow, w.AppWindow = true, true }, want: true},
		{name: "owned window marked as app window", change: func(w *window) { w.Owned, w.AppWindow = true, true }, want: true},
		{name: "desktop (Progman)", change: func(w *window) { w.Class = "Progman" }, want: false},
		{name: "desktop (WorkerW)", change: func(w *window) { w.Class = "WorkerW" }, want: false},
		{name: "taskbar", change: func(w *window) { w.Class = "Shell_TrayWnd" }, want: false},
		{name: "second-monitor taskbar", change: func(w *window) { w.Class = "Shell_SecondaryTrayWnd" }, want: false},
		{name: "lightsout's own console", change: func(w *window) { w.OwnConsole = true }, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := appWindow()
			tt.change(&w)
			if got := onTaskbar(w); got != tt.want {
				t.Errorf("onTaskbar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppName(t *testing.T) {
	tests := []struct {
		name string
		w    window
		want string
	}{
		{name: "file description", w: window{Exe: `C:\Program Files\Google\Chrome\Application\chrome.exe`, Description: "Google Chrome", Title: "New Tab"}, want: "Google Chrome"},
		{name: "description with spaces around", w: window{Exe: `C:\x\app.exe`, Description: "  My App  "}, want: "My App"},
		{name: "no description: exe name", w: window{Exe: `C:\Tools\Some Tool\sometool.exe`, Title: "Untitled"}, want: "sometool"},
		{name: "forward slashes", w: window{Exe: "C:/Tools/other.EXE"}, want: "other"},
		{name: "Store app: window title", w: window{Exe: `C:\Windows\System32\ApplicationFrameHost.exe`, Description: "Application Frame Host", Title: "Calculator"}, want: "Calculator"},
		{name: "exe unreadable (e.g. elevated): window title", w: window{Title: "Administrator: Registry Editor"}, want: "Administrator: Registry Editor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appName(tt.w); got != tt.want {
				t.Errorf("appName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Windows of the same app are grouped, keeping their order. File Explorer is
// not an app to quit (its folder windows are closed separately), and windows
// that aren't on the taskbar are left out.
func TestGroupApps(t *testing.T) {
	chrome := `C:\Program Files\Google\Chrome\Application\chrome.exe`
	windows := []window{
		{Handle: 10, PID: 100, Title: "Inbox", Exe: chrome, Description: "Google Chrome", Visible: true},
		{Handle: 11, Title: "Docs", Exe: `C:\Windows\explorer.exe`, Class: "CabinetWClass", Description: "Windows Explorer", Visible: true},
		{Handle: 12, PID: 100, Title: "News", Exe: chrome, Description: "Google Chrome", Visible: true},
		{Handle: 13, Title: "Save as", Exe: chrome, Description: "Google Chrome", Visible: true, Owned: true},
		{Handle: 14, PID: 200, Title: "Untitled - Notepad", Exe: `C:\Windows\notepad.exe`, Description: "Notepad", Visible: true},
		{Handle: 15, Title: "Program Manager", Exe: `C:\Windows\explorer.exe`, Class: "Progman", Visible: true},
	}

	got := groupApps(windows)

	want := map[string]App{
		"Google Chrome": {Windows: []uintptr{10, 12}, PIDs: []int{100, 100}},
		"Notepad":       {Windows: []uintptr{14}, PIDs: []int{200}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groupApps() = %v, want %v", got, want)
	}
}

func TestExplorerFolders(t *testing.T) {
	explorer := `C:\Windows\explorer.exe`
	windows := []window{
		{Handle: 20, Title: "Documents", Exe: explorer, Class: "CabinetWClass", Visible: true},
		{Handle: 21, Title: "Program Manager", Exe: explorer, Class: "Progman", Visible: true},
		{Handle: 22, Title: "", Exe: explorer, Class: "Shell_TrayWnd", Visible: true},
		{Handle: 23, Title: "Downloads", Exe: `C:\Other\explorer-clone.exe`, Class: "CabinetWClass", Visible: true},
		{Handle: 24, Title: "Pictures", Exe: explorer, Class: "CabinetWClass", Visible: true},
	}

	if got, want := explorerFolders(windows), []uintptr{20, 24}; !reflect.DeepEqual(got, want) {
		t.Errorf("explorerFolders() = %v, want %v", got, want)
	}
}

func TestCloseWindows(t *testing.T) {
	tests := []struct {
		name          string
		handles       []uintptr
		requestErr    error
		openChecks    int // windows report open for this many checks; -1 means forever
		wantErr       bool
		wantRequested []uintptr
	}{
		{name: "all windows close", handles: []uintptr{1, 2}, openChecks: 2, wantRequested: []uintptr{1, 2}},
		{name: "a window stays open (save dialog)", handles: []uintptr{1}, openChecks: -1, wantErr: true, wantRequested: []uintptr{1}},
		{name: "close request refused", handles: []uintptr{1, 2}, requestErr: errors.New("access denied"), wantErr: true, wantRequested: []uintptr{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requested []uintptr
			request := func(h uintptr) error {
				requested = append(requested, h)
				return tt.requestErr
			}
			checks := 0
			exists := func(uintptr) bool {
				checks++
				return tt.openChecks < 0 || checks <= tt.openChecks
			}

			// Long enough for several polls at pollInterval.
			err := closeWindows(tt.handles, request, exists, 5*pollInterval)

			if (err != nil) != tt.wantErr {
				t.Errorf("closeWindows() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(requested, tt.wantRequested) {
				t.Errorf("requested close for %v, want %v", requested, tt.wantRequested)
			}
		})
	}
}
