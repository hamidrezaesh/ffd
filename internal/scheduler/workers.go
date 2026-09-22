package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

func clamp(value, minValue, maxValue int64) int64 {
	if value < minValue {
		return minValue
	}

	if value > maxValue {
		return maxValue
	}

	return value
}

func nWorkers(
	url string,
	workers int,
	startByte int64,
	endByte int64,
	client *HeaderClient,
	progress *tracker.Progress,
) (<-chan Chunk, <-chan error) {
	out := make(chan Chunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if workers <= 0 {
			errCh <- fmt.Errorf("workers must be greater than 0")
			return
		}

		if startByte < 0 || startByte > endByte {
			errCh <- fmt.Errorf("invalid byte range")
			return
		}

		if client == nil {
			client = &HeaderClient{
				Base: http.DefaultClient,
			}
		}

		totalSize := endByte - startByte + 1

		if int64(workers) > totalSize {
			workers = int(totalSize)
		}

		var wg sync.WaitGroup
		var errOnce sync.Once

		sendError := func(err error) {
			errOnce.Do(func() {
				errCh <- err
			})
		}

		baseSize := totalSize / int64(workers)
		remainder := totalSize % int64(workers)

		nextStart := startByte

		for i := 0; i < workers; i++ {
			workerSize := baseSize

			if int64(i) < remainder {
				workerSize++
			}

			workerStart := nextStart
			workerEnd := workerStart + workerSize - 1

			nextStart = workerEnd + 1

			wg.Add(1)

			go func(start, end int64) {
				defer wg.Done()

				req, err := http.NewRequest(
					http.MethodGet,
					url,
					nil,
				)
				if err != nil {
					sendError(err)
					return
				}

				req.Header.Set(
					"Range",
					"bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10),
				)

				resp, err := client.Do(req)
				if err != nil {
					sendError(err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusPartialContent {
					sendError(fmt.Errorf(
						"server returned status %d, expected %d",
						resp.StatusCode,
						http.StatusPartialContent,
					))
					return
				}

				buf := make([]byte, 32*1024)
				offset := start

				for {
					n, readErr := resp.Body.Read(buf)

					if n > 0 {
						data := make([]byte, n)
						copy(data, buf[:n])

						out <- Chunk{
							Bytes:  data,
							Offset: offset,
						}

						if progress != nil {
							progress.AddDownloaded(int64(n))
						}

						offset += int64(n)
					}

					if readErr != nil {
						if readErr != io.EOF {
							sendError(readErr)
						}

						return
					}
				}
			}(workerStart, workerEnd)
		}

		wg.Wait()
	}()

	return out, errCh
}

func (s *Scheduler) testWorkers(
	startByte int64,
	emit func(Chunk),
) (int, int64, error) {
	workerCounts := []int{
		8,
		16,
		32,
		64,
	}

	testSize := clamp(
		s.TotalSize/50,
		2*1024*1024,
		200*1024*1024,
	)

	var bestWorkers int
	var bestSpeed float64

	for _, workerCount := range workerCounts {
		if startByte >= s.TotalSize {
			break
		}

		testEnd := startByte + testSize

		if testEnd > s.TotalSize {
			testEnd = s.TotalSize
		}

		ranges, err := Split(
			startByte,
			testEnd-1,
			1,
			0,
		)
		if err != nil {
			return 0, startByte, err
		}

		testRange := ranges[0]

		startTime := time.Now()

		chunks, errors := nWorkers(
			s.URL,
			workerCount,
			testRange.Start,
			testRange.End,
			s.Client,
			s.Progress,
		)

		for chunks != nil || errors != nil {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					chunks = nil
					continue
				}

				if emit != nil {
					emit(chunk)
				}

			case err, ok := <-errors:
				if !ok {
					errors = nil
					continue
				}

				if err != nil {
					return 0, startByte, err
				}
			}
		}

		elapsed := time.Since(startTime).Seconds()

		if elapsed <= 0 {
			elapsed = 0.000001
		}

		speed := float64(testRange.End-testRange.Start+1) / elapsed

		if speed > bestSpeed {
			bestSpeed = speed
			bestWorkers = workerCount
		}

		startByte = testRange.End + 1
	}

	if bestWorkers == 0 {
		bestWorkers = 1
	}

	return bestWorkers, startByte, nil
}
