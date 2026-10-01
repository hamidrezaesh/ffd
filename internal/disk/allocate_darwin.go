//go:build darwin

package disk

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func Allocate(f *os.File, offset int64, size int64) error {
	store := unix.Fstore_t{
		Flags:   unix.F_ALLOCATECONTIG,
		Posmode: unix.F_PEOFPOSMODE,
		Offset:  offset,
		Length:  size,
	}

	_, err := unix.FcntlInt(
		f.Fd(),
		unix.F_PREALLOCATE,
		int(uintptr(unsafe.Pointer(&store))),
	)

	return err
}
