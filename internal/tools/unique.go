package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func UniqueFilename(path, filename, ext string) string {
	candidate := filename + ext

	if _, err := os.Stat(filepath.Join(path, candidate)); os.IsNotExist(err) {
		return filename
	}

	for i := 1; ; i++ {
		candidate = fmt.Sprintf("%s(%d)%s", filename, i, ext)

		_, err := os.Stat(filepath.Join(path, candidate))
		if os.IsNotExist(err) {
			return fmt.Sprintf("%s (%d)", filename, i)
		}
	}
}
