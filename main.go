package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/config"
	"github.com/abhisheksinha1506/lightsout/pkg/input"
	"github.com/abhisheksinha1506/lightsout/pkg/output"
	"github.com/abhisheksinha1506/lightsout/pkg/version"
	"github.com/briandowns/spinner"
	"github.com/gleich/statuser/v2"
)

// Each supported operating system has a platform_<os>.go file providing:
//
//	platformSupported  true on macOS, Linux and Windows
//	listApps()         the running apps' names, and how to quit one by name
//	limitsWarning()    a warning about apps lightsout can't see, or ""
//	afterQuit()        cleanup once every app was asked to quit
func main() {
	statuser.Emojis = false
	if !platformSupported {
		statuser.ErrorMsg("This application only works on macOS, Linux and Windows", 1)
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
	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Getting list of applications"
	s.Start()
	apps, quit, err := listApps()
	s.Stop()
	if err != nil {
		statuser.Error("Failed to get the running applications", err, 1)
	}
	if warning := limitsWarning(); warning != "" {
		statuser.Warning(warning)
	}

	// Getting executing terminal
	cleanedApps := input.ExecutingTerm(apps, ignoredApps)
	fmt.Println("")

	// Quitting applications
	failed := quitApps(cleanedApps, quit)
	afterQuit()

	fmt.Println()
	if len(failed) > 0 {
		statuser.ErrorMsg("Failed to quit: "+strings.Join(failed, ", "), 1)
	}
	statuser.Success("Quit All Applications! Have a good day :)")
}

// appNames returns the keys of a map from app name to its quit handles, sorted.
func appNames[T any](apps map[string]T) []string {
	names := make([]string, 0, len(apps))
	for name := range apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
