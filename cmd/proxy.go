package cmd

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/hamidrezaesh/ffd/internal/proxy"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the ffd forward proxy",
	Run: func(cmd *cobra.Command, args []string) {
		// Parse upstream proxy.
		var proxyServer *url.URL

		if upstreamProxyServerString != "" {
			var err error
			proxyServer, err = url.Parse(upstreamProxyServerString)
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}
		}

		// Prepare headers.
		headers := scheduler.Headers{}

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

		proxy.Start(proxy.Server{
			Addr:              fmt.Sprintf("127.0.0.1:%v", port),
			Proxy:             proxyServer,
			Headers:           headers,
			CookieJar:         jar,
			MaxRetries:        maxRetries,
			MaxWorkers:        maxWorkers,
			MaxChunks:         maxChunks,
			PreferredProtocol: 0,
		})
	},
}
