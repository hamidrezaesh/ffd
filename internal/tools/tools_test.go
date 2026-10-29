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

func TestUrlNormalizer(t *testing.T) {
	urls := []string{"https://localhost:8080/file",
		"127.0.0.1/file",
		"http://localhost/file",
		"https://127.0.0.1/file",
		"192.168.1.1/file",
		"http://192.168.1.1/file",
		"https://192.168.1.1/file",
		"http://example.com/file",
		"https://example.com/file",
		"example.com/file",
	}

	var out string

	for _, url := range urls {
		out, _ = NormalizeURL(url)
		t.Logf("URL: %s -> %s", url, out)
	}
}
