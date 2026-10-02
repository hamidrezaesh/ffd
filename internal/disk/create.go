//go:build !windows

package disk

import (
	"errors"
	"os"
	"path/filepath"
)

func Create(f FileInfo) (*os.File, error) {
	dir := filepath.Dir(f.Path)

	// Check if file exists.
	_, err := os.Stat(f.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Check free disk space.
	free, err := FreeSpace(dir)
	if err != nil {
		return nil, err
	}

	if free < f.TotalSize {
		return nil, errors.New("not enough disk space")
	}

	// Create directory if it doesn't exist.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	var file *os.File

	if os.IsNotExist(err) {
		// File doesn't exist.
		file, err = os.OpenFile(
			f.Path,
			os.O_RDWR|os.O_CREATE,
			0644,
		)
	} else {
		// File already exists
		file, err = os.OpenFile(
			f.Path,
			os.O_RDWR,
			0644,
		)
	}

	if err != nil {
		return nil, err
	}

	// Ensure the file has the expected size without destroying existing data.
	if err := file.Truncate(f.TotalSize); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}
