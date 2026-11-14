//go:build !darwin && !linux && !windows

package main

import "errors"

// lightsout doesn't support this operating system; main exits early.
const platformSupported = false

func listApps() ([]string, func(app string) error, error) {
	return nil, nil, errors.New("unsupported operating system")
}

func limitsWarning() string { return "" }

func afterQuit() {}
