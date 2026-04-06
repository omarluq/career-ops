// Package vinfo holds build-time version information injected via ldflags.
package vinfo

import "fmt"

// Version, Commit, and BuildDate are set at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// String returns a formatted version string.
func String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
}
