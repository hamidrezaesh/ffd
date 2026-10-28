package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hamidrezaesh/ffd/internal/disk"
	"github.com/hamidrezaesh/ffd/internal/metadata"
	"github.com/hamidrezaesh/ffd/internal/scheduler"
	"github.com/hamidrezaesh/ffd/internal/state"
	"github.com/hamidrezaesh/ffd/internal/tools"
	"github.com/hamidrezaesh/ffd/internal/tracker"
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

func hashURL(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])
}

func Download(
	req Request,
	maxRetries int,
	maxWorkers int,
	maxChunks int,
	preferredProtocol int,
	proxyServer *url.URL,
	overwrite bool,
	minFileSizeToTest int64,
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

	// Use the URL hash for the temporary download and its state.
	urlHash := hashURL(req.URL)
	tempPath := filepath.Join(req.Path, urlHash+".part")
	statePath := state.GetStatePath(tempPath)

	// Check whether this download can be resumed.
	var downloadState *state.State
	resumeExisting := false

	if overwrite {
		_ = os.Remove(statePath)
		_ = os.Remove(tempPath)
	} else {
		loadedState, err := state.Load(statePath)
		if err == nil {
			_, fileErr := os.Stat(tempPath)
			if fileErr == nil {
				downloadState = loadedState
				resumeExisting = true
			} else if !os.IsNotExist(fileErr) {
				return nil, fileErr
			} else {
				// The state exists, but its temporary file does not.
				_ = os.Remove(statePath)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}

	var (
		startByte     int64
		resp          *http.Response
		md            metadata.Metadata
		err           error
		initialChunks []scheduler.Chunk
	)

	if resumeExisting {
		// Don't start another GET when resuming an existing partial file.
		resp, err = client.Head(req.URL)
		if err != nil {
			return nil, err
		}

		md, err = metadata.FetchMetadata(req.URL, resp)
		if err != nil {
			_ = resp.Body.Close()
			return nil, err
		}

		_ = resp.Body.Close()

		saved := downloadState.Data()
		if saved.URL != req.URL {
			return nil, fmt.Errorf("saved download URL does not match")
		}
		if saved.TotalSize != md.TotalSize {
			return nil, fmt.Errorf("remote file size changed; cannot safely resume")
		}
	} else {
		// Download initial bytes while fetching metadata.
		startByte, resp, md, err = scheduler.DownloadInitial(
			client,
			req.URL,
			func(chunk scheduler.Chunk) error {
				initialChunks = append(initialChunks, chunk)
				return nil
			},
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	// Choose the final filename.
	filename := md.Filename
	if req.Filename != "" {
		filename = req.Filename
	}

	if resumeExisting {
		// Keep the original final filename across restarts.
		filename = downloadState.Data().Filename
	}

	if err := tools.Filename(filename); err != nil {
		return nil, err
	}

	// Only choose a new unique name for a fresh download.
	if !overwrite && !resumeExisting {
		filename = tools.UniqueFilename(req.Path, filename, md.Ext)
	}

	finalPath := filepath.Join(req.Path, filename+md.Ext)

	// The downloader writes to the hash-based temporary file.
	fileInfo := disk.FileInfo{
		Filename:  urlHash + ".part",
		Path:      tempPath,
		TotalSize: md.TotalSize,
	}

	if md.TotalSize <= minFileSizeToTest {
		preferredProtocol = 1
		maxWorkers = 8
		maxChunks = 12
	}

	// Initialize state for a fresh download.
	if !resumeExisting {
		downloadState, err = state.Init(statePath, state.DownloadState{
			URL:       req.URL,
			Filename:  filename,
			TotalSize: md.TotalSize,
			Chunks:    []state.ChunkState{},
		})
		if err != nil {
			return nil, err
		}
	}

	download, err := newDownload(fileInfo)
	if err != nil {
		return nil, err
	}

	// Save the initial bytes and their resume position.
	if !resumeExisting {
		for _, chunk := range initialChunks {
			if err := download.Write(chunk); err != nil {
				_ = download.Close()
				return nil, err
			}
		}

		if startByte > 0 {
			downloadState.SetPlan([]state.ChunkState{
				{
					Index:      0,
					Offset:     0,
					Size:       md.TotalSize,
					Downloaded: startByte,
				},
			})

			if err := downloadState.Save(); err != nil {
				_ = download.Close()
				return nil, err
			}
		}
	}

	// Initialize progress with bytes already downloaded.
	progress := tracker.New(md.TotalSize)

	for _, chunk := range downloadState.Data().Chunks {
		if chunk.Downloaded > 0 {
			progress.AddDownloaded(chunk.Downloaded)
		}
	}

	// Convert saved chunks into scheduler resume chunks.
	resumeChunks := make(
		[]scheduler.ResumeChunk,
		0,
		len(downloadState.Data().Chunks),
	)

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

				if !chunk.Test {
					downloadState.AddChunk(state.ChunkState{
						Index:      chunk.Index,
						Offset:     chunk.RangeStart,
						Size:       chunk.RangeSize,
						Downloaded: 0,
					})

					downloaded := chunk.Offset -
						chunk.RangeStart +
						int64(len(chunk.Bytes))

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

		// On overwrite, replace the old final file only after the download
		// has completed successfully.
		if overwrite {
			if err := os.Remove(finalPath); err != nil &&
				!os.IsNotExist(err) {
				result.Done <- err
				return
			}
		}

		// Publish the completed file under its final filename.
		if err := os.Rename(tempPath, finalPath); err != nil {
			result.Done <- err
			return
		}

		// Remove resume state after the rename succeeds.
		if err := downloadState.RemoveState(); err != nil &&
			!os.IsNotExist(err) {
			result.Done <- err
			return
		}

		result.Done <- nil
	}()

	return result, nil
}
