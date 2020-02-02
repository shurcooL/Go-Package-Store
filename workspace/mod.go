package workspace

import (
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

func parsePseudoVersion(v string) (base, timestamp, rev, build string, ok bool) {
	if !isPseudoVersion(v) {
		return "", "", "", "", false
	}
	build = semver.Build(v)
	v = strings.TrimSuffix(v, build)
	j := strings.LastIndex(v, "-")
	v, rev = v[:j], v[j+1:]
	i := strings.LastIndex(v, "-")
	if j := strings.LastIndex(v, "."); j > i {
		base = v[:j] // "vX.Y.Z-pre.0" or "vX.Y.(Z+1)-0"
		timestamp = v[j+1:]
	} else {
		base = v[:i] // "vX.0.0"
		timestamp = v[i+1:]
	}
	return base, timestamp, rev, build, true
}

// isPseudoVersion reports whether v is a pseudo-version.
func isPseudoVersion(v string) bool {
	return strings.Count(v, "-") >= 2 && semver.IsValid(v) && pseudoVersionRE.MatchString(v)
}

var pseudoVersionRE = regexp.MustCompile(`^v[0-9]+\.(0\.0-|\d+\.\d+-([^+]*\.)?0\.)\d{14}-[A-Za-z0-9]+(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
