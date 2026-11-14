package main

import (
	"os"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/linux"
	"github.com/abhisheksinha1506/lightsout/pkg/proc"
)

const platformSupported = true

// newPlatform lists the apps with open windows and quits one by closing each
// of its windows. In a Wayland session it warns that native Wayland windows
// are invisible to wmctrl.
func newPlatform() platform {
	var running map[string]linux.App
	p := platform{
		list: func() ([]app, error) {
			var err error
			running, err = linux.Applications()
			apps := make([]app, 0, len(running))
			for name, a := range running {
				apps = append(apps, app{name: name, pids: a.PIDs})
			}
			return apps, err
		},
		quit:      func(name string, timeout time.Duration) error { return linux.Close(running[name].Windows, timeout) },
		afterQuit: func() {},
		ancestors: proc.Ancestors(os.Getpid(), proc.ParentPID),
	}
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		p.warning = "Wayland session: only apps running through XWayland can be listed and closed."
	}
	return p
}
