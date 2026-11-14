package main

import (
	"github.com/abhisheksinha1506/lightsout/pkg/windows"
	"github.com/gleich/statuser/v2"
)

const platformSupported = true

// Turn on ANSI colors before anything is printed.
func init() { windows.EnableColors() }

// listApps returns the apps on the taskbar and quits one by closing each of
// its windows.
func listApps() ([]string, func(app string) error, error) {
	apps, err := windows.Applications()
	if err != nil {
		return nil, nil, err
	}
	return appNames(apps), func(app string) error { return windows.Close(apps[app]) }, nil
}

func limitsWarning() string { return "" }

// afterQuit closes all File Explorer folder windows; File Explorer itself
// (the taskbar and desktop) keeps running.
func afterQuit() {
	if err := windows.CloseExplorerWindows(); err != nil {
		statuser.Warning("Failed to close File Explorer windows:", err)
		return
	}
	statuser.Success("Closed all File Explorer windows")
}
