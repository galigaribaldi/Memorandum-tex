package contract

import (
	"strconv"
	"strings"
)

// Semver is a MAJOR.MINOR.PATCH version (§2: no prerelease, no metadata).
type Semver struct {
	Major, Minor, Patch int
}

// ParseSemver parses s; ok is false if s is not valid semver.
func ParseSemver(s string) (v Semver, ok bool) {
	if !ValidSemver(s) {
		return Semver{}, false
	}
	parts := strings.Split(s, ".")
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil { // only on overflow: the pattern allows digits only
			return Semver{}, false
		}
		nums[i] = n
	}
	return Semver{nums[0], nums[1], nums[2]}, true
}

// Compare returns -1, 0 or +1 as v is older than, equal to or newer than w.
func (v Semver) Compare(w Semver) int {
	for _, d := range [3]int{v.Major - w.Major, v.Minor - w.Minor, v.Patch - w.Patch} {
		switch {
		case d < 0:
			return -1
		case d > 0:
			return 1
		}
	}
	return 0
}

// String formats the version.
func (v Semver) String() string {
	return strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
}

// MinimumCreativeInk parses requires.creativeInk (">=X.Y.Z"); ok is false
// when the constraint is malformed.
func MinimumCreativeInk(constraint string) (min Semver, ok bool) {
	if !reCreativeInk.MatchString(constraint) {
		return Semver{}, false
	}
	return ParseSemver(strings.TrimPrefix(constraint, ">="))
}
