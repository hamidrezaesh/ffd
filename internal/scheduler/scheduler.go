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
	Client       *HeaderClient
	MaxRetries   int
	MinFileSize  int64
	Progress     *tracker.Progress
	URL          string
	TotalSize    int64
	MaxWorkers   int
	MaxChunks    int
	ResumeChunks []ResumeChunk
}

/*
fetchChunks downloads either saved resume chunks or newly split ranges.
*/
func (s *Scheduler) fetchChunks(
	startByte int64,
) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		var ranges []ResumeChunk

		// Use saved ranges when resuming.
		if len(s.ResumeChunks) > 0 {
			ranges = s.ResumeChunks
		} else {
			if startByte >= s.TotalSize {
				return
			}

			splitRanges, err := Split(
				startByte,
				s.TotalSize-1,
				s.MaxChunks,
				s.MinFileSize,
			)
			if err != nil {
				errCh <- err
				return
			}

			ranges = make([]ResumeChunk, 0, len(splitRanges))

			for i, r := range splitRanges {
				ranges = append(ranges, ResumeChunk{
					Index: i,
					Range: r,
				})
			}
		}

		// Remove fully downloaded chunks.
		pending := make([]ResumeChunk, 0, len(ranges))

		for _, resume := range ranges {
			rangeSize := resume.Range.End - resume.Range.Start + 1

			if resume.Downloaded >= rangeSize {
				continue
			}

			pending = append(pending, resume)
		}

		if len(pending) == 0 {
			return
		}

		jobs := make(chan ResumeChunk)

		var wg sync.WaitGroup

		workerCount := s.MaxWorkers
		if workerCount > len(pending) {
			workerCount = len(pending)
		}

		if workerCount <= 0 {
			errCh <- fmt.Errorf("no workers available")
			return
		}

		workerChunks := make(chan Chunk, workerCount*2)

		// Start workers.
		for i := 0; i < workerCount; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for resume := range jobs {
					task := Task{
						URL:    s.URL,
						Range:  resume.Range,
						Index:  resume.Index,
						Client: s.Client,
					}

					if err := Worker(
						task,
						resume,
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
			}()
		}

		// Feed jobs.
		go func() {
			defer close(jobs)

			for _, resume := range pending {
				jobs <- resume
			}
		}()

		// Wait for workers and close their output.
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
It uses testProtocol, testWorkers and fetchChunks to download the file.
*/
func Download(
	url string,
	totalSize int64,
	proxyServer *url.URL,
	headers Headers,
	jar http.CookieJar,
	progress *tracker.Progress,
	acceptRange bool,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
	preferredProtocol int,
	resumeChunks []ResumeChunk,
) (<-chan Chunk, <-chan error) {
	scheduler := &Scheduler{
		MaxRetries:   maxRetries,
		MinFileSize:  5 * 1024 * 1024,
		Progress:     progress,
		URL:          url,
		TotalSize:    totalSize,
		MaxWorkers:   maxWorkers,
		MaxChunks:    maxChunks,
		ResumeChunks: resumeChunks,
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

		// If the server does not support ranges.
		if !acceptRange {
			baseClient := &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        4,
					MaxIdleConnsPerHost: 2,
					MaxConnsPerHost:     2,
					IdleConnTimeout:     90 * time.Second,
					ForceAttemptHTTP2:   true,
				},
				Jar: jar,
			}

			if proxyServer != nil {
				baseClient.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyServer)
			}

			scheduler.Client = &HeaderClient{
				Base:    baseClient,
				Headers: headers,
			}

			// Non-range downloads cannot resume individual chunks.
			resume := ResumeChunk{
				Index: 0,
				Range: ByteRange{
					Start: 0,
					End:   scheduler.TotalSize - 1,
				},
			}

			// If there is saved progress, start after it.
			if len(resumeChunks) > 0 {
				resume = resumeChunks[0]
			}

			task := Task{
				URL:    scheduler.URL,
				Range:  resume.Range,
				Index:  resume.Index,
				Client: scheduler.Client,
			}

			workerChunks := make(chan Chunk)

			go func() {
				defer close(workerChunks)

				if err := Worker(
					task,
					resume,
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

		/*
			Resume existing chunks directly.

			Their ranges are already known, so there is no reason
			to run protocol or worker tests again.
		*/
		if len(scheduler.ResumeChunks) > 0 {
			scheduler.Client = newClient(
				preferredProtocol,
				proxyServer,
				headers,
				jar,
			)

			// If no protocol was explicitly selected, use the
			// normal HTTP/2 client rather than testing again.
			if preferredProtocol == 0 {
				scheduler.Client = newClient(
					2,
					proxyServer,
					headers,
					jar,
				)
			}

			if scheduler.MaxWorkers <= 0 {
				scheduler.MaxWorkers = 1
			}

			chunks, errors := scheduler.fetchChunks(0)

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

			return
		}

		startByte := int64(0)
		var protocol int

		// Detect best protocol if it isn't specified by the user.
		if preferredProtocol == 0 {
			finalProtocol, nextByte, err := scheduler.testProtocol(
				proxyServer,
				jar,
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
		scheduler.Client = newClient(
			protocol,
			proxyServer,
			headers,
			jar,
		)

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

		chunks, errors := scheduler.fetchChunks(startByte)

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
