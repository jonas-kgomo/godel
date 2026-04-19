package ui

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os/exec"
	"runtime"
	"sync"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	signals "github.com/coregx/signals"
	"github.com/intercode/godel/pkg/input"
)

// Widget is the core interface for Gödel UI components in CogentCore.
type Widget interface {
	Build(parent core.Widget) core.Widget
}

// WidgetFunc allows using functions as widgets.
type WidgetFunc func(parent core.Widget) core.Widget

func (f WidgetFunc) Build(parent core.Widget) core.Widget {
	return f(parent)
}

// WidgetBuilder is a function that populates a CogentCore Body.
type WidgetBuilder func(body *core.Body)

var GlobalInput *input.State

// Color is a standard Go color.
type Color = color.Color

// Common color helpers mapping to CogentCore
func RGB(r, g, b uint8) Color     { return color.RGBA{r, g, b, 255} }
func RGBA(r, g, b, a uint8) Color { return color.RGBA{r, g, b, a} }
func Hex(h string) Color {
	c, err := colors.FromString(h)
	if err != nil {
		return color.RGBA{0, 0, 0, 255}
	}
	return c
}

// --- Widget Registry for Simulation ---

var (
	widgetRegistry = make(map[string]core.Widget)
	registryMu     sync.RWMutex
)

func RegisterWidget(id string, w core.Widget) {
	registryMu.Lock()
	defer registryMu.Unlock()
	widgetRegistry[id] = w
}

func GetWidgetByID(id string) core.Widget {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return widgetRegistry[id]
}

// --- Signals / State Management mappings ---

func NewSignal[T any](initial T) signals.Signal[T] {
	return signals.New[T](initial)
}

func NewComputed[T any](fn func() T, deps ...any) signals.ReadonlySignal[T] {
	return signals.Computed[T](fn, deps...)
}

// --- UIBox: The primary layout container ---

type UIBox struct {
	children []Widget
	dir      styles.Directions
	styles   []func(s *styles.Style)
	id       string
	onClick  func()
}

func (u *UIBox) Build(parent core.Widget) core.Widget {
	ly := core.NewFrame(parent)
	ly.Style(func(s *styles.Style) {
		s.Direction = u.dir
		for _, f := range u.styles {
			f(s)
		}
	})

	if u.id != "" {
		RegisterWidget(u.id, ly)
	}

	if u.onClick != nil {
		ly.OnClick(func(e events.Event) {
			u.onClick()
		})
	}

	for _, child := range u.children {
		if child != nil {
			child.Build(ly)
		}
	}
	return ly
}

func (u *UIBox) Background(c Color) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Background = colors.Uniform(c)
	})
	return u
}

func (u *UIBox) Padding(v float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Padding.Set(units.Dp(v))
	})
	return u
}

func (u *UIBox) PaddingXY(x, y float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Padding.Left.Set(x, units.UnitDp)
		s.Padding.Right.Set(x, units.UnitDp)
		s.Padding.Top.Set(y, units.UnitDp)
		s.Padding.Bottom.Set(y, units.UnitDp)
	})
	return u
}

func (u *UIBox) Gap(v float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Gap.Set(units.Dp(v))
	})
	return u
}

func (u *UIBox) Width(v float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Min.X.Set(units.Dp(v))
	})
	return u
}

func (u *UIBox) Height(v float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Min.Y.Set(units.Dp(v))
	})
	return u
}

func (u *UIBox) Rounded(v float32) *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Border.Radius.Set(units.Dp(v))
	})
	return u
}

func (u *UIBox) OnClick(fn func()) *UIBox {
	u.onClick = fn
	return u
}

func (u *UIBox) ID(id string) *UIBox {
	u.id = id
	return u
}

func (u *UIBox) Alignment(a int) *UIBox {
	// CogentCore uses Justify and Align
	return u
}

func (u *UIBox) Shadow(v float32) *UIBox { return u }

func (u *UIBox) Expand() Widget {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	return u
}

func (u *UIBox) Center() *UIBox {
	u.styles = append(u.styles, func(s *styles.Style) {
		s.Justify.Content = styles.Center
		s.Align.Content = styles.Center
		s.Justify.Items = styles.Center
		s.Align.Items = styles.Center
	})
	return u
}

// --- Layout Functions ---

func VStack(children ...Widget) *UIBox {
	return &UIBox{children: children, dir: styles.Column}
}

func HStack(children ...Widget) *UIBox {
	return &UIBox{children: children, dir: styles.Row}
}

func Container(children ...Widget) *UIBox {
	return &UIBox{children: children, dir: styles.Column}
}

func Spacer(w, h float32) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		sp := core.NewFrame(parent)
		sp.Style(func(s *styles.Style) {
			if w == 0 && h == 0 {
				s.Grow.Set(1, 1)
			} else {
				if w > 0 {
					s.Min.X.Set(units.Dp(w))
				}
				if h > 0 {
					s.Min.Y.Set(units.Dp(h))
				}
			}
		})
		return sp
	})
}

// --- Basic Widgets ---

type TextWidget struct {
	text   string
	signal signals.ReadonlySignal[string]
	styles []func(s *styles.Style)
}

func (t *TextWidget) Build(parent core.Widget) core.Widget {
	txt := core.NewText(parent)
	if t.signal != nil {
		txt.SetText(t.signal.Get())
		// Reactive update
		t.signal.On(func(val string) {
			txt.SetText(val)
			txt.Update()
		})
	} else {
		txt.SetText(t.text)
	}

	txt.Style(func(s *styles.Style) {
		for _, f := range t.styles {
			f(s)
		}
	})
	return txt
}

func (t *TextWidget) FontSize(v float32) *TextWidget {
	t.styles = append(t.styles, func(s *styles.Style) {
		s.Font.Size.Set(units.Dp(v))
	})
	return t
}

func (t *TextWidget) Color(c Color) *TextWidget {
	t.styles = append(t.styles, func(s *styles.Style) {
		s.Color = colors.Uniform(c)
	})
	return t
}

func (t *TextWidget) Bold() *TextWidget {
	t.styles = append(t.styles, func(s *styles.Style) {
		s.Font.Weight = styles.WeightBold
	})
	return t
}

func Label(text string) *TextWidget {
	return &TextWidget{text: text}
}

func LabelSignal(sig signals.ReadonlySignal[string]) *TextWidget {
	return &TextWidget{signal: sig}
}

// --- Interactive Widgets ---

type ButtonConfig struct {
	Label   string
	OnClick func(context.Context) error
	Style   signals.ReadonlySignal[ButtonStyle]
}

type ButtonStyle struct {
	BackgroundColor Color
	TextColor       Color
}

func Button(cfg ButtonConfig) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		btn := core.NewButton(parent).SetText(cfg.Label)
		btn.OnClick(func(e events.Event) {
			if cfg.OnClick != nil {
				_ = cfg.OnClick(context.Background())
			}
		})
		if cfg.Style != nil {
			btn.Style(func(s *styles.Style) {
				st := cfg.Style.Get()
				if st.BackgroundColor != nil {
					s.Background = colors.Uniform(st.BackgroundColor)
				}
				if st.TextColor != nil {
					s.Color = colors.Uniform(st.TextColor)
				}
			})
			cfg.Style.On(func(st ButtonStyle) {
				btn.Update()
			})
		}
		return btn
	})
}

type CheckBoxConfig struct {
	Label    string
	Checked  bool
	OnChange func(context.Context, bool) error
}

func CheckBox(cfg CheckBoxConfig) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		cb := core.NewSwitch(parent).SetText(cfg.Label)
		// cb.SetChecked(cfg.Checked) // Check if SetChecked exists
		cb.OnChange(func(e events.Event) {
			if cfg.OnChange != nil {
				_ = cfg.OnChange(context.Background(), cb.IsChecked())
			}
		})
		return cb
	})
}

type TextInputConfig struct {
	Placeholder string
	Value       signals.Signal[string]
	OnChange    func(context.Context, string) error
}

func TextInput(cfg TextInputConfig) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		tf := core.NewTextField(parent).SetPlaceholder(cfg.Placeholder)
		if cfg.Value != nil {
			tf.SetText(cfg.Value.Get())
			cfg.Value.On(func(val string) {
				if tf.Text() != val {
					tf.SetText(val)
					tf.Update()
				}
			})
		}
		tf.OnChange(func(e events.Event) {
			val := tf.Text()
			if cfg.Value != nil {
				cfg.Value.Set(val)
			}
			if cfg.OnChange != nil {
				_ = cfg.OnChange(context.Background(), val)
			}
		})
		return tf
	})
}

func Slider(cfg SliderConfig) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		s := core.NewSlider(parent)
		s.SetMin(cfg.Min).SetMax(cfg.Max).SetValue(cfg.Value)
		s.OnChange(func(e events.Event) {
			if cfg.OnChange != nil {
				_ = cfg.OnChange(context.Background(), s.Value)
			}
		})
		return s
	})
}

type SliderConfig struct {
	Value    float32
	Min      float32
	Max      float32
	OnChange func(context.Context, float32) error
}

// --- Helper Functions ---

func Image(w, h float32) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		img := core.NewImage(parent)
		img.Style(func(s *styles.Style) {
			s.Min.X.Set(units.Dp(w))
			s.Min.Y.Set(units.Dp(h))
		})
		return img
	})
}

func ScrollView(child Widget) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		sv := core.NewFrame(parent)
		sv.Style(func(s *styles.Style) {
			s.Overflow.Set(styles.OverflowAuto)
		})
		if child != nil {
			child.Build(sv)
		}
		return sv
	})
}

func Dynamic(sig signals.ReadonlySignal[Widget]) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		// In CogentCore, we can use a frame and clear it
		frame := core.NewFrame(parent)
		
		update := func() {
			frame.DeleteChildren()
			w := sig.Get()
			if w != nil {
				w.Build(frame)
			}
			frame.Update()
		}
		
		update()
		sig.On(func(w Widget) {
			update()
		})
		
		return frame
	})
}

func WithID(id string, w Widget) Widget {
	return WidgetFunc(func(parent core.Widget) core.Widget {
		cw := w.Build(parent)
		RegisterWidget(id, cw)
		return cw
	})
}

const (
	AlignCenter = 1
)

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
