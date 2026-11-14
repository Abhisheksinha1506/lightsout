package main

import (
	"os"

	"github.com/abhisheksinha1506/lightsout/pkg/linux"
)

const platformSupported = true

// listApps returns the apps with open windows and quits one by closing each
// of its windows.
func listApps() ([]string, func(app string) error, error) {
	apps, err := linux.Applications()
	if err != nil {
		return nil, nil, err
	}
	return appNames(apps), func(app string) error { return linux.Close(apps[app]) }, nil
}

// limitsWarning explains that native Wayland windows are invisible to wmctrl.
func limitsWarning() string {
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return "Wayland session: only apps running through XWayland can be listed and closed."
	}
	return ""
}

func afterQuit() {}
