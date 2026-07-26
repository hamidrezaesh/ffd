//go:build windows

package disk

import "golang.org/x/sys/windows"

func FreeSpace(path string) (int64, error) {
	var freeBytesAvailable uint64

	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	err = windows.GetDiskFreeSpaceEx(
		ptr,
		&freeBytesAvailable,
		nil,
		nil,
	)
	if err != nil {
		return 0, err
	}

	return int64(freeBytesAvailable), nil
}
