package version

import (
	"strconv"
	"strings"
)

// newer reports whether release version latest is newer than local, comparing
// "vMAJOR.MINOR.PATCH" as numbers. A prerelease (such as "v1.2.0-rc1") is
// older than the same version without a suffix. If either version can't be
// read, it returns false, so an odd version never triggers an update notice.
func newer(latest, local string) bool {
	l, lPre, ok := parseVersion(latest)
	if !ok {
		return false
	}
	c, cPre, ok := parseVersion(local)
	if !ok {
		return false
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	// Same numbers: only a final release is newer than its prerelease.
	return !lPre && cPre
}

// parseVersion reads "v1.2.3" or "1.2.3", optionally followed by a
// prerelease suffix such as "-rc1". ok is false if it isn't a version.
func parseVersion(v string) (numbers [3]int, prerelease bool, ok bool) {
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexByte(v, '-'); i >= 0 {
		v, prerelease = v[:i], true
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return numbers, false, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return numbers, false, false
		}
		numbers[i] = n
	}
	return numbers, prerelease, true
}
