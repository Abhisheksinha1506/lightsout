package main

import "github.com/abhisheksinha1506/lightsout/pkg/macos"

const platformSupported = true

// listApps returns the apps in the Dock and quits one by asking each of its
// processes to quit.
func listApps() ([]string, func(app string) error, error) {
	apps, err := macos.Applications()
	if err != nil {
		return nil, nil, err
	}
	return appNames(apps), func(app string) error { return macos.Quit(apps[app]) }, nil
}

func limitsWarning() string { return "" }

// afterQuit closes all Finder windows; Finder itself keeps running.
func afterQuit() { macos.CloseFinder() }
