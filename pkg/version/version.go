package version

import (
	"fmt"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
	"github.com/gleich/release"
	"github.com/gleich/statuser/v2"
)

// Version of lightsout, matching its git tag (e.g. v5.1.2). Release builds set
// it with -ldflags (see .goreleaser.yml):
//
//	-X github.com/abhisheksinha1506/lightsout/pkg/version.Version=<tag>
//
// Other builds keep devVersion.
var Version = devVersion

// Version of builds not made from a release tag
const devVersion = "dev"

// updateCheckTimeout is the longest each update-check request may take.
var updateCheckTimeout = 5 * time.Second

// Check for an update for the CLI. Dev builds are not a release, so they skip
// the check. A failed check only prints a warning so that lightsout still runs
// when GitHub is unreachable or rate limited.
func CheckForUpdate() {
	if Version == devVersion {
		return
	}

	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Checking for update"
	s.Start()

	// release.Check uses http.Get, which has no timeout, so a stalled network
	// would hang lightsout here. lightsout makes no other HTTP requests, so
	// setting the default client's timeout is safe.
	http.DefaultClient.Timeout = updateCheckTimeout

	const repoURL = "https://github.com/abhisheksinha1506/lightsout"
	isOutdated, version, err := release.Check(Version, repoURL)
	s.Stop()
	if err != nil {
		statuser.Warning("Failed to check for update, skipping:", err)
		return
	}
	if isOutdated {
		fmt.Println("\nPLEASE UPDATE TO NEW VERSION", version)
		fmt.Println("Get update from", repoURL, "or with your package manager")
		fmt.Println()
	}
}
