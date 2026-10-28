//go:build benchmark

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	sizes []string
	runs  int
	local bool
)

var rootCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark FFD against aria2 and wget",
	Long:  "Compare download performance using files hosted by testfile.to.",
	RunE:  runBenchmark,
}

func init() {
	rootCmd.Flags().StringSliceVar(
		&sizes,
		"sizes",
		[]string{"10mb"},
		"File sizes to benchmark (1mb, 10mb, 100mb, 1gb, 10gb)",
	)

	rootCmd.Flags().IntVar(
		&runs,
		"runs",
		3,
		"Number of benchmark runs per downloader",
	)

	rootCmd.Flags().BoolVar(
		&local,
		"local",
		false,
		"Use local benchmark files from 127.0.0.1:8080",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
