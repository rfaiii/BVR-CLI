package version

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

const defaultVersion = "1.2.1"

var (
	Version = defaultVersion
	Commit  = "unknown"
	// BuildID is a unique identifier for this build. For release builds it
	// equals Commit; for development builds (go run / go build without
	// ldflags) it is derived from the executable's modification time, which
	// changes on every recompilation.
	BuildID = ""
)

// A user may install bvr using `go install github.com/richavery/bvr-cli@latest`.
// without -ldflags, in which case the version above is unset. As a workaround
// we use the embedded build version that *is* set when using `go install` (and
// is only set for `go install` and not for `go build`).
func init() {
	// If Version was explicitly overridden via ldflags, keep that value.
	// Otherwise, try to use the Go module version from build info.
	if Version == defaultVersion {
		info, ok := debug.ReadBuildInfo()
		if ok {
			mainVersion := info.Main.Version
			if mainVersion != "" && mainVersion != "(devel)" {
				// Prefer real release versions over Go pseudo-versions like
				// v1.1.3-0.20260814083926-cb74185e261b+dirty.
				if i := strings.Index(mainVersion, "-"); i > 0 {
					if len(mainVersion) > i+1 && mainVersion[i+1] >= '0' && mainVersion[i+1] <= '9' {
						// Pseudo-version detected; keep defaultVersion.
					} else {
						Version = mainVersion
					}
				} else {
					Version = mainVersion
				}
			}
		}
	}
	// Beta/dev tags (e.g. "1.1.6:beta_v2") must win over VCS-stamped build
	// info (e.g. "v1.1.6+dirty") so a dev build always identifies itself
	// with its explicit test tag.
	if defaultVersion != "" && strings.Contains(defaultVersion, ":") && !strings.Contains(Version, ":") {
		Version = defaultVersion
	}

	// Derive BuildID when not set via ldflags.
	if BuildID == "" {
		BuildID = deriveBuildID()
	}
}

// deriveBuildID uses the running executable's modification time as a unique
// build fingerprint. This changes on every recompilation (including `go run`),
// making it reliable for detecting stale servers during development.
func deriveBuildID() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	fi, err := os.Stat(exe)
	if err != nil {
		return "unknown"
	}
	return strconv.FormatInt(fi.ModTime().UnixNano(), 36)
}

// ShortVersion returns a display-friendly version string. For Go
// pseudo-versions (e.g. v0.87.1-0.20260731174531-4d...), it returns just the
// semver prefix (v0.87.1). For prereleases like v1.0.0-alpha.1 it returns
// them unchanged. For "devel" and other short strings it returns them
// unchanged.
func ShortVersion() string {
	v := Version
	if i := strings.Index(v, "-"); i > 0 {
		// Go pseudo-versions look like vX.Y.Z-N.timestamp-commit where N is
		// a digit. Prereleases look like vX.Y.Z-alpha.1 where the part after
		// the hyphen starts with a letter.
		if len(v) > i+1 && v[i+1] >= '0' && v[i+1] <= '9' {
			return v[:i]
		}
	}
	return v
}
