package scheduler

type ByteRange struct {
	Start int64
	End   int64
}

func Split(totalSize int64, chunks int, minFileSize int64) (map[int]ByteRange, error) {
	ranges := make(map[int]ByteRange)

	if totalSize <= minFileSize {
		ranges[1] = ByteRange{
			Start: 0,
			End:   totalSize - 1,
		}
		return ranges, nil
	}

	base := totalSize / int64(chunks)

	start := int64(0)

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
