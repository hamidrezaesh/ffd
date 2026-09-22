package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hamidrezaesh/ffd/internal/scheduler"
)

func TestDownloadWithHeaders(t *testing.T) {
	data := []byte("Hello from FFD test server!")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("========== FFD TEST SERVER ==========")
		fmt.Println("Method:", r.Method)
		fmt.Println("URL:", r.URL)
		fmt.Println("Protocol:", r.Proto)
		fmt.Println("Headers:")

		for key, values := range r.Header {
			fmt.Printf("  %s: %v\n", key, values)
		}

		fmt.Println("=====================================")

		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	req := Request{
		URL:      server.URL,
		Path:     t.TempDir(),
		Filename: "test",
		Headers: scheduler.Headers{
			{
				Key:   "X-Test",
				Value: "hello",
			},
			{
				Key:   "X-FFD-Test",
				Value: "true",
			},
		},
	}

	result, err := Download(
		req,
		1,
		1,
		1,
		1,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := <-result.Done; err != nil {
		t.Fatal(err)
	}
}
