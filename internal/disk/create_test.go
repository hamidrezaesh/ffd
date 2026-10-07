//go:build !windows

package disk

import (
	"os"
	"testing"
)

func TestCreateAndAllocate(t *testing.T) {
	tempDir := t.TempDir()

	path := tempDir + "/example.test"

	f := FileInfo{
		Filename:  "example.test",
		Path:      path,
		TotalSize: 1_200_000,
	}

	// Create
	file, err := Create(f)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if file == nil {
		t.Fatal("Create() returned nil file")
	}

	// Check initial size
	info, err := os.Stat(f.Path)
	if err != nil {
		t.Fatalf("failed to stat created file: %v", err)
	}

	// Allocate
	if err := Allocate(file, 0, 100); err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}

	// Check final size
	info, err = os.Stat(f.Path)
	if err != nil {
		t.Fatalf("failed to stat allocated file: %v", err)
	}

	if info.Size() != f.TotalSize {
		t.Errorf(
			"allocated file size = %d, want %d",
			info.Size(),
			f.TotalSize,
		)
	}

	t.Logf(
		"Created and allocated %q: initial size = 0, final size = %d",
		f.Path,
		info.Size(),
	)

	if err := file.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}
}
