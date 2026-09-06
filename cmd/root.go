package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/hamidrezaesh/ffd/internal/engine"
	"github.com/hamidrezaesh/ffd/internal/formatter"
	"github.com/hamidrezaesh/ffd/internal/proxy"
	"github.com/spf13/cobra"
)

var commandsHelp string = `usage: ffd [URL]...[OPTION]

Commands:
proxy              Start the ffd forward proxy
Example: ffd proxy

update             Update ffd to the latest version
Example: ffd update

Startup:
-h, --help	Show help

-v, --version	Show ffd version

Options:
-o, --output NAME	Save the file with a custom filename
Example: ffd <URL> -o my-file

-w, --wait SECONDS	Wait before starting the download
Example: ffd <URL> -w 100

-p, --path PATH	save the file to a custom directory (default .)
Example: ffd <URL> -p /path/to/your/folder

-r --max-retries NUMBER_OF_RETRIES	Total retries after connection failed (default 4)
Example: ffd <URL> -r 10

-W --max-workers NUMBER_OF_WORKERS	Total concurrent workers (default 8)
Example: ffd <URL> -W 10

-c --max-chunks NUMBER_OF_CHUNKS	Total parts of download (default 12)
Example: ffd <URL> -c 20

--protocol PROTOCOL	Protocol to use for download (default 'auto')
Example: ffd <URL> --protocol http2
`

var proxyHelp string = `usage: ffd proxy [OPTION]

Startup:
-h, --help        Show help

Options:
--port PORT   Port to run the proxy on (default 8000)
Example: ffd proxy --port 9000
`

var (
	output     string
	wait       int
	path       string
	maxRetries int
	maxWorkers int
	maxChunks  int
	port       int
	protocol   string
)

func downloadUrl(req engine.Request) error {
	var useProtocol int
	switch protocol {
	case "auto":
		useProtocol = 0
	case "http1":
		useProtocol = 1
	case "http2":
		useProtocol = 2
	case "http3":
		useProtocol = 3
	}

	startTime := time.Now()

	// Start download.
	result, err := engine.Download(req, maxRetries, maxWorkers, maxChunks, useProtocol)

	if err != nil {
		fmt.Printf(
			"\r\033[KDownload failed: %v\n",
			err,
		)

		return err
	}

	fmt.Println("Downloading...")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	first := true
	for {
		select {
		case <-ticker.C:
			progress := result.Progress.Info()
			if first {
				fmt.Printf(
					"%s\t%.2f%%\n%s left - %s / %s | %s/s",
					result.Filename,
					progress.Percent,
					progress.TimeLeft,
					formatter.Bytes(progress.Downloaded),
					formatter.Bytes(progress.Total),
					formatter.Bytes(progress.Speed),
				)
				first = false
				continue
			}

			fmt.Printf(
				"\033[1A\r\033[K%s\t%.2f%%\n\033[K%s left - %s / %s | %s/s",
				result.Filename,
				progress.Percent,
				progress.TimeLeft,
				formatter.Bytes(progress.Downloaded),
				formatter.Bytes(progress.Total),
				formatter.Bytes(progress.Speed),
			)
		case err := <-result.Done:
			if err != nil {
				fmt.Printf(
					"\n\033[KDownload failed: %v\n",
					err,
				)
				return err
			}

			duration := time.Since(startTime).Round(time.Second)

			fmt.Printf(
				"\033[1A\r\033[K%s - 100%%\n\033[KDownload completed in %s\n",
				result.Filename,
				duration,
			)
			return nil
		}
	}
}

type Release struct {
	TagName string `json:"tag_name"`
}

func GetLatestVersion() (string, error) {
	url := "https://api.github.com/repos/hamidrezaesh/ffd/releases/latest"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release Release

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

var rootCmd = &cobra.Command{
	Use:     "ffd [URL] [OPTIONS]",
	Short:   "Fast, multi-segment data fetcher",
	Version: version,
	Args:    cobra.ArbitraryArgs,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(commandsHelp)
			return
		}

		// Countdown
		if wait > 0 {
			for i := wait; i > 0; i-- {
				fmt.Printf(
					"\r\033[KStarting download in %d seconds...",
					i,
				)

				time.Sleep(time.Second)
			}

			fmt.Print("\r\033[K")
		}

		for _, url := range args {
			req := engine.Request{
				URL:      url,
				Path:     path,
				Filename: output,
			}

			err := downloadUrl(req)
			if err != nil {
				fmt.Printf("Error downloading: %v\n", err)
				return
			}
		}
	},
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the ffd forward proxy",
	Run: func(cmd *cobra.Command, args []string) {
		proxy.Start(port)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ffd",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// detecting os
			log.Print("Detecting OS...")
			os := runtime.GOOS
			log.Printf("OS: %v\n", os)

			// checking version
			latestVersion, err := GetLatestVersion()
			if err != nil {
				log.Fatal(err)
			}
			if latestVersion == version {
				log.Printf("ffd is up to date (%v)\n", version)
				return
			}

			log.Printf("updating to the latest version (%v)...\n", latestVersion)

			var cmd *exec.Cmd

			if os == "windows" {
				cmd = exec.Command(
					"powershell",
					"-Command",
					"Set-ExecutionPolicy -Scope CurrentUser RemoteSigned; irm https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.ps1 | iex",
				)
			} else if os == "linux" || os == "darwin" {
				cmd = exec.Command(
					"sh",
					"-c",
					"curl -fsSL https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.sh | sh",
				)
			} else {
				log.Fatalf("Unsupported OS: %v", runtime.GOOS)
			}

			err = cmd.Run()
			if err != nil {
				log.Fatalf("Error while updating ffd: %v\n", err)
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&wait, "wait", "w", 0, "Wait before starting download (seconds)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Custom Filename")
	rootCmd.Flags().StringVarP(&path, "path", "p", ".", "Download Directory")
	rootCmd.Flags().IntVarP(&maxRetries, "max-retries", "r", 4, "Total retries after connection failed")
	rootCmd.Flags().IntVarP(&maxWorkers, "max-workers", "W", 8, "Total concurrent workers")
	rootCmd.Flags().IntVarP(&maxChunks, "max-chunks", "c", 12, "Total parts of download")
	rootCmd.Flags().StringVarP(&protocol, "protocol", "", "auto", "Protocol to use")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(commandsHelp)
	})

	proxyCmd.Flags().IntVarP(&port, "port", "", 8000, "Port of the proxy")

	proxyCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(proxyHelp)
	})

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(updateCmd)
}
