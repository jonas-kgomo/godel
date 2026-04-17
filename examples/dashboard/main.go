package main

import (
	"context"
	"log"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/ui"
)

func main() {
	myApp := app.New(
		app.WithTitle("Gödel Dashboard Example"),
		app.WithSize(900, 700),
		app.WithContinuousRender(false),
	)

	// Application state
	userName := ui.NewSignal("Gödel Developer")
	notificationsEnabled := ui.NewSignal(true)
	volumeLevel := ui.NewSignal(float32(50.0))

	myApp.OnReady(func(ctx context.Context) error {

		// Side pane layout
		sidePane := ui.VStack(
			ui.Label("Menu").Bold().FontSize(20).Color(ui.RGB(60, 60, 60)),
			ui.Spacer(0, 20),
			ui.Button(ui.ButtonConfig{
				Label: "Home",
			}),
			ui.Button(ui.ButtonConfig{
				Label: "Settings",
				OnClick: func(ctx context.Context) error {
					// We removed ShowMessageDialog from ui.go, skipping it here for now
					return nil
				},
			}),
			ui.Button(ui.ButtonConfig{
				Label: "Analytics",
			}),
		).Padding(24).Width(250).Background(ui.RGB(240, 240, 245))

		// Main content area
		mainContent := ui.VStack(
			ui.Label("Dashboard").Bold().FontSize(32),
			ui.Spacer(0, 20),
			
			ui.Label("Profile Settings").Bold().FontSize(18).Color(ui.RGB(100, 100, 100)),
			ui.Spacer(0, 10),
			
			// Form fields
			ui.HStack(
				ui.Container(ui.Label("Name:")).Width(100),
				ui.TextInput(ui.TextInputConfig{
					Placeholder: "Enter your name...",
					OnChange: func(ctx context.Context, val string) error {
						userName.Set(val)
						return nil
					},
				}),
			).Gap(16),
			
			ui.Spacer(0, 10),
			
			ui.HStack(
				ui.Container(ui.Label("Notifications:")).Width(100),
				ui.CheckBox(ui.CheckBoxConfig{
					Checked: notificationsEnabled.Get(),
					OnChange: func(ctx context.Context, checked bool) error {
						notificationsEnabled.Set(checked)
						return nil
					},
				}),
			).Gap(16),

			ui.Spacer(0, 10),

			ui.HStack(
				ui.Container(ui.Label("Volume:")).Width(100),
				ui.Container(ui.Slider(ui.SliderConfig{
					Min: 0,
					Max: 100,
					Value: volumeLevel.Get(),
					OnChange: func(ctx context.Context, val float32) error {
						volumeLevel.Set(val)
						return nil
					},
				})).Width(200),
			).Gap(16),
			
			ui.Spacer(0, 40),
			
			// Reactive output summary
			ui.Container(
				ui.VStack(
					ui.Label("Summary Configuration:").Bold(),
					ui.Spacer(0, 10),
					ui.LabelSignal(ui.NewComputed(func() string {
						return "Welcome, " + userName.Get()
					})),
					ui.LabelSignal(ui.NewComputed(func() string {
						status := "Disabled"
						if notificationsEnabled.Get() {
							status = "Enabled"
						}
						return "Notifications are " + status
					})),
				).Padding(20).Background(ui.RGB(220, 240, 220)).Rounded(10),
			),

		).Padding(40)

		// Root Layout (SidePane + MainContent)
		root := ui.HStack(
			sidePane,
			mainContent,
		).Background(ui.RGB(255, 255, 255))

		myApp.SetRoot(root)
		return nil
	})

	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}
