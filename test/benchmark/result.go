//go:build benchmark

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var downloaderNames = []string{"ffd", "aria2", "wget"}

type downloaderSummary struct {
	Average time.Duration
	Count   int
	Errors  []string
}

func aggregateResults(results []DownloaderResult) map[string]downloaderSummary {
	totals := make(map[string]time.Duration)
	summaries := make(map[string]downloaderSummary)
	for _, r := range results {
		name := canonicalDownloader(r.Name)
		if name == "" {
			continue
		}
		s := summaries[name]
		if r.Error != nil {
			s.Errors = append(s.Errors, r.Error.Error())
		} else if r.Elapsed > 0 {
			totals[name] += r.Elapsed
			s.Count++
		} else {
			s.Errors = append(s.Errors, "successful result had no positive elapsed time")
		}
		summaries[name] = s
	}
	for name, s := range summaries {
		if s.Count > 0 {
			s.Average = time.Duration(int64(totals[name]) / int64(s.Count))
		}
		summaries[name] = s
	}
	return summaries
}

func canonicalDownloader(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ffd":
		return "ffd"
	case "aria2", "aria2c":
		return "aria2"
	case "wget":
		return "wget"
	default:
		return ""
	}
}

func writeResults(results []BenchmarkResult) error {
	var report strings.Builder
	report.WriteString("# FFD Benchmark Results\n\n")
	report.WriteString("Generated: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	report.WriteString("| File Size | FFD | aria2 | wget | FFD vs aria2 | FFD vs wget |\n")
	report.WriteString("|---|---:|---:|---:|---:|---:|\n")

	for _, result := range results {
		summaries := aggregateResults(result.Results)
		times := make(map[string]time.Duration)
		for _, name := range downloaderNames {
			if s, ok := summaries[name]; ok && s.Count > 0 {
				times[name] = s.Average
			}
		}
		report.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			result.Size.Name, formatTime(times, "ffd"), formatTime(times, "aria2"), formatTime(times, "wget"),
			formatSpeed(times, "aria2"), formatSpeed(times, "wget")))
	}

	// Keep failure diagnostics in the report instead of silently dropping failed trials.
	var failures []string
	for _, result := range results {
		for _, r := range result.Results {
			if r.Error != nil {
				failures = append(failures, fmt.Sprintf("- **%s / %s:** %v", result.Size.Name, canonicalOrOriginal(r.Name), r.Error))
			}
		}
	}
	if len(failures) > 0 {
		report.WriteString("\n## Failed measurements\n\n")
		for _, failure := range failures {
			report.WriteString(failure + "\n")
		}
	}

	content := report.String()
	// Print the exact report table before writing the file.
	fmt.Print("\n" + content)
	outputPath, err := filepath.Abs("RESULT.md")
	if err != nil {
		return fmt.Errorf("resolve report path: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}
	fmt.Printf("\nResults written to %s\n", outputPath)
	return nil
}

func canonicalOrOriginal(name string) string {
	if canonical := canonicalDownloader(name); canonical != "" {
		return canonical
	}
	return name
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
	if !ffdOK || !otherOK || other <= 0 {
		return "N/A"
	}
	// Positive means FFD is faster; negative means FFD is slower.
	percent := (float64(other) - float64(ffd)) / float64(other) * 100
	return fmt.Sprintf("%+.2f%%", percent)
}

func sortedFailureNames(m map[string]downloaderSummary) []string {
	var names []string
	for name, s := range m {
		if len(s.Errors) > 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
