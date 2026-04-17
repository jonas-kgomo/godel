// Package config parses godel.toml project configuration files.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// GodelConfig represents the full godel.toml configuration.
type GodelConfig struct {
	App         AppConfig         `toml:"app"`
	Build       BuildConfig       `toml:"build"`
	Platforms   PlatformsConfig   `toml:"platforms"`
	Theme       ThemeConfig       `toml:"theme"`
	Development DevelopmentConfig `toml:"development"`
	Release     ReleaseConfig     `toml:"release"`
	Renderer    RendererConfig    `toml:"renderer"`
	Windows     WindowsConfig     `toml:"windows"`
	MacOS       MacOSConfig       `toml:"macos"`
	Linux       LinuxConfig       `toml:"linux"`
}

// AppConfig holds application metadata.
type AppConfig struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Author      string `toml:"author"`
	License     string `toml:"license"`
}

// BuildConfig holds build settings.
type BuildConfig struct {
	Main     string `toml:"main"`
	Output   string `toml:"output"`
	Icon     string `toml:"icon"`
	BundleID string `toml:"bundle_id"`
	Width    int    `toml:"width"`
	Height   int    `toml:"height"`
}

// PlatformsConfig specifies which platforms to target.
type PlatformsConfig struct {
	Windows bool `toml:"windows"`
	MacOS   bool `toml:"macos"`
	Linux   bool `toml:"linux"`
	Android bool `toml:"android"`
	IOS     bool `toml:"ios"`
}

// ThemeConfig holds theme settings.
type ThemeConfig struct {
	Primary      string `toml:"primary"`
	Secondary    string `toml:"secondary"`
	Accent       string `toml:"accent"`
	DesignSystem string `toml:"design_system"` // material3, fluent, cupertino, custom
}

// DevelopmentConfig holds dev mode settings.
type DevelopmentConfig struct {
	HotReload bool   `toml:"hot_reload"`
	Debug     bool   `toml:"debug"`
	LogLevel  string `toml:"log_level"`
	Renderer  string `toml:"renderer"` // vulkan, metal, dx12, software
}

// ReleaseConfig holds release build settings.
type ReleaseConfig struct {
	Strip    bool     `toml:"strip"`
	Optimize bool     `toml:"optimize"`
	Targets  []string `toml:"target"`
}

// RendererConfig holds rendering settings.
type RendererConfig struct {
	Backend          string `toml:"backend"` // auto, vulkan, metal, dx12, software
	VulkanValidation bool   `toml:"vulkan_validation"`
	FrameRate        int    `toml:"frame_rate"`
	VSync            bool   `toml:"vsync"`
}

// WindowsConfig holds Windows-specific settings.
type WindowsConfig struct {
	Icon        string `toml:"icon"`
	Sign        bool   `toml:"sign"`
	Certificate string `toml:"certificate"`
}

// MacOSConfig holds macOS-specific settings.
type MacOSConfig struct {
	Icon                 string `toml:"icon"`
	Sign                 bool   `toml:"sign"`
	TeamID               string `toml:"team_id"`
	ProvisioningProfile  string `toml:"provisioning_profile"`
}

// LinuxConfig holds Linux-specific settings.
type LinuxConfig struct {
	Icon       string   `toml:"icon"`
	Categories []string `toml:"categories"`
}

// Load reads and parses a godel.toml file.
func Load(path string) (*GodelConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	cfg := &GodelConfig{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config syntax: %w", err)
	}

	// Apply defaults
	if cfg.Build.Main == "" {
		cfg.Build.Main = "main.go"
	}
	if cfg.Build.Width == 0 {
		cfg.Build.Width = 800
	}
	if cfg.Build.Height == 0 {
		cfg.Build.Height = 600
	}
	if cfg.Renderer.FrameRate == 0 {
		cfg.Renderer.FrameRate = 60
	}
	if cfg.Theme.DesignSystem == "" {
		cfg.Theme.DesignSystem = "material3"
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "0.1.0"
	}

	return cfg, nil
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *GodelConfig {
	return &GodelConfig{
		App: AppConfig{
			Name:    "My Gödel App",
			Version: "0.1.0",
		},
		Build: BuildConfig{
			Main:   "main.go",
			Output: "myapp",
			Width:  800,
			Height: 600,
		},
		Platforms: PlatformsConfig{
			Windows: true,
			MacOS:   true,
			Linux:   true,
		},
		Theme: ThemeConfig{
			Primary:      "#007AFF",
			Accent:       "#FF3B30",
			DesignSystem: "material3",
		},
		Development: DevelopmentConfig{
			HotReload: true,
			Debug:     true,
			LogLevel:  "debug",
		},
		Renderer: RendererConfig{
			Backend:   "auto",
			FrameRate: 60,
			VSync:     true,
		},
	}
}
