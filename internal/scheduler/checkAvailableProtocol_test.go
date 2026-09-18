package scheduler

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func TestCheckHTTP1(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	client := &http.Client{
		Transport: &http.Transport{
			TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
		},
		Timeout: 5 * time.Second,
	}

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/1.1 available: %v", supported)
}

func TestCheckHTTP2(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
		},
		Timeout: 5 * time.Second,
	}

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/2 available: %v", supported)
}

func TestCheckHTTP3(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	transport := &http3.Transport{}
	defer transport.Close()

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/3 available: %v", supported)
}
