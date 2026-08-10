package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

type Task struct {
	URL    string
	Range  ByteRange
	Index  int
	Client *http.Client
}

type Chunk struct {
	Index  int
	Offset int64
	Bytes  []byte
}

func fetchFromOffset(
	t Task,
	progress *tracker.Progress,
	offset *int64,
	chunks chan<- Chunk,
) error {
	req, err := http.NewRequest("GET", t.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Set(
		"Range",
		fmt.Sprintf("bytes=%d-%d", *offset, t.Range.End),
	)

	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected HTTP status: %v", resp.Status)
	}

	buf := make([]byte, 128*1024)

	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			data := append([]byte(nil), buf[:n]...)

			chunks <- Chunk{
				Index:  t.Index,
				Offset: *offset,
				Bytes:  data,
			}

			progress.AddDownloaded(int64(n))
			*offset += int64(n)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func Worker(
	t Task,
	progress *tracker.Progress,
	maxRetries int,
	chunks chan<- Chunk,
) error {
	if maxRetries < 0 {
		maxRetries = 4
	}

	offset := t.Range.Start

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fetchFromOffset(
			t,
			progress,
			&offset,
			chunks,
		)

		if err == nil {
			return nil
		}

		if attempt == maxRetries {
			return fmt.Errorf(
				"download failed after %d attempts: %w",
				attempt+1,
				err,
			)
		}

		time.Sleep(3 * time.Second)
	}

	return nil
}
