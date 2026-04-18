---
title:  "Gödel Architecture"
description: "Technical Architecture Deep-Dive"
---

*Understanding the rendering engine, GPU pipeline, and performance characteristics.*
 

## Table of Contents
1. [GPU Rendering Pipeline](#gpu-rendering-pipeline)
2. [SDF Rendering System](#sdf-rendering-system)
3. [Widget Tree & Layout](#widget-tree--layout)
4. [Event Loop & Scheduling](#event-loop--scheduling)
5. [Memory Management](#memory-management)
6. [Concurrency Guarantees](#concurrency-guarantees)
7. [Performance Profiling](#performance-profiling)

 

## GPU Rendering Pipeline

### Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│              (Go business logic & state)                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   Widget Tree (Retained)                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Root Container                                        │   │
│  │  ├─ Button (state: hovered, pressed)                │   │
│  │  ├─ TextField (state: focused, text value)          │   │
│  │  └─ List (state: scrollOffset)                      │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Layout Pass (Measured Once)                     │
│  1. Measure: Calculate size constraints                      │
│  2. Arrange: Position children                              │
│  3. Cache: Store layout (reuse if not dirty)               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Draw Call Generation                            │
│  1. Traverse widget tree (depth-first)                      │
│  2. Emit draw calls for each widget                         │
│  3. Batch by material/shader                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│         GPU Command Buffer Recording                         │
│  1. Bind render pass                                        │
│  2. Set viewport / scissor                                  │
│  3. Issue batched draw calls                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│             GPU Execution (Vulkan/Metal/DX12)               │
│  1. Vertex shader: Position transformation                  │
│  2. Rasterization: Fragment generation                      │
│  3. Fragment shader: SDF evaluation, color output           │
│  4. Blending: Alpha composition                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│          Frame Buffer → Screen (vsync or immediate)         │
└─────────────────────────────────────────────────────────────┘
```

### Frame Time Budget (60fps = 16.7ms)

```
Target: 16.7ms per frame @ 60fps

┌────────────────────────────────┐
│ Frame 60  (16.7ms budget)      │
├────────────────────────────────┤
│ Event polling           &lt;0.5ms │  Non-blocking OS event poll
│ Event processing        &lt;0.5ms │  Call handlers, queue callbacks
│ Layout (if dirty)        &lt;2ms  │  Measure + arrange (cached)
│ Draw call generation     &lt;1ms  │  Traverse tree, batch
│ GPU submission           &lt;0.5ms│  Record commands, submit
│ GPU render              ~12ms  │  Actual rasterization (hardware-dependent)
│ Frame flip/vsync         &lt;0.2ms│  Present to screen
├────────────────────────────────┤
│ Total:                 ~16.7ms │
└────────────────────────────────┘

Actual breakdown (typical):
- CPU-side work: 4-5ms
- GPU-side work: 11-12ms (bottleneck is GPU, not CPU)
```

### GPU Backend Abstraction

**API Layer (Gödel abstraction):**
```go
type Renderer interface {
    // Lifecycle
    Init(window Window) error
    Shutdown() error
    
    // Frame management
    BeginFrame(clearColor Color) error
    EndFrame() error
    
    // Command recording
    BindRenderPass(pass *RenderPass) error
    SetViewport(rect Rect) error
    SetScissor(rect Rect) error
    
    // Draw calls
    DrawQuad(quad *DrawCall) error
    DrawPath(path *DrawCall) error
    DrawText(text *DrawCall) error
    DrawMesh(mesh *DrawCall) error
    
    // Batching
    Flush() error
    
    // Shader management
    BindShader(name string) error
    SetShaderParam(name string, value interface{}) error
}
```

**Implementation per platform:**
```
Vulkan impl (VkRenderer):
├─ Instance, Device, Queue setup
├─ Command pools, command buffers
├─ Render passes, framebuffers
└─ Pipeline caching

Metal impl (MTLRenderer):
├─ MTLDevice, MTLCommandQueue
├─ MTLRenderPassDescriptor
├─ Pipeline state objects
└─ Texture caches

Direct3D 12 impl (D3D12Renderer):
├─ ID3D12Device, ID3D12CommandQueue
├─ Root signatures, pipeline state
├─ Descriptor heaps
└─ Fence synchronization
```

### Draw Call Batching

**Naive approach (slow):**
```
Widget tree traversal:
  └─ Button:         1 draw call (rectangle)
  └─ Button:         1 draw call (text)
  └─ TextInput:      1 draw call (border)
  └─ TextInput:      1 draw call (text)
  └─ Label:          1 draw call (text)
Total: 5 draw calls per frame
```

**Optimized approach (Gödel):**
```
Draw call generation:
  1. Collect all draw calls
  2. Sort by:
     - Shader (SDF, gradient, text, etc.)
     - Blend mode (opaque, transparent)
     - Texture (atlas, custom)
  3. Emit batched commands

Result:
  - Batch 1 (SDF shader):      3 rects (buttons, input borders)
  - Batch 2 (Text shader):     3 texts
Total: 2 draw calls (vs 5)

Speedup: 2.5x fewer GPU commands
```

---

## SDF Rendering System

### What is SDF?

**Signed Distance Field:** A texture where each pixel stores the shortest distance to the nearest edge.

```
Regular texture (raster):      SDF texture (distance field):
┌─────────────────────┐        ┌─────────────────────┐
│ ░░░░░░░░░░░░░░░░░░░│        │ 99 99 99 99 99 99  │
│ ░░░░░░░░░░░░░░░░░░░│        │ 50 10  5  5 10 50  │
│ ░░░░░█████░░░░░░░░░│        │ 10  2  1  1  2 10  │
│ ░░░░░█████░░░░░░░░░│        │ 10  2  1  1  2 10  │
│ ░░░░░░░░░░░░░░░░░░░│        │ 50 10  5  5 10 50  │
└─────────────────────┘        └─────────────────────┘
(256×256 pixels)               (lower resolution, smaller)
```

**Advantages:**
1. **Infinite scalability:** Distance field is mathematically continuous.
2. **Compact storage:** Lower resolution needed (16×16 works for most icons).
3. **Anti-aliasing:** Built-in smooth edges via distance-based blending.
4. **Effects:** Glow, outline, shadow all from same SDF value.

### SDF Rendering Pipeline

**Step 1: Offline Asset Generation**

During build time, convert vector assets to SDF:
```
TTF/SVG assets
    ↓
msdfgen (multi-channel distance field tool)
    ↓
SDF atlas texture (PNG with distance values)
    ↓
Embed in binary as constant data
```

**Build script example:**
```bash
#!/bin/bash
# scripts/compile_assets.sh

for font in assets/fonts/*.ttf; do
    msdfgen "$font" \
        --size 32 \
        --output assets/generated/$(basename $font .ttf).sdf.png
done

for icon in assets/icons/*.svg; do
    msdfgen "$icon" \
        --size 64 \
        --output assets/generated/$(basename $icon .svg).sdf.png
done

# Embed in Go
go run cmd/embed-assets/main.go \
    --input assets/generated/ \
    --output pkg/render/assets.go
```

**Step 2: GPU Rendering**

Vertex shader (position transformation):
```glsl
#version 450

layout(location = 0) in vec2 position;
layout(location = 1) in vec2 texCoord;

layout(push_constant) uniform PushConstants {
    mat4 projection;
    vec4 rect;
    vec4 color;
} pc;

layout(location = 0) out vec2 vTexCoord;
layout(location = 1) out vec4 vColor;

void main() {
    gl_Position = pc.projection * vec4(position, 0.0, 1.0);
    vTexCoord = texCoord;
    vColor = pc.color;
}
```

Fragment shader (SDF evaluation):
```glsl
#version 450

layout(location = 0) in vec2 vTexCoord;
layout(location = 1) in vec4 vColor;

layout(set = 0, binding = 0) uniform sampler2D sdfTexture;

layout(location = 0) out vec4 outColor;

void main() {
    // Sample SDF texture
    float sdfDist = texture(sdfTexture, vTexCoord).r;
    
    // Convert from 0-1 range to distance
    // (depends on atlas packing)
    float distance = sdfDist * 2.0 - 1.0;
    
    // Smooth step to create anti-aliased edge
    float smoothWidth = fwidth(distance);
    float alpha = smoothstep(-smoothWidth, smoothWidth, distance);
    
    // Optional: add glow or shadow
    float glow = exp(-distance * distance * 4.0);
    
    outColor = vec4(vColor.rgb, vColor.a * alpha);
}
```

### Performance vs. Mesh Rendering

**SDF Rendering:**
```
Pros:
├─ Perfect scaling (infinitely smooth)
├─ Tiny asset footprint (1 atlas = 100+ icons)
├─ Great for vector UI (buttons, icons, text)
└─ Consistent across platforms

Cons:
├─ Ray marching is iterative (slower than pre-baked)
└─ Not ideal for complex 3D geometry

Typical cost:
├─ SDF text (100 glyphs):    ~0.5-1.0ms per frame
├─ SDF shapes (50 elements): ~0.2-0.5ms per frame
└─ Total for typical UI:     ~1-2ms fragment shader time
```

**Mesh Rendering (fallback):**
```
Pros:
├─ Direct triangle rasterization (very fast)
├─ Optimized by 30 years of GPU evolution
└─ Good for 3D content

Cons:
├─ Large asset files (1 atlas = fewer icons)
├─ Blurry when scaled
└─ Platform-dependent rasterization

Typical cost:
├─ Mesh text (100 glyphs):    ~0.1-0.3ms per frame
├─ Mesh shapes (50 elements): ~0.05-0.1ms per frame
└─ But asset file 10x larger
```

**Hybrid approach (Gödel):**
```
Default: Use SDF for everything (text, icons, shapes)
├─ Asset size: minimal
├─ Scaling: perfect
├─ Performance: good for UI (1-2ms total)

Fallback: Mesh for 3D or complex scenes
├─ Auto-detect complex geometry
├─ Render 3D with mesh, 2D with SDF
└─ Best of both worlds
```

### Text Rendering with SDF

**Challenge:** Crisp, scalable text at any DPI.

**Solution: Multi-Channel SDF (MSDF)**

```
Regular SDF:              MSDF (3-channel):
┌──────────────┐         ┌──────────────┐
│ 0.5 0.3 0.1  │         │ R:0.5 G:0.3  │
│ 0.5 0.0 0.1  │    →    │ B:0.1 R:0.5  │
│ 0.5 0.3 0.1  │         │ G:0.0 B:0.1  │
└──────────────┘         └──────────────┘

MSDF allows much higher accuracy with lower resolution
```

**Text pipeline:**

1. **Build time:** Convert TTF font to MSDF atlas
   ```bash
   msdfgen myfont.ttf --size 32 --type msdf
   # Output: font_atlas.png (contains all glyphs as MSDF)
   ```

2. **Runtime:** Render glyphs using MSDF shader
   ```go
   type TextRenderer struct {
       fontAtlas *Texture         // MSDF texture
       glyphData map[rune]Glyph   // Glyph metrics
       shader    *Shader          // MSDF fragment shader
   }
   
   func (r *TextRenderer) RenderText(text string, pos Point, size float32) {
       for _, ch := range text {
           glyph := r.glyphData[ch]
           quad := Quad{
               Position: pos,
               TexCoord: glyph.TexRect,
               Size: glyph.Size * (size / 32), // Scale
           }
           r.DrawQuad(quad)
           pos.X += glyph.Advance * (size / 32)
       }
   }
   ```

3. **Result:** Crisp text at any size, any DPI

---

## Widget Tree & Layout

### Widget Tree Structure

```go
type Widget interface {
    // Measurement phase
    Measure(constraints Constraints) Size
    
    // Arrangement phase
    Layout(rect Rect)
    
    // Rendering phase
    Draw(ctx DrawContext) error
    
    // Event handling
    OnEvent(e *Event) bool
}

type Container struct {
    children    []Widget
    constraints Constraints
    rect        Rect          // Computed in Layout()
    isDirty     bool
}
```

### Layout Algorithm (Flexbox-style)

**Two-pass layout (like Flutter):**

```
Pass 1: Measure (bottom-up)
┌─────────────────────────────┐
│ Container                   │ Asks children: "How big are you?"
├─────────────────────────────┤
│ Button ("Click")  [50×20]   │ Reports: 50×20
│ Label ("Text")    [100×16]  │ Reports: 100×16
│ Spacer           [flex]     │ Reports: flexible
└─────────────────────────────┘
Container concludes: "I need at least 100×40"

Pass 2: Arrange (top-down)
Container is given: Rect{x: 0, y: 0, w: 200, h: 100}
└─ Distribute 200×100 among children
   ├─ Button:  Rect{x: 0, y: 0, w: 50, h: 20}
   ├─ Label:   Rect{x: 0, y: 20, w: 100, h: 16}
   └─ Spacer:  Rect{x: 0, y: 36, w: 200, h: 64} (flex takes remainder)
```

### Layout Caching

**Optimization: Don't re-layout if nothing changed**

```go
type Widget struct {
    layoutCache   *LayoutResult
    layoutVersion int64
    isDirty       bool
}

func (w *Widget) Layout(constraints Constraints) {
    // Check if we can reuse cached layout
    if !w.isDirty && w.cachedConstraints == constraints {
        return  // Use cached layout
    }
    
    // Otherwise, re-layout children
    for _, child := range w.children {
        child.Layout(childConstraints)
    }
    
    // Cache result
    w.layoutCache = computeLayout()
    w.layoutVersion = globalVersion
    w.isDirty = false
}

// When state changes, mark as dirty
func (w *Widget) setState(newState interface{}) {
    w.state = newState
    w.isDirty = true
    w.parent.invalidateLayout()  // Propagate up
}
```

### Early Exit for Unchanged Subtrees

```
Global change: "Counter changed from 5 to 6"
                ↓
┌─────────────────────────────────┐
│ Root Container                  │ Marked dirty
├─────────────────────────────────┤
│ Header (independent) ────────→  │ Reuse cached layout ✓
│ Content Panel ──────────────┐   │ (no change)
│  ├─ Sidebar ─────→          │   │ Reuse ✓
│  └─ Main View               │   │
│      ├─ Counter (changed)   │   │ Recompute layout ✗
│      └─ Other widgets ────→ │   │ Reuse ✓
└─────────────────────────────────┘

Result: Only 2 widgets re-laid out (not 7)
```

---

## Event Loop & Scheduling

### Main Event Loop

```go
func (app *App) eventLoop(ctx context.Context) error {
    var nextFrameTime time.Time
    targetFrameTime := time.Second / 60 // 16.7ms for 60fps
    
    for {
        select {
        case <-ctx.Done():
            return nil
            
        default:
            // 1. Poll OS events (non-blocking)
            startTime := time.Now()
            events := pollOSEvents(0) // timeout=0 = non-blocking
            
            for _, event := range events {
                app.handleOSEvent(event)
            }
            
            // 2. Process callbacks from goroutines
            app.taskQueue.DrainAll(func(task func()) {
                task()
            })
            
            // 3. Layout if needed
            if app.layoutDirty {
                app.layoutTree()
                app.layoutDirty = false
            }
            
            // 4. Render
            if err := app.renderer.BeginFrame(app.clearColor); err != nil {
                return err
            }
            
            if err := app.drawTree(app.root); err != nil {
                return err
            }
            
            if err := app.renderer.EndFrame(); err != nil {
                return err
            }
            
            // 5. Sleep to cap frame rate
            elapsed := time.Since(startTime)
            sleepTime := targetFrameTime - elapsed
            if sleepTime > 0 {
                time.Sleep(sleepTime)
                nextFrameTime = time.Now()
            } else {
                // Skipped frame (heavy workload)
                log.Printf("Skipped frame: took %.1fms", elapsed.Seconds()*1000)
            }
        }
    }
}
```

### Task Queue (Safe from Goroutines)

```go
type TaskQueue struct {
    mu    sync.Mutex
    tasks []func()
}

// Safe to call from any goroutine
func (q *TaskQueue) Enqueue(task func()) {
    q.mu.Lock()
    q.tasks = append(q.tasks, task)
    q.mu.Unlock()
}

// Called from main thread only
func (q *TaskQueue) DrainAll(fn func(func())) {
    q.mu.Lock()
    tasks := q.tasks
    q.tasks = q.tasks[:0]  // Reset
    q.mu.Unlock()
    
    for _, task := range tasks {
        fn(task)
    }
}

// Example: goroutine updating UI
go func() {
    data := fetchData()  // Network I/O (blocks goroutine, not main)
    
    app.QueueCallback(func() {
        // This runs on main thread, safe to update UI
        state.Data = data
        app.RequestRedraw()
    })
}()
```

### Context-Based Cancellation

```go
// Cancel all tasks when app closes
ctx, cancel := context.WithCancel(context.Background())

app.OnClose(func() error {
    cancel()  // Signals all goroutines
    return nil
})

// Goroutine respects cancellation
go func() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return  // Cleanup and exit
        case <-ticker.C:
            // Do work
        }
    }
}()
```

---

## Memory Management

### Allocation Patterns

**Typical memory usage for app with 1000 widgets:**

```
Component                      Size         Count    Total
────────────────────────────────────────────────────────────
Widget struct (base)           ~200 bytes   1000     ~200 KB
Layout cache                   ~100 bytes   1000     ~100 KB
Event handlers                 ~80 bytes    2000     ~160 KB
Draw calls (per frame)         ~50 bytes    5000     ~250 KB
Texture atlases (SDF)          ~2 MB        1        ~2 MB
GPU command buffer             ~4 MB        1        ~4 MB
────────────────────────────────────────────────────────────
Typical total:                                      ~6.8 MB
```

### Garbage Collection Impact

**GC pause times:**
```
App with ~5000 widgets:
├─ Allocation rate:    ~5 MB/s (during UI updates)
├─ GC cycle frequency: ~10-20ms @ 60fps (every 0.5-1 frame)
├─ GC pause:           &lt;0.5ms (if generation0 only)
└─ Result:             No visible stuttering
```

**Mitigation strategies:**
1. **Object pooling** for frequently allocated widgets
2. **Reuse allocations** in layout phase
3. **Disable GC** during performance-critical sections (rare)

```go
// Reusable layout context
type LayoutContext struct {
    rects      []Rect
    sizes      []Size
    constraints []Constraints
}

func (ctx *LayoutContext) Reset() {
    ctx.rects = ctx.rects[:0]
    ctx.sizes = ctx.sizes[:0]
    ctx.constraints = ctx.constraints[:0]
}

// Reuse across frames
layoutCtx := NewLayoutContext()
for frame := 0; frame < maxFrames; frame++ {
    layoutCtx.Reset()
    app.layoutTree(layoutCtx)
}
```

---

## Concurrency Guarantees

### Thread Safety Rules

```
═════════════════════════════════════════════════════════════
Gödel Concurrency Model:

1. UI THREAD (main loop)
   - Owns: Widget tree, state, renderer
   - Can: Read/write state directly
   - Cannot: Block (will freeze UI)

2. WORKER GOROUTINES
   - Can: Do I/O, compute, long operations
   - Cannot: Directly mutate widget state
   - Must: Use app.QueueCallback() to update UI

3. SYNCHRONIZATION
   - Main thread runs all UI updates
   - Goroutines queue callbacks
   - No locks needed (single-threaded UI)
═════════════════════════════════════════════════════════════
```

### Race Condition Prevention

**BAD: Race condition**
```go
go func() {
    data := fetchData()
    state.Data = data  // ❌ Data race! Main thread might read data simultaneously
}()
```

**GOOD: Safe callback**
```go
go func() {
    data := fetchData()
    app.QueueCallback(func() {
        // ✅ Safe: Runs on main thread, after previous frame done
        state.Data = data
        app.RequestRedraw()
    })
}()
```

### Testing Concurrency

```go
// Stress test: 1000 goroutines updating state
func TestConcurrentUpdates(t *testing.T) {
    app := godel.New()
    state := state.NewSignal(0)
    
    for i := 0; i < 1000; i++ {
        go func(id int) {
            for j := 0; j < 100; j++ {
                app.QueueCallback(func() {
                    state.Set(state.Get() + 1)
                })
            }
        }(i)
    }
    
    // All updates should be ordered and safe
    time.Sleep(2 * time.Second)
    assert.Equal(t, 1000*100, state.Get())
}
```

---

## Performance Profiling

### Built-in Profiling

```bash
# Frame time profiling
godel profile --frames 600 --output profile.json

# Output: frame_times.json
{
  "frames": [
    {
      "id": 0,
      "total_ms": 16.2,
      "phases": {
        "event_poll_ms": 0.3,
        "event_process_ms": 0.2,
        "layout_ms": 0.8,
        "draw_call_gen_ms": 0.5,
        "gpu_submit_ms": 0.4,
        "gpu_render_ms": 13.5,
        "frame_present_ms": 0.5
      }
    },
    ...
  ]
}
```

### Memory Profiling

```bash
# Heap allocation profiling
godel profile --memory --duration 30s --output heap.pprof

# Analyze with pprof
go tool pprof heap.pprof
(pprof) top10
(pprof) list MyWidget.Layout
```

### GPU Profiling (Platform-Specific)

```bash
# On macOS (Metal)
xcode-select --install
# Run in Xcode GPU Debugger

# On Linux (Vulkan)
renderdoc  # Free, open-source GPU profiler

# On Windows (Direct3D)
PIX on Windows  # Microsoft's GPU debugger
```

### Custom Profiling Markers

```go
// Mark sections of code for analysis
app.ProfileBegin("fetch_data")
data := fetchFromAPI()
app.ProfileEnd("fetch_data")

// Output:
// fetch_data: 245.3ms (over 10 calls, avg: 24.5ms)
```

---

## Optimization Tips

### 1. Reduce Widget Count
```go
// ❌ Bad: Create 10,000 individual labels
for i := 0; i < 10000; i++ {
    ui.Label{Text: fmt.Sprintf("Item %d", i)}
}

// ✅ Good: Use ListView with virtualization
ui.ListView{
    Items: data,
    ViewportHeight: 600,
    ItemHeight: 40,
    // Only renders ~15 visible items at a time
    ItemBuilder: func(item interface{}) ui.Widget { ... },
}
```

### 2. Cache Complex Layouts
```go
// ❌ Bad: Recompute layout every frame
func (w *Widget) Layout(rect Rect) {
    expensiveLayout(w.children)  // O(N^2) every frame
}

// ✅ Good: Cache if constraints unchanged
func (w *Widget) Layout(rect Rect) {
    if w.cachedRect == rect {
        return  // Reuse cached layout
    }
    expensiveLayout(w.children)
    w.cachedRect = rect
}
```

### 3. Batch State Updates
```go
// ❌ Bad: Triggers redraw 100 times
for i := 0; i < 100; i++ {
    items.Set(append(items.Get(), newItem()))
    // Each Set() triggers layout + render
}

// ✅ Good: Batch updates, single redraw
state.Batch(func() {
    allItems := items.Get()
    for i := 0; i < 100; i++ {
        allItems = append(allItems, newItem())
    }
    items.Set(allItems)  // Single trigger
})
```

### 4. Use Signals Efficiently
```go
// ❌ Bad: Recompute on every store access
counter.Map(func(v int) string {
    return fmt.Sprintf("Count: %d", v)  // Recomputed, even if counter didn't change
})

// ✅ Good: Computed signal
computed := state.Computed(func() string {
    return fmt.Sprintf("Count: %d", counter.Get())
}).ListenTo(counter)  // Only recomputes when counter changes
```

---

## Benchmarks & Targets

### Expected Performance

```
Operation               Latency      Notes
────────────────────────────────────────────────────────
Button click            &lt;1ms         Direct function call
Text input (keystroke)  &lt;2ms         Character insertion + redraw
List scroll (1000 items) &lt;16ms        GPU-accelerated, virtual list
Modal dialog open       &lt;50ms        First-time render
App startup            &lt;300ms        Binary load + initialization

Memory
────────────────────────────────────────────────────────
Hello World app        3-5 MB
Complex dashboard      20-30 MB
With custom assets     +asset size
```

### Comparison Matrix

```
Framework    Latency    Memory    Binary    Consistency
─────────────────────────────────────────────────────────
Gödel        &lt;1-2ms     Low       5-20MB    Excellent
Flutter      &lt;1ms       Medium    40-80MB   Excellent
Tauri        10-40ms    Medium    10-15MB   Fair
Wails        10-50ms    Medium    20-40MB   Poor (WebView)
Electron     5-20ms     High      150+MB    Fair
```

---

**Deep understanding of Gödel's internals will help you build fast, reliable desktop applications. Happy coding! 🚀**
