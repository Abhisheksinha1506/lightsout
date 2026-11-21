package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/briandowns/spinner"
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

const (
	repoURL          = "https://github.com/abhisheksinha1506/lightsout"
	latestReleaseURL = "https://api.github.com/repos/abhisheksinha1506/lightsout/releases/latest"
)

// CheckForUpdate checks GitHub for a newer release. Dev builds are not a
// release, so they skip the check. A failed check never stops lightsout: if
// GitHub can't be reached it is skipped silently, and if GitHub answered but
// the check failed (for example rate limiting) it prints a warning.
func CheckForUpdate() {
	if Version == devVersion {
		return
	}

	s := spinner.New(spinner.CharSets[13], 30*time.Millisecond)
	s.Suffix = " Checking for update"
	s.Start()
	latest, err := latestVersion(&http.Client{Timeout: updateCheckTimeout}, latestReleaseURL)
	s.Stop()

	var unreachable *url.Error
	switch {
	case errors.As(err, &unreachable):
		return
	case err != nil:
		statuser.Warning("Failed to check for update, skipping:", err)
		return
	}
	if newer(latest, Version) {
		fmt.Println("\nPLEASE UPDATE TO NEW VERSION", latest)
		fmt.Println("Get update from", repoURL, "or with your package manager")
		fmt.Println()
	}
}

// latestVersion asks the GitHub releases API at releaseURL for the latest
// release and returns its tag (such as "v1.2.3").
//
// It returns a *url.Error if the server can't be reached or doesn't answer
// in time, and another error if it answers with a status other than 200 or
// with a body that has no tag.
func latestVersion(client *http.Client, releaseURL string) (string, error) {
	resp, err := client.Get(releaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck // nothing to do if closing fails

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("invalid answer from GitHub: %w", err)
	}
	if release.TagName == "" {
		return "", errors.New("GitHub's latest release has no tag")
	}
	return release.TagName, nil
}
