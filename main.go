package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/config"
	"github.com/abhisheksinha1506/lightsout/pkg/input"
	"github.com/abhisheksinha1506/lightsout/pkg/linux"
	"github.com/abhisheksinha1506/lightsout/pkg/macos"
	"github.com/abhisheksinha1506/lightsout/pkg/output"
	"github.com/abhisheksinha1506/lightsout/pkg/version"
	"github.com/briandowns/spinner"
	"github.com/gleich/statuser/v2"
)

func main() {
	statuser.Emojis = false
	operatingSystem := runtime.GOOS
	if operatingSystem != "darwin" && operatingSystem != "linux" {
		statuser.ErrorMsg("This application only works on macOS and linux", 1)
	}

	var configContents config.Conf
	if config.Exists() {
		config.Read(&configContents)
	}

	// Ignoring apps
	var ignoredApps []string
	if config.Exists() {
		ignoredApps = configContents.IgnoredApps
	}
	ignoredApps = append(ignoredApps, os.Args[1:]...)

	output.Title()

	if !configContents.IgnoreUpdates {
		version.CheckForUpdate()
	}

	// Getting running applications, and how to quit each one
	var apps []string
	var quit func(app string) error
	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Getting list of applications"
	s.Start()
	switch operatingSystem {
	case "darwin":
		macApps, err := macos.Applications()
		if err != nil {
			statuser.Error("Failed to get macos applications", err, 1)
		}
		for app := range macApps {
			apps = append(apps, app)
		}
		quit = func(app string) error { return macos.Quit(macApps[app]) }
	case "linux":
		linuxApps, err := linux.Applications()
		if err != nil {
			statuser.Error("Failed to get linux applications", err, 1)
		}
		for app := range linuxApps {
			apps = append(apps, app)
		}
		quit = func(app string) error { return linux.Close(linuxApps[app]) }
	}
	s.Stop()
	if operatingSystem == "linux" && os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		statuser.Warning("Wayland session: only apps running through XWayland can be listed and closed.")
	}

	// Getting executing terminal
	cleanedApps := input.ExecutingTerm(apps, ignoredApps)
	fmt.Println("")

	// Quitting applications
	failed := quitApps(cleanedApps, quit)
	if operatingSystem == "darwin" {
		macos.CloseFinder()
	}

	fmt.Println()
	if len(failed) > 0 {
		statuser.ErrorMsg("Failed to quit: "+strings.Join(failed, ", "), 1)
	}
	statuser.Success("Quit All Applications! Have a good day :)")
}

// quitApps quits each app in order using quit, showing a spinner per app.
// A failure is printed as a warning and skipped so the remaining apps are
// still quit. It returns the apps that failed to quit, in order.
func quitApps(apps []string, quit func(app string) error) []string {
	var failed []string
	for _, app := range apps {
		s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
		s.Suffix = " Quitting " + app
		s.Start()
		err := quit(app)
		s.Stop()
		if err != nil {
			statuser.Warning("Failed to quit "+app+":", err)
			failed = append(failed, app)
			continue
		}
		statuser.Success("Quit " + app)
	}
	return failed
}
