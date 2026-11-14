// Command releasenotes prints the GitHub release notes for one tag, taken from
// that version's section in CHANGELOG.md. The release workflow runs it before
// GoReleaser:
//
//	go run ./tools/releasenotes v1.0.4 v1.0.3 > notes.md
//
// The second argument, the previous tag, is optional. It exits with status 1
// if CHANGELOG.md has no notes for the tag, so a release can't go out without
// them.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const repoURL = "https://github.com/abhisheksinha1506/lightsout"

// installSteps is appended to every release's notes.
const installSteps = "### Install or upgrade\n" +
	"\n" +
	"```bash\n" +
	"brew install abhisheksinha1506/tap/lightsout\n" +
	"brew upgrade abhisheksinha1506/tap/lightsout\n" +
	"```\n" +
	"\n" +
	"While the repository is private, export `HOMEBREW_GITHUB_API_TOKEN` (a token\n" +
	"that can read it) before installing or upgrading. On Linux, download the\n" +
	"archive for your architecture below.\n"

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: releasenotes <tag> [previous-tag]")
		os.Exit(2)
	}
	previous := ""
	if len(os.Args) == 3 {
		previous = os.Args[2]
	}

	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		fmt.Fprintln(os.Stderr, "releasenotes:", err)
		os.Exit(1)
	}
	notes, err := releaseNotes(string(changelog), os.Args[1], previous)
	if err != nil {
		fmt.Fprintln(os.Stderr, "releasenotes:", err)
		os.Exit(1)
	}
	fmt.Print(notes)
}

// releaseNotes returns the release notes for tag (for example "v1.0.4"): the
// body of the "## [1.0.4]" section of changelog, the install steps and, if
// previous is not empty, a link to the full diff since that tag.
//
// It returns an error if the section is missing or empty.
func releaseNotes(changelog, tag, previous string) (string, error) {
	version := strings.TrimPrefix(tag, "v")
	section, err := changelogSection(changelog, version)
	if err != nil {
		return "", err
	}

	notes := section + "\n\n" + installSteps
	if previous != "" {
		notes += "\n**Full changelog:** " + repoURL + "/compare/" + previous + "..." + tag + "\n"
	}
	return notes, nil
}

// changelogSection returns the trimmed text under the "## [version]" heading,
// up to the next "## " heading or the link definitions at the end of the file.
func changelogSection(changelog, version string) (string, error) {
	heading := "## [" + version + "]"
	var body []string
	found := false
	for _, line := range strings.Split(changelog, "\n") {
		if !found {
			found = strings.HasPrefix(line, heading)
			continue
		}
		if strings.HasPrefix(line, "## ") || isLinkDefinition(line) {
			break
		}
		body = append(body, line)
	}

	if !found {
		return "", fmt.Errorf("CHANGELOG.md has no %q section", heading)
	}
	section := strings.TrimSpace(strings.Join(body, "\n"))
	if section == "" {
		return "", errors.New("the " + heading + " section of CHANGELOG.md is empty")
	}
	return section, nil
}

// isLinkDefinition reports whether line is a Markdown link definition such as
// "[1.0.2]: https://...", as found at the end of the changelog.
func isLinkDefinition(line string) bool {
	return strings.HasPrefix(line, "[") && strings.Contains(line, "]: ")
}
