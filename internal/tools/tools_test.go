package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilename(t *testing.T) {
	testFilename := "check123..<>"
	err := Filename(testFilename)
	t.Logf("Filename error: %v", err)
}

func TestUnique(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "Movie.mkv"), nil, 0600)

	got := UniqueFilename(dir, "Movie", ".mkv")
	t.Logf("Unique filename: %s", got)
}
