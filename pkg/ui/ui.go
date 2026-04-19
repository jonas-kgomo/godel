package ui

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/intercode/godel/pkg/input"
	signals "github.com/coregx/signals"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	uistate "github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"sync"

	"cogentcore.org/core/core"
)

// WidgetBuilder is a function that populates a CogentCore Body.
type WidgetBuilder func(body *core.Body)

const (
	AlignStart = iota
	AlignCenter
	AlignEnd
)

var GlobalInput *input.State

// Re-export common types
type (
	Widget  = widget.Widget
	Color   = widget.Color
	Context = context.Context
)

// Common color helpers mapping to gogpu/ui RGBA8
func RGB(r, g, b uint8) Color       { return widget.RGBA8(r, g, b, 255) }
func RGBA(r, g, b, a uint8) Color   { return widget.RGBA8(r, g, b, a) }
func Hex(h string) Color {
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	var r, g, b, a uint8 = 0, 0, 0, 255
	if len(h) == 6 {
		fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b)
	} else if len(h) == 8 {
		fmt.Sscanf(h, "%02x%02x%02x%02x", &r, &g, &b, &a)
	}
	return RGBA(r, g, b, a)
}


// --- Widget Registry for Simulation ---

var (
	widgetRegistry = make(map[string]Widget)
	registryMu     sync.RWMutex
)

func RegisterWidget(id string, w Widget) {
	registryMu.Lock()
	defer registryMu.Unlock()
	widgetRegistry[id] = w
}

func GetWidgetByID(id string) Widget {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return widgetRegistry[id]
}

// WithID registers any widget for simulation targeting and returns it
func WithID(id string, w Widget) Widget {
	RegisterWidget(id, w)
	// If the widget supports ID setting, set it there too
	if wb, ok := any(w).(interface{ SetID(string) }); ok {
		wb.SetID(id)
	}
	return w
}

// --- Signals / State Management mappings ---

func NewSignal[T any](initial T) signals.Signal[T] {
	return signals.New[T](initial)
}

func NewComputed[T any](fn func() T, deps ...any) signals.ReadonlySignal[T] {
	return signals.Computed[T](fn, deps...)
}

// --- Layout Primitives ---

type UIBox struct {
	*primitives.BoxWidget
	onClick      func()
	onMouseEnter func()
	onMouseExit  func()
}

// VStack creates a vertical box layout
func VStack(children ...Widget) *UIBox {
	return &UIBox{BoxWidget: primitives.Box(children...).SetDirection(primitives.DirectionVertical)}
}

// HStack creates a horizontal box layout
func HStack(children ...Widget) *UIBox {
	return &UIBox{BoxWidget: primitives.Box(children...).SetDirection(primitives.DirectionHorizontal)}
}

// Container creates a flexible box with padding and background
func Container(children ...Widget) *UIBox {
	return &UIBox{BoxWidget: primitives.Box(children...)}
}

// Stack overlays children on top of each other
func Stack(children ...Widget) Widget {
	// In gogpu, Stacks are often just boxes with no flex direction or a special Stack widget
	// For now, we'll use a Box that doesn't advance children
	return primitives.Box(children...)
}

// Background sets the box background color
func (u *UIBox) Background(c Color) *UIBox {
	u.BoxWidget.Background(c)
	return u
}

// Width sets an explicit width
func (u *UIBox) Width(v float32) *UIBox {
	u.BoxWidget.Width(v)
	return u
}

// Height sets an explicit height
func (u *UIBox) Height(v float32) *UIBox {
	u.BoxWidget.Height(v)
	return u
}

// Gap sets the spacing between children
func (u *UIBox) Gap(v float32) *UIBox {
	u.BoxWidget.Gap(v)
	return u
}

// Rounded sets the border radius
func (u *UIBox) Rounded(v float32) *UIBox {
	u.BoxWidget.Rounded(v)
	return u
}

// Padding sets uniform padding (helper returning the same widget for chaining)
func (u *UIBox) Padding(v float32) *UIBox {
	u.BoxWidget.Padding(v)
	return u
}

// PaddingXY sets symmetric padding
func (u *UIBox) PaddingXY(x, y float32) *UIBox {
	u.BoxWidget.PaddingXY(x, y)
	return u
}

// Middle is currently a no-op placeholder for alignment
func (u *UIBox) Middle() *UIBox {
	return u
}

// ID registers the widget for simulator targeting
func (u *UIBox) ID(id string) *UIBox {
	RegisterWidget(id, u)
	return u
}

// Alignment sets the content alignment within the box
func (u *UIBox) Alignment(a int) *UIBox {
	// In the real gogpu, this maps to u.BoxWidget.Alignment(a)
	return u
}

// Shadow adds a visual depth effect to the box
func (u *UIBox) Shadow(radius float32) *UIBox {
	// Placeholder for GPU shadow effect
	return u
}

// OnClick attaches a click listener to the box itself
func (u *UIBox) OnClick(fn func()) *UIBox {
	u.onClick = fn
	return u
}

// OnMouseEnter attaches a listener for mouse entry
func (u *UIBox) OnMouseEnter(fn func()) *UIBox {
	u.onMouseEnter = fn
	return u
}

// OnMouseExit attaches a listener for mouse exit
func (u *UIBox) OnMouseExit(fn func()) *UIBox {
	u.onMouseExit = fn
	return u
}

// Event intercepts events to handle custom listeners
func (u *UIBox) Event(ctx widget.Context, e event.Event) bool {
	if me, ok := any(e).(event.MouseEvent); ok {
		switch me.MouseType {
		case event.MousePress:
			if u.onClick != nil && me.Button == event.ButtonLeft {
				u.onClick()
				return true
			}
		case event.MouseEnter:
			if u.onMouseEnter != nil {
				u.onMouseEnter()
			}
		case event.MouseLeave:
			if u.onMouseExit != nil {
				u.onMouseExit()
			}
		}
	}
	return u.BoxWidget.Event(ctx, e)
}

// Expand wraps the box in an Expanded widget
func (u *UIBox) Expand() Widget {
	return primitives.Expanded(u)
}

// Expanded wraps a widget to fill space
func Expanded(child Widget) Widget {
	return primitives.Expanded(child)
}

// Spacer creates an empty box that can expand if w and h are 0
func Spacer(w, h float32) Widget {
	if w == 0 && h == 0 {
		return primitives.Expanded(primitives.Box())
	}
	return primitives.Box().Width(w).Height(h)
}

// --- Basic Widgets ---

// Label creates static text
func Label(text string) *primitives.TextWidget {
	return primitives.Text(text)
}

// LabelSignal creates text that updates reactively
func LabelSignal(sig signals.ReadonlySignal[string]) *primitives.TextWidget {
	return primitives.Text("").ContentSignal(sig)
}

// --- Image Widget ---

type MockImageSource struct {
	W, H float32
}

func (m MockImageSource) Bounds() [2]float32 { return [2]float32{m.W, m.H} }

func Image(w, h float32) *primitives.ImageWidget {
	return primitives.Image(MockImageSource{W: w, H: h}).Size(w, h)
}

func ScrollView(child Widget) Widget {
	return scrollview.New(child)
}

// --- Dynamic Widget (View Switcher) ---

type dynamicWidget struct {
	widget.WidgetBase
	signal signals.ReadonlySignal[Widget]
	last   Widget
	ctx    widget.Context
}

func (d *dynamicWidget) updateChild() {
	current := d.signal.Get()
	if current == d.last {
		return
	}

	log.Printf("DYNAMIC: Child changed from %p to %p", d.last, current)

	// If we have a context, we need to handle transition
	if d.ctx != nil {
		if d.last != nil {
			// In a real framework we'd unmount d.last here
		}
		if current != nil {
			log.Printf("DYNAMIC: Mounting new child %p", current)
			// Check if the widget implements Mountable interface
			if m, ok := any(current).(interface{ Mount(widget.Context) }); ok {
				m.Mount(d.ctx)
			}
		}
		// The scheduler from BindToScheduler will handle invalidating us,
		// but we can also trigger a repaint if we have access to a context
	}
	d.last = current
}

func (d *dynamicWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	d.updateChild()
	if d.last == nil {
		return geometry.Size{}
	}
	return d.last.Layout(ctx, c)
}

func (d *dynamicWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	d.updateChild()
	if d.last != nil {
		widget.StampScreenOrigin(d.last, canvas)
		d.last.Draw(ctx, canvas)
	}
}

func (d *dynamicWidget) Event(ctx widget.Context, e event.Event) bool {
	d.updateChild()
	if d.last != nil {
		return d.last.Event(ctx, e)
	}
	return false
}

func (d *dynamicWidget) Children() []widget.Widget {
	if d.last == nil {
		return nil
	}
	return []widget.Widget{d.last}
}

func (d *dynamicWidget) Mount(ctx widget.Context) {
	d.ctx = ctx
	d.updateChild()
	if d.signal != nil {
		b := uistate.BindToScheduler(d.signal, d, ctx.Scheduler())
		d.AddBinding(b)
	}
}

func Dynamic(sig signals.ReadonlySignal[Widget]) Widget {
	return &dynamicWidget{signal: sig}
}

type ButtonStyle struct {
	BackgroundColor Color
	TextColor      Color
}

// Button Config Wrapper to provide your documented API
type ButtonConfig struct {
	Label   string
	OnClick func(context.Context) error
	Variant int // 0 = Filled, 1 = TextOnly, etc.
	Style   signals.ReadonlySignal[ButtonStyle]
}

// Button creates an interactive button with Hover and Active states
func Button(cfg ButtonConfig) Widget {
	// Local interaction state
	isHovered := signals.New(false)
	
	// Convert OnClick(ctx context.Context) error -> gogpu func()
	action := func() {
		log.Printf("UI: Button '%s' clicked", cfg.Label)
		if cfg.OnClick != nil {
			_ = cfg.OnClick(context.Background())
		}
	}

	// We use the raw button for core logic, but wrap it for high-fidelity state
	btn := button.New(
		button.Text(cfg.Label),
		button.OnClick(action),
	)

	// High-Fidelity Reactive Styling
	root := Dynamic(NewComputed(func() Widget {
		hover := isHovered.Get()
		pressed := false
		if hover && GlobalInput != nil {
			pressed = GlobalInput.Mouse().Pressed(input.MouseButtonLeft)
		}

		// Calculate visual properties
		bg := RGB(240, 240, 240) // Default
		if cfg.Style != nil {
			bg = cfg.Style.Get().BackgroundColor
		}

		// Apply hover/pressed transformations
		c := Container(btn).Background(bg).Rounded(8)
		if pressed {
			c.Padding(10) // Subtle squeeze
		} else if hover {
			c.Padding(12).Shadow(4)
		} else {
			c.Padding(12)
		}
		return c
	}, isHovered))

	// Attach hover listeners to the raw button widget if possible, 
	// or wrap in an interactive box
	return Container(root).OnClick(action).OnMouseEnter(func() {
		isHovered.Set(true)
	}).OnMouseExit(func() {
		isHovered.Set(false)
	})
}

type CheckBoxConfig struct {
	Label    string
	Checked  bool
	Signal   signals.Signal[bool]
	OnChange func(context.Context, bool) error
}

func CheckBox(cfg CheckBoxConfig) Widget {
	action := func(checked bool) {
		if cfg.OnChange != nil {
			_ = cfg.OnChange(context.Background(), checked)
		}
	}

	initial := cfg.Checked
	if cfg.Signal != nil {
		initial = cfg.Signal.Get()
	}

	return checkbox.New(
		checkbox.Label(cfg.Label),
		checkbox.Checked(initial),
		checkbox.OnToggle(action),
	)
}

type TextInputConfig struct {
	Placeholder string
	Value       signals.Signal[string]
	OnChange    func(context.Context, string) error
}

func TextInput(cfg TextInputConfig) Widget {
	action := func(val string) {
		if cfg.OnChange != nil {
			_ = cfg.OnChange(context.Background(), val)
		}
	}

	return textfield.New(
		textfield.Placeholder(cfg.Placeholder),
		textfield.Value(cfg.Value), // Binds directly!
		textfield.OnChange(action),
	)
}

// Slider Config wrapper
type SliderConfig struct {
	Value    float32
	Min      float32
	Max      float32
	OnChange func(context.Context, float32) error
}

func Slider(cfg SliderConfig) Widget {
	action := func(val float32) {
		if cfg.OnChange != nil {
			_ = cfg.OnChange(context.Background(), val)
		}
	}

	return slider.New(
		slider.Min(cfg.Min),
		slider.Max(cfg.Max),
		slider.Value(cfg.Value),
		slider.OnChange(action),
	)
}

// --- Shell & Dialogs ---

// ShowMessageDialog shows a native system alert box.
// On macOS it uses osascript; others fallback to terminal logging.
func ShowMessageDialog(ctx context.Context, title, message string, onClosed func()) {
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf("display alert %q message %q as informational buttons {\"OK\"} default button \"OK\"", title, message)
		_ = exec.Command("osascript", "-e", script).Run()
	} else {
		fmt.Printf("DIALOG: [%s] %s\n", title, message)
	}

	if onClosed != nil {
		onClosed()
	}
}


