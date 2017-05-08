package version

import (
	"bytes"
	"fmt"
)

var (
	// GitCommit is the git commit that was compiled. This will be filled in by the compiler
	// in the Makefile from Git short SHA-1 of HEAD commit.
	GitCommit string

	// Version is the main version number that is being run at the moment. This will be
	// filled in by the compiler in the Makefile from latest.go
	Version string

	// VersionPrerelease is a pre-release marker for the  If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	// This will be filled in by the compiler in the Makefile from latest.go
	VersionPrerelease string
)

// Println prints the version using the output of String()
func Println() {
	fmt.Println(String())
}

// String return the version as it will be show in the terminal
func String() string {
	var version bytes.Buffer
	fmt.Fprintf(&version, "Servus v%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&version, "-%s", VersionPrerelease)
		if GitCommit != "" {
			fmt.Fprintf(&version, " (%s)", GitCommit)
		}
	}

	return version.String()
}
