// Package app provides the core application lifecycle for Gödel anew,
// wrapped entirely around CogentCore's native WebGPU systems.
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"

	"github.com/intercode/godel/pkg/config"
	"github.com/intercode/godel/pkg/input"
	"github.com/intercode/godel/pkg/ui"
)

// App is the top-level Gödel application binding for CogentCore
type App struct {
	mu sync.Mutex

	cfg *Config

	body *core.Body

	devtools *DevTools

	onReady    []func(context.Context) error
	onClose    []func(context.Context) error
	onSimulate []func(context.Context) error

	taskQueue chan func()
	inputState *input.State

	// Simulator vars
	simActive bool
	simStatus string
	simReport []string

	nextEventID atomic.Int64
}

// Config holds the app configuration
type Config struct {
	Title         string
	Width         int
	Height        int
	Resizable     bool
	VSync         bool
	FrameRate     int
	EventDriven   bool 
	MinWidth      int
	MinHeight     int
	ConfigFile    string
}

func DefaultConfig() *Config {
	return &Config{
		Title:       "Gödel App",
		Width:       800,
		Height:      600,
		Resizable:   true,
		VSync:       true,
		FrameRate:   60,
		EventDriven: true,
	}
}


func New(opts ...Option) *App {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	app := &App{
		cfg:        cfg,
		taskQueue:  make(chan func(), 1024),
		devtools:   NewDevTools(),
		inputState: input.New(),
	}

	if os.Getenv("GODEL_DEBUG") == "1" {
		app.devtools.Enabled = true
	}

	return app
}

func NewFromConfig(path string) (*App, error) {
	godelCfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", path, err)
	}

	opts := []Option{
		WithTitle(godelCfg.App.Name),
		WithSize(godelCfg.Build.Width, godelCfg.Build.Height),
	}

	return New(opts...), nil
}

func (a *App) OnReady(fn func(context.Context) error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onReady = append(a.onReady, fn)
}

func (a *App) OnClose(fn func(context.Context) error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onClose = append(a.onClose, fn)
}

func (a *App) OnSimulate(fn func(context.Context) error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onSimulate = append(a.onSimulate, fn)
}

// SetRoot takes our declarative ui.Widget or ui.WidgetBuilder and binds it to the core Body
func (a *App) SetRoot(root any) {
	if a.body == nil {
		return
	}
	switch r := root.(type) {
	case ui.Widget:
		r.Build(a.body)
	case ui.WidgetBuilder:
		r(a.body)
	case func(body *core.Body):
		r(a.body)
	default:
		log.Printf("godel: unknown root type: %T", root)
	}
}

func (a *App) QueueCallback(fn func()) {
	select {
	case a.taskQueue <- fn:
	default:
		log.Println("godel: task queue full")
	}
}

func (a *App) RequestRedraw() {
	if a.body != nil {
		a.body.Update()
	}
}

// Runtime entry
func (a *App) Run(ctx ...context.Context) error {
	appCtx := context.Background()
	if len(ctx) > 0 {
		appCtx = ctx[0]
	}

	a.body = core.NewBody(a.cfg.Title)

	// Fire OnReady hooks where SetRoot is naturally called
	for _, fn := range a.onReady {
		if err := fn(appCtx); err != nil {
			return fmt.Errorf("godel: OnReady hook failed: %w", err)
		}
	}

	// In CogentCore, we intercept shortcuts using window events via an overall binding.
	// We bind to the body to simulate our DevTools toggles
	a.body.OnKeyChord(func(e events.Event) {
		// Just intercept events. For now, we simulate basic intercepts.
		// a.dequeueTaskQueue() is missing here, usually this is a loop or custom ticker.
	})

	// Setup task drainage ticker since CogentCore owns the main loop
	go func() {
		for {
			a.drainTaskQueue()
			time.Sleep(16 * time.Millisecond) // ~60 FPS queue drainage
		}
	}()

	a.body.RunMainWindow()

	// Clean exit
	for _, fn := range a.onClose {
		_ = fn(appCtx)
	}
	return nil
}

func (a *App) drainTaskQueue() {
	for {
		select {
		case task := <-a.taskQueue:
			task()
		default:
			return
		}
	}
}

// Simulator Subsystems (Stubs for direct CogentCore Mapping)
func (a *App) SimulateClick(x, y float64) {}
func (a *App) SimulateType(text string) {}
func (a *App) AutoExplore() {}
func (a *App) SimulateClickOn(id string) {}
func (a *App) SetSimStatus(status string) {}
func (a *App) LogSimStep(msg string) {}
func (a *App) LogSimWarning(msg string) {}
func (a *App) SaveReport() error { return nil }
