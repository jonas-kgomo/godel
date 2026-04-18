package godelui

import (
	"context"

	"github.com/intercode/godel/pkg/ui"
)

// Re-export common types from base ui for convenience
type (
	Widget  = ui.Widget
	Color   = ui.Color
	Context = ui.Context
)

// Brand colors from our Gödel UI mockup
var (
	ColorPrimary   = ui.RGB(99, 102, 241)  // Indigo 500
	ColorSecondary = ui.RGB(168, 85, 247)  // Purple 500
	ColorAccent    = ui.RGB(236, 72, 153)  // Pink 500
	ColorBackground = ui.RGB(15, 15, 20)   // Dark Slate
	ColorSurface   = ui.RGBA(30, 30, 40, 180) // Glassy surface
)

// --- Premium Components ---

// Card represents a Shadcn-style container
func Card(title string, children ...Widget) Widget {
	header := ui.VStack(
		ui.Label(title).Bold().FontSize(18).Color(ui.RGB(240, 240, 240)),
		ui.Spacer(0, 12),
	)
	
	content := ui.VStack(children...)
	
	return ui.VStack(header, content).
		Padding(24).
		Background(ColorSurface).
		Rounded(12)
}

// Badge is a small status indicator
func Badge(text string, color Color) Widget {
	return ui.Container(
		ui.Label(text).FontSize(11).Bold().Color(ui.RGB(255, 255, 255)),
	).PaddingXY(8, 4).Background(color).Rounded(20)
}

// PrimaryButton creates a branded indigo button
func PrimaryButton(label string, onClick func(context.Context) error) Widget {
	// Custom styling for buttons will be implemented by overriding the painter
	// For now we use the standard button with our brand color
	return ui.Button(ui.ButtonConfig{
		Label:   label,
		OnClick: onClick,
	})
}
