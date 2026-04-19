package app

import (
	"sync"
	"time"
)

type DevTools struct {
	mu         sync.Mutex
	Enabled    bool
	Logs       []string
	MaxLogs    int
	EventTrace []EventMark
	FocusedID  string
	UITree     string
	ActiveTab  int
}

type EventMark struct {
	ID        int64
	Timestamp time.Time
	Type      string
	Target    string
	Consumed  bool
	Details   string
}

func NewDevTools() *DevTools {
	return &DevTools{
		MaxLogs: 50,
	}
}

func (d *DevTools) AddLog(msg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Logs = append(d.Logs, msg)
	if len(d.Logs) > d.MaxLogs {
		d.Logs = d.Logs[1:]
	}
}

func (d *DevTools) TraceEvent(mark EventMark) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.EventTrace = append(d.EventTrace, mark)
	if len(d.EventTrace) > 20 {
		d.EventTrace = d.EventTrace[1:]
	}
}

func (d *DevTools) SetFocus(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.FocusedID = id
}

func (d *DevTools) Toggle() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Enabled = !d.Enabled
}
