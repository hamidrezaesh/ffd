package tracker

import (
	"sync"
	"sync/atomic"
	"time"
)

type ProgressInfo struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Speed      int64   `json:"speed"`
	Percent    float64 `json:"percent"`
}

type Progress struct {
	Downloaded atomic.Int64
	Speed      atomic.Int64
	Total      int64

	done     chan struct{}
	stopOnce sync.Once
}

func New(total int64) *Progress {
	return &Progress{
		Total: total,
		done:  make(chan struct{}),
	}
}

func (p *Progress) AddDownloaded(n int64) {
	p.Downloaded.Add(n)
}

func (p *Progress) Start() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var previous int64

		for {
			select {
			case <-ticker.C:
				current := p.Downloaded.Load()

				p.Speed.Store(current - previous)

				previous = current

			case <-p.done:
				return
			}
		}
	}()
}

func (p *Progress) Stop() {
	p.stopOnce.Do(func() {
		close(p.done)
	})
}

func (p *Progress) SetDownloaded(n int64) {
	p.Downloaded.Store(n)
}

func (p *Progress) GetDownloaded() int64 {
	return p.Downloaded.Load()
}

func (p *Progress) GetSpeed() int64 {
	return p.Speed.Load()
}

func (p *Progress) GetTotal() int64 {
	return p.Total
}

func (p *Progress) GetRemaining() int64 {
	return p.Total - p.Downloaded.Load()
}

func (p *Progress) GetPercent() float64 {
	if p.Total == 0 {
		return 0
	}

	downloaded := p.Downloaded.Load()

	if downloaded >= p.Total {
		return 100
	}

	return float64(downloaded) / float64(p.Total) * 100
}

func (p *Progress) Info() ProgressInfo {
	return ProgressInfo{
		Downloaded: p.GetDownloaded(),
		Total:      p.GetTotal(),
		Speed:      p.GetSpeed(),
		Percent:    p.GetPercent(),
	}
}
