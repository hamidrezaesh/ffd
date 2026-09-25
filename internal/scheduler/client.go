package scheduler

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go/http3"
)

type Header struct {
	Key   string
	Value string
}

type Headers []Header

type HeaderClient struct {
	Base    *http.Client
	Headers Headers
}

func (c *HeaderClient) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *HeaderClient) Do(req *http.Request) (*http.Response, error) {
	for _, header := range c.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	return c.Base.Do(req)
}

/*
newClient is responsible for creating an HTTP client based on finalProtocol.
It is used by Download to return the selected HTTP client.
*/
func newClient(
	protocol int,
	proxyServer *url.URL,
	headers Headers,
	jar http.CookieJar,
) *HeaderClient {
	var client *http.Client

	switch protocol {
	case 1: // HTTP/1.1
		transport := &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			IdleConnTimeout:     90 * time.Second,
			TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
		}

		if proxyServer != nil {
			transport.Proxy = http.ProxyURL(proxyServer)
		}

		client = &http.Client{
			Transport: transport,
			Jar:       jar,
		}

	case 2: // HTTP/2
		transport := &http.Transport{
			MaxIdleConns:        128,
			MaxIdleConnsPerHost: 64,
			MaxConnsPerHost:     64,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		}

		if proxyServer != nil {
			transport.Proxy = http.ProxyURL(proxyServer)
		}

		client = &http.Client{
			Transport: transport,
			Jar:       jar,
		}

	case 3: // HTTP/3
		transport := &http3.Transport{}

		client = &http.Client{
			Transport: transport,
			Jar:       jar,
		}

	default:
		client = http.DefaultClient
	}

	return &HeaderClient{
		Base:    client,
		Headers: headers,
	}
}
