package main

import "testing"

// The bump follows the biggest change since the last tag: a breaking change
// bumps the major version, a feature the minor, a fix the patch, and anything
// else releases nothing.
func TestNextVersion(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     string
	}{
		{"no commits", nil, ""},
		{"docs, ci and chore only", []string{"docs: fix a typo", "ci: cache Go", "chore: tidy"}, ""},
		{"Dependabot subject", []string{"Bump golang.org/x/sys from 0.49.0 to 0.50.0"}, ""},
		{"fix", []string{"docs: x", "fix: keep the terminal open"}, "v1.2.4"},
		{"perf", []string{"perf: list windows once"}, "v1.2.4"},
		{"scoped fix", []string{"fix(linux): find xterm windows"}, "v1.2.4"},
		{"feat beats fix", []string{"fix: a", "feat: add --quiet"}, "v1.3.0"},
		{"scoped feat", []string{"feat(macos): skip Finder"}, "v1.3.0"},
		{"breaking with !", []string{"feat: a", "feat!: drop the old config"}, "v2.0.0"},
		{"scoped breaking with !", []string{"refactor(config)!: rename keys"}, "v2.0.0"},
		{"breaking footer", []string{"fix: a\n\nBREAKING CHANGE: the config moved"}, "v2.0.0"},
		{"breaking footer with hyphen", []string{"feat: a\n\nBREAKING-CHANGE: x"}, "v2.0.0"},
		{"feat only in the body", []string{"docs: list changes\n\nfeat: not a subject"}, ""},
		{"type needs a colon", []string{"feature request template"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextVersion("v1.2.3", tt.messages)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("nextVersion(v1.2.3, %q) = %q, want %q", tt.messages, got, tt.want)
			}
		})
	}
}

func TestNextVersionRejectsOddTags(t *testing.T) {
	for _, tag := range []string{"", "1.2.3", "v1.2", "v1.2.x", "v1.2.3-rc1"} {
		if _, err := nextVersion(tag, []string{"fix: a"}); err == nil {
			t.Errorf("nextVersion(%q) gave no error", tag)
		}
	}
}
