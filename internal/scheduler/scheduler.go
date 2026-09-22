package scheduler

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

/*
Scheduler is responsible for choosing the protocol, choosing workers,
and downloading the actual file.
*/
type Scheduler struct {
	Client      *http.Client
	MaxRetries  int
	MinFileSize int64
	Progress    *tracker.Progress
	URL         string
	TotalSize   int64
	MaxWorkers  int
	MaxChunks   int
}

/*
fetchRest downloads the remaining part of the file.
*/
func (s *Scheduler) fetchRest(
	startByte int64,
) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if startByte >= s.TotalSize {
			return
		}

		ranges, err := Split(
			startByte,
			s.TotalSize-1,
			s.MaxChunks,
			s.MinFileSize,
		)

		if err != nil {
			errCh <- err
			return
		}

		jobs := make(chan ByteRange)

		var wg sync.WaitGroup

		workerCount := s.MaxWorkers
		if workerCount > len(ranges) {
			workerCount = len(ranges)
		}

		if workerCount == 0 {
			errCh <- fmt.Errorf("no ranges to download")
			return
		}

		workerChunks := make(chan Chunk, s.MaxWorkers*2)

		// Start workers.
		for i := 0; i < workerCount; i++ {
			wg.Add(1)

			go func(workerID int) {
				defer wg.Done()

				for job := range jobs {
					task := Task{
						URL:    s.URL,
						Range:  job,
						Index:  workerID,
						Client: s.Client,
					}

					if err := Worker(
						task,
						s.Progress,
						s.MaxRetries,
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
Download is the main component of the scheduler.
it uses testProtocol and testWorkers and fetchRest to download the file while downloading it.
*/
func Download(
	url string,
	totalSize int64,
	proxyServer *url.URL,
	progress *tracker.Progress,
	acceptRange bool,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
	preferredProtocol int,
) (<-chan Chunk, <-chan error) {
	scheduler := &Scheduler{
		MaxRetries:  maxRetries,
		MinFileSize: 5 * 1024 * 1024,
		Progress:    progress,
		URL:         url,
		TotalSize:   totalSize,
		MaxWorkers:  maxWorkers,
		MaxChunks:   maxChunks,
	}

	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if scheduler.TotalSize <= 0 {
			errCh <- fmt.Errorf("invalid total size")
			return
		}

		// If server does not support ranges.
		if !acceptRange {
			task := Task{
				URL: scheduler.URL,
				Range: ByteRange{
					Start: 0,
					End:   scheduler.TotalSize - 1,
				},
				Index: 0,
				Client: &http.Client{
					Transport: &http.Transport{
						MaxIdleConns:        4,
						MaxIdleConnsPerHost: 2,
						MaxConnsPerHost:     2,
						IdleConnTimeout:     90 * time.Second,
						ForceAttemptHTTP2:   true,
						Proxy:               http.ProxyURL(proxyServer),
					},
				},
			}

			scheduler.Client = task.Client

			workerChunks := make(chan Chunk)

			go func() {
				defer close(workerChunks)

				if err := Worker(
					task,
					scheduler.Progress,
					scheduler.MaxRetries,
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

		// Detect best protocol if it isn't specified by the user.
		if preferredProtocol == 0 {
			finalProtocol, nextByte, err := scheduler.testProtocol(
				proxyServer,
				startByte,
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

		// Make a client.
		scheduler.Client = newClient(protocol, proxyServer)
		if scheduler.Client == nil {
			scheduler.Client = http.DefaultClient
		}

		// Automatically determine worker count.
		if scheduler.MaxWorkers <= 0 {
			selectedWorkers, nextByte, err := scheduler.testWorkers(
				startByte,
				func(chunk Chunk) {
					out <- chunk
				},
			)

			if err != nil {
				errCh <- err
				return
			}

			scheduler.MaxWorkers = selectedWorkers
			startByte = nextByte
		}

		if scheduler.MaxWorkers <= 0 {
			scheduler.MaxWorkers = 1
		}

		if scheduler.MaxChunks <= 0 {
			scheduler.MaxChunks = scheduler.MaxWorkers * 2
		}

		// Tests downloaded the entire file.
		if startByte >= scheduler.TotalSize {
			return
		}

		chunks, errors := scheduler.fetchRest(startByte)

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
