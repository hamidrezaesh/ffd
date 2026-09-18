package metadata

import (
	"net/http"
	"testing"
)

func TestFetchMetadata(t *testing.T) {
	test_url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso" // ubuntu iso

	resp, err := http.Get(test_url)
	if err != nil {
		t.Fatalf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected HTTP status: %s", resp.Status)
	}

	metadata, err := FetchMetadata(test_url, resp)
	if err != nil {
		t.Fatalf("FetchMetadata failed: %v", err)
	}

	t.Logf("Metadata: %+v", metadata)
}
