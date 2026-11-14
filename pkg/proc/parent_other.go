//go:build !darwin && !linux && !windows

package proc

import "errors"

// ParentPID is not supported on this operating system.
func ParentPID(int) (int, error) { return 0, errors.New("unsupported operating system") }
