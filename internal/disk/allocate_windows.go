//go:build windows

package disk

import (
	"fmt"
	"math"
	"os"
)

/*
Allocate validates the requested range.
Windows does not provide the same arbitrary-range allocation
semantics as POSIX through FileAllocationInfo. Let WriteAt
allocate storage as chunks are written.
*/

func Allocate(file *os.File, offset, size int64) error {
	if file == nil {
		return fmt.Errorf("nil file")
	}

	if offset < 0 || size < 0 || offset > math.MaxInt64-size {
		return fmt.Errorf("invalid allocation range")
	}

	return nil
}
