package proc

import "golang.org/x/sys/unix"

// ParentPID returns the parent process ID of pid.
func ParentPID(pid int) (int, error) {
	info, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return 0, err
	}
	return int(info.Eproc.Ppid), nil
}
