//go:build benchmark

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const ffdPath = "../../build/ffd"

func runBenchmark(cmd *cobra.Command, args []string) error {
	if runs < 1 {
		return fmt.Errorf("runs must be at least 1")
	}

	// Build FFD before running any benchmarks.
	if err := buildFFD(); err != nil {
		return err
	}

	var results []BenchmarkResult

	for _, size := range sizes {
		file, ok := fileSizes[size]
		if !ok {
			return fmt.Errorf("unsupported file size: %s", size)
		}

		fmt.Printf("\nBenchmarking %s...\n", file.Name)

		// Store results for this file size across all trials.
		benchmarkResult := BenchmarkResult{
			Size: file,
		}

		for i := 1; i <= runs; i++ {
			fmt.Printf("\nRun %d/%d\n", i, runs)

			for _, downloader := range []string{"ffd", "aria2", "wget"} {
				if downloader != "ffd" {
					binary := downloader
					if binary == "aria2" {
						binary = "aria2c"
					}

					if _, err := exec.LookPath(binary); err != nil {
						fmt.Printf("  %s: not installed, skipping\n", downloader)

						benchmarkResult.Results = append(
							benchmarkResult.Results,
							DownloaderResult{
								Name: downloader,
								Error: fmt.Errorf(
									"%s is not installed",
									binary,
								),
							},
						)
						continue
					}
				}

				dir := filepath.Join(
					os.TempDir(),
					"ffd-benchmark",
					fmt.Sprintf("%s-%d-%s", size, i, downloader),
				)

				start := time.Now()
				err := download(downloader, file.URL, dir)
				elapsed := time.Since(start)

				benchmarkResult.Results = append(
					benchmarkResult.Results,
					DownloaderResult{
						Name:    downloader,
						Elapsed: elapsed,
						Error:   err,
					},
				)

				if err != nil {
					fmt.Printf("  %s: failed: %v\n", downloader, err)
					continue
				}

				fmt.Printf("  %s: %s\n",
					downloader,
					elapsed.Round(time.Millisecond),
				)
			}
		}

		results = append(results, benchmarkResult)
	}

	// Write the final report.
	if err := writeResults(results); err != nil {
		return fmt.Errorf("write benchmark results: %w", err)
	}

	fmt.Println("\nResults written to RESULT.md")
	return nil
}

func buildFFD() error {
	fmt.Println("Building FFD...")

	if err := os.MkdirAll("../../build", 0755); err != nil {
		return fmt.Errorf("create build directory: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", "build/ffd", ".")
	cmd.Dir = "../.."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build FFD; benchmark aborted: %w", err)
	}

	fmt.Println("FFD build successful.")
	return nil
}

func download(downloader, url, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	output := filepath.Join(dir, "download")

	// Remove the previous download so every trial starts fresh.
	if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous output: %w", err)
	}

	var cmd *exec.Cmd

	switch downloader {
	case "ffd":
		path, err := getFFDPath()
		if err != nil {
			return err
		}

		cmd = exec.Command(
			path,
			"--overwrite",
			url,
			"-o", "download",
		)
		cmd.Dir = dir

	case "aria2":
		cmd = exec.Command(
			"aria2c",
			"--allow-overwrite=true",
			"--auto-file-renaming=false",
			"-d", dir,
			"-o", filepath.Base(output),
			url,
		)

	case "wget":
		cmd = exec.Command(
			"wget",
			"-O", output,
			url,
		)

	default:
		return fmt.Errorf("unknown downloader: %s", downloader)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getFFDPath() (string, error) {
	path, err := filepath.Abs("../../build/ffd")
	if err != nil {
		return "", err
	}
	return path, nil
}
