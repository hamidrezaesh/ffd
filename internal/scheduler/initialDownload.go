package scheduler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hamidrezaesh/ffd/internal/metadata"
)

type initialHeadResult struct {
	resp *http.Response
	err  error
}

type initialMetadataResult struct {
	resp *http.Response
	md   metadata.Metadata
	err  error
}

func DownloadInitial(
	client *HeaderClient,
	rawURL string,
	emit func(Chunk) error,
) (
	startByte int64,
	resp *http.Response,
	md metadata.Metadata,
	err error,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	headCh := make(chan initialHeadResult, 1)
	metadataCh := make(chan initialMetadataResult, 1)
	chunks := make(chan Chunk)
	var offset int64

	// 1. Send HEAD.
	go func() {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodHead,
			rawURL,
			nil,
		)
		if err != nil {
			headCh <- initialHeadResult{err: err}
			return
		}

		resp, err := client.Do(req)
		headCh <- initialHeadResult{
			resp: resp,
			err:  err,
		}
	}()

	// 2. Parse metadata as soon as HEAD returns.
	go func() {
		head := <-headCh
		if head.err != nil {
			metadataCh <- initialMetadataResult{err: head.err}
			return
		}

		md, err := metadata.FetchMetadata(rawURL, head.resp)
		if err != nil {
			_ = head.resp.Body.Close()
		}

		metadataCh <- initialMetadataResult{
			resp: head.resp,
			md:   md,
			err:  err,
		}
	}()

	// 3. Start GET independently of HEAD.
	go func() {
		defer close(chunks)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			rawURL,
			nil,
		)
		if err != nil {
			return
		}

		getResp, err := client.Do(req)
		if err != nil {
			return
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			return
		}

		buffer := make([]byte, 32*1024)
		var downloaded int64

		for {
			n, readErr := getResp.Body.Read(buffer)

			if n > 0 {
				data := append([]byte(nil), buffer[:n]...)

				chunks <- Chunk{
					Index:      0,
					RangeStart: 0,
					Offset:     downloaded,
					Bytes:      data,
				}

				downloaded += int64(n)
			}

			if readErr != nil {
				return
			}
		}
	}()

	// Consume every received chunk before handing off.
	consume := func(chunk Chunk) error {
		if emit != nil {
			if err := emit(chunk); err != nil {
				return err
			}
		}

		offset += int64(len(chunk.Bytes))
		return nil
	}

	drain := func() error {
		for chunk := range chunks {
			if err := consume(chunk); err != nil {
				return err
			}
		}
		return nil
	}

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				chunks = nil
				continue
			}

			if err := consume(chunk); err != nil {
				cancel()
				_ = drain()
				return offset, nil, metadata.Metadata{}, err
			}

		case result := <-metadataCh:
			if result.err != nil {
				cancel()
				_ = drain()
				return offset, nil, metadata.Metadata{}, result.err
			}

			// Stop GET and collect any bytes it already read.
			cancel()

			if err := drain(); err != nil {
				return offset, nil, metadata.Metadata{}, err
			}

			if offset > result.md.TotalSize {
				_ = result.resp.Body.Close()
				return offset, nil, metadata.Metadata{},
					fmt.Errorf("initial download exceeded content length")
			}

			return offset, result.resp, result.md, nil
		}
	}
}
