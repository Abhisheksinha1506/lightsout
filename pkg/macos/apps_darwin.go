package macos

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// appListScript prints every running app as a JSON array of
// {"name", "pid", "policy", "bundle"} objects. It uses NSWorkspace, which
// needs no Automation permission. The JavaScript bridge returns
// activationPolicy as a string, so numbers are converted explicitly.
const appListScript = `ObjC.import("AppKit");
const apps = $.NSWorkspace.sharedWorkspace.runningApplications;
const out = [];
for (let i = 0; i < apps.count; i++) {
  const a = apps.objectAtIndex(i);
  out.push({
    name: ObjC.unwrap(a.localizedName) || "",
    pid: Number(a.processIdentifier),
    policy: Number(a.activationPolicy),
    bundle: ObjC.unwrap(a.bundleIdentifier) || "",
  });
}
JSON.stringify(out);`

// quitScript asks the app whose process ID is its first argument to quit, the
// same request as choosing Quit from the app's menu. It prints "sent",
// "refused", or "gone" if the app is no longer running.
const quitScript = `function run(argv) {
  ObjC.import("AppKit");
  const app = $.NSRunningApplication.runningApplicationWithProcessIdentifier(parseInt(argv[0], 10));
  if (app.isNil()) { return "gone"; }
  return app.terminate ? "sent" : "refused";
}`

// Applications returns the apps shown in the Dock, except Finder, as a map
// from app name to the process IDs running under that name.
func Applications() (map[string][]int, error) {
	out, err := exec.Command("osascript", "-l", "JavaScript", "-e", appListScript).Output()
	if err != nil {
		return nil, err
	}
	return parseApps(out)
}

// Quit asks each process in pids to quit, as if the user chose Quit from the
// app's menu, then waits up to timeout for all of them to exit.
//
// It never force quits: an app that stays open, for example because it is
// asking the user to save, is left running and reported as an error.
func Quit(pids []int, timeout time.Duration) error {
	return quit(pids, requestQuit, isRunning, timeout)
}

// requestQuit sends one quit request through quitScript. The process ID is
// passed as a script argument, never inserted into the script text.
func requestQuit(pid int) error {
	out, err := exec.Command("osascript", "-l", "JavaScript", "-e", quitScript, strconv.Itoa(pid)).Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) == "refused" {
		return errors.New("the app refused to quit")
	}
	return nil
}

// isRunning reports whether a process with this ID still exists.
func isRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
