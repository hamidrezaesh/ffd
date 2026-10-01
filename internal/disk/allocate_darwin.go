//go:build darwin

package disk

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func Allocate(f *os.File, offset int64, size int64) error {
	var store unix.Fstore_t

	store.FstFlags = unix.F_ALLOCATECONTIG
	store.FstPosmode = unix.F_PEOFFSET
	store.FstOffset = offset
	store.FstLength = size

	_, err := unix.Fcntl(
		int(f.Fd()),
		unix.F_PREALLOCATE,
		uintptr(unsafe.Pointer(&store)),
	)

	return err
}
