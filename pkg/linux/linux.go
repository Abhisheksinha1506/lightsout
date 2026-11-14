package linux

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

const pollInterval = 100 * time.Millisecond

// App is one app with open windows.
type App struct {
	Windows []string // window IDs
	PIDs    []int    // process IDs that own the windows, where known
}

// Applications returns the apps with open windows, as a map from each
// window's WM_CLASS class name to that app's windows and process IDs.
//
// Desktop and panel windows are skipped, as are windows without a WM_CLASS.
// It needs the wmctrl and xprop tools, and only sees X11 windows (including
// XWayland), not native Wayland ones.
func Applications() (map[string]App, error) {
	for _, tool := range []string{"wmctrl", "xprop"} {
		if _, err := exec.LookPath(tool); err != nil {
			return nil, fmt.Errorf("%s is required to list apps; install it with your package manager", tool)
		}
	}

	ids, err := openWindowIDs()
	if err != nil {
		return nil, err
	}
	apps := map[string]App{}
	for _, id := range ids {
		out, err := exec.Command("xprop", "-id", id, "WM_CLASS", "_NET_WM_WINDOW_TYPE", "_NET_WM_PID").Output()
		if err != nil {
			continue // the window closed while we were listing
		}
		class, pid, ok := parseXprop(string(out))
		if !ok {
			continue
		}
		app := apps[class]
		app.Windows = append(app.Windows, id)
		if pid > 0 {
			app.PIDs = append(app.PIDs, pid)
		}
		apps[class] = app
	}
	return apps, nil
}

// Close asks each window in windowIDs to close, as if the user clicked its
// close button, then waits up to timeout for all of them to go away.
//
// It never kills a process: a window that stays open, for example because the
// app is asking the user to save, is left alone and reported as an error.
func Close(windowIDs []string, timeout time.Duration) error {
	return closeWindows(windowIDs, requestClose, openWindowIDs, timeout)
}

// closeWindows is Close with the close request, the window listing and the
// timeout passed in, so it can be tested without a desktop.
func closeWindows(ids []string, request func(id string) error, open func() ([]string, error), timeout time.Duration) error {
	for _, id := range ids {
		if err := request(id); err != nil {
			return err
		}
	}

	closed := wait.Until(timeout, pollInterval, func() bool {
		current, err := open()
		if err != nil {
			return false
		}
		stillOpen := map[string]bool{}
		for _, id := range current {
			stillOpen[id] = true
		}
		for _, id := range ids {
			if stillOpen[id] {
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

// requestClose asks the window manager to close one window by its ID.
func requestClose(id string) error {
	return exec.Command("wmctrl", "-i", "-c", id).Run()
}

// openWindowIDs returns the IDs of all windows the window manager manages.
func openWindowIDs() ([]string, error) {
	out, err := exec.Command("wmctrl", "-l").Output()
	if err != nil {
		return nil, err
	}
	return parseWindowIDs(string(out)), nil
}

// parseWindowIDs returns the window ID (first column) of each `wmctrl -l` line.
func parseWindowIDs(out string) []string {
	var ids []string
	for _, line := range strings.Split(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			ids = append(ids, fields[0])
		}
	}
	return ids
}

// parseXprop reads `xprop -id <id> WM_CLASS _NET_WM_WINDOW_TYPE _NET_WM_PID`
// output and returns the window's class name (the last quoted WM_CLASS
// value) and the process ID that owns it, or 0 if the window doesn't say.
//
// ok is false for windows to leave alone: no WM_CLASS, or a desktop or dock
// (panel) window type.
func parseXprop(out string) (class string, pid int, ok bool) {
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "WM_CLASS("):
			// WM_CLASS(STRING) = "instance", "Class": quoted values sit at odd
			// indexes when split on quotes.
			parts := strings.Split(line, `"`)
			for i := 1; i < len(parts); i += 2 {
				class = parts[i]
			}
		case strings.HasPrefix(line, "_NET_WM_WINDOW_TYPE("):
			if strings.Contains(line, "_NET_WM_WINDOW_TYPE_DESKTOP") || strings.Contains(line, "_NET_WM_WINDOW_TYPE_DOCK") {
				return "", 0, false
			}
		case strings.HasPrefix(line, "_NET_WM_PID("):
			// _NET_WM_PID(CARDINAL) = 1234
			if _, value, found := strings.Cut(line, "= "); found {
				pid, _ = strconv.Atoi(strings.TrimSpace(value))
			}
		}
	}
	return class, pid, class != ""
}
