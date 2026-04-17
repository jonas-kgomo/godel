package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godel",
	Short: "Gödel cross-platform framework CLI",
	Long: `Gödel is a GPU-accelerated desktop UI framework for Go.
Use this CLI to initialize, develop, profile, and build your applications.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add commands here as we build them out
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(buildCmd)
}

// --- Init Command ---

var templateFlag string

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Gödel project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("Initializing Gödel project '%s' using template '%s'...\n", projectName, templateFlag)
		// Scaffolding logic goes here
		// 1. mkdir projectName
		// 2. generate go.mod
		// 3. generate main.go
		// 4. generate godel.toml
		log.Println("Project initialized successfully.")
	},
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "basic", "Project template (basic, dashboard, plugin-system)")
}

// --- Dev Command ---

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run the application with hot reload",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Gödel development server...")
		fmt.Println("Watching files in the current directory...")
		// fsnotify watcher logic goes here
		// launch `go run main.go`
		// restart process on file changes
	},
}

// --- Build Command ---

var targetFlag string
var releaseFlag string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile the application for distribution",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Building Gödel application for target: %s (Release: %s)\n", targetFlag, releaseFlag)
		// cross-compilation logic goes here
		// parsing godel.toml
		// generating assets
		// running go build with tags/ldflags
	},
}

func init() {
	buildCmd.Flags().StringVar(&targetFlag, "target", "current", "Target platform (current, linux, macos, windows, all)")
	buildCmd.Flags().StringVar(&releaseFlag, "release", "false", "Build optimized release binary")
}
