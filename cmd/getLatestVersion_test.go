package cmd

import (
	"testing"
)

func TestGetLatestVersion(t *testing.T){
	latestVersion, err := getLatestVersion()
	if err != nil {
		t.Fatalf("\nError: %v\n", err)
	}

	t.Logf("\nLatest Version On Github: %v\n", latestVersion)
}
