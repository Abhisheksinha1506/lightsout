package proc

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ParentPID returns the parent process ID of pid, from a snapshot of all
// processes.
func ParentPID(pid int) (int, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot) //nolint:errcheck // nothing to do if closing fails

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if int(entry.ProcessID) == pid {
			return int(entry.ParentProcessID), nil
		}
	}
	return 0, fmt.Errorf("process %d not found", pid)
}
