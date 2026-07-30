package engine

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/metadata"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/hamidrezaesh/ffd/internal/tracker"
	"github.com/hamidrezaesh/ffd/internal/validator"
)

type Request struct {
	URL      string
	Path     string
	Filename string
}

type Result struct {
	Metadata metadata.Metadata
	Filename string
	Progress *tracker.Progress
	Done     chan error
}

func Download(req Request, maxRetries int, maxWorkers int, maxChunks int) (*Result, error) {
	if maxRetries == 0 {
		maxRetries = 4
	}
	if maxWorkers == 0 {
		maxWorkers = 8
	}
	if maxChunks == 0 {
		maxChunks = 12
	}
	if maxWorkers > maxChunks {
		maxWorkers = maxChunks
	}

	transport := &http.Transport{
		MaxIdleConns:        16,
		MaxIdleConnsPerHost: 16,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Get response
	resp, err := client.Head(req.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Fetch metadata
	md, err := metadata.FetchMetadata(req.URL, resp)
	if err != nil {
		return nil, err
	}

	// Use custom filename if provided
	filename := md.Filename

	if req.Filename != "" {
		filename = req.Filename
	}

	// Validate filename
	if err := validator.Filename(filename); err != nil {
		return nil, err
	}

	// Allocate file
	fileInfo := disk.FileInfo{
		Filename:  filename + md.Ext,
		Path:      filepath.Join(req.Path, filename+md.Ext),
		TotalSize: md.TotalSize,
	}

	file, err := disk.Allocate(fileInfo)
	if err != nil {
		return nil, err
	}

	// Create progress tracker
	progress := tracker.New(md.TotalSize)
	progress.Start()

	result := &Result{
		Metadata: md,
		Filename: filename,
		Progress: progress,
		Done:     make(chan error, 1),
	}

	go func() {
		defer file.Close()
		defer progress.Stop()

		err := scheduler.Download(
			req.URL,
			file,
			md.TotalSize,
			md.AcceptRanges,
			5*1024*1024,
			maxWorkers,
			maxChunks,
			client,
			progress,
			maxRetries,
		)

		result.Done <- err
		close(result.Done)
	}()

	return result, nil
}
