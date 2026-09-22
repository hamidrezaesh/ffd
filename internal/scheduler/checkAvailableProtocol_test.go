package scheduler

import (
	"testing"
)

func TestCheckHTTP1(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	client := newClient(1, nil, nil)

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/1.1 available: %v", supported)
}

func TestCheckHTTP2(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	client := newClient(2, nil, nil)

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/2 available: %v", supported)
}

func TestCheckHTTP3(t *testing.T) {
	url := "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso"

	client := newClient(3, nil, nil)

	supported := checkAvailableProtocol(url, client)

	t.Logf("HTTP/3 available: %v", supported)
}
