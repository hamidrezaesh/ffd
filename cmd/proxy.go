package cmd

import (
	"fmt"
	"net/url"

	"github.com/hamidrezaesh/ffd/internal/proxy"
	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the ffd forward proxy",
	Run: func(cmd *cobra.Command, args []string) {
		var proxyServer *url.URL
		var err error
		if upstreamProxyServerString != "" {
			proxyServer, err = url.Parse(upstreamProxyServerString)
			if err != nil {
				fmt.Printf(
					"\r\033[KFailed: %v\n",
					err,
				)
				return
			}
		}
		proxy.Start(port, proxyServer)
	},
}
