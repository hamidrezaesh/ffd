package engine

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/metadata"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/hamidrezaesh/ffd/internal/state"
	"github.com/hamidrezaesh/ffd/internal/tracker"
	"github.com/hamidrezaesh/ffd/internal/validator"
)

type Request struct {
	URL      string
	Path     string
	Filename string
	Headers  scheduler.Headers
	Jar      http.CookieJar
}

type Result struct {
	Metadata metadata.Metadata
	Filename string
	Progress *tracker.Progress
	Done     chan error
}

func Download(
	req Request,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
	preferredProtocol int,
	proxyServer *url.URL,
) (*Result, error) {
	if maxRetries == 0 {
		maxRetries = 4
	}

	if maxWorkers > maxChunks {
		maxWorkers = maxChunks
	}

	client := &scheduler.HeaderClient{
		Base: &http.Client{
			Jar: req.Jar,
		},
		Headers: req.Headers,
	}

	// Get a response.
	resp, err := client.Head(req.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Fetch metadata.
	md, err := metadata.FetchMetadata(req.URL, resp)
	if err != nil {
		return nil, err
	}

	// Use custom filename if provided.
	filename := md.Filename
	if req.Filename != "" {
		filename = req.Filename
	}

	// Validate filename.
	if err := validator.Filename(filename); err != nil {
		return nil, err
	}

	// define path
	filePath := filepath.Join(req.Path, filename+md.Ext)

	fileInfo := disk.FileInfo{
		Filename:  filename + md.Ext,
		Path:      filePath,
		TotalSize: md.TotalSize,
	}

	// check and initial state
	statePath := state.GetStatePath(filePath)
	downloadState, err := state.Load(statePath)
	if err != nil {
		if os.IsNotExist(err) { // if state doesnt exists, initial it.
			downloadState, err = state.Init(statePath, state.DownloadState{
				URL:       req.URL,
				Filename:  filename,
				TotalSize: fileInfo.TotalSize,
				Chunks:    []state.ChunkState{},
			})
		}

		if err != nil {
			return nil, err
		}
	}

	download, err := newDownload(fileInfo)
	if err != nil {
		return nil, err
	}

	// initial progress tracker
	progress := tracker.New(md.TotalSize)

	// add downloaded bytes to tracker
	for _, chunk := range downloadState.Data().Chunks {
		if chunk.Downloaded > 0 {
			progress.AddDownloaded(chunk.Downloaded)
		}
	}

	// save downloaded chunks to resumeChunks
	resumeChunks := make([]scheduler.ResumeChunk, 0, len(downloadState.Data().Chunks))

	for _, chunk := range downloadState.Data().Chunks {
		resumeChunks = append(resumeChunks, scheduler.ResumeChunk{
			Index: chunk.Index,
			Range: scheduler.ByteRange{
				Start: chunk.Offset,
				End:   chunk.Offset + chunk.Size - 1,
			},
			Downloaded: chunk.Downloaded,
		})
	}

	// start tracker
	progress.Start()

	result := &Result{
		Metadata: md,
		Filename: filename,
		Progress: progress,
		Done:     make(chan error, 1),
	}

	go func() {
		defer progress.Stop()

		chanChunks, chanErr := scheduler.Download(
			req.URL,
			md.TotalSize,
			proxyServer,
			req.Headers,
			req.Jar,
			progress,
			md.AcceptRanges,
			maxRetries,
			maxWorkers,
			maxChunks,
			preferredProtocol,
			resumeChunks,
			func(plan []scheduler.ResumeChunk) error {
				chunks := make([]state.ChunkState, 0, len(plan))

				for _, r := range plan {
					chunks = append(chunks, state.ChunkState{
						Index:      r.Index,
						Offset:     r.Range.Start,
						Size:       r.Range.End - r.Range.Start + 1,
						Downloaded: 0,
					})
				}

				downloadState.SetPlan(chunks)

				return downloadState.Save()
			},
		)

		for chanChunks != nil || chanErr != nil {
			select {
			case chunk, ok := <-chanChunks:
				if !ok {
					chanChunks = nil
					continue
				}

				if err := download.Write(chunk); err != nil {
					_ = download.Close()
					result.Done <- err
					return
				}

				// if chunks are test chunks, don't add them into state
				if !chunk.Test {
					downloadState.AddChunk(state.ChunkState{
						Index:      chunk.Index,
						Offset:     chunk.RangeStart,
						Size:       chunk.RangeSize,
						Downloaded: 0,
					})

					downloaded := chunk.Offset - chunk.RangeStart + int64(len(chunk.Bytes))

					if err := downloadState.UpdateChunk(
						chunk.RangeStart,
						downloaded,
					); err != nil {
						_ = download.Close()
						result.Done <- err
						return
					}

					if err := downloadState.Save(); err != nil {
						_ = download.Close()
						result.Done <- err
						return
					}
				}

			case err, ok := <-chanErr:
				if !ok {
					chanErr = nil
					continue
				}

				if err != nil {
					_ = download.Close()
					result.Done <- err
					return
				}
			}
		}

		if err := download.Close(); err != nil {
			result.Done <- err
			return
		}

		result.Done <- nil
	}()

	return result, nil
}
