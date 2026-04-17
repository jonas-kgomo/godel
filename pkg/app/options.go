package app

// Option is a functional option for configuring the App.
type Option func(*Config)

// WithTitle sets the window title.
func WithTitle(title string) Option {
	return func(c *Config) {
		c.Title = title
	}
}

// WithSize sets the initial window dimensions.
func WithSize(width, height int) Option {
	return func(c *Config) {
		c.Width = width
		c.Height = height
	}
}

// WithMinSize sets the minimum window dimensions.
func WithMinSize(width, height int) Option {
	return func(c *Config) {
		c.MinWidth = width
		c.MinHeight = height
	}
}

// WithResizable sets whether the window can be resized.
func WithResizable(resizable bool) Option {
	return func(c *Config) {
		c.Resizable = resizable
	}
}

// WithConfig loads configuration from a godel.toml file.
func WithConfig(path string) Option {
	return func(c *Config) {
		c.ConfigFile = path
	}
}

// WithFrameRate sets the target frame rate (default 60).
func WithFrameRate(fps int) Option {
	return func(c *Config) {
		c.FrameRate = fps
	}
}

// WithContinuousRender enables continuous rendering (vs event-driven).
// Event-driven (default) uses 0% CPU when idle.
// Continuous is useful for animations or real-time content.
func WithContinuousRender(continuous bool) Option {
	return func(c *Config) {
		c.EventDriven = !continuous
	}
}

// WithVSync enables or disables vertical sync.
func WithVSync(vsync bool) Option {
	return func(c *Config) {
		c.VSync = vsync
	}
}
