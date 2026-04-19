package main

import (
	"context"
	"fmt"
	"log"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/shell"
	"github.com/intercode/godel/pkg/ui"
)

func main() {
	myApp := app.New(
		app.WithTitle("Gödel Features Demo"),
		app.WithSize(600, 400),
	)

	myApp.OnReady(func(ctx context.Context) error {

		content := ui.VStack(
			ui.Label("Native Shell Integrations").Bold().FontSize(24),
			ui.Spacer(0, 20),

			ui.Button(ui.ButtonConfig{
				Label: "Send System Notification",
				OnClick: func(ctx context.Context) error {
					err := shell.Notify("Gödel Engine", "This is a native macOS notification!")
					if err != nil {
						log.Printf("Notification failed: %v", err)
					}
					return err
				},
			}),

			ui.Spacer(0, 10),

			ui.Button(ui.ButtonConfig{
				Label: "Select File (Native Picker)",
				OnClick: func(ctx context.Context) error {
					path, err := shell.SelectFile("Choose a file for Gödel")
					if err != nil {
						log.Printf("File picker error: %v", err)
						return nil
					}
					if path == "" {
						return nil // User cancelled
					}
					ui.ShowMessageDialog(ctx, "File Selected", "You chose: "+path, nil)
					return nil
				},
			}),

			ui.Spacer(0, 10),

			ui.Button(ui.ButtonConfig{
				Label: "Open Documentation (Browser)",
				OnClick: func(ctx context.Context) error {
					return shell.OpenURL("https://github.com/intercode/godel")
				},
			}),

			ui.Spacer(0, 10),

			ui.HStack(
				ui.Button(ui.ButtonConfig{
					Label: "Activate System Tray",
					OnClick: func(ctx context.Context) error {
						shell.TrayIcon("Gödel Engine", nil)
						return shell.Notify("Tray Active", "Minimized to system tray (Simulated)")
					},
				}),
				ui.WithID("btn-dialog", ui.Button(ui.ButtonConfig{
					Label: "Show App Dialog",
					OnClick: func(ctx context.Context) error {
						ui.ShowMessageDialog(ctx, "Info", "Native dialogs are coming in v0.2!", nil)
						return nil
					},
				})),
			).Gap(10),

			ui.Spacer(0, 30),
			ui.Label("These features run Zero-CGO using system bridges.").FontSize(12).Color(ui.RGB(150, 150, 150)),
		).Padding(40).Background(ui.RGB(255, 255, 255))

		myApp.SetRoot(content)

		myApp.OnSimulate(func(ctx context.Context) error {
			myApp.LogSimStep("FEATURES TOUR")
			myApp.SimulateClickOn("btn-dialog")
			return nil
		})

		return nil
	})

	fmt.Println("🚀 Running Features Demo...")
	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}
