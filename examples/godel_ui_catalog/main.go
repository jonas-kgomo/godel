package main

import (
	"context"
	"fmt"
	"log"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/godelui"
	"github.com/intercode/godel/pkg/ui"
)

func main() {
	myApp := app.New(
		app.WithTitle("Gödel UI Component Catalog"),
		app.WithSize(1100, 800),
	)

	myApp.OnReady(func(ctx context.Context) error {
		
		// Header Section
		pageHeader := ui.VStack(
			ui.HStack(
				ui.Label("Gödel").Bold().FontSize(32).Color(godelui.ColorPrimary),
				ui.Label(" UI").Bold().FontSize(32).Color(ui.RGB(200, 200, 200)),
				ui.Spacer(20, 0),
				godelui.Badge("v0.1.0-alpha", godelui.ColorSecondary),
			).Middle(),
			ui.Label("The high-performance component system for Go-native desktop apps.").Color(ui.RGB(150, 150, 150)),
		).Padding(40)

		// Grid Content
		grid := ui.HStack(
			// Left Column: Inputs & Controls
			ui.VStack(
				godelui.Card("Inputs & Controls",
					ui.Label("Name").FontSize(12).Color(ui.RGB(120, 120, 120)),
					ui.TextInput(ui.TextInputConfig{Placeholder: "Enter your name..."}),
					ui.Spacer(0, 16),
					
					ui.Label("Volume").FontSize(12).Color(ui.RGB(120, 120, 120)),
					ui.Slider(ui.SliderConfig{Min: 0, Max: 100, Value: 50}),
					ui.Spacer(0, 16),
					
					ui.HStack(
						ui.CheckBox(ui.CheckBoxConfig{Checked: true}),
						ui.Label("Enable GPU Acceleration").Color(ui.RGB(200, 200, 200)),
					).Gap(8),
				),
				
				ui.Spacer(0, 24),
				
				godelui.Card("Buttons Gallery",
					ui.HStack(
						godelui.PrimaryButton("Primary Action", nil),
						ui.Button(ui.ButtonConfig{Label: "Outline"}),
					).Gap(12),
					ui.Spacer(0, 16),
					ui.Button(ui.ButtonConfig{Label: "Large Destructive Action"}),
				),
			).Width(400),

			ui.Spacer(32, 0),

			// Right Column: Dashboard Elements
			ui.VStack(
				godelui.Card("Performance Metrics (GPU)",
					ui.Label("Real-time telemetry from Apple M1 pipeline").FontSize(12).Color(ui.RGB(100, 100, 100)),
					ui.Spacer(0, 20),
					ui.HStack(
						metricItem("GPU Load", "68%", godelui.ColorPrimary),
						metricItem("VRAM", "4.2GB", godelui.ColorSecondary),
						metricItem("FPS", "60.0", ui.RGB(34, 197, 94)),
					).Gap(20),
					ui.Spacer(0, 20),
					// Simulated Chart area
					ui.Container().Height(120).Background(ui.RGBA(255, 255, 255, 10)).Rounded(8),
				),

				ui.Spacer(0, 24),

				godelui.Card("Recent Activity",
					ui.VStack(
						activityItem("Renderer initialized", "2m ago"),
						activityItem("Shader compiled (SDF)", "5m ago"),
						activityItem("File 'main.go' saved", "12m ago"),
					).Gap(12),
				),
			).Expand(),
		).PaddingXY(40, 0)

		// Page Assembly
		root := ui.VStack(
			pageHeader,
			grid,
		).Background(godelui.ColorBackground)

		myApp.SetRoot(root)
		return nil
	})

	fmt.Println("✨ Launching Gödel UI Catalog...")
	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func metricItem(label, value string, color ui.Color) ui.Widget {
	return ui.VStack(
		ui.Label(label).FontSize(12).Color(ui.RGB(150, 150, 150)),
		ui.Label(value).Bold().FontSize(24).Color(color),
	).Expand()
}

func activityItem(msg, time string) ui.Widget {
	return ui.HStack(
		ui.Container().Width(8).Height(8).Background(godelui.ColorPrimary).Rounded(4),
		ui.Spacer(8, 0),
		ui.Label(msg).Color(ui.RGB(200, 200, 200)),
		ui.Spacer(0, 0),
		ui.Label(time).FontSize(11).Color(ui.RGB(100, 100, 100)),
	).Middle()
}
