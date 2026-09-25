package cmd

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hamidrezaesh/ffd/internal/config"
	"github.com/hamidrezaesh/ffd/internal/engine"
	"github.com/hamidrezaesh/ffd/internal/formatter"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/hamidrezaesh/ffd/internal/version"
	"github.com/spf13/cobra"
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
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	// get proxy server
	var proxyServer *url.URL
	var err error
	if downloadProxyServerString != "" {
		proxyServer, err = url.Parse(downloadProxyServerString)
		if err != nil {
			fmt.Printf(
				"\r\033[KDownload failed: %v\n",
				err,
			)
			return err
		}
	}

	startTime := time.Now()

	// Start download.
	result, err := engine.Download(req, maxRetries, maxWorkers, maxChunks, useProtocol, proxyServer)

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

var rootCmd = &cobra.Command{
	Use:     "ffd [URL] [OPTIONS]",
	Short:   "Fast, multi-segment data fetcher",
	Version: version.Version,
	Args:    cobra.ArbitraryArgs,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(commandsHelp)
			return
		}

		// Load config.
		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Set headers.
		headers := scheduler.Headers{}

		for _, header := range cfg.Headers {
			headers = append(headers, scheduler.Header{
				Key:   header.Key,
				Value: header.Value,
			})
		}

		for _, value := range requestHeaders {
			parts := strings.SplitN(value, ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				fmt.Printf("Invalid header: %s\n", value)
				return
			}

			headers = append(headers, scheduler.Header{
				Key:   strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}

		// Create cookie jar.
		jar, err := cookiejar.New(nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Set cookies.
		for _, value := range requestCookies {
			parts := strings.SplitN(value, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				fmt.Printf("Invalid cookie: %s\n", value)
				return
			}

			cookie := &http.Cookie{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			}

			for _, rawURL := range args {
				u, err := url.Parse(rawURL)
				if err != nil {
					fmt.Printf("Invalid URL: %s\n", rawURL)
					return
				}

				jar.SetCookies(u, []*http.Cookie{cookie})
			}
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
				Headers:  headers,
				Jar:      jar,
			}

			err := downloadUrl(req)
			if err != nil {
				fmt.Printf("Error downloading: %v\n", err)
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
	rootCmd.Flags().IntVarP(&maxRetries, "max-retries", "r", 4, "Total retries after connection failed")
	rootCmd.Flags().IntVarP(&maxWorkers, "max-workers", "W", 8, "Total concurrent workers")
	rootCmd.Flags().IntVarP(&maxChunks, "max-chunks", "c", 12, "Total parts of download")
	rootCmd.Flags().StringVarP(&protocol, "protocol", "", "auto", "Protocol to use")
	rootCmd.Flags().StringVar(&downloadProxyServerString, "set-proxy", "", "Set proxy server")

	rootCmd.Flags().StringArrayVar(
		&requestHeaders,
		"header",
		nil,
		"Add HTTP header (Header: Value)",
	)

	rootCmd.Flags().StringArrayVar(
		&requestCookies,
		"cookie",
		nil,
		"Add HTTP cookie (name=value)",
	)

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(commandsHelp)
	})

	proxyCmd.Flags().IntVarP(&port, "port", "", 8000, "Port of the proxy")
	proxyCmd.Flags().StringVar(&upstreamProxyServerString, "set-proxy", "", "Set upstream proxy server")

	proxyCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(proxyHelp)
	})

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(updateCmd)
}
