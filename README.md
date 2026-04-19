# Gödel: Cross-Platform Development Engine

Gödel is a GPU-accelerated desktop UI framework for Go, designed for building lean, consistent, cross-platform Native applications. It leverages the rendering and widget power of `gogpu/ui` as its foundation, and wraps it with an opinionated, developer-first platform offering CLI tooling, native OS integration, and hot-reload.

## Installation & Setup

1. **Clone the repository and install dependencies:**
   ```bash
   cd path/to/godel
   go mod tidy
   ```

2. **Install the `godel` CLI tool:**
   The `godel` CLI abstracts away Go internals (it automatically handles `CGO_ENABLED=0` for you!). Install it by running:
   ```bash
   go install ./cmd/godel
   ```
   **Important:** Ensure your Go path is available in your terminal. Add this to your `~/.zshrc` or `~/.bashrc` if you see a "command not found" error:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

## Running Applications

You can run the built-in examples or any `main.go` file directly using `godel run`. You never need to worry about CGO flags again!

```bash
# Run the Hello World example
godel run examples/hello-world/main.go

# Run the complex Dashboard example
godel run examples/dashboard/main.go
```

## Using the Gödel CLI

The Gödel developer platform layer simplifies the build process:

```bash
# Initialize a new project
godel init my-app

# Run the project in development mode (hot reloading)
godel dev

# Build the project for release
godel build --release
```

## Testing & Benchmarking

*(Note: Test and benchmark suites are currently being built out.)*

For internal engine tests only (requires manual flags):
```bash
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go test -bench=. -benchmem ./...
```




How to make 

```bash
export PATH=$PATH:/Users/jonaskgomo/go/bin
 

echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```