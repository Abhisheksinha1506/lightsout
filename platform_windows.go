package main

import (
	"os"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/proc"
	"github.com/abhisheksinha1506/lightsout/pkg/windows"
	"github.com/gleich/statuser/v2"
)

const platformSupported = true

// Turn on ANSI colors before anything is printed.
func init() { windows.EnableColors() }

// newPlatform lists the apps on the taskbar, quits one by closing each of its
// windows, and closes all File Explorer folder windows at the end (File
// Explorer itself, the taskbar and desktop, keeps running).
func newPlatform() platform {
	var running map[string]windows.App
	return platform{
		list: func() ([]app, error) {
			var err error
			running, err = windows.Applications()
			apps := make([]app, 0, len(running))
			for name, a := range running {
				apps = append(apps, app{name: name, pids: a.PIDs})
			}
			return apps, err
		},
		quit:          func(name string, timeout time.Duration) error { return windows.Close(running[name].Windows, timeout) },
		afterQuit:     closeExplorerWindows,
		afterQuitNote: "close all File Explorer windows",
		ancestors:     proc.Ancestors(os.Getpid(), proc.ParentPID),
	}
}

func closeExplorerWindows() {
	if err := windows.CloseExplorerWindows(); err != nil {
		statuser.Warning("Failed to close File Explorer windows:", err)
		return
	}
	statuser.Success("Closed all File Explorer windows")
}
