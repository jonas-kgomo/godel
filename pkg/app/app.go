// Package app provides the core application lifecycle for Gödel.
// It wraps gogpu/gogpu for windowing and gogpu/ui for the widget toolkit,
// exposing a clean, opinionated API for desktop application development.
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/render"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"

	"github.com/intercode/godel/pkg/config"
)

// App is the top-level Gödel application.
// It owns the window, the UI tree, the event loop, and the task queue.
type App struct {
	mu sync.Mutex

	// Configuration
	cfg *Config

	// gogpu internals
	gogpuApp *gogpu.App
	uiApp    *uiapp.App
	canvas   *ggcanvas.Canvas

	// Lifecycle hooks
	onReady    []func(context.Context) error
	onClose    []func(context.Context) error

	// Task queue for safe goroutine → main thread communication
	taskQueue chan func()

	// Root widget
	root widget.Widget
}

// Config holds the app configuration, either from code or godel.toml.
type Config struct {
	Title         string
	Width         int
	Height        int
	MinWidth      int
	MinHeight     int
	Resizable     bool
	VSync         bool
	FrameRate     int
	ConfigFile    string // path to godel.toml
	EventDriven   bool   // false = continuous render, true = 0% CPU idle
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Title:       "Gödel App",
		Width:       800,
		Height:      600,
		MinWidth:    400,
		MinHeight:   300,
		Resizable:   true,
		VSync:       true,
		FrameRate:   60,
		EventDriven: true, // 0% CPU when idle
	}
}

// New creates a new Gödel application with the given options.
func New(opts ...Option) *App {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	app := &App{
		cfg:       cfg,
		taskQueue: make(chan func(), 1024),
	}

	return app
}

// NewFromConfig creates a new app from a godel.toml configuration file.
func NewFromConfig(path string) (*App, error) {
	godelCfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", path, err)
	}

	opts := []Option{
		WithTitle(godelCfg.App.Name),
		WithSize(godelCfg.Build.Width, godelCfg.Build.Height),
	}

	if godelCfg.Theme.DesignSystem != "" {
		// Theme will be applied after app creation
	}

	return New(opts...), nil
}

// OnReady registers a callback invoked when the app is fully initialized
// and ready to build the UI. This is where you set the root widget.
func (a *App) OnReady(fn func(context.Context) error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onReady = append(a.onReady, fn)
}

// OnClose registers a callback invoked when the app is about to close.
// Use this for cleanup, saving state, etc.
func (a *App) OnClose(fn func(context.Context) error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onClose = append(a.onClose, fn)
}

// SetRoot sets the root widget of the application.
func (a *App) SetRoot(w widget.Widget) {
	a.root = w
	if a.uiApp != nil {
		a.uiApp.SetRoot(w)
	}
}

// QueueCallback safely schedules a function to run on the main thread.
// This is the only safe way to update UI state from a goroutine.
//
// Example:
//
//	go func() {
//	    data := fetchFromAPI()
//	    app.QueueCallback(func() {
//	        state.Data.Set(data)      // safe: runs on main thread
//	    })
//	}()
func (a *App) QueueCallback(fn func()) {
	select {
	case a.taskQueue <- fn:
	default:
		log.Println("godel: task queue full, dropping callback")
	}
}

// RequestRedraw schedules a redraw for the next frame.
func (a *App) RequestRedraw() {
	if a.gogpuApp != nil {
		a.gogpuApp.RequestRedraw()
	}
}

// Run starts the application event loop. This blocks until the window is closed.
func (a *App) Run(ctx ...context.Context) error {
	appCtx := context.Background()
	if len(ctx) > 0 {
		appCtx = ctx[0]
	}

	// Create the gogpu application (windowing + GPU)
	gogpuCfg := gogpu.DefaultConfig().
		WithTitle(a.cfg.Title).
		WithSize(a.cfg.Width, a.cfg.Height).
		WithContinuousRender(!a.cfg.EventDriven)

	a.gogpuApp = gogpu.NewApp(gogpuCfg)

	// Create the UI app (widget toolkit)
	a.uiApp = uiapp.New(
		uiapp.WithWindowProvider(a.gogpuApp),
		uiapp.WithPlatformProvider(a.gogpuApp),
		uiapp.WithEventSource(a.gogpuApp.EventSource()),
		uiapp.WithTheme(theme.DefaultLight()),
	)

	// Fire OnReady hooks
	for _, fn := range a.onReady {
		if err := fn(appCtx); err != nil {
			return fmt.Errorf("godel: OnReady hook failed: %w", err)
		}
	}

	// Set root if it was set before Run()
	if a.root != nil {
		a.uiApp.SetRoot(a.root)
	}

	// Internal metrics
	var frameTimes []time.Duration

	// Register draw handler
	a.gogpuApp.OnDraw(func(dc *gogpu.Context) {
		start := time.Now()
		a.drainTaskQueue()
		w, h := dc.Width(), dc.Height()

		if a.canvas == nil {
			var err error
			a.canvas, err = ggcanvas.New(a.gogpuApp.GPUContextProvider(), w, h)
			if err != nil {
				log.Printf("godel: canvas creation failed: %v", err)
				return
			}
		} else if a.canvas.Width() != w || a.canvas.Height() != h {
			a.canvas.Resize(w, h)
		}

		// Process UI frame (layout, state, events)
		a.uiApp.Frame()

		// Render
		sv := dc.SurfaceView()
		sw, sh := dc.SurfaceSize()
		gg.SetAcceleratorSurfaceTarget(sv, sw, sh)

		a.canvas.Draw(func(cc *gg.Context) {
			// Clear background
			cc.SetRGBA(0.94, 0.94, 0.94, 1)
			cc.DrawRectangle(0, 0, float64(w), float64(h))
			cc.Fill()

			// Render UI tree
			a.uiApp.Window().DrawTo(render.NewCanvas(cc, w, h))
		})

		_ = a.canvas.RenderDirect(sv, sw, sh)

		// Collect metrics if we're in a special mode (env var check for simplicity)
		if os.Getenv("GODEL_BENCHMARK") == "1" {
			frameTimes = append(frameTimes, time.Since(start))
			// Only keep last 100 for windowing or similar
			if len(frameTimes) > 1000 {
				frameTimes = frameTimes[1:]
			}
		}
	})

	// Register close handler
	a.gogpuApp.OnClose(func() {
		for _, fn := range a.onClose {
			if err := fn(appCtx); err != nil {
				log.Printf("godel: OnClose hook error: %v", err)
			}
		}
		gg.CloseAccelerator()
	})

	// Block on the event loop
	return a.gogpuApp.Run()
}

// drainTaskQueue runs all pending callbacks on the main thread.
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
