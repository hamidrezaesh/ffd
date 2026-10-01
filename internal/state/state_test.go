package state

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func printJSON(t *testing.T, v any) {
	t.Helper()

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestState(t *testing.T) {
	tmpDir := t.TempDir()

	statePath := filepath.Join(tmpDir, "test.bin.ffd")

	state := DownloadState{
		Filename:  "test.bin",
		URL:       "https://example.com/test.bin",
		TotalSize: 300 * 1024 * 1024,
		Chunks:    []ChunkState{},
	}

	// 1. Init
	if err := Init(statePath, state); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(statePath)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, loaded)

	// 2. Add first chunk
	if err := AddChunk(statePath, ChunkState{
		Index:      0,
		Offset:     0,
		Size:       100 * 1024 * 1024,
		Downloaded: 0,
	}); err != nil {
		t.Fatal(err)
	}

	loaded, err = Load(statePath)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, loaded)

	// 3. Add second chunk
	if err := AddChunk(statePath, ChunkState{
		Index:      1,
		Offset:     100 * 1024 * 1024,
		Size:       100 * 1024 * 1024,
		Downloaded: 0,
	}); err != nil {
		t.Fatal(err)
	}

	loaded, err = Load(statePath)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, loaded)

	// 4. Update first chunk
	if err := UpdateChunk(statePath, 0, 50*1024*1024); err != nil {
		t.Fatal(err)
	}

	loaded, err = Load(statePath)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, loaded)
}
