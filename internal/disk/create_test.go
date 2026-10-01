package disk

import (
	"testing"
	"os"
)

func TestAllocate(t *testing.T){
	f := FileInfo {
		Filename: "example.test",
		Path: "example.test",
		TotalSize: int64(1200000),
	}

	file, err := Allocate(f)
	if err != nil{
		t.Fatalf("Error: %v", err)
	}

	i, _ := os.Stat("example.test")
	size := i.Size()

	t.Logf("\nFilename: %v\nPath: %v\nExpectedSize: %v\nFile: %v\nFinalSize: %v", f.Filename, f.Path, f.TotalSize, file, size)

	_ = os.Remove("example.test")
	t.Log("example file removed\n")
}
