//go:build benchmark

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const ffdPath = "../../build/ffd"

func runBenchmark(cmd *cobra.Command, args []string) error {
	if runs < 1 {
		return fmt.Errorf("runs must be at least 1")
	}
	if len(sizes) == 0 {
		return fmt.Errorf("at least one --sizes value is required")
	}

	if err := buildFFD(); err != nil {
		return err
	}

	timePath, err := findTimeExecutable()
	if err != nil {
		return err
	}

	var results []BenchmarkResult

	source := "remote"
	if local {
		source = "local"
	}
	fmt.Printf("Benchmark source: %s\n", source)

	for _, size := range sizes {
		file, ok := fileSizes[strings.ToLower(strings.TrimSpace(size))]
		if !ok {
			return fmt.Errorf("unsupported file size: %s", size)
		}

		url := file.GetURL(local)
		if url == "" {
			return fmt.Errorf(
				"no URL configured for %s (local mode: %t)",
				size,
				local,
			)
		}

		fmt.Printf("\nBenchmarking %s...\n", file.Name)
		br := BenchmarkResult{Size: file}

		for i := 1; i <= runs; i++ {
			fmt.Printf("\nRun %d/%d\n", i, runs)

			for _, downloader := range downloaderNames {
				dir := filepath.Join(
					os.TempDir(),
					"ffd-benchmark",
					fmt.Sprintf("%s-%d-%s", size, i, downloader),
				)

				elapsed, runErr := downloadTimed(
					timePath,
					downloader,
					url,
					dir,
				)

				br.Results = append(br.Results, DownloaderResult{
					Name:    downloader,
					Elapsed: elapsed,
					Error:   runErr,
				})

				if runErr != nil {
					fmt.Printf("  %s: FAILED: %v\n", downloader, runErr)
					continue
				}

				fmt.Printf(
					"  %s: %s\n",
					downloader,
					elapsed.Round(time.Millisecond),
				)
			}
		}

		results = append(results, br)
	}

	if err := writeResults(results); err != nil {
		return fmt.Errorf("write benchmark results: %w", err)
	}

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

func findTimeExecutable() (string, error) {
	if info, err := os.Stat("/usr/bin/time"); err == nil && !info.IsDir() {
		return "/usr/bin/time", nil
	}

	path, err := exec.LookPath("time")
	if err != nil {
		return "", fmt.Errorf(
			"GNU time executable not found (checked /usr/bin/time and PATH): %w",
			err,
		)
	}

	// Verify support for GNU time's -f and -o options.
	probe := exec.Command(
		path,
		"-f", "%e total",
		"-o", os.DevNull,
		"true",
	)

	if err := probe.Run(); err != nil {
		return "", fmt.Errorf(
			"time executable %q does not support GNU time -f/-o options: %w",
			path,
			err,
		)
	}

	return path, nil
}

func downloadTimed(
	timePath string,
	downloader string,
	url string,
	dir string,
) (elapsed time.Duration, retErr error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("create download directory: %w", err)
	}

	// Always remove this run's downloaded files, including on failure.
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Warning: failed to clean up %s: %v\n",
				dir,
				err,
			)
		}
	}()

	output := filepath.Join(dir, "download")

	if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("remove previous output: %w", err)
	}

	var command *exec.Cmd

	switch downloader {
	case "ffd":
		path, err := getFFDPath()
		if err != nil {
			return 0, err
		}

		command = exec.Command(
			path,
			"--overwrite",
			url,
			"-o", "download",
		)
		command.Dir = dir

	case "aria2":
		path, err := exec.LookPath("aria2c")
		if err != nil {
			return 0, fmt.Errorf("aria2c is not installed: %w", err)
		}

		command = exec.Command(
			path,
			"--allow-overwrite=true",
			"--auto-file-renaming=false",
			"-d", dir,
			"-o", filepath.Base(output),
			url,
		)

	case "wget":
		path, err := exec.LookPath("wget")
		if err != nil {
			return 0, fmt.Errorf("wget is not installed: %w", err)
		}

		command = exec.Command(path, "-O", output, url)

	default:
		return 0, fmt.Errorf("unknown downloader: %s", downloader)
	}

	// Store timing output separately so downloader output cannot corrupt it.
	timingFile, err := os.CreateTemp("", "ffd-benchmark-time-*.txt")
	if err != nil {
		return 0, fmt.Errorf("create timing file: %w", err)
	}

	timingPath := timingFile.Name()

	if err := timingFile.Close(); err != nil {
		_ = os.Remove(timingPath)
		return 0, fmt.Errorf("close timing file: %w", err)
	}

	defer func() {
		if err := os.Remove(timingPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(
				os.Stderr,
				"Warning: failed to remove timing file %s: %v\n",
				timingPath,
				err,
			)
		}
	}()

	args := []string{
		"-f", "%e total",
		"-o", timingPath,
		command.Path,
	}
	args = append(args, command.Args[1:]...)

	timed := exec.Command(timePath, args...)
	timed.Dir = command.Dir
	timed.Stdout = os.Stdout
	timed.Stderr = os.Stderr

	if err := timed.Run(); err != nil {
		return readElapsed(
			timingPath,
			fmt.Errorf("%s exited unsuccessfully: %w", downloader, err),
		)
	}

	return readElapsed(timingPath, nil)
}

func readElapsed(path string, commandErr error) (time.Duration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if commandErr != nil {
			return 0, fmt.Errorf(
				"%v; read timing output: %w",
				commandErr,
				err,
			)
		}
		return 0, fmt.Errorf("read timing output: %w", err)
	}

	elapsed, parseErr := parseElapsed(string(data))

	if commandErr != nil {
		if parseErr != nil {
			return 0, fmt.Errorf(
				"%v; parse timing output: %w",
				commandErr,
				parseErr,
			)
		}
		return elapsed, commandErr
	}

	if parseErr != nil {
		return 0, fmt.Errorf("parse timing output: %w", parseErr)
	}

	return elapsed, nil
}

func parseElapsed(output string) (time.Duration, error) {
	fields := strings.Fields(output)

	if len(fields) != 2 || fields[1] != "total" {
		return 0, fmt.Errorf(
			"expected '<seconds> total', got %q",
			strings.TrimSpace(output),
		)
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("invalid elapsed seconds %q", fields[0])
	}

	if seconds > float64((1<<63-1)/int64(time.Second)) {
		return 0, fmt.Errorf("elapsed time overflows time.Duration")
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func getFFDPath() (string, error) {
	return filepath.Abs(ffdPath)
}
