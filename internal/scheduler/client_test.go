package scheduler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "hello" {
			t.Error("header was not set")
		} else {
			t.Log("header was set")
		}
	}))
	defer server.Close()

	client := &HeaderClient{
		Base: http.DefaultClient,
		Headers: Headers{
			{
				Key:   "X-Test",
				Value: "hello",
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
