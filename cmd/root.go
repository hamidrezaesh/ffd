package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/hamidrezaesh/ffd/internal/engine"
	"github.com/hamidrezaesh/ffd/internal/formatter"
	"github.com/spf13/cobra"
)

var commandsHelp string = `usage: ffd [URL]...[OPTION]

Startup:
-v, --version	Show the version of ffd
-h, --help	Show help

Options:
-o, --output NAME	Save the file with a custom filename
Example: ffd <URL> -o my-file

-w, --wait SECONDS	Wait before starting the download
Example: ffd <URL> -w 100

-p, --path PATH	save the file to a custom directory
Example: ffd <URL> -p /path/to/your/folder

Examples:
ffd https://example.com/file.zip
ffd https://example.com/file.zip -o my-file
ffd https://example.com/file.zip -w 100
ffd https://example.com/file.zip -p ~/Downloads
`

var (
	output string
	wait   int
	path   string
)

var rootCmd = &cobra.Command{
	Use:   "ffd [URL] [OPTIONS]",
	Short: "Fast, multi-segment data fetcher",
	Args:  cobra.ExactArgs(1),

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

		req := engine.Request{
			URL:      args[0],
			Path:     path,
			Filename: output,
		}

		startTime := time.Now()

		// Start download.
		result, err := engine.Download(req)

		if err != nil {
			fmt.Printf(
				"\r\033[KDownload failed: %v\n",
				err,
			)

			return
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
					return
				}

				duration := time.Since(startTime).Round(time.Second)

				fmt.Printf(
					"\033[1A\r\033[K%s - 100%%\n\033[KDownload completed in %s\n",
					result.Filename,
					duration,
				)
				return
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

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(commandsHelp)
	})
}
