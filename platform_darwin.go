package main

import (
	"os"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/macos"
	"github.com/abhisheksinha1506/lightsout/pkg/proc"
)

const platformSupported = true

// newPlatform lists the apps in the Dock, quits one by asking each of its
// processes to quit, and closes all Finder windows at the end (Finder itself
// keeps running).
func newPlatform() platform {
	var running map[string][]int
	return platform{
		list: func() ([]app, error) {
			var err error
			running, err = macos.Applications()
			apps := make([]app, 0, len(running))
			for name, pids := range running {
				apps = append(apps, app{name: name, pids: pids})
			}
			return apps, err
		},
		quit:          func(name string, timeout time.Duration) error { return macos.Quit(running[name], timeout) },
		afterQuit:     macos.CloseFinder,
		afterQuitNote: "close all Finder windows",
		ancestors:     proc.Ancestors(os.Getpid(), proc.ParentPID),
	}
}
