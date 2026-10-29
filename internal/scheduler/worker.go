package scheduler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
)

type Task struct {
	URL    string
	ETag   string
	Range  ByteRange
	Index  int
	Client *HeaderClient
}

type Chunk struct {
	Index      int
	RangeStart int64
	RangeSize  int64
	Offset     int64
	Bytes      []byte
	Test       bool
}

type ResumeChunk struct {
	Index      int
	Range      ByteRange
	Downloaded int64
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

	requestStart := *offset

	req.Header.Set(
		"Range",
		fmt.Sprintf("bytes=%d-%d", requestStart, t.Range.End),
	)

	// Protect this segment against remote file changes.
	if t.ETag != "" {
		if strings.HasPrefix(t.ETag, "W/") {
			return fmt.Errorf("cannot use weak ETag for If-Match")
		}
		req.Header.Set("If-Match", t.ETag)
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		return fmt.Errorf("remote file changed: ETag mismatch")
	}

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected HTTP status: %v", resp.Status)
	}

	contentRange := resp.Header.Get("Content-Range")
	expectedPrefix := fmt.Sprintf("bytes %d-", requestStart)

	if !strings.HasPrefix(contentRange, expectedPrefix) {
		return fmt.Errorf(
			"unexpected Content-Range: %q, expected start %d",
			contentRange,
			requestStart,
		)
	}

	buf := make([]byte, 128*1024)

	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			data := append([]byte(nil), buf[:n]...)

			chunks <- Chunk{
				Index:      t.Index,
				RangeStart: t.Range.Start,
				RangeSize:  t.Range.End - t.Range.Start + 1,
				Offset:     *offset,
				Bytes:      data,
				Test:       false,
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
	resume ResumeChunk,
	progress *tracker.Progress,
	maxRetries int,
	chunks chan<- Chunk,
) error {
	if maxRetries < 0 {
		maxRetries = 4
	}

	offset := resume.Range.Start + resume.Downloaded

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
