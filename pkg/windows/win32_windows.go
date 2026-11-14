//go:build windows

package windows

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	win "golang.org/x/sys/windows"
)

// Win32 functions that golang.org/x/sys/windows doesn't wrap.
var (
	user32   = win.NewLazySystemDLL("user32.dll")
	kernel32 = win.NewLazySystemDLL("kernel32.dll")

	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindow            = user32.NewProc("GetWindow")
	procGetWindowLongPtrW    = user32.NewProc("GetWindowLongPtrW")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procGetConsoleWindow     = kernel32.NewProc("GetConsoleWindow")
)

const (
	gwOwner        = 4            // GW_OWNER
	gwlExStyle     = ^uintptr(19) // GWL_EXSTYLE, which is -20
	wsExToolWindow = 0x00000080   // WS_EX_TOOLWINDOW
	wsExAppWindow  = 0x00040000   // WS_EX_APPWINDOW
	wmClose        = 0x0010       // WM_CLOSE
)

// Applications returns the apps on the taskbar, except File Explorer, as a
// map from app name to that app's windows and process IDs.
func Applications() (map[string]App, error) {
	windows, err := topLevelWindows()
	if err != nil {
		return nil, err
	}
	return groupApps(windows), nil
}

// Close asks each window in handles to close, as if the user clicked its X
// button, then waits up to timeout for all of them to go away. A window
// that stays open, for example to ask the user to save, is left alone and
// reported as an error.
func Close(handles []uintptr, timeout time.Duration) error {
	return closeWindows(handles, postClose, windowExists, timeout)
}

// CloseExplorerWindows asks every File Explorer folder window to close. It
// doesn't wait; File Explorer itself (the taskbar and desktop) keeps running.
func CloseExplorerWindows() error {
	windows, err := topLevelWindows()
	if err != nil {
		return err
	}
	for _, h := range explorerFolders(windows) {
		if err := postClose(h); err != nil {
			return err
		}
	}
	return nil
}

// EnableColors turns on ANSI escape sequences for the console, so colored
// output and the banner render in classic consoles too. Failures are ignored:
// output still works, just without colors.
func EnableColors() {
	stdout := win.Handle(os.Stdout.Fd())
	var mode uint32
	if win.GetConsoleMode(stdout, &mode) == nil {
		_ = win.SetConsoleMode(stdout, mode|win.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}

// topLevelWindows reads every top-level window on the desktop.
func topLevelWindows() ([]window, error) {
	var handles []win.HWND
	callback := syscall.NewCallback(func(h win.HWND, _ uintptr) uintptr {
		handles = append(handles, h)
		return 1 // keep enumerating
	})
	if err := win.EnumWindows(callback, nil); err != nil {
		return nil, fmt.Errorf("listing windows: %w", err)
	}

	console, _, _ := procGetConsoleWindow.Call()
	descriptions := map[string]string{} // FileDescription per executable
	windows := make([]window, 0, len(handles))
	for _, h := range handles {
		w := readWindow(h)
		w.OwnConsole = uintptr(h) == console
		if w.Exe != "" {
			if _, ok := descriptions[w.Exe]; !ok {
				descriptions[w.Exe] = fileDescription(w.Exe)
			}
			w.Description = descriptions[w.Exe]
		}
		windows = append(windows, w)
	}
	return windows, nil
}

// readWindow reads what onTaskbar and appName need to know about one window.
func readWindow(h win.HWND) window {
	owner, _, _ := procGetWindow.Call(uintptr(h), gwOwner)
	exStyle, _, _ := procGetWindowLongPtrW.Call(uintptr(h), gwlExStyle)
	var cloaked uint32
	_ = win.DwmGetWindowAttribute(h, win.DWMWA_CLOAKED, unsafe.Pointer(&cloaked), uint32(unsafe.Sizeof(cloaked)))
	var pid uint32
	_, _ = win.GetWindowThreadProcessId(h, &pid)

	return window{
		Handle:     uintptr(h),
		PID:        int(pid),
		Title:      windowText(h),
		Class:      className(h),
		Exe:        processExe(pid),
		Visible:    win.IsWindowVisible(h),
		Owned:      owner != 0,
		ToolWindow: exStyle&wsExToolWindow != 0,
		AppWindow:  exStyle&wsExAppWindow != 0,
		Cloaked:    cloaked != 0,
	}
}

// windowText returns a window's title. For windows of other processes it
// reads the stored caption without messaging the window, so a hung app can't
// block it.
func windowText(h win.HWND) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(h))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	copied, _, _ := procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return win.UTF16ToString(buf[:copied])
}

// className returns a window's class name, or "" if it can't be read.
func className(h win.HWND) string {
	buf := make([]uint16, 256)
	n, err := win.GetClassName(h, &buf[0], int32(len(buf)))
	if err != nil {
		return ""
	}
	return win.UTF16ToString(buf[:n])
}

// processExe returns the full path of a process's executable, or "" if it
// can't be read (for example a protected system process).
func processExe(pid uint32) string {
	process, err := win.OpenProcess(win.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer win.CloseHandle(process) //nolint:errcheck // nothing useful to do if closing fails

	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	if err := win.QueryFullProcessImageName(process, 0, &buf[0], &size); err != nil {
		return ""
	}
	return win.UTF16ToString(buf[:size])
}

// fileDescription returns the FileDescription from an executable's version
// information (such as "Google Chrome"), or "" if it has none.
func fileDescription(exe string) string {
	size, err := win.GetFileVersionInfoSize(exe, nil)
	if err != nil || size == 0 {
		return ""
	}
	info := make([]byte, size)
	if err := win.GetFileVersionInfo(exe, 0, size, unsafe.Pointer(&info[0])); err != nil {
		return ""
	}

	// The strings are stored per language and code page, listed under
	// \VarFileInfo\Translation; use the first one.
	var translation *[2]uint16
	var length uint32
	if err := win.VerQueryValue(unsafe.Pointer(&info[0]), `\VarFileInfo\Translation`, unsafe.Pointer(&translation), &length); err != nil || length < 4 {
		return ""
	}
	key := fmt.Sprintf(`\StringFileInfo\%04x%04x\FileDescription`, translation[0], translation[1])
	var description *uint16
	if err := win.VerQueryValue(unsafe.Pointer(&info[0]), key, unsafe.Pointer(&description), &length); err != nil || length == 0 {
		return ""
	}
	return win.UTF16PtrToString(description)
}

// postClose sends a window WM_CLOSE, the message its X button sends.
func postClose(h uintptr) error {
	ok, _, err := procPostMessageW.Call(h, wmClose, 0, 0)
	if ok != 0 {
		return nil
	}
	if errors.Is(err, win.ERROR_ACCESS_DENIED) {
		// Windows blocks messages to apps running with higher rights.
		return errors.New("it runs as administrator; close it yourself")
	}
	return err
}

// windowExists reports whether a window handle still refers to a window.
func windowExists(h uintptr) bool {
	return win.IsWindow(win.HWND(h))
}
