package scheduler

type ByteRange struct {
	Start int64
	End   int64
}

func Split(totalSize int64, startByte int64, chunks int, minFileSize int64) (map[int]ByteRange, error) {
	ranges := make(map[int]ByteRange)

	if totalSize <= minFileSize {
		ranges[1] = ByteRange{
			Start: startByte,
			End:   totalSize - 1,
		}
		return ranges, nil
	}

	remaining := totalSize - startByte
	base := remaining / int64(chunks)

	start := startByte

	for i := 1; i <= chunks; i++ {
		end := start + base - 1

		if i == chunks {
			end = totalSize - 1
		}

		ranges[i] = ByteRange{
			Start: start,
			End:   end,
		}

		start = end + 1
	}

	return ranges, nil
}
