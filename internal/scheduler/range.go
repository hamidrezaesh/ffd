package scheduler

import "fmt"

type ByteRange struct {
	Start int64
	End   int64
}

func Split(
	startByte int64,
	endByte int64,
	chunks int,
	minFileSize int64,
) (map[int]ByteRange, error) {
	ranges := make(map[int]ByteRange)

	if startByte < 0 || endByte < startByte {
		return nil, fmt.Errorf("invalid byte range")
	}

	totalSize := endByte - startByte + 1

	if totalSize <= minFileSize || chunks <= 1 {
		ranges[1] = ByteRange{
			Start: startByte,
			End:   endByte,
		}
		return ranges, nil
	}

	base := totalSize / int64(chunks)
	start := startByte

	for i := 1; i <= chunks; i++ {
		end := start + base - 1

		if i == chunks {
			end = endByte
		}

		ranges[i] = ByteRange{
			Start: start,
			End:   end,
		}

		start = end + 1
	}

	return ranges, nil
}
