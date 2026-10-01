//go:build windows

/*
Windows does not support the same range-based allocation strategy
used on POSIX systems. We therefore download each range into a
temporary file, then merge the ranges into the final file in order.
*/

package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
)

type download struct {
	fileInfo disk.FileInfo
	files    map[int64]*os.File
	paths    map[int64]string
}

func newDownload(fileInfo disk.FileInfo) (*download, error) {
	return &download{
		fileInfo: fileInfo,
		files:    make(map[int64]*os.File),
		paths:    make(map[int64]string),
	}, nil
}

func (d *download) Write(chunk scheduler.Chunk) error {
	partDir := d.fileInfo.Path

	if _, err := os.Stat(partDir); os.IsNotExist(err) {
		if err := os.MkdirAll(partDir, 0755); err != nil {
			return err
		}
	}

	partFile, exists := d.files[chunk.RangeStart]

	if !exists {
		partPath := filepath.Join(
			partDir,
			fmt.Sprintf(
				"part-%d.bin",
				chunk.RangeStart,
			),
		)

		partFile, err := disk.AllocateChunk(
			partPath,
			int64(len(chunk.Bytes)),
		)
		if err != nil {
			return err
		}

		d.files[chunk.RangeStart] = partFile
		d.paths[chunk.RangeStart] = partPath
	}

	_, err := partFile.Write(chunk.Bytes)
	return err
}

func (d *download) Close() error {
	for _, file := range d.files {
		if err := file.Close(); err != nil {
			return err
		}
	}

	rangeStarts := make([]int64, 0, len(d.paths))

	for rangeStart := range d.paths {
		rangeStarts = append(rangeStarts, rangeStart)
	}

	sort.Slice(
		rangeStarts,
		func(i, j int) bool {
			return rangeStarts[i] < rangeStarts[j]
		},
	)

	for _, rangeStart := range rangeStarts {
		if err := disk.Append(
			d.fileInfo.Path,
			d.paths[rangeStart],
		); err != nil {
			return err
		}
	}

	return os.Remove(d.fileInfo.Path)
}
