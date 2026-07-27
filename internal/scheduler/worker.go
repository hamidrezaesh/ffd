package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

type Task struct {
	URL    string
	Range  ByteRange
	File   *os.File
	Client *http.Client
}

func fetchFromOffset(t Task, progress *tracker.Progress, offset *int64) error {
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
			_, wErr := t.File.WriteAt(buf[:n], *offset)
			if wErr != nil {
				return wErr
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

func Worker(t Task, progress *tracker.Progress, maxRetries int) error {
	if maxRetries <= -1 {
		maxRetries = 4
	}

	offset := t.Range.Start

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt == maxRetries {
			return fmt.Errorf("download failed after %d retries.", maxRetries)
		}

		err := fetchFromOffset(t, progress, &offset)

		if err == nil {
			return nil
		}

		time.Sleep(time.Second * 3)
	}

	return nil
}
