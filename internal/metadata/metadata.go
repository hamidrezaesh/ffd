package metadata

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Metadata struct {
	URL          string
	TotalSize    int64
	AcceptRanges bool
	Filename     string
	Ext          string
}

func extractor(filename string) (string, string) {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	return name, ext
}

func resolve(finalURL string, resp *http.Response) (string, string, error) {
	// Content-Disposition
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			filename := params["filename"]
			if filename != "" {
				filename, ext := extractor(filename)
				return filename, ext, nil
			}
		}
	}

	// URL Path (Alternative)
	parsedURL, err := url.Parse(finalURL)
	if err == nil {
		base := path.Base(parsedURL.Path)
		if strings.Contains(base, ".") {
			filename, ext := extractor(base)
			return filename, ext, nil
		}
	}

	// Content-Type (Alternative 2)
	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		exts, _ := mime.ExtensionsByType(ct)
		if len(exts) > 0 {
			return "download_file", exts[0], nil
		}
	}

	return "download_file", ".bin", nil
}

func FetchMetadata(url string, resp *http.Response) (Metadata, error) {
	// Check status
	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Check if this is actually a downloadable file
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return Metadata{}, fmt.Errorf("URL does not point to a downloadable file")
	}

	// Parse Content-Length
	if resp.Header.Get("Content-Length") == "" {
		return Metadata{}, fmt.Errorf("server did not provide Content-Length")
	}
	totalSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return Metadata{}, err
	}

	// Check Accept-Ranges
	acceptRanges := resp.Header.Get("Accept-Ranges") == "bytes"

	// extract fileinfo
	filename, ext, err := resolve(url, resp)
	if err != nil {
		return Metadata{}, err
	}

	return Metadata{
		URL:          url,
		TotalSize:    totalSize,
		AcceptRanges: acceptRanges,
		Filename:     filename,
		Ext:          ext,
	}, nil
}
