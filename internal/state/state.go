package state

type DownloadState struct {
	URL       string       `json:"url"`
	Filename  string       `json:"filename"`
	TotalSize int64        `json:"total_size"`
	Protocol  string       `json:"protocol"`
	Chunks    []ChunkState `json:"chunks"`
}

type ChunkState struct {
	Index      int   `json:"index"`
	Offset     int64 `json:"offset"`
	Size       int64 `json:"size"`
	Downloaded int64 `json:"downloaded"`
}
