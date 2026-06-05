package version

import "runtime"

// Build-time variables (set via -ldflags)
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// Info holds version information.
type Info struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
}

// Get returns the current version info.
func Get() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
	}
}
