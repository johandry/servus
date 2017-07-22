package version_test

import (
	"testing"

	"github.com/johandry/servus/version"
)

func TestVersionString(t *testing.T) {
	version.Version = "1.0"
	version.VersionPrerelease = "dev"
	version.GitCommit = "ab12cd3"
	actualVer := version.String()

	expectedVer := "Servus v1.0-dev (ab12cd3)"

	if actualVer != expectedVer {
		t.Fatalf("Expected %s, but got %s", expectedVer, actualVer)
	}

}
