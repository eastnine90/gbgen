package buildinfo

import "runtime/debug"

// These can be set via -ldflags at build time.
// Example:
//
//	go build -ldflags "-X github.com/eastnine90/gbgen/internal/buildinfo.Version=v0.1.0"
var (
	Version    = "dev"
	Commit     = ""
	CommitDate = "" // unix seconds (string) or RFC3339; depends on your build pipeline
	TreeState  = "" // e.g. "clean" or "dirty"
)

func init() {
	// Prefer explicit -ldflags overrides if present.
	if Version != "" && Version != "dev" {
		return
	}

	// Fall back to Go build info.
	if info, ok := debug.ReadBuildInfo(); ok {
		// info.Main.Version is the module version used to build the binary.
		// For go install ...@vX.Y.Z this is vX.Y.Z.
		v := info.Main.Version

		Version = v

		// Populate commit/date if present.
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				Commit = s.Value
			case "vcs.time":
				CommitDate = s.Value
			}
		}
	}
}
