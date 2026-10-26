//go:build benchmark

/*
FFD Benchmark

Compare FFD, aria2, and wget using remote test files or a local benchmark server.
Run commands from ffd/test/benchmark.

Create temp files:
mkdir -p testdata

dd if=/dev/urandom of=testdata/1mb.bin bs=1M count=1 status=progress
dd if=/dev/urandom of=testdata/10mb.bin bs=1M count=10 status=progress
dd if=/dev/urandom of=testdata/100mb.bin bs=1M count=100 status=progress
dd if=/dev/urandom of=testdata/1gb.bin bs=1M count=1024 status=progress
dd if=/dev/urandom of=testdata/10gb.bin bs=1M count=10240 status=progress

Build: go build -o benchmark -tags benchmark .
Run:   ./benchmark --local --runs 3 --sizes 1mb,10mb
*/
package main

import "time"

type FileSize struct {
	Name     string
	URL      string
	LocalURL string
}

type DownloaderResult struct {
	Name    string
	Elapsed time.Duration
	Error   error
}

type BenchmarkResult struct {
	Size    FileSize
	Results []DownloaderResult
}

func (f FileSize) GetURL(local bool) string {
	if local {
		return f.LocalURL
	}
	return f.URL
}

var fileSizes = map[string]FileSize{
	"1mb":   {Name: "1 MB", URL: "https://testfile.to/dl/1mb", LocalURL: "http://127.0.0.1:8080/1mb.bin"},
	"10mb":  {Name: "10 MB", URL: "https://testfile.to/dl/10mb", LocalURL: "http://127.0.0.1:8080/10mb.bin"},
	"100mb": {Name: "100 MB", URL: "https://testfile.to/dl/100mb", LocalURL: "http://127.0.0.1:8080/100mb.bin"},
	"1gb":   {Name: "1 GB", URL: "https://testfile.to/dl/1gb", LocalURL: "http://127.0.0.1:8080/1gb.bin"},
	"10gb":  {Name: "10 GB", URL: "https://testfile.to/dl/10gb", LocalURL: "http://127.0.0.1:8080/10gb.bin"},
}
