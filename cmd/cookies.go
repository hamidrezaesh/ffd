package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type netscapeCookie struct {
	Cookie *http.Cookie
	Domain string
}

func parseCookieFile(path string) ([]netscapeCookie, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cookies []netscapeCookie

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		httpOnly := false

		if strings.HasPrefix(line, "#HttpOnly_") {
			httpOnly = true
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) != 7 {
			return nil, fmt.Errorf("invalid cookie file line: %q", line)
		}

		domain := fields[0]
		includeSubdomains := fields[1]
		path := fields[2]
		secure := fields[3]
		expiration := fields[4]
		name := fields[5]
		value := fields[6]

		if domain == "" || name == "" {
			return nil, fmt.Errorf("invalid cookie file line: %q", line)
		}

		if includeSubdomains != "TRUE" && includeSubdomains != "FALSE" {
			return nil, fmt.Errorf(
				"invalid include-subdomains value %q in line: %q",
				includeSubdomains,
				line,
			)
		}

		if secure != "TRUE" && secure != "FALSE" {
			return nil, fmt.Errorf(
				"invalid secure value %q in line: %q",
				secure,
				line,
			)
		}

		expiresUnix, err := strconv.ParseInt(expiration, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid cookie expiration %q: %w",
				expiration,
				err,
			)
		}

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			Secure:   secure == "TRUE",
			HttpOnly: httpOnly,
		}

		if expiresUnix > 0 {
			cookie.Expires = time.Unix(expiresUnix, 0)
		}

		cookies = append(cookies, netscapeCookie{
			Cookie: cookie,
			Domain: domain,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading cookie file: %w", err)
	}

	return cookies, nil
}

func loadCookieFile(
	jar http.CookieJar,
	path string,
	targetURL *url.URL,
) error {
	cookies, err := parseCookieFile(path)
	if err != nil {
		return err
	}

	for _, cookie := range cookies {
		jar.SetCookies(targetURL, []*http.Cookie{cookie.Cookie})
	}

	return nil
}
