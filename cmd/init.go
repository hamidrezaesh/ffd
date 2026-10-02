package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Flags().IntVarP(&wait, "wait", "w", 0, "Wait before starting download (seconds)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Custom Filename")
	rootCmd.Flags().StringVarP(&path, "path", "p", ".", "Download Directory")
	rootCmd.Flags().IntVarP(&maxRetries, "max-retries", "r", 4, "Total retries after connection failed")
	rootCmd.Flags().IntVarP(&maxWorkers, "max-workers", "W", 8, "Total concurrent workers")
	rootCmd.Flags().IntVarP(&maxChunks, "max-chunks", "c", 12, "Total parts of download")
	rootCmd.Flags().StringVarP(&protocol, "protocol", "", "auto", "Protocol to use")
	rootCmd.Flags().StringVar(&downloadProxyServerString, "set-proxy", "", "Set proxy server")
	rootCmd.Flags().BoolVarP(&overwrite, "overwrite", "", false, "Overwrite downloaded file and start fresh")

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

	rootCmd.Flags().StringVar(
		&requestCookieFile,
		"cookie-file",
		"",
		"Load HTTP cookies from a Netscape cookie file",
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
