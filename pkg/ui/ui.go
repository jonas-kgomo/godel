package ui

import (
	"context"

	state "github.com/coregx/signals"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
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
func Hex(h string) Color            { /* TODO implement hex parser */ return RGB(0, 0, 0) }


// --- Signals / State Management mappings ---

func NewSignal[T any](initial T) state.Signal[T] {
	return state.NewSignal(initial)
}

func NewComputed[T any](fn func() T) state.Computed[T] {
	return state.NewComputed(fn)
}

// --- Layout Primitives ---

// VStack creates a vertical box layout
func VStack(children ...Widget) *primitives.BoxWidget {
	return primitives.Box(children...).SetDirection(primitives.DirectionVertical)
}

// HStack creates a horizontal box layout
func HStack(children ...Widget) *primitives.BoxWidget {
	return primitives.Box(children...).SetDirection(primitives.DirectionHorizontal)
}

// Container creates a flexible box with padding and background
func Container(children ...Widget) *primitives.BoxWidget {
	return primitives.Box(children...)
}

// Spacer creates an empty box that expands
func Spacer(w, h float32) *primitives.BoxWidget {
	// Not an exact equivalent to a generic flex spacer yet,
	// but serves as fixed-size spacing for now.
	return primitives.Box().Size(w, h)
}

// --- Basic Widgets ---

// Label creates static text
func Label(text string) *primitives.TextWidget {
	return primitives.Text(text)
}

// LabelSignal creates text that updates reactively
func LabelSignal(sig state.ReadonlySignal[string]) *primitives.TextWidget {
	// primitives.NewText("").ContentSignal(sig)
	// fallback for exact API match depending on gogpu version
	return primitives.Text("").ContentSignal(sig)
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
		button.Label(cfg.Label),
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
		checkbox.OnChange(action),
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

// --- Modal / Dialog ---

// ShowMessageDialog maps your simple dialog API to gogpu/ui core/dialog
func ShowMessageDialog(ctx context.Context, title, message string, onOK func()) {
	d := dialog.Alert(title, message, func() {
		if onOK != nil {
			onOK()
		}
	})
	d.Show(ctx)
}
