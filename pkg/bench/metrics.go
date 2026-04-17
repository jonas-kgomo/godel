package bench

import (
	"runtime"
	"time"
)

// Metrics represents performance data for a Gödel application.
type Metrics struct {
	StartupTime      time.Duration
	MeanFrameTime    time.Duration
	MaxFrameTime     time.Duration
	MemoryRSS        uint64 // in bytes
	BinarySize       int64  // in bytes
	CgoCalls         int64
}

// Collector gathers metrics during app execution.
type Collector struct {
	startTime time.Time
	frames    []time.Duration
}

func NewCollector() *Collector {
	return &Collector{
		startTime: time.Now(),
	}
}

func (c *Collector) RecordFrame(d time.Duration) {
	c.frames = append(c.frames, d)
}

func (c *Collector) Stats() Metrics {
	var total time.Duration
	var max time.Duration
	for _, f := range c.frames {
		total += f
		if f > max {
			max = f
		}
	}

	mean := time.Duration(0)
	if len(c.frames) > 0 {
		mean = total / time.Duration(len(c.frames))
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return Metrics{
		StartupTime:   time.Since(c.startTime),
		MeanFrameTime: mean,
		MaxFrameTime:  max,
		MemoryRSS:     m.Sys,
		CgoCalls:      runtime.NumCgoCall(),
	}
}
