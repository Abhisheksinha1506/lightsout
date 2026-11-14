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
	"io"
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
	"On Windows, use Scoop:\n" +
	"\n" +
	"```powershell\n" +
	"scoop bucket add abhisheksinha1506 https://github.com/abhisheksinha1506/scoop-bucket\n" +
	"scoop install abhisheksinha1506/lightsout\n" +
	"```\n" +
	"\n" +
	"On Linux, install the `.deb` or `.rpm` package below. Or download the archive for your system below.\n"

func main() {
	if err := run(os.Args[1:], os.ReadFile, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "releasenotes:", err)
		os.Exit(1)
	}
}

// run prints the release notes for args (<tag> [previous-tag]) to out,
// reading CHANGELOG.md with readFile.
func run(args []string, readFile func(string) ([]byte, error), out io.Writer) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("usage: releasenotes <tag> [previous-tag]")
	}
	previous := ""
	if len(args) == 2 {
		previous = args[1]
	}

	changelog, err := readFile("CHANGELOG.md")
	if err != nil {
		return err
	}
	notes, err := releaseNotes(string(changelog), args[0], previous)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, notes)
	return err
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
