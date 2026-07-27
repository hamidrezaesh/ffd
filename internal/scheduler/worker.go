package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

type Task struct {
	URL    string
	Range  ByteRange
	File   *os.File
	Client *http.Client
}

func Worker(t Task, progress *tracker.Progress) error {
	req, err := http.NewRequest("GET", t.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Set(
		"Range",
		fmt.Sprintf("bytes=%d-%d", t.Range.Start, t.Range.End),
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

	offset := t.Range.Start

	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			_, wErr := t.File.WriteAt(buf[:n], offset)
			if wErr != nil {
				return wErr
			}

			progress.AddDownloaded(int64(n))

			offset += int64(n)
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
