package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/intercode/godel/pkg/app"
	"github.com/intercode/godel/pkg/ui"
)

type FeedItem struct {
	ID       int
	Title    string
	Author   string
	Metadata string
	Likes    int
}

func main() {
	myApp := app.New(
		app.WithTitle("Gödel Premium Dashboard"),
		app.WithSize(1100, 800),
	)

	// --- Global State ---
	activeTab := ui.NewSignal("home")
	
	// Home State
	userName := ui.NewSignal("Gödel Developer")
	
	// Feed State
	searchText := ui.NewSignal("")
	selectedItem := ui.NewSignal[*FeedItem](nil)
	allItems := []FeedItem{
		{1, "Ethereal Mountains", "Alex Riverside", "4032 x 3024", 124},
		{2, "Neon Cybercity", "Luna Vector", "1920 x 1080", 89},
		{3, "Golden Hour Forest", "Oak Stream", "5120 x 2880", 256},
		{4, "Arctic Silence", "Frost Walker", "3840 x 2160", 42},
		{5, "Desert Mirage", "Sand Piper", "6000 x 4000", 178},
		{6, "Deep Ocean Blue", "Marina Deep", "8000 x 4500", 512},
	}
	feedItems := ui.NewSignal(allItems)
	
	// Live Updates Simulation
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			items := feedItems.Get()
			if len(items) > 0 {
				items[0].Likes += 5
				feedItems.Set(items)
			}
		}
	}()

	myApp.OnReady(func(ctx context.Context) error {

		// --- View Builders ---

		buildSidebar := func() ui.Widget {
			navButton := func(label, id string) ui.Widget {
				return ui.Button(ui.ButtonConfig{
					Label: label,
					OnClick: func(ctx context.Context) error {
						activeTab.Set(id)
						return nil
					},
				})
			}

			return ui.VStack(
				ui.Label("GÖDEL OPS").Bold().FontSize(22).Color(ui.RGB(99, 102, 241)),
				ui.Spacer(0, 40),
				navButton("Dashboard", "home"),
				ui.Spacer(0, 10),
				navButton("Nature Feed", "feed"),
				ui.Spacer(0, 10),
				navButton("Typography", "typo"),
				ui.Spacer(0, 10),
				navButton("Analytics", "analytics"),
				ui.Spacer(0, 0), // Filler spacer
				ui.Label("v0.2.7-beta").FontSize(10).Color(ui.RGB(150, 150, 150)),
			).Padding(24).Width(220).Background(ui.RGB(245, 246, 250))
		}

		buildHomeView := func() ui.Widget {
			return ui.VStack(
				ui.Label("System Overview").Bold().FontSize(32),
				ui.Spacer(0, 20),
				ui.Container(
					ui.VStack(
						ui.Label("Welcome back,").FontSize(16).Color(ui.RGB(100, 100, 100)),
						ui.LabelSignal(ui.NewComputed(func() string { return userName.Get() })).Bold().FontSize(24),
					),
				).Padding(20).Background(ui.RGB(240, 240, 255)).Rounded(12),
				ui.Spacer(0, 30),
				ui.Label("Quick Actions").Bold().FontSize(18),
				ui.Spacer(0, 12),
				ui.HStack(
					ui.Button(ui.ButtonConfig{Label: "Refresh Data"}),
					ui.Spacer(12, 0),
					ui.Button(ui.ButtonConfig{Label: "Export Logs"}),
				),
			).Padding(40)
		}

		buildFeedView := func() ui.Widget {
			filteredItems := ui.NewComputed(func() []FeedItem {
				search := strings.ToLower(searchText.Get())
				if search == "" {
					return feedItems.Get()
				}
				var filtered []FeedItem
				for _, item := range feedItems.Get() {
					if strings.Contains(strings.ToLower(item.Title), search) || 
					   strings.Contains(strings.ToLower(item.Author), search) {
						filtered = append(filtered, item)
					}
				}
				return filtered
			})

			detailPanel := ui.NewComputed(func() ui.Widget {
				item := selectedItem.Get()
				if item == nil {
					return ui.VStack(
						ui.Spacer(0, 100),
						ui.Label("Select an item to view details").Color(ui.RGB(150, 150, 150)),
					).Padding(20)
				}
				
				return ui.VStack(
					ui.Label(item.Title).Bold().FontSize(24),
					ui.Label("Photo by "+item.Author).Color(ui.RGB(100, 100, 100)),
					ui.Spacer(0, 20),
					ui.Image(400, 250).Rounded(12),
					ui.Spacer(0, 20),
					ui.Label("Metadata").Bold().FontSize(16),
					ui.Label("Resolution: "+item.Metadata),
					ui.Label(fmt.Sprintf("Global Likes: %d", item.Likes)),
					ui.Spacer(0, 20),
					ui.Button(ui.ButtonConfig{Label: "Download Original"}),
				).Padding(20)
			})

			listView := ui.NewComputed(func() ui.Widget {
				items := filteredItems.Get()
				var rows []ui.Widget
				for _, item := range items {
					i := item // capture
					row := ui.Button(ui.ButtonConfig{
						Label: fmt.Sprintf("%s (%s)", i.Title, i.Author),
						OnClick: func(ctx context.Context) error {
							selectedItem.Set(&i)
							return nil
						},
					})
					rows = append(rows, row, ui.Spacer(0, 8))
				}
				return ui.VStack(rows...)
			})

			return ui.HStack(
				ui.VStack(
					ui.Label("Nature Gallery").Bold().FontSize(28),
					ui.Spacer(0, 16),
					ui.TextInput(ui.TextInputConfig{
						Placeholder: "Search nature...",
						OnChange: func(ctx context.Context, s string) error {
							searchText.Set(s)
							return nil
						},
					}),
					ui.Spacer(0, 16),
					ui.ScrollView(ui.Dynamic(listView)),
				).Width(350).Padding(24),
				
				ui.Container(ui.Dynamic(detailPanel)).Background(ui.RGB(250, 250, 252)),
			)
		}

		buildTypoView := func() ui.Widget {
			return ui.VStack(
				ui.Label("SDF Typography Demo").Bold().FontSize(32),
				ui.Label("Infinite resolution glyphs powered by Signed Distance Fields.").FontSize(16).Color(ui.RGB(100, 100, 100)),
				ui.Spacer(0, 40),
				
				ui.Label("Scale Comparison").Bold(),
				ui.Spacer(0, 20),
				ui.Label("Tiny Text (10px)").FontSize(10),
				ui.Label("Standard Interface (14px)").FontSize(14),
				ui.Label("Subheading (22px)").FontSize(22),
				ui.Label("Large Headline (48px)").FontSize(48).Bold().Color(ui.RGB(99, 102, 241)),
				ui.Label("Hero Impact (96px)").FontSize(96).Bold(),
				
				ui.Spacer(0, 40),
				ui.Label("The Benefit:").Bold(),
				ui.Label("Unlike raster fonts, SDF fonts maintain sharp edges at any zoom level without pixelation or heavy memory overhead."),
			).Padding(40)
		}

		// Main Switcher
		currentView := ui.NewComputed(func() ui.Widget {
			tab := activeTab.Get()
			switch tab {
			case "feed": return buildFeedView()
			case "typo": return buildTypoView()
			default: return buildHomeView()
			}
		})

		root := ui.HStack(
			buildSidebar(),
			ui.Container(ui.Dynamic(currentView)),
		).Background(ui.RGB(255, 255, 255))

		myApp.SetRoot(root)
		return nil
	})

	if err := myApp.Run(); err != nil {
		log.Fatal(err)
	}
}
