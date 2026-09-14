package scheduler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tracker"
	"github.com/quic-go/quic-go/http3"
)

/*
checkAvailableProtocol checks whether the server accepts a download
request using the specified HTTP protocol.

It returns true if the request succeeds and false if it fails.
*/

func checkAvailableProtocol(url string, client *http.Client) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusPartialContent
}

/*
testProtocol is responsible for testing HTTP/1.1, HTTP/2 and HTTP/3 speeds while downloading parts of
the file.
it tests a portion of the file with each protocol and returns the preferred protocol.
when speeds are within ~5% of each other, it prefers HTTP/3 over HTTP/2 and HTTP/2 over HTTP/1.1.
*/

func testProtocol(
	url string,
	totalSize int64,
	startByte int64,
	progress *tracker.Progress,
	emit func(Chunk),
) (int, int64, error) {
	if totalSize <= 0 {
		return 0, 0, fmt.Errorf("invalid total size")
	}

	const minTestSize int64 = 2 * 1024 * 1024
	const maxTestSize int64 = 200 * 1024 * 1024

	testProtocols := []int{
		1,
		2,
		3,
	}

	testSize := clamp(
		totalSize/50,
		minTestSize,
		maxTestSize,
	)

	if testSize*int64(len(testProtocols)) > totalSize {
		testSize = totalSize / int64(len(testProtocols))
	}

	if testSize <= 0 {
		return 0, 0, nil
	}

	var speed1 float64 = 0.0
	var speed2 float64 = 0.0
	var speed3 float64 = 0.0

	for _, protocol := range testProtocols {
		if startByte >= totalSize {
			break
		}
		testEnd := startByte + testSize
		if testEnd > totalSize {
			testEnd = totalSize
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

		testRange := ranges[1]

		transport := &http.Transport{
			MaxIdleConns:        16,
			MaxIdleConnsPerHost: 16,
			MaxConnsPerHost:     16,
		}

		var client *http.Client

		switch protocol {
		case 1:
			transport.TLSNextProto =
				map[string]func(string, *tls.Conn) http.RoundTripper{}

			client = &http.Client{
				Transport: transport,
			}

		case 2:
			transport.ForceAttemptHTTP2 = true

			client = &http.Client{
				Transport: transport,
			}

		case 3:
			h3Transport := &http3.Transport{}

			client = &http.Client{
				Transport: h3Transport,
				Timeout:   200 * time.Millisecond,
			}
		}

		supported := checkAvailableProtocol(url, client)
		if !supported {
			continue
		}

		testStart := time.Now()

		chunks, errors := nWorkers(
			url,
			8,
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

		switch protocol {
		case 1:
			speed1 = speed
		case 2:
			speed2 = speed
		case 3:
			speed3 = speed
		}

		startByte = testRange.End + 1
	}

	var finalProtocol int

	if speed3 >= speed2*0.95 {
		finalProtocol = 3
	} else if speed2 >= speed1*0.95 {
		finalProtocol = 2
	} else {
		finalProtocol = 1
	}

	return finalProtocol, startByte, nil
}
