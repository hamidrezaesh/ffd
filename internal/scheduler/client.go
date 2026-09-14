package scheduler

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
)

/*
newClient is responsible for creating an http client base on finalProtocol. it will use by Download
to return best http client.
*/

func newClient(protocol int) *http.Client {
	switch protocol {
	case 1: // HTTP/1.1
		transport := &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			IdleConnTimeout:     90 * time.Second,
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
		}

		return &http.Client{
			Transport: transport,
		}

	case 2: // HTTP/2
		transport := &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		}

		return &http.Client{
			Transport: transport,
		}

	case 3: // HTTP/3
		transport := &http3.Transport{}

		return &http.Client{
			Transport: transport,
		}
	}

	return http.DefaultClient
}
