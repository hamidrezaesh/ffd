package disk

import (
	"errors"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Filename  string
	Path      string
	TotalSize int64
}

func Allocate(f FileInfo) (*os.File, error) {
	dir := filepath.Dir(f.Path)

	// Check free disk space
	free, err := FreeSpace(dir)
	if err != nil {
		return nil, err
	}

	if free < f.TotalSize {
		return nil, errors.New("not enough disk space")
	}

	// Create directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Create file
	file, err := os.Create(f.Path)
	if err != nil {
		return nil, err
	}

	// Pre-allocate file size
	if err := file.Truncate(f.TotalSize); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}
