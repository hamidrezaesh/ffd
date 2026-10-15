//go:build windows

/*
Windows does not support the same range-based allocation strategy
used on POSIX systems. We attempt whole-file allocation on the first
write, then write chunks directly to their final offsets.
*/

package engine

import (
	"io"
	"os"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
)

type download struct {
	file     *os.File
	fileSize int64

	allocateAvailable  bool
	wholeFileAllocated bool
}

func newDownload(fileInfo disk.FileInfo) (*download, error) {
	file, err := disk.Create(fileInfo)
	if err != nil {
		return nil, err
	}

	return &download{
		file:              file,
		fileSize:          fileInfo.TotalSize,
		allocateAvailable: true,
	}, nil
}

func (d *download) Write(chunk scheduler.Chunk) error {
	n, err := d.file.WriteAt(chunk.Bytes, chunk.Offset)
	if err != nil {
		return err
	}

	if n != len(chunk.Bytes) {
		return io.ErrShortWrite
	}

	return nil
}

func (d *download) Close() error {
	return d.file.Close()
}
