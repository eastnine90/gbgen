package buildinfo

// These can be set via -ldflags at build time.
// Example:
//
//	go build -ldflags "-X github.com/eastnine90/gbgen/internal/buildinfo.Version=v0.1.0"
var (
	Version = "dev"
)
