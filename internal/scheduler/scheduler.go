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
	maxChunks int,
	client *http.Client,
	progress *tracker.Progress,
	maxRetries int,
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
		ranges, err = Split(totalSize, maxChunks, minFileSize)
		if err != nil {
			return err
		}
	}

	errCh = make(chan error, len(ranges))
	jobs := make(chan ByteRange)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for job := range jobs {
				t := Task{
					URL:    url,
					Range:  job,
					File:   file,
					Client: client,
				}

				if workerErr := Worker(t, progress, maxRetries); workerErr != nil {
					errCh <- workerErr
				}
			}
		}()
	}

	for _, r := range ranges {
		jobs <- r
	}

	close(jobs)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
