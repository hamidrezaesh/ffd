//go:build linux

package disk

import (
	"os"

	"golang.org/x/sys/unix"
)

func Allocate(file *os.File, offset int64, size int64) error {
	return unix.Fallocate(
		int(file.Fd()),
		0,
		offset,
		size,
		)
}