package cmd

import (
	"testing"
)

func TestGetLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GitHub API test in short mode")
	}

	latestVersion, err := getLatestVersion()
	if err != nil {
		t.Fatalf("getLatestVersion() error: %v", err)
	}

	t.Logf("Latest version on GitHub: %v", latestVersion)
}
