//go:build windows

package windows

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/abhisheksinha1506/lightsout/pkg/wait"
)

// End-to-end check of the real Win32 code: open Notepad, find it in
// Applications, close it with Close. It opens and closes a real window, so it
// only runs when LIGHTSOUT_WINDOWS_E2E=1 (as in CI).
func TestRealNotepadIsListedAndCloses(t *testing.T) {
	if os.Getenv("LIGHTSOUT_WINDOWS_E2E") != "1" {
		t.Skip("set LIGHTSOUT_WINDOWS_E2E=1 to run (opens and closes Notepad)")
	}

	notepad := exec.Command("notepad.exe")
	if err := notepad.Start(); err != nil {
		t.Fatal(err)
	}
	// Only matters if the test fails before Notepad was closed.
	t.Cleanup(func() { _ = notepad.Process.Kill() })

	var handles []uintptr
	found := wait.Until(15*time.Second, 200*time.Millisecond, func() bool {
		apps, err := Applications()
		if err != nil {
			return false
		}
		for name, h := range apps {
			if strings.Contains(strings.ToLower(name), "notepad") {
				handles = h
				return true
			}
		}
		return false
	})
	if !found {
		apps, err := Applications()
		names := make([]string, 0, len(apps))
		for name := range apps {
			names = append(names, name)
		}
		t.Fatalf("Notepad not listed after 15s; listed apps: %q (err %v)", names, err)
	}

	if err := Close(handles); err != nil {
		t.Fatalf("Close(%v) = %v", handles, err)
	}
}
