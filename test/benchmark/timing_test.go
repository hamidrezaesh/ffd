//go:build benchmark

package main

import (
	"errors"
	"testing"
	"time"
)

func TestParseElapsed(t *testing.T) {
	got, err := parseElapsed("1.486 total\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1486*time.Millisecond {
		t.Fatalf("got %v, want 1.486s", got)
	}
}

func TestParseElapsedRejectsDownloaderNoise(t *testing.T) {
	for _, input := range []string{"downloaded 1.486 MB", "1.486", "1.486 total\n2.0 total", "NaN total", "-1 total", "1.2 seconds"} {
		if _, err := parseElapsed(input); err == nil {
			t.Errorf("parseElapsed(%q) unexpectedly succeeded", input)
		}
	}
}

func TestAggregateResults(t *testing.T) {
	got := aggregateResults([]DownloaderResult{
		{Name: "ffd", Elapsed: time.Second},
		{Name: "ffd", Elapsed: 3 * time.Second},
		{Name: "aria2c", Elapsed: 2 * time.Second},
		{Name: "wget", Error: errors.New("failed")},
	})
	if got["ffd"].Average != 2*time.Second || got["ffd"].Count != 2 {
		t.Fatalf("bad FFD aggregate: %+v", got["ffd"])
	}
	if got["aria2"].Average != 2*time.Second || got["aria2"].Count != 1 {
		t.Fatalf("bad aria2 aggregate: %+v", got["aria2"])
	}
	if got["wget"].Count != 0 || len(got["wget"].Errors) != 1 {
		t.Fatalf("bad wget aggregate: %+v", got["wget"])
	}
}
