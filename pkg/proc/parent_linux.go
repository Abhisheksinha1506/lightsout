package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParentPID returns the parent process ID of pid, read from /proc.
func ParentPID(pid int) (int, error) {
	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}
	return parseStatPPID(string(stat))
}

// parseStatPPID reads the parent PID from a /proc/<pid>/stat line. The
// command name in parentheses can contain spaces and ")", so fields are
// counted from the last ")".
func parseStatPPID(stat string) (int, error) {
	end := strings.LastIndexByte(stat, ')')
	if end < 0 {
		return 0, fmt.Errorf("unexpected /proc stat format: %q", stat)
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected /proc stat format: %q", stat)
	}
	return strconv.Atoi(fields[1]) // fields[0] is the state
}
