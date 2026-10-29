package tools

/* this function will validate the input url. for example if
the url doesn't start with http or https, this will add it.
if url starts with 127* or 192* or localhost* it will add http (and
https or 192*).
*/

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	hasScheme := strings.Contains(raw, "://")

	if !hasScheme {
		host := raw
		if i := strings.IndexByte(host, '/'); i >= 0 {
			host = host[:i]
		}
		if i := strings.IndexByte(host, '?'); i >= 0 {
			host = host[:i]
		}
		if i := strings.IndexByte(host, '#'); i >= 0 {
			host = host[:i]
		}

		// Remove user info, if present.
		if i := strings.LastIndexByte(host, '@'); i >= 0 {
			host = host[i+1:]
		}

		// Extract hostname, including hosts with ports.
		hostname := host
		if h, _, err := net.SplitHostPort(host); err == nil {
			hostname = h
		} else if strings.HasPrefix(host, "[") {
			if end := strings.IndexByte(host, ']'); end >= 0 {
				hostname = host[1:end]
			}
		}

		isLocal := strings.EqualFold(hostname, "localhost")
		if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
			isLocal = true
		}

		scheme := "https://"
		if isLocal {
			scheme = "http://"
		}
		raw = scheme + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("URL must contain a hostname")
	}

	// Local addresses use HTTP, even if HTTPS was supplied.
	hostname := u.Hostname()
	isLocal := strings.EqualFold(hostname, "localhost")
	if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
		isLocal = true
	}
	if isLocal {
		u.Scheme = "http"
	}

	return u.String(), nil
}

func validateURL(raw string) (string, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("URL must contain a hostname")
	}
	return u.String(), nil
}
