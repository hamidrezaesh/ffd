package scheduler

import (
	"testing"
	"time"

	"github.com/hamidrezaesh/ffd/internal/tools"
)

func TestTestProtocol(t *testing.T) {
	s := Scheduler{
		Client:       nil,
		MaxRetries:   6,
		MinFileSize:  10 * 1024 * 1024,
		Progress:     nil,
		URL:          "https://releases.ubuntu.com/26.04.1/ubuntu-26.04.1-desktop-amd64.iso",
		TotalSize:    6482409472,
		MaxWorkers:   0,
		MaxChunks:    0,
		ResumeChunks: nil,
		OnPlan:       nil,
	}

	timeout := tools.Clamp(s.TotalSize/(100*1024), 100, 1000)
	t.Logf("Timeout: %v\n", timeout)

	start := time.Now()

	finalProtocol, startByte, err := s.testProtocol(nil, nil, 0, nil)
	if err != nil {
		t.Fatalf("TestProtocol returned an error: %v", err)
	}

	elapsed := time.Since(start)

	t.Logf("TestProtocol took %v and downloaded %v bytes and the final protocol is %v", elapsed, startByte, finalProtocol)
}
