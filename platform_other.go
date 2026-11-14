//go:build !darwin && !linux && !windows

package main

import "errors"

// lightsout doesn't support this operating system; main exits early.
const platformSupported = false

func newPlatform() platform {
	return platform{
		list:      func() ([]app, error) { return nil, errors.New("unsupported operating system") },
		afterQuit: func() {},
	}
}
