//go:build windows

package disk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsCreateAllocateWriteRead(t *testing.T) {
	const totalSize int64 = 1 << 20 // 1 MiB

	path := filepath.Join(t.TempDir(), "download.bin")

	// 1. Create the destination file.
	file, err := Create(FileInfo{
		Filename:  "download.bin",
		Path:      path,
		TotalSize: totalSize,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	defer file.Close()

	// 2. Allocate the entire file.
	if err := Allocate(file, 0, totalSize); err != nil {
		t.Fatalf("Allocate() error: %v", err)
	}

	// Allocation must not change the logical file size.
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat() after Allocate(): %v", err)
	}
	if info.Size() != totalSize {
		t.Fatalf("file size after Allocate() = %d, want %d",
			info.Size(), totalSize)
	}

	// 3. Repeated allocation must also succeed.
	if err := Allocate(file, 0, totalSize); err != nil {
		t.Fatalf("second Allocate() error: %v", err)
	}

	// 4. Write deterministic test data across the whole file.
	want := make([]byte, int(totalSize))
	for i := range want {
		want[i] = byte(i % 251)
	}

	n, err := file.WriteAt(want, 0)
	if err != nil {
		t.Fatalf("WriteAt() error: %v", err)
	}
	if n != len(want) {
		t.Fatalf("WriteAt() wrote %d bytes, want %d", n, len(want))
	}

	if err := file.Sync(); err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	// 5. Read the file back and verify every byte.
	got := make([]byte, len(want))
	n, err = file.ReadAt(got, 0)
	if err != nil {
		t.Fatalf("ReadAt() error: %v", err)
	}
	if n != len(want) {
		t.Fatalf("ReadAt() read %d bytes, want %d", n, len(want))
	}

	if !bytes.Equal(got, want) {
		t.Fatal("file contents differ from the written data")
	}

	// 6. Verify the final file size.
	info, err = os.Stat(path)
	if err != nil {
		t.Fatalf("final Stat() error: %v", err)
	}
	if info.Size() != totalSize {
		t.Fatalf("final file size = %d, want %d",
			info.Size(), totalSize)
	}
}
