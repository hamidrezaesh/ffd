//go:build benchmark

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func writeResults(results []BenchmarkResult) error {
	var report strings.Builder

	report.WriteString("# FFD Benchmark Results\n\n")
	report.WriteString("Generated: " + time.Now().Format("2006-01-01 15:02") + "\n\n")
	report.WriteString("| File Size | FFD | aria2 | wget | FFD vs aria2 | FFD vs wget |\n")
	report.WriteString("|---|---:|---:|---:|---:|---:|\n")

	for _, result := range results {
		times := make(map[string]time.Duration)
		for _, r := range result.Results {
			if r.Error == nil {
				times[r.Name] = r.Elapsed
			}
		}

		report.WriteString(fmt.Sprintf(
			"| %s | %s | %s | %s | %s | %s |\n",
			result.Size.Name,
			formatTime(times, "ffd"),
			formatTime(times, "aria2"),
			formatTime(times, "wget"),
			formatSpeed(times, "aria2"),
			formatSpeed(times, "wget"),
		))
	}

	return os.WriteFile("RESULT.md", []byte(report.String()), 0644)
}

func formatTime(times map[string]time.Duration, name string) string {
	t, ok := times[name]
	if !ok {
		return "N/A"
	}
	return t.Round(time.Millisecond).String()
}

func formatSpeed(times map[string]time.Duration, competitor string) string {
	ffd, ffdOK := times["ffd"]
	other, otherOK := times[competitor]

	if !ffdOK || !otherOK || other == 0 {
		return "N/A"
	}

	percent := (float64(other-ffd) / float64(other)) * 100
	return fmt.Sprintf("%+.2f%%", percent)
}
