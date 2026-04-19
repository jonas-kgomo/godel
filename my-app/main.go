package main

import (
	"context"
	"log"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/ui"
)

func main() {
	myApp := app.New(
		app.WithTitle("my-app"),
		app.WithSize(800, 600),
	)

	myApp.OnReady(func(ctx context.Context) error {
		root := ui.Container(
			ui.Label("Welcome to my-app").FontSize(32).Bold(),
			ui.Spacer(0, 20),
			ui.Label("Edit main.go to start building your application."),
			ui.Spacer(0, 40),
			ui.Button(ui.ButtonConfig{
				Label: "Click Me",
				OnClick: func(ctx context.Context) error {
					log.Println("Button clicked!")
					return nil
				},
			}),
		).Padding(50).Center()

		myApp.SetRoot(root)
		return nil
	})

	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}
