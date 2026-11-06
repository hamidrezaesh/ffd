package main

var benchmarkHelp string = `usage: ./benchmark [OPTION]

FFD Benchmark

A benchmarking tool to compare FFD, aria2, and wget download performance
using test files hosted on testfile.to.

Startup:
-h, --help             Show benchmark help

Options:
--runs NUMBER      Number of benchmark runs per file size (default 3)
Example: ./benchmark --runs 5

--sizes SIZES      Comma-separated file sizes to benchmark
Example: benchmark --sizes 1mb,10mb,100mb

--tools TOOLS	Comma-separated tools to benchmark
Example: ./benchmark --tools aria2,ffd

Output:
RESULT.md              Benchmark results, performance comparisons, and speed differences

Build:
go build -o benchmark -tags benchmark .

Note:
Run all commands from ffd/test/benchmark.
`
