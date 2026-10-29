package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func etagForFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return `"` + hex.EncodeToString(hash.Sum(nil)) + `"`, nil
}

func fileHandler(root string) http.Handler {
	files := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			files.ServeHTTP(w, r)
			return
		}

		// Resolve the requested path under the configured directory.
		name := path.Clean("/" + r.URL.Path)
		filePath := root + string(os.PathSeparator) +
			strings.TrimPrefix(name, "/")

		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}

		etag, err := etagForFile(filePath)
		if err != nil {
			http.Error(w, "failed to calculate ETag", http.StatusInternalServerError)
			return
		}

		w.Header().Set("ETag", etag)
		w.Header().Set("Accept-Ranges", "bytes")

		// If-Match requires a strong ETag match.
		if value := r.Header.Get("If-Match"); value != "" &&
			value != "*" && !matchesETag(value, etag) {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}

		// If-Range: when the validator differs, ignore Range and send
		// the complete representation (HTTP 200), never partial bytes.
		if value := r.Header.Get("If-Range"); value != "" &&
			value != etag {
			r2 := r.Clone(r.Context())
			r2.Header = r.Header.Clone()
			r2.Header.Del("Range")
			r2.Header.Del("If-Range")
			files.ServeHTTP(w, r2)
			return
		}

		files.ServeHTTP(w, r)
	})
}

func matchesETag(header, current string) bool {
	// If-Match uses strong comparison. Weak tags never match.
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		if strings.HasPrefix(candidate, "W/") {
			continue
		}
		if candidate == current {
			return true
		}
	}
	return false
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	dir := flag.String("dir", "testdata", "Directory containing test files")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", fileHandler(*dir))

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Serving %s on http://%s", *dir, *addr)
	log.Fatal(server.ListenAndServe())
}
