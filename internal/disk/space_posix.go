package disk

import (
	"golang.org/x/sys/unix"
)

func FreeSpace(path string) (int64, error) {
	var stat unix.Statfs_t

	if err := unix.Statfs(path, &stat); err != nil {
		return 0, err
	}

	free := stat.Bavail * uint64(stat.Bsize)
	return int64(free),  nil
}
