// Package version holds build information injected at link time via ldflags.
package version

import "fmt"

// These values are overridden at build time by the Makefile:
//
//	-X github.com/fabianoflorentino/srectl/internal/version.Version=$(VERSION)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String renders the full build information in a single line.
func String() string {
	return fmt.Sprintf("srectl %s (commit %s, built %s)", Version, Commit, Date)
}
