// Package windows lists and politely closes apps on Microsoft Windows.
//
// The Win32 calls are in win32_windows.go. This file holds the logic that
// compiles and is tested on every OS: deciding which windows are apps, naming
// them, and the close-and-wait loop.
package windows

import (
	"errors"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

const (
	pollInterval = 100 * time.Millisecond

	explorerExe   = "explorer.exe"
	frameHostExe  = "applicationframehost.exe" // hosts Store apps' windows
	explorerClass = "CabinetWClass"            // a File Explorer folder window
)

// shellClasses are the desktop and taskbar windows, which are never apps.
var shellClasses = map[string]bool{
	"Progman":                true,
	"WorkerW":                true,
	"Shell_TrayWnd":          true,
	"Shell_SecondaryTrayWnd": true,
}

// App is one app on the taskbar.
type App struct {
	Windows []uintptr // window handles, in window order
	PIDs    []int     // process IDs that own the windows
}

// window describes one top-level window, as read from Windows.
type window struct {
	Handle      uintptr
	PID         int
	Title       string
	Class       string
	Exe         string // full path of the owning process's executable, "" if unreadable
	Description string // the executable's version-info FileDescription, "" if none
	Visible     bool
	Owned       bool // has an owner window, like a dialog
	ToolWindow  bool // WS_EX_TOOLWINDOW: floating palettes, hidden from the taskbar
	AppWindow   bool // WS_EX_APPWINDOW: forced onto the taskbar
	Cloaked     bool // hidden by the window manager, e.g. a suspended Store app
	OwnConsole  bool // the console window lightsout itself runs in
}

// onTaskbar reports whether w is an app window that would show on the
// taskbar, using the same rules as the taskbar.
func onTaskbar(w window) bool {
	if !w.Visible || w.Title == "" || w.Cloaked || w.OwnConsole || shellClasses[w.Class] {
		return false
	}
	return (!w.Owned && !w.ToolWindow) || w.AppWindow
}

// appName returns the name lightsout shows for w's app: the executable's
// description (such as "Google Chrome"), or its file name without ".exe".
// Store apps all run inside ApplicationFrameHost.exe, so they are named by
// their window title (such as "Calculator"), as is any window whose
// executable can't be read.
func appName(w window) string {
	exe := exeName(w.Exe)
	switch {
	case exe == "" || exe == frameHostExe:
		return w.Title
	case strings.TrimSpace(w.Description) != "":
		return strings.TrimSpace(w.Description)
	default:
		return strings.TrimSuffix(exe, ".exe")
	}
}

// exeName returns the lowercase file name of an executable path, which may
// use either slash.
func exeName(path string) string {
	return strings.ToLower(path[strings.LastIndexAny(path, `\/`)+1:])
}

// groupApps returns the apps on the taskbar, except File Explorer, as a map
// from app name to that app's windows (in window order) and process IDs.
func groupApps(windows []window) map[string]App {
	apps := map[string]App{}
	for _, w := range windows {
		if !onTaskbar(w) || exeName(w.Exe) == explorerExe {
			continue
		}
		name := appName(w)
		app := apps[name]
		app.Windows = append(app.Windows, w.Handle)
		if w.PID > 0 {
			app.PIDs = append(app.PIDs, w.PID)
		}
		apps[name] = app
	}
	return apps
}

// explorerFolders returns the handles of File Explorer's folder windows.
func explorerFolders(windows []window) []uintptr {
	var handles []uintptr
	for _, w := range windows {
		if exeName(w.Exe) == explorerExe && w.Class == explorerClass {
			handles = append(handles, w.Handle)
		}
	}
	return handles
}

// closeWindows asks each window in handles to close, then waits up to
// timeout for all of them to stop existing. request and exists are passed in
// so it can be tested without real windows.
//
// It never kills a process: a window that stays open, for example because
// the app is asking the user to save, is left alone and reported as an error.
func closeWindows(handles []uintptr, request func(handle uintptr) error, exists func(handle uintptr) bool, timeout time.Duration) error {
	for _, h := range handles {
		if err := request(h); err != nil {
			return err
		}
	}

	closed := wait.Until(timeout, pollInterval, func() bool {
		for _, h := range handles {
			if exists(h) {
				return false
			}
		}
		return true
	})
	if !closed {
		return errors.New("still open; it may be waiting for you to save")
	}
	return nil
}
