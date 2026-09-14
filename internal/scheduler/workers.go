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

/*
nWorkers is responsible for downloading a part of file using input workers.
it downloads the part and return downloaded bytes and speed.
*/

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
	startByte int64,
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

	if testSize*int64(len(testWorkerCounts)) > totalSize {
		testSize = totalSize / int64(len(testWorkerCounts))
	}

	if testSize <= 0 {
		return 1, 0, nil
	}

	speeds := make([]float64, 0, len(testWorkerCounts))

	for _, workerCount := range testWorkerCounts {
		if startByte >= totalSize {
			break
		}

		testEnd := startByte + testSize
		switch {
		case testEnd > totalSize:
			testEnd = totalSize
		}

		ranges, err := Split(
			testEnd,
			startByte,
			1,
			0,
		)
		if err != nil {
			return 0, startByte, err
		}

		testRange := ranges[1]

		testStart := time.Now()

		chunks, errors := nWorkers(
			url,
			workerCount,
			testRange.Start,
			testRange.End,
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

		startByte = testRange.End + 1
	}

	if len(speeds) == 0 {
		return 1, 0, nil
	}

	bestIndex := 0

	for i := 1; i < len(speeds); i++ {
		if speeds[i] > speeds[bestIndex] {
			bestIndex = i
		}
	}

	return testWorkerCounts[bestIndex], startByte, nil
}
