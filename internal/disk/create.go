//go:build !windows

package disk

import (
	"errors"
	"os"
	"path/filepath"
)

func Create(f FileInfo) (*os.File, error) {
	dir := filepath.Dir(f.Path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	free, err := FreeSpace(dir)
	if err != nil {
		return nil, err
	}

	if free < f.TotalSize {
		return nil, errors.New("not enough disk space")
	}

	_, err = os.Stat(f.Path)

	var file *os.File

	if os.IsNotExist(err) {
		file, err = os.OpenFile(
			f.Path,
			os.O_RDWR|os.O_CREATE,
			0644,
		)
	} else if err == nil {
		file, err = os.OpenFile(
			f.Path,
			os.O_RDWR,
			0644,
		)
	} else {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	if err := file.Truncate(f.TotalSize); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}
