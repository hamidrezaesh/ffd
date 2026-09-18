package disk

import (
	"path/filepath"
	"testing"
)

func TestFreeSpace(t *testing.T) {
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	free, err := FreeSpace(dir)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Logf("Free Space: %v bytes", free)
}
