package scheduler

import (
	"net/http"
	"sync"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

func Download(
	url string,
	totalSize int64,
	acceptRanges bool,
	minFileSize int64,
	maxWorkers int,
	maxChunks int,
	client *http.Client,
	progress *tracker.Progress,
	maxRetries int,
) (<-chan Chunk, <-chan error) {

	out := make(chan Chunk)
	workerChunks := make(chan Chunk, maxWorkers*2)
	errCh := make(chan error, 1)

	var ranges map[int]ByteRange
	var err error

	// Decide ranges.
	if !acceptRanges {
		ranges = map[int]ByteRange{
			1: {
				Start: 0,
				End:   totalSize - 1,
			},
		}
	} else {
		ranges, err = Split(
			totalSize,
			maxChunks,
			minFileSize,
		)
		if err != nil {
			close(out)
			close(errCh)
			errCh <- err
			return out, errCh
		}
	}

	jobs := make(chan ByteRange)

	var workers sync.WaitGroup

	// Start workers.
	workerCount := maxWorkers

	if workerCount > len(ranges) {
		workerCount = len(ranges)
	}

	for i := 0; i < workerCount; i++ {
		workers.Add(1)

		go func() {
			defer workers.Done()

			for job := range jobs {
				t := Task{
					URL:    url,
					Range:  job,
					Index:  i,
					Client: client,
				}

				if workerErr := Worker(
					t,
					progress,
					maxRetries,
					workerChunks,
				); workerErr != nil {
					select {
					case errCh <- workerErr:
					default:
					}
				}
			}
		}()
	}

	// Send jobs.
	go func() {
		defer close(jobs)

		for _, r := range ranges {
			jobs <- r
		}
	}()

	// Wait for workers, then close their output.
	go func() {
		workers.Wait()
		close(workerChunks)
	}()

	// Forward chunks.
	go func() {
		defer close(out)
		defer close(errCh)

		for chunk := range workerChunks {
			out <- chunk
		}
	}()

	return out, errCh
}
