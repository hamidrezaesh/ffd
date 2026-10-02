package state

import (
	"encoding/json"
	"fmt"
	"os"
)

type State struct {
	path string
	data DownloadState
}

func GetStatePath(filePath string) string {
	return filePath + ".ffd"
}

func Init(path string, data DownloadState) (*State, error) {
	state := &State{
		path: path,
		data: data,
	}

	if err := state.Save(); err != nil {
		return nil, err
	}

	return state, nil
}

func Load(path string) (*State, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data DownloadState

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid state file: %w", err)
	}

	return &State{
		path: path,
		data: data,
	}, nil
}

func (s *State) AddChunk(chunk ChunkState) {
	for _, existing := range s.data.Chunks {
		if existing.Index == chunk.Index {
			return
		}
	}

	s.data.Chunks = append(s.data.Chunks, chunk)
}

func (s *State) UpdateChunk(index int, downloaded int64) error {
	for i := range s.data.Chunks {
		if s.data.Chunks[i].Index != index {
			continue
		}

		if downloaded < 0 || downloaded > s.data.Chunks[i].Size {
			return fmt.Errorf("invalid downloaded value: %d", downloaded)
		}

		s.data.Chunks[i].Downloaded = downloaded
		return nil
	}

	return fmt.Errorf("chunk not found: %d", index)
}

func (s *State) Save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *State) Data() DownloadState {
	return s.data
}
