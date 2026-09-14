package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

/*
Scheduler is responsible for choose protocol, choose workers
and download the actual file.
*/

/*
After testing protocol and workers count, we use fetchRest to download rest of the file
from the last downloaded byte
*/

func fetchRest(
	url string,
	totalSize int64,
	startByte int64,
	minFileSize int64,
	maxWorkers int,
	maxChunks int,
	client *http.Client,
	progress *tracker.Progress,
	maxRetries int,
) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if startByte >= totalSize {
			return
		}

		if maxWorkers <= 0 {
			maxWorkers = 1
		}

		if maxChunks <= 0 {
			maxChunks = maxWorkers * 2
		}

		ranges, err := Split(
			startByte,
			totalSize-1,
			maxChunks,
			minFileSize,
		)

		if err != nil {
			errCh <- err
			return
		}

		jobs := make(chan ByteRange)

		var wg sync.WaitGroup

		workerCount := maxWorkers

		if workerCount > len(ranges) {
			workerCount = len(ranges)
		}

		if workerCount == 0 {
			errCh <- fmt.Errorf("no ranges to download")
			return
		}

		workerChunks := make(chan Chunk, maxWorkers*2)

		// Start workers.
		for i := 0; i < workerCount; i++ {
			wg.Add(1)

			go func(workerID int) {
				defer wg.Done()

				for job := range jobs {
					task := Task{
						URL:    url,
						Range:  job,
						Index:  workerID,
						Client: client,
					}

					if err := Worker(
						task,
						progress,
						maxRetries,
						workerChunks,
					); err != nil {
						select {
						case errCh <- err:
						default:
						}

						return
					}
				}
			}(i)
		}

		// Feed jobs.
		go func() {
			defer close(jobs)

			for _, r := range ranges {
				jobs <- r
			}
		}()

		// Wait for workers and then close their output.
		go func() {
			wg.Wait()
			close(workerChunks)
		}()

		// Forward worker chunks.
		for chunk := range workerChunks {
			out <- chunk
		}
	}()

	return out, errCh
}

/*
Download is the main component of scheduler. it tests protocol and workers using
testWorkers and testProtocol, then download the rest of file using fetchRest.finally it
returns the downloaded bytes and error (if exists)
*/

func Download(
	url string,
	totalSize int64,
	progress *tracker.Progress,
	acceptRange bool,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
	preferredProtocol int,
) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if totalSize <= 0 {
			errCh <- fmt.Errorf("invalid total size")
			return
		}

		// if server does not support ranges.
		if !acceptRange {
			task := Task{
				URL: url,
				Range: ByteRange{
					Start: 0,
					End:   totalSize - 1,
				},
				Index: 0,
				Client: &http.Client{
					Transport: &http.Transport{
						MaxIdleConns:        4,
						MaxIdleConnsPerHost: 2,
						MaxConnsPerHost:     2,
						IdleConnTimeout:     90 * time.Second,
						ForceAttemptHTTP2:   true,
					},
				},
			}

			workerChunks := make(chan Chunk)

			go func() {
				defer close(workerChunks)

				if err := Worker(
					task,
					progress,
					maxRetries,
					workerChunks,
				); err != nil {
					select {
					case errCh <- err:
					default:
					}
				}
			}()

			for chunk := range workerChunks {
				out <- chunk
			}

			return
		}

		startByte := int64(0)
		var protocol int

		// detect best protocol if it isn't specified by user
		if preferredProtocol == 0 {
			finalProtocol, nextByte, err := testProtocol(
				url,
				totalSize,
				startByte,
				progress,
				func(chunk Chunk) {
					out <- chunk
				},
			)
			if err != nil {
				errCh <- err
				return
			}

			startByte = nextByte
			protocol = finalProtocol
		} else {
			protocol = preferredProtocol
		}

		// make a client
		client := newClient(protocol)
		if client == nil {
			client = http.DefaultClient
		}

		// Automatically determine worker count.
		if maxWorkers <= 0 {
			selectedWorkers, nextByte, err := testWorkers(
				url,
				totalSize,
				startByte,
				client,
				progress,
				func(chunk Chunk) {
					out <- chunk
				},
			)

			if err != nil {
				errCh <- err
				return
			}

			maxWorkers = selectedWorkers
			startByte = nextByte
		}

		if maxWorkers <= 0 {
			maxWorkers = 1
		}

		if maxChunks <= 0 {
			maxChunks = maxWorkers * 2
		}

		// Tests downloaded the entire file.
		if startByte >= totalSize {
			return
		}

		chunks, errors := fetchRest(
			url,
			totalSize,
			startByte,
			5*1024*1024,
			maxWorkers,
			maxChunks,
			client,
			progress,
			maxRetries,
		)

		for chunks != nil || errors != nil {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					chunks = nil
					continue
				}

				out <- chunk

			case err, ok := <-errors:
				if !ok {
					errors = nil
					continue
				}

				if err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return out, errCh
}
