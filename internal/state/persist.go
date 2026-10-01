package state

import (
	"encoding/json"
	"fmt"
	"os"
)

func Init(path string, state DownloadState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func Load(statePath string) (*DownloadState, error) {
	file, err := os.Open(statePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var state DownloadState

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("invalid state file: %w", err)
	}

	return &state, nil
}

func AddChunk(path string, chunk ChunkState) error {
	state, err := Load(path)
	if err != nil {
		return err
	}

	state.Chunks = append(state.Chunks, chunk)

	return save(path, *state)
}

func UpdateChunk(statePath string, index int, downloaded int64) error {
	state, err := Load(statePath)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(state.Chunks) {
		return fmt.Errorf("chunk index out of range: %d", index)
	}

	if downloaded < 0 || downloaded > state.Chunks[index].Size {
		return fmt.Errorf("invalid downloaded value: %d", downloaded)
	}

	state.Chunks[index].Downloaded = downloaded

	return save(statePath, *state)
}

func save(path string, state DownloadState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
