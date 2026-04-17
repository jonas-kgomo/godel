# Gödel: Implementation Quick-Start & Project Templates

---

## Table of Contents
1. [Project Initialization Templates](#project-initialization-templates)
2. [Core API Quick Reference](#core-api-quick-reference)
3. [Build Configuration](#build-configuration)
4. [Plugin Development](#plugin-development)
5. [Deployment Checklist](#deployment-checklist)

---

## Project Initialization Templates

### Template 1: Basic Desktop App

**Run:** `godel init my-app --template basic`

**Generated structure:**
```
my-app/
├── main.go
├── go.mod
├── go.sum
├── Makefile
├── godel.toml
└── assets/
    └── icons/
        ├── app.png (256x256)
        └── app-dark.png (256x256)
```

**`godel.toml` (configuration):**
```toml
[app]
name = "My App"
version = "0.1.0"
description = "A beautiful Gödel app"
author = "Your Name"

[build]
main = "main.go"
output = "my-app"
icon = "assets/icons/app.png"
bundle_id = "com.example.myapp"  # macOS/iOS

[platforms]
windows = true
macos = true
linux = true

[theme]
primary = "#007AFF"
accent = "#FF3B30"
design_system = "material3"  # or "fluent", "cupertino"

[development]
hot_reload = true
debug = true
```

**`main.go` (starter):**
```go
package main

import (
    "context"
    "log"
    
    "github.com/yourusername/godel/pkg/app"
    "github.com/yourusername/godel/pkg/ui"
)

func main() {
    ctx := context.Background()
    
    // Create app instance
    myApp := app.New(
        app.WithConfig("godel.toml"),
        app.WithSize(800, 600),
        app.WithMinSize(400, 300),
    )
    
    // Setup UI
    myApp.OnReady(func(ctx context.Context) error {
        root := ui.Container{
            Padding: ui.EdgeInsets{All: 20},
            Children: []ui.Widget{
                ui.Label{
                    Text: "Welcome to Gödel",
                    FontSize: 32,
                    FontWeight: ui.FontBold,
                },
            },
        }
        myApp.SetRoot(root)
        return nil
    })
    
    // Handle app lifecycle
    myApp.OnClose(func(ctx context.Context) error {
        log.Println("App closing...")
        return nil
    })
    
    // Run app
    if err := myApp.Run(ctx); err != nil {
        log.Fatalf("App error: %v", err)
    }
}
```

**`Makefile`:**
```makefile
.PHONY: dev build clean help

help:
	@echo "Gödel Development Commands"
	@echo "make dev       - Run app with hot reload"
	@echo "make build     - Build release binary"
	@echo "make clean     - Remove build artifacts"

dev:
	godel dev

build:
	godel build --release

build-all:
	godel build --target linux,macos,windows --release

clean:
	rm -rf dist/ *.exe *.app
	go clean

test:
	go test ./...

bench:
	go test -bench=. -benchmem ./...
```

---

### Template 2: Data-Intensive Dashboard

**Run:** `godel init dashboard --template dashboard`

**Structure additions:**
```
dashboard/
├── main.go
├── internal/
│   ├── models/
│   │   ├── metric.go
│   │   └── alert.go
│   ├── data/
│   │   ├── loader.go
│   │   └── cache.go
│   └── state/
│       └── app_state.go
├── plugins/
│   └── datasource/
│       ├── prometheus.go
│       └── postgres.go
└── assets/
    └── icons/
```

**`internal/state/app_state.go`:**
```go
package state

import (
    "github.com/yourusername/godel/pkg/state"
    "myapp/internal/models"
)

type AppState struct {
    // Signals (reactive state)
    Metrics      state.Signal[[]models.Metric]
    Alerts       state.Signal[[]models.Alert]
    SelectedTab  state.Signal[string]
    RefreshRate  state.Signal[time.Duration]
    
    // Computed values
    CriticalCount state.Computed[int]
    AverageLoad   state.Computed[float64]
}

func NewAppState() *AppState {
    s := &AppState{
        Metrics:     state.NewSignal([]models.Metric{}),
        Alerts:      state.NewSignal([]models.Alert{}),
        SelectedTab: state.NewSignal("overview"),
        RefreshRate: state.NewSignal(30 * time.Second),
    }
    
    // Computed: count critical alerts
    s.CriticalCount = state.Computed(func() int {
        alerts := s.Alerts.Get()
        count := 0
        for _, a := range alerts {
            if a.Severity == models.SeverityCritical {
                count++
            }
        }
        return count
    }).ListenTo(s.Alerts)
    
    return s
}
```

**`main.go` (dashboard):**
```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/yourusername/godel/pkg/app"
    "github.com/yourusername/godel/pkg/ui"
    "myapp/internal/state"
    "myapp/internal/data"
)

func main() {
    ctx := context.Background()
    myApp := app.New(app.WithSize(1400, 900))
    appState := state.NewAppState()
    
    myApp.OnReady(func(ctx context.Context) error {
        // Start data loader (concurrent)
        startDataLoader(ctx, myApp, appState)
        
        // Build UI
        root := buildDashboard(appState)
        myApp.SetRoot(root)
        return nil
    })
    
    if err := myApp.Run(ctx); err != nil {
        log.Fatal(err)
    }
}

func startDataLoader(ctx context.Context, app *app.App, state *state.AppState) {
    ticker := time.NewTicker(state.RefreshRate.Get())
    defer ticker.Stop()
    
    go func() {
        for {
            select {
            case <-ticker.C:
                // Load data (non-blocking)
                go func() {
                    metrics, err := data.FetchMetrics(ctx)
                    if err != nil {
                        log.Printf("Error fetching metrics: %v", err)
                        return
                    }
                    app.QueueCallback(func() {
                        state.Metrics.Set(metrics)
                    })
                }()
            case <-ctx.Done():
                return
            }
        }
    }()
}

func buildDashboard(state *state.AppState) ui.Widget {
    return ui.HStack(
        buildSidebar(state),
        buildMainPanel(state),
    )
}

func buildSidebar(state *state.AppState) ui.Widget {
    return ui.VStack(
        ui.Label{Text: "Dashboard", FontSize: 24, FontWeight: ui.FontBold},
        ui.Spacer(h: 20),
        buildTabButton(state, "overview", "Overview"),
        buildTabButton(state, "alerts", "Alerts"),
        buildTabButton(state, "settings", "Settings"),
    )
}

func buildMainPanel(state *state.AppState) ui.Widget {
    return state.SelectedTab.Map(func(tab string) ui.Widget {
        switch tab {
        case "overview":
            return buildOverviewPanel(state)
        case "alerts":
            return buildAlertsPanel(state)
        default:
            return ui.Label{Text: "Settings (not implemented)"}
        }
    })
}

func buildOverviewPanel(state *state.AppState) ui.Widget {
    return ui.GridView{
        Columns: 2,
        Spacing: 16,
        Items: state.Metrics.Get(),
        ItemBuilder: func(m interface{}) ui.Widget {
            metric := m.(models.Metric)
            return buildMetricCard(metric)
        },
    }
}

func buildAlertsPanel(state *state.AppState) ui.Widget {
    return ui.ListView{
        Items: state.Alerts.Get(),
        ItemBuilder: func(a interface{}) ui.Widget {
            alert := a.(models.Alert)
            return buildAlertRow(alert)
        },
    }
}

func buildTabButton(state *state.AppState, tabID string, label string) ui.Widget {
    isActive := state.SelectedTab.Map(func(tab string) bool {
        return tab == tabID
    })
    
    return ui.Button{
        Label: label,
        Variant: isActive.Map(func(active bool) ui.ButtonVariant {
            if active {
                return ui.ButtonPrimary
            }
            return ui.ButtonSecondary
        }),
        OnClick: func(ctx context.Context) error {
            state.SelectedTab.Set(tabID)
            return nil
        },
    }
}

func buildMetricCard(m models.Metric) ui.Widget {
    return ui.Card{
        Padding: 16,
        Children: []ui.Widget{
            ui.Label{Text: m.Name, FontSize: 16, FontWeight: ui.FontSemiBold},
            ui.Spacer(h: 8),
            ui.Label{
                Text: fmt.Sprintf("%.2f %s", m.Value, m.Unit),
                FontSize: 28,
                FontWeight: ui.FontBold,
                Color: m.StatusColor(),
            },
        },
    }
}

func buildAlertRow(a models.Alert) ui.Widget {
    return ui.HStack(
        ui.Container{
            Width: 4,
            Background: a.SeverityColor(),
        },
        ui.VStack(
            ui.Label{Text: a.Title, FontWeight: ui.FontSemiBold},
            ui.Label{Text: a.Message, FontSize: 12, Color: ui.ColorGray},
        ),
    )
}
```

---

### Template 3: Plugin-Based Architecture

**Run:** `godel init plugin-system --template plugin-system`

**Structure:**
```
plugin-system/
├── main.go
├── plugins/
│   ├── analytics/
│   │   ├── analytics.go
│   │   └── go.mod
│   ├── auth/
│   │   ├── auth.go
│   │   └── go.mod
│   └── storage/
│       ├── storage.go
│       └── go.mod
└── go.mod
```

**`plugins/analytics/analytics.go`:**
```go
package analytics

import (
    "context"
    "log"
    
    "github.com/yourusername/godel/pkg/app"
    "github.com/yourusername/godel/pkg/plugin"
)

type AnalyticsPlugin struct {
    enabled bool
    apiKey  string
}

func (p *AnalyticsPlugin) Name() string {
    return "analytics"
}

func (p *AnalyticsPlugin) Version() string {
    return "1.0.0"
}

func (p *AnalyticsPlugin) Init(ctx context.Context, app *app.App) error {
    log.Println("Initializing analytics plugin...")
    
    // Register event listener
    app.OnEvent("ui:interaction", func(e *app.Event) error {
        return p.trackEvent(e)
    })
    
    return nil
}

func (p *AnalyticsPlugin) Hooks() plugin.PluginHooks {
    return plugin.PluginHooks{
        OnReady: func(ctx context.Context) error {
            log.Println("Analytics plugin ready")
            return nil
        },
        OnShutdown: func(ctx context.Context) error {
            log.Println("Analytics plugin shutting down")
            return nil
        },
    }
}

func (p *AnalyticsPlugin) trackEvent(e *app.Event) error {
    // Send event to analytics service
    log.Printf("Tracking event: %+v", e)
    return nil
}
```

**`main.go` (with plugins):**
```go
package main

import (
    "context"
    "log"
    
    "github.com/yourusername/godel/pkg/app"
    "github.com/yourusername/godel/pkg/plugin"
    "myapp/plugins/analytics"
    "myapp/plugins/auth"
    "myapp/plugins/storage"
)

func main() {
    ctx := context.Background()
    myApp := app.New()
    
    // Register plugins
    pluginRegistry := plugin.NewRegistry()
    pluginRegistry.Register(&analytics.AnalyticsPlugin{})
    pluginRegistry.Register(&auth.AuthPlugin{})
    pluginRegistry.Register(&storage.StoragePlugin{})
    
    // Load plugins
    if err := pluginRegistry.InitAll(ctx, myApp); err != nil {
        log.Fatalf("Plugin initialization failed: %v", err)
    }
    
    myApp.OnReady(func(ctx context.Context) error {
        // Build UI...
        return nil
    })
    
    if err := myApp.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

---

## Core API Quick Reference

### App Lifecycle

```go
// Create app
app := godel.New(
    godel.WithTitle("My App"),
    godel.WithSize(800, 600),
    godel.WithMinSize(400, 300),
    godel.WithResizable(true),
)

// Lifecycle hooks
app.OnReady(func(ctx context.Context) error {
    // App is ready, build UI
    return nil
})

app.OnClose(func(ctx context.Context) error {
    // App closing, cleanup
    return nil
})

// Run
app.Run(ctx)

// Request redraw
app.RequestRedraw()

// Queue callback (safe from goroutines)
app.QueueCallback(func() {
    // Update UI state
})
```

### Widgets & Layout

```go
// Basic widgets
ui.Label{Text: "Hello", FontSize: 14}
ui.Button{Label: "Click", OnClick: func(...) error { ... }}
ui.TextInput{Placeholder: "Enter text", OnChange: func(...) error { ... }}
ui.CheckBox{Checked: true, OnChange: func(...) error { ... }}
ui.RadioButton{Selected: true, OnChange: func(...) error { ... }}
ui.Slider{Value: 0.5, Min: 0, Max: 1, OnChange: func(...) error { ... }}

// Layout containers
ui.VStack(child1, child2)  // Vertical
ui.HStack(child1, child2)  // Horizontal
ui.Container{
    Padding: ui.EdgeInsets{All: 16},
    Background: ui.Color{R: 1, G: 0, B: 0, A: 1},
    Children: []ui.Widget{...},
}

// List / Grid
ui.ListView{
    Items: data,
    ItemBuilder: func(item interface{}) ui.Widget {
        return ui.Label{Text: item.(string)}
    },
}

ui.GridView{
    Columns: 3,
    Spacing: 16,
    Items: data,
    ItemBuilder: func(item interface{}) ui.Widget { ... },
}

// Spacers
ui.Spacer(h: 20)  // Vertical spacer
ui.Spacer(w: 20)  // Horizontal spacer
```

### State Management

```go
// Signal (reactive value)
counter := state.NewSignal(0)
counter.Get()                    // Read
counter.Set(5)                   // Write
counter.Map(func(v int) string {
    return fmt.Sprintf("Count: %d", v)
})                               // Transform

// Computed value
computed := state.Computed(func() int {
    return counter.Get() * 2
}).ListenTo(counter)             // Re-compute on change

// Conditional rendering
state.If(
    signal,                       // Condition
    ui.Label{Text: "Yes"},        // True widget
    ui.Label{Text: "No"},         // False widget
)

// Batch updates
state.Batch(func() {
    signal1.Set(val1)
    signal2.Set(val2)
})                               // Single redraw
```

### Event Handling

```go
// Button click
ui.Button{
    Label: "Click",
    OnClick: func(ctx context.Context) error {
        log.Println("Clicked!")
        return nil
    },
}

// Text input change
ui.TextInput{
    OnChange: func(ctx context.Context, text string) error {
        state.InputValue.Set(text)
        return nil
    },
}

// Global events
app.OnEvent("custom:event", func(e *app.Event) error {
    log.Printf("Event: %+v", e.Data)
    return nil
})

// Emit custom event
app.EmitEvent("custom:event", map[string]interface{}{
    "key": "value",
})
```

### Theming

```go
// Apply theme
app.ApplyTheme(theme.Material3Dark)

// Theme interface
type Theme interface {
    Name() string
    Colors() ThemeColors
    Typography() ThemeTypography
    Shapes() ThemeShapes
}

// Custom theme
customTheme := &MyTheme{
    primaryColor: color.RGBA{0, 122, 255, 255},
    accentColor: color.RGBA{255, 59, 48, 255},
}
app.ApplyTheme(customTheme)
```

### Native OS Integration

```go
// File dialog
files, err := app.ShowOpenFileDialog(
    godel.WithTitle("Open File"),
    godel.WithFilters([]string{".txt", ".md"}),
)

// Folder dialog
folder, err := app.ShowFolderDialog()

// Message dialog
result, err := app.ShowMessageDialog(
    godel.WithTitle("Confirm"),
    godel.WithMessage("Are you sure?"),
    godel.WithButtons([]string{"Yes", "No"}),
)

// Notification
app.SendNotification(&godel.Notification{
    Title: "Alert",
    Body: "Something happened",
    Icon: "assets/icons/alert.png",
})

// Tray icon
app.SetTrayIcon(&godel.TrayIcon{
    Icon: "assets/icons/tray.png",
    Menu: []godel.MenuItem{
        {Label: "Show", OnClick: func() { app.Show() }},
        {Label: "Exit", OnClick: func() { app.Quit() }},
    },
})
```

---

## Build Configuration

### `godel.toml` Reference

```toml
[app]
name = "My Application"
version = "0.1.0"
description = "A Gödel app"
author = "Your Name"
license = "MIT"

[build]
main = "main.go"
output = "my-app"
icon = "assets/icons/app.png"
bundle_id = "com.example.myapp"

[platforms]
windows = true
macos = true
linux = true
android = false
ios = false

[windows]
icon = "assets/icons/app-windows.ico"
sign = false
certificate = ""  # Path to .pfx for signing

[macos]
icon = "assets/icons/app-macos.icns"
sign = true
team_id = "ABCDEF1234"
provisioning_profile = ""

[linux]
icon = "assets/icons/app-linux.png"
categories = ["Development", "Utility"]

[theme]
primary = "#007AFF"
secondary = "#FF9500"
accent = "#FF3B30"
design_system = "material3"  # material3, fluent, cupertino, custom

[development]
hot_reload = true
debug = true
log_level = "debug"
renderer = "vulkan"  # vulkan, metal, dx12, software

[release]
strip = true
optimize = true
target = ["linux", "macos", "windows"]

[renderer]
backend = "vulkan"
vulkan_validation = false
frame_rate = 60
vsync = true
```

### Build Commands

```bash
# Development with hot reload
godel dev

# Build release (current platform)
godel build --release

# Build for specific platforms
godel build --target linux,macos,windows --release

# Build with custom output
godel build --output ~/dist/ --name "MyApp-v1.0"

# Cross-compile (requires Docker for Linux targets from macOS/Windows)
godel build --target linux --cross

# Sign release (macOS/Windows)
godel build --sign --certificate /path/to/cert

# Analyze binary size
godel build --analyze-size
```

---

## Plugin Development

### Plugin Structure

**`plugins/my-plugin/main.go`:**
```go
package main

import (
    "context"
    
    "github.com/yourusername/godel/pkg/app"
    "github.com/yourusername/godel/pkg/plugin"
)

type MyPlugin struct {
    // Plugin state
    config map[string]string
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Init(ctx context.Context, app *app.App) error {
    // Initialize plugin
    // Register event handlers, load configuration, etc.
    return nil
}

func (p *MyPlugin) Hooks() plugin.PluginHooks {
    return plugin.PluginHooks{
        OnReady: func(ctx context.Context) error { ... },
        OnShutdown: func(ctx context.Context) error { ... },
        OnConfigChange: func(cfg map[string]string) error { ... },
    }
}
```

**`plugins/my-plugin/go.mod`:**
```
module github.com/yourusername/godel-plugins/my-plugin

go 1.21

require github.com/yourusername/godel v0.1.0
```

### Loading Plugins

```go
// Static loading (linked at compile time)
import (
    _ "myapp/plugins/analytics"
    _ "myapp/plugins/auth"
)

registry := plugin.Registry()
registry.InitAll(ctx, app)

// Or dynamic loading (runtime)
registry.Load(ctx, "./plugins/my-plugin.so")
```

---

## Deployment Checklist

### Before Release

- [ ] Test on Windows, macOS, Linux
- [ ] Test on different screen DPI (96, 144, 192)
- [ ] Test dark mode support
- [ ] Performance profile (60fps check)
- [ ] Memory leak check (run for 30+ mins)
- [ ] Accessibility check (keyboard navigation, screen reader)
- [ ] Security audit (no hardcoded secrets, validate inputs)
- [ ] Update version number
- [ ] Update `CHANGELOG.md`
- [ ] Sign binaries (Windows, macOS)
- [ ] Create installer (MSI for Windows, DMG for macOS)

### Release Process

```bash
# 1. Tag release
git tag v1.0.0
git push origin v1.0.0

# 2. Build all platforms
godel build --target linux,macos,windows --release --sign

# 3. Create checksums
sha256sum dist/* > SHA256SUMS
gpg --sign SHA256SUMS

# 4. Create GitHub release
# - Upload binaries
# - Upload checksums
# - Add release notes

# 5. Announce
# - Twitter/social media
# - Gödel community forum
# - Product Hunt (if major release)
```

### Distribution

**Windows:**
```bash
# Create MSI installer (optional)
godel build --installer msi

# Or distribute as standalone executable
dist/my-app-windows-amd64.exe
```

**macOS:**
```bash
# Create .app bundle
godel build --bundle

# Create DMG (optional)
godel build --installer dmg

# Or distribute as .app
dist/my-app-macos-arm64.app
```

**Linux:**
```bash
# Create AppImage
godel build --installer appimage

# Or distribute as AppImage
dist/my-app-linux-x86_64.AppImage

# Or .deb/.rpm
godel build --installer deb,rpm
```

---

## Performance Benchmarks

Expected performance targets:

```
Frame Time (60fps = 16.7ms budget):
├─ Logic execution     <1ms (user code)
├─ Widget layout       <2ms (50+ widgets)
├─ GPU rendering       <13ms (100+ elements)
└─ Total              ~16ms

Memory (native path):
├─ Base framework      ~10MB
├─ Per widget          ~1-2KB
├─ Total for 1000 UI   ~20MB

Binary Sizes:
├─ Hello World (native)    ~3.5MB
├─ Complex app (native)    ~5-8MB
├─ With web support        ~15-20MB
├─ Electron equivalent      ~150MB
```

Measure your app:
```bash
godel bench --duration 60s
# Output: FPS, frame time percentiles, memory usage
```

---

## Troubleshooting

### Common Issues

**Issue: Build fails on Linux (Vulkan not found)**
```bash
# Install Vulkan SDK
sudo apt install vulkan-tools libvulkan-dev

# Or use software renderer
godel dev --renderer=software
```

**Issue: App crashes on startup**
```bash
# Run with debug output
godel dev --debug

# Check logs
tail -f ~/.godel/logs/my-app.log
```

**Issue: High memory usage**
```bash
# Profile memory
godel profile --memory

# Check for goroutine leaks
godel profile --goroutines
```

**Issue: Slow rendering**
```bash
# Check GPU utilization
godel profile --gpu

# Reduce widget count or use virtualization (ListView)
```

---

**Happy building with Gödel! 🚀**
