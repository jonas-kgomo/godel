package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/ui"
)

func main() {
	// Use our new framework wrapper
	myApp := app.New(
		app.WithTitle("Gödel + WebGPU Core Example"),
		app.WithSize(600, 600),
		app.WithContinuousRender(false), // 0% CPU idle!
	)

	// Our reactive state (mirrored from coregx/signals)
	counter := ui.NewSignal(0)
	lastUpdate := ui.NewSignal(time.Now().Format("15:04:05"))
	
	// App Lifecycle
	myApp.OnReady(func(ctx context.Context) error {
		
		// Build the UI tree using our primitives
		root := ui.Container(
			ui.Label("Hello from Gödel!").FontSize(28).Bold(),
			ui.Spacer(0, 20),
			
			// Reactive text binding
			ui.LabelSignal(ui.NewComputed(func() string {
				return fmt.Sprintf("Counter Value: %d", counter.Get())
			})).FontSize(24),
			
			ui.Spacer(0, 20),
			
			ui.HStack(
				ui.Button(ui.ButtonConfig{
					Label: "- Decrement",
					OnClick: func(ctx context.Context) error {
						// State updates trigger automatic repaints in Cogent Core via Gödel's binding
						counter.Set(counter.Get() - 1)
						lastUpdate.Set(time.Now().Format("15:04:05"))
						return nil
					},
				}),
				ui.Spacer(20, 0),
				ui.Button(ui.ButtonConfig{
					Label: "+ Increment",
					OnClick: func(ctx context.Context) error {
						counter.Set(counter.Get() + 1)
						lastUpdate.Set(time.Now().Format("15:04:05"))
						return nil
					},
				}),
			).Gap(16),
			
			ui.Spacer(0, 40),
			
			// Background task example (safe concurrency)
			ui.Button(ui.ButtonConfig{
				Label: "Simulate Expensive Task",
				OnClick: func(ctx context.Context) error {
					go func() {
						time.Sleep(2 * time.Second)
						
						// SAFELY update state from goroutine using QueueCallback
						myApp.QueueCallback(func() {
							counter.Set(counter.Get() + 100)
							lastUpdate.Set(time.Now().Format("15:04:05 (Async)"))
						})
					}()
					return nil
				},
			}),

			ui.Spacer(0, 20),
			
			ui.LabelSignal(ui.NewComputed(func() string {
				return "Last updated: " + lastUpdate.Get()
			})).Color(ui.RGB(100, 100, 100)),
			
		).Padding(40).Gap(10).Background(ui.RGB(255, 255, 255))
		
		myApp.SetRoot(root)
		return nil
	})

	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}
