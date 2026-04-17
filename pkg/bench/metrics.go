package bench

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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
	PID       int
	startTime time.Time
	frames    []time.Duration
}

func NewCollector() *Collector {
	return &Collector{
		PID:       os.Getpid(),
		startTime: time.Now(),
	}
}

// GetRSS returns the current RSS of the process in bytes.
func GetRSS(pid int) uint64 {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "rss=")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			rssStr := strings.TrimSpace(out.String())
			if rss, err := strconv.ParseUint(rssStr, 10, 64); err == nil {
				return rss * 1024 // ps returns in KB
			}
		}
	}
	// Fallback to runtime stats if ps fails or on windows
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Sys
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

	return Metrics{
		StartupTime:   time.Since(c.startTime),
		MeanFrameTime: mean,
		MaxFrameTime:  max,
		MemoryRSS:     GetRSS(c.PID),
		CgoCalls:      0, // Logic to read NumCgoCall if needed
	}
}
