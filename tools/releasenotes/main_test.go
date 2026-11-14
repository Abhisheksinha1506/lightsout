package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

const changelog = `# Changelog

Intro text.

## [Unreleased]

## [1.0.30] - 2026-12-01

### Fixed

- A later fix.

## [1.0.2] - 2026-11-14

### Added

- Choice 0.

### Fixed

- No prompt when every app is kept.

## [1.0.0] - 2026-11-14

First release.

[Unreleased]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.2...HEAD
[1.0.2]: https://github.com/abhisheksinha1506/lightsout/compare/v1.0.0...v1.0.2
`

// The notes are the version's section, the install steps and, when the
// previous tag is known, a link to the full diff.
func TestReleaseNotes(t *testing.T) {
	got, err := releaseNotes(changelog, "v1.0.2", "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	wantSection := "### Added\n\n- Choice 0.\n\n### Fixed\n\n- No prompt when every app is kept.\n"
	if !strings.HasPrefix(got, wantSection) {
		t.Errorf("notes do not start with the 1.0.2 section:\n%s", got)
	}
	for _, want := range []string{
		"brew install abhisheksinha1506/tap/lightsout",
		"**Full changelog:** https://github.com/abhisheksinha1506/lightsout/compare/v1.0.0...v1.0.2\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("notes missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"## [", "A later fix", "First release"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("notes contain %q from another section:\n%s", unwanted, got)
		}
	}
}

// The last section is followed by the link definitions, which are not notes.
func TestReleaseNotesLastSection(t *testing.T) {
	got, err := releaseNotes(changelog, "v1.0.0", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "First release.\n") {
		t.Errorf("notes do not start with the 1.0.0 section:\n%s", got)
	}
	if strings.Contains(got, "[Unreleased]:") || strings.Contains(got, "[1.0.2]:") {
		t.Errorf("notes contain link definitions:\n%s", got)
	}
	if strings.Contains(got, "Full changelog") {
		t.Errorf("notes link a diff although no previous tag was given:\n%s", got)
	}
}

// A release must not go out without notes.
func TestReleaseNotesErrors(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{name: "version not in changelog", tag: "v9.9.9"},
		{name: "only a longer version matches the prefix", tag: "v1.0.3"},
		{name: "section is empty", tag: "Unreleased"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := releaseNotes(changelog, tt.tag, ""); err == nil {
				t.Fatalf("releaseNotes(%q) = %q, want an error", tt.tag, got)
			}
		})
	}
}

// run reads CHANGELOG.md and prints the notes, and rejects bad arguments or
// an unreadable changelog.
func TestRun(t *testing.T) {
	notes := "# Changelog\n\n## [1.2.0] - 2026-01-01\n\n- New thing.\n"
	read := func(string) ([]byte, error) { return []byte(notes), nil }

	var out strings.Builder
	if err := run([]string{"v1.2.0", "v1.1.0"}, read, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "- New thing.") || !strings.Contains(out.String(), "v1.1.0...v1.2.0") {
		t.Errorf("run() output:\n%s", out.String())
	}

	for _, args := range [][]string{nil, {"a", "b", "c"}} {
		if err := run(args, read, io.Discard); err == nil {
			t.Errorf("run(%q) = nil error, want usage error", args)
		}
	}
	failRead := func(string) ([]byte, error) { return nil, errors.New("no file") }
	if err := run([]string{"v1.2.0"}, failRead, io.Discard); err == nil {
		t.Error("run() with unreadable changelog = nil error")
	}
	if err := run([]string{"v9.9.9"}, read, io.Discard); err == nil {
		t.Error("run() for a missing version = nil error")
	}
}
