// Command nextversion prints the tag of the next release, worked out from the
// Conventional Commits since the latest release tag. The release job runs it
// on every push to main:
//
//	go run ./tools/nextversion
//
// A breaking change ("feat!:", "BREAKING CHANGE:") bumps the major version,
// "feat:" the minor version and "fix:" or "perf:" the patch version. If there
// is nothing to release (only docs, ci, chore and similar commits), it prints
// nothing. It must run in a clone that has the tags.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "nextversion:", err)
		os.Exit(1)
	}
}

// run reads the latest release tag and the commits since it from git, and
// prints the next tag, if any, to stdout.
func run() error {
	// Release tags only: prereleases such as v1.2.0-rc1 are skipped.
	last, err := git("describe", "--tags", "--abbrev=0", "--match", "v[0-9]*", "--exclude", "*-*")
	if err != nil {
		return err
	}
	last = strings.TrimSpace(last)

	// %x00 ends each message, so multi-line bodies stay together.
	log, err := git("log", "--format=%B%x00", last+"..HEAD")
	if err != nil {
		return err
	}
	var messages []string
	for _, m := range strings.Split(log, "\x00") {
		if m = strings.TrimSpace(m); m != "" {
			messages = append(messages, m)
		}
	}

	next, err := nextVersion(last, messages)
	if err != nil {
		return err
	}
	if next != "" {
		fmt.Println(next)
	}
	return nil
}

// git runs git with args and returns its output, with stderr in the error.
func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

var (
	tagPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	// "type(scope)!:" in a subject marks a breaking change.
	breakingSubject = regexp.MustCompile(`^\w+(\([^)]*\))?!:`)
	// A "BREAKING CHANGE:" footer on its own line in the body.
	breakingFooter = regexp.MustCompile(`(?m)^BREAKING[ -]CHANGE:`)
	featSubject    = regexp.MustCompile(`^feat(\([^)]*\))?:`)
	fixSubject     = regexp.MustCompile(`^(fix|perf)(\([^)]*\))?:`)
)

// nextVersion returns the tag that follows last ("vMAJOR.MINOR.PATCH") given
// the full messages of the commits since it, or "" if none of them is a
// breaking change, feature or fix. It returns an error if last is not a
// release tag.
func nextVersion(last string, messages []string) (string, error) {
	m := tagPattern.FindStringSubmatch(last)
	if m == nil {
		return "", fmt.Errorf("latest tag %q is not vMAJOR.MINOR.PATCH", last)
	}
	var major, minor, patch int
	if _, err := fmt.Sscan(m[1]+" "+m[2]+" "+m[3], &major, &minor, &patch); err != nil {
		return "", fmt.Errorf("latest tag %q: %w", last, err)
	}

	const (
		none = iota
		bumpPatch
		bumpMinor
		bumpMajor
	)
	bump := none
	for _, msg := range messages {
		subject, _, _ := strings.Cut(msg, "\n")
		switch {
		case breakingSubject.MatchString(subject) || breakingFooter.MatchString(msg):
			bump = max(bump, bumpMajor)
		case featSubject.MatchString(subject):
			bump = max(bump, bumpMinor)
		case fixSubject.MatchString(subject):
			bump = max(bump, bumpPatch)
		}
	}

	switch bump {
	case bumpMajor:
		return fmt.Sprintf("v%d.0.0", major+1), nil
	case bumpMinor:
		return fmt.Sprintf("v%d.%d.0", major, minor+1), nil
	case bumpPatch:
		return fmt.Sprintf("v%d.%d.%d", major, minor, patch+1), nil
	}
	return "", nil
}
