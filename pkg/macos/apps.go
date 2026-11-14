package macos

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

// The code that talks to macOS (Applications, Quit) is in apps_darwin.go. This
// file holds the logic that compiles and is tested on every OS.

const (
	// regularPolicy is NSApplicationActivationPolicyRegular: apps shown in the
	// Dock. Other policies are menu bar agents and system processes such as
	// loginwindow, which must never be quit.
	regularPolicy = 0
	finderBundle  = "com.apple.finder"
	// quitTimeout is how long Quit waits for an app to exit.
	quitTimeout  = 10 * time.Second
	pollInterval = 100 * time.Millisecond
)

type runningApp struct {
	Name   string `json:"name"`
	PID    int    `json:"pid"`
	Policy int    `json:"policy"`
	Bundle string `json:"bundle"`
}

// parseApps reads appListScript's output and keeps named Dock apps other than
// Finder, grouping process IDs by app name.
func parseApps(out []byte) (map[string][]int, error) {
	var running []runningApp
	if err := json.Unmarshal(out, &running); err != nil {
		return nil, fmt.Errorf("reading the app list: %w", err)
	}

	apps := map[string][]int{}
	for _, app := range running {
		if app.Policy != regularPolicy || app.Bundle == finderBundle || app.Name == "" || app.PID <= 0 {
			continue
		}
		apps[app.Name] = append(apps[app.Name], app.PID)
	}
	return apps, nil
}

// quit is Quit with the quit request, the running check and the timeout
// passed in, so it can be tested without quitting real apps.
func quit(pids []int, request func(pid int) error, running func(pid int) bool, timeout time.Duration) error {
	for _, pid := range pids {
		if err := request(pid); err != nil {
			return err
		}
	}

	exited := wait.Until(timeout, pollInterval, func() bool {
		for _, pid := range pids {
			if running(pid) {
				return false
			}
		}
		return true
	})
	if !exited {
		return errors.New("still open; it may be waiting for you to save")
	}
	return nil
}
