package linux

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

const (
	// closeTimeout is how long Close waits for an app's windows to go away.
	closeTimeout = 10 * time.Second
	pollInterval = 100 * time.Millisecond
)

// Applications returns the apps with open windows, as a map from each
// window's WM_CLASS class name to the IDs of that app's windows.
//
// Desktop and panel windows are skipped, as are windows without a WM_CLASS.
// It needs the wmctrl and xprop tools, and only sees X11 windows (including
// XWayland), not native Wayland ones.
func Applications() (map[string][]string, error) {
	for _, tool := range []string{"wmctrl", "xprop"} {
		if _, err := exec.LookPath(tool); err != nil {
			return nil, fmt.Errorf("%s is required to list apps; install it with your package manager", tool)
		}
	}

	ids, err := openWindowIDs()
	if err != nil {
		return nil, err
	}
	apps := map[string][]string{}
	for _, id := range ids {
		out, err := exec.Command("xprop", "-id", id, "WM_CLASS", "_NET_WM_WINDOW_TYPE").Output()
		if err != nil {
			continue // the window closed while we were listing
		}
		if class, ok := parseXprop(string(out)); ok {
			apps[class] = append(apps[class], id)
		}
	}
	return apps, nil
}

// Close asks each window in windowIDs to close, as if the user clicked its
// close button, then waits up to closeTimeout for all of them to go away.
//
// It never kills a process: a window that stays open, for example because the
// app is asking the user to save, is left alone and reported as an error.
func Close(windowIDs []string) error {
	return closeWindows(windowIDs, requestClose, openWindowIDs, closeTimeout)
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

// parseXprop reads `xprop -id <id> WM_CLASS _NET_WM_WINDOW_TYPE` output and
// returns the window's class name, the last quoted WM_CLASS value.
//
// ok is false for windows to leave alone: no WM_CLASS, or a desktop or dock
// (panel) window type.
func parseXprop(out string) (class string, ok bool) {
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
				return "", false
			}
		}
	}
	return class, class != ""
}
