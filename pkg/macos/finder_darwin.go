package macos

import (
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/gleich/statuser/v2"
)

// CloseFinder closes all open Finder windows. A failure only prints a warning, so the
// summary of apps that could not be quit is still shown afterwards.
func CloseFinder() {
	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Closing Finder windows"
	s.Start()
	err := exec.Command("osascript", "-e", `tell application "Finder" to close windows`).Run()
	s.Stop()
	if err != nil {
		statuser.Warning("Failed to close Finder windows:", err)
		return
	}
	statuser.Success("Closed all Finder windows")
}
