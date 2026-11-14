package version

import "testing"

// Only a genuinely newer release may trigger "PLEASE UPDATE". Versions compare
// as numbers, not text, and a prerelease is older than its final release.
func TestNewer(t *testing.T) {
	tests := []struct {
		latest, local string
		want          bool
	}{
		{"v1.1.1", "v1.1.0", true},
		{"v1.2.0", "v1.1.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.10.0", "v1.9.0", true}, // as text, "v1.10.0" < "v1.9.0"
		{"v1.1.0", "v1.1.0", false},
		{"v1.1.0", "v1.1.1", false},     // local is newer, e.g. testing a fix
		{"v1.1.0", "v1.2.0-rc1", false}, // local prerelease of a newer version
		{"v1.2.0", "v1.2.0-rc1", true},  // the final release of a local prerelease
		{"1.1.1", "v1.1.0", true},       // a missing "v" is fine
		{"", "v1.1.0", false},           // no answer from GitHub
		{"latest", "v1.1.0", false},     // unreadable versions never nag
		{"v1.1.1", "dev", false},
	}
	for _, tt := range tests {
		if got := newer(tt.latest, tt.local); got != tt.want {
			t.Errorf("newer(%q, %q) = %v, want %v", tt.latest, tt.local, got, tt.want)
		}
	}
}
