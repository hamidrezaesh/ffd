//go:build benchmark

/*
FFD Benchmark

A lightweight benchmarking tool to compare FFD, aria2, and wget download performance
using test files hosted on testfile.to. It runs configurable trials for different
file sizes, measures download times, and generates a `RESULT.md` report with performance comparisons
and speed differences.

Guide:
Note: Run all below commands in ffd/test/benchmark

Build benchmark:

	go build -o benchmark -tags benchmark .

Run benchmark:

	./benchmark --runs TOTAL_OF_RUNS --sizes SIZES

Example:
	./benchmark --runs 3 --sizes 1mb,10mb

That will run benchmark with 1mb and 10mb test file, which runs 3 times.
*/

package main

import "time"

type FileSize struct {
	Name string
	URL  string
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

var fileSizes = map[string]FileSize{
	"1mb":   {Name: "1 MB", URL: "https://testfile.to/dl/1mb"},
	"10mb":  {Name: "10 MB", URL: "https://testfile.to/dl/10mb"},
	"100mb": {Name: "100 MB", URL: "https://testfile.to/dl/100mb"},
	"1gb":   {Name: "1 GB", URL: "https://testfile.to/dl/1gb"},
	"10gb":  {Name: "10 GB", URL: "https://testfile.to/dl/10gb"},
}
