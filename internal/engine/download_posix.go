//go:build linux || darwin

/*
POSIX systems support allocating arbitrary ranges of a file.
We create the final file immediately, then allocate each range
as its data arrives and write it directly at its correct offset.
*/

package engine

import (
	"os"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
)

type download struct {
	file *os.File
}

func newDownload(fileInfo disk.FileInfo) (*download, error) {
	file, err := disk.Create(fileInfo)
	if err != nil {
		return nil, err
	}

	return &download{
		file: file,
	}, nil
}

func (d *download) Write(chunk scheduler.Chunk) error {
	if err := disk.Allocate(
		d.file,
		chunk.Offset,
		int64(len(chunk.Bytes)),
	); err != nil {
		return err
	}

	_, err := d.file.WriteAt(
		chunk.Bytes,
		chunk.Offset,
	)

	return err
}

func (d *download) Close() error {
	return d.file.Close()
}
