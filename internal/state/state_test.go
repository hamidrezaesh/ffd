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

	initialState := DownloadState{
		Filename:  "test.bin",
		URL:       "https://example.com/test.bin",
		TotalSize: 300 * 1024 * 1024,
		Protocol:  "HTTP/1.1",
		Chunks:    []ChunkState{},
	}

	// 1. Init
	state, err := Init(statePath, initialState)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, state.Data())

	// 2. Add first chunk
	state.AddChunk(ChunkState{
		Index:      0,
		Offset:     0,
		Size:       100 * 1024 * 1024,
		Downloaded: 0,
	})

	if err := state.Save(); err != nil {
		t.Fatal(err)
	}

	printJSON(t, state.Data())

	// 3. Add second chunk
	state.AddChunk(ChunkState{
		Index:      1,
		Offset:     100 * 1024 * 1024,
		Size:       100 * 1024 * 1024,
		Downloaded: 0,
	})

	if err := state.Save(); err != nil {
		t.Fatal(err)
	}

	printJSON(t, state.Data())

	// 4. Add existing chunk
	state.AddChunk(ChunkState{
		Index:      0,
		Offset:     0,
		Size:       100 * 1024 * 1024,
		Downloaded: 123,
	})

	if len(state.Data().Chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(state.Data().Chunks))
	}

	// 5. Update first chunk
	if err := state.UpdateChunk(0, 50*1024*1024); err != nil {
		t.Fatal(err)
	}

	if err := state.Save(); err != nil {
		t.Fatal(err)
	}

	printJSON(t, state.Data())

	// 6. Load state again
	loaded, err := Load(statePath)
	if err != nil {
		t.Fatal(err)
	}

	printJSON(t, loaded.Data())
}
