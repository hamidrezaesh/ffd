package proxy

import (
	"io"
	"net/http"
	"strconv"

	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/hamidrezaesh/ffd/internal/tracker"
)

const (
	accelerationThreshold int64 = 5 * 1024 * 1024
	minFileSize           int64 = 5 * 1024 * 1024

	maxWorkers = 8
	maxChunks  = 12
	maxRetries = 4
)

func (s *Server) proxyNormal(
	w http.ResponseWriter,
	req *http.Request,
) {
	resp, err := s.Client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) proxyAccelerated(
	w http.ResponseWriter,
	req *http.Request,
	headers http.Header,
	totalSize int64,
) {
	progress := tracker.New(totalSize)
	progress.Start()
	defer progress.Stop()

	chunks, errors := scheduler.Download(
		req.URL.String(),
		totalSize,
		true,
		minFileSize,
		maxWorkers,
		maxChunks,
		s.Client,
		progress,
		maxRetries,
	)

	// Copy upstream response metadata
	copyHeaders(w.Header(), headers)

	w.Header().Set(
		"Content-Length",
		strconv.FormatInt(totalSize, 10),
	)

	w.WriteHeader(http.StatusOK)

	// write chunks as they arrive
	pending := make(map[int64][]byte)
	nextOffset := int64(0)

	for chunks != nil || errors != nil {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				chunks = nil
				continue
			}

			pending[chunk.Offset] = chunk.Bytes

			for {
				data, ok := pending[nextOffset]
				if !ok {
					break
				}

				if _, err := w.Write(data); err != nil {
					return
				}

				delete(pending, nextOffset)
				nextOffset += int64(len(data))
			}

		case err, ok := <-errors:
			if !ok {
				errors = nil
				continue
			}

			if err != nil {
				return
			}
		}
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		switch http.CanonicalHeaderKey(key) {
		case "Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"Te",
			"Trailer",
			"Transfer-Encoding",
			"Upgrade":
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
