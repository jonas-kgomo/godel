package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	signals "github.com/coregx/signals"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	uistate "github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

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


// --- Signals / State Management mappings ---

func NewSignal[T any](initial T) signals.Signal[T] {
	return signals.New[T](initial)
}

func NewComputed[T any](fn func() T) signals.ReadonlySignal[T] {
	return signals.Computed[T](fn)
}

// --- Layout Primitives ---

type UIBox struct {
	*primitives.BoxWidget
}

// VStack creates a vertical box layout
func VStack(children ...Widget) *UIBox {
	return &UIBox{primitives.Box(children...).SetDirection(primitives.DirectionVertical)}
}

// HStack creates a horizontal box layout
func HStack(children ...Widget) *UIBox {
	return &UIBox{primitives.Box(children...).SetDirection(primitives.DirectionHorizontal)}
}

// Container creates a flexible box with padding and background
func Container(children ...Widget) *UIBox {
	return &UIBox{primitives.Box(children...)}
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

	// If we have a context, we need to handle transition
	if d.ctx != nil {
		if d.last != nil {
			// In a real framework we'd unmount d.last here
		}
		if current != nil {
			current.Mount(d.ctx)
		}
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
	if d.last != nil {
		widget.StampScreenOrigin(d.last, canvas)
		d.last.Draw(ctx, canvas)
	}
}

func (d *dynamicWidget) Event(ctx widget.Context, e event.Event) bool {
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

// Button Config Wrapper to provide your documented API
type ButtonConfig struct {
	Label   string
	OnClick func(context.Context) error
	Variant int // 0 = Filled, 1 = TextOnly, etc.
}

// Button creates an interactive button
func Button(cfg ButtonConfig) Widget {
	// Convert OnClick(ctx context.Context) error -> gogpu func()
	action := func() {
		if cfg.OnClick != nil {
			_ = cfg.OnClick(context.Background()) // Ignore error for now
		}
	}

	btn := button.New(
		button.Text(cfg.Label),
		button.OnClick(action),
	)

	// In real implementation we'd map Variant to button.VariantFilled etc
	return btn
}

// CheckBox Config wrapper
type CheckBoxConfig struct {
	Checked  bool
	OnChange func(context.Context, bool) error
}

func CheckBox(cfg CheckBoxConfig) Widget {
	action := func(checked bool) {
		if cfg.OnChange != nil {
			_ = cfg.OnChange(context.Background(), checked)
		}
	}

	return checkbox.New(
		checkbox.Checked(cfg.Checked),
		checkbox.OnToggle(action),
	)
}

// TextInput Config wrapper
type TextInputConfig struct {
	Placeholder string
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


