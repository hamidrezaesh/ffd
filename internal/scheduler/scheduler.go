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
	client *http.Client,
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
			client = http.DefaultClient
		}

		totalSize := endByte - startByte + 1

		if int64(workers) > totalSize {
			workers = int(totalSize)
		}

		var wg sync.WaitGroup
		var errOnce sync.Once

		sendError := func(err error) {
			if err == nil {
				return
			}

			errOnce.Do(func() {
				errCh <- err
			})
		}

		// Divide the requested range between workers.
		base := totalSize / int64(workers)
		remainder := totalSize % int64(workers)

		nextStart := startByte

		for i := 0; i < workers; i++ {
			workerSize := base

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
					"bytes="+
						strconv.FormatInt(start, 10)+
						"-"+
						strconv.FormatInt(end, 10),
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

/*
testWorkers test multiple workers' speed while downloading the file and return the best workers for download
*/

func testWorkers(
	url string,
	totalSize int64,
	client *http.Client,
	progress *tracker.Progress,
	emit func(Chunk),
) (int, int64, error) {
	if totalSize <= 0 {
		return 0, 0, fmt.Errorf("invalid total size")
	}

	const minTestSize int64 = 2 * 1024 * 1024
	const maxTestSize int64 = 200 * 1024 * 1024

	testWorkerCounts := []int{
		8,
		16,
		32,
		64,
	}

	testSize := clamp(
		totalSize/50,
		minTestSize,
		maxTestSize,
	)

	// Make sure all tests fit inside the file.
	if testSize*int64(len(testWorkerCounts)) > totalSize {
		testSize = totalSize / int64(len(testWorkerCounts))
	}

	if testSize <= 0 {
		return 1, 0, nil
	}

	speeds := make([]float64, 0, len(testWorkerCounts))

	startByte := int64(0)

	for _, workerCount := range testWorkerCounts {
		if startByte >= totalSize {
			break
		}

		endByte := startByte + testSize - 1

		if endByte >= totalSize {
			endByte = totalSize - 1
		}

		testStart := time.Now()

		chunks, errors := nWorkers(
			url,
			workerCount,
			startByte,
			endByte,
			client,
			progress,
		)

		var downloaded int64

		for chunks != nil || errors != nil {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					chunks = nil
					continue
				}

				downloaded += int64(len(chunk.Bytes))

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

		elapsed := time.Since(testStart).Seconds()

		speed := float64(0)

		if elapsed > 0 {
			speed = float64(downloaded) / elapsed
		}

		speeds = append(speeds, speed)

		startByte = endByte + 1
	}

	if len(speeds) == 0 {
		return 1, 0, nil
	}

	// Find fastest worker count.
	bestIndex := 0

	for i := 1; i < len(speeds); i++ {
		if speeds[i] > speeds[bestIndex] {
			bestIndex = i
		}
	}

	return testWorkerCounts[bestIndex], startByte, nil
}

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
			totalSize,
			startByte,
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

func Download(
	url string,
	totalSize int64,
	progress *tracker.Progress,
	acceptRange bool,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
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

		transport := &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		}

		client := &http.Client{
			Transport: transport,
		}

		if client == nil {
			client = http.DefaultClient
		}

		// if server does not support ranges.
		if !acceptRange {
			task := Task{
				URL: url,
				Range: ByteRange{
					Start: 0,
					End:   totalSize - 1,
				},
				Index:  0,
				Client: client,
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

		// Automatically determine worker count.
		if maxWorkers <= 0 {
			selectedWorkers, nextByte, err := testWorkers(
				url,
				totalSize,
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
