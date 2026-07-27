package scheduler

import (
	"net/http"
	"os"
	"sync"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

func Download(
	url string,
	file *os.File,
	totalSize int64,
	acceptRanges bool,
	minFileSize int64,
	maxWorkers int,
	client *http.Client,
	progress *tracker.Progress,
	maxRetires int,
) error {

	var (
		ranges map[int]ByteRange
		wg     sync.WaitGroup
		errCh  chan error
		err    error
	)

	// Decide ranges
	if !acceptRanges {
		ranges = map[int]ByteRange{
			1: {
				Start: 0,
				End:   totalSize - 1,
			},
		}
	} else {
		ranges, err = Split(totalSize, maxWorkers, minFileSize)
		if err != nil {
			return err
		}
	}

	errCh = make(chan error, len(ranges))

	for _, r := range ranges {
		wg.Add(1)

		go func(r ByteRange) {
			defer wg.Done()

			task := Task{
				URL:    url,
				Range:  r,
				File:   file,
				Client: client,
			}

			if err := Worker(task, progress, maxRetires); err != nil {
				errCh <- err
			}
		}(r)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
