package main

import (
	"fmt"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
)

func main() {
	b := core.NewBody("Extended WebGPU UI Demo")

	// Title
	core.NewText(b).SetText("CogentCore UI Demo")

	// Text input
	tf := core.NewTextField(b).SetPlaceholder("Type something...")
	result := core.NewText(b).SetText("Your input will appear here")

	tf.OnChange(func(e events.Event) {
		result.SetText("You typed: " + tf.Text())
		b.Update()
	})

	// Button
	btnResult := core.NewText(b).SetText("Button not clicked yet")
	btn := core.NewButton(b).SetText("Click Me")

	btn.OnClick(func(e events.Event) {
		btnResult.SetText("Button clicked")
		b.Update()
	})

	// Checkbox (In CogentCore this is a Switch)
	swResult := core.NewText(b).SetText("Switch is off")
	sw := core.NewSwitch(b).SetText("Enable feature")

	sw.OnChange(func(e events.Event) {
		if sw.IsChecked() {
			swResult.SetText("Switch is on")
		} else {
			swResult.SetText("Switch is off")
		}
		b.Update()
	})

	// Slider
	sliderResult := core.NewText(b).SetText("Slider value: 0")
	slider := core.NewSlider(b)
	slider.SetMin(0).SetMax(100).SetValue(0)

	slider.OnChange(func(e events.Event) {
		val := slider.Value
		sliderResult.SetText(fmt.Sprintf("Slider value: %.0f", val))
		b.Update()
	})

	// Dropdown (Select is Chooser in CogentCore)
	chooserResult := core.NewText(b).SetText("Selected: none")
	
	chooser := core.NewChooser(b).SetStrings("Option A", "Option B", "Option C")
	
	chooser.OnChange(func(e events.Event) {
		chooserResult.SetText(fmt.Sprintf("Selected: %v", chooser.CurrentItem.Value))
		b.Update()
	})

	// Multi-line text area (TextEditor or Multi-line TextField usually, we'll use a standard TextField here or TextEditor)
	// We'll use TextField but with a larger size or a different widget if available. 
	// For now let's just make another Textfield
	textArea := core.NewTextField(b).SetPlaceholder("Write something longer...")
	textAreaResult := core.NewText(b).SetText("Text area content will appear here")

	textArea.OnChange(func(e events.Event) {
		textAreaResult.SetText("Content: " + textArea.Text())
		b.Update()
	})

	// Run app
	b.RunMainWindow()
}
