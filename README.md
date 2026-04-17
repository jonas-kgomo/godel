# Gödel: Cross-Platform Development Engine

Gödel is a GPU-accelerated desktop UI framework for Go, designed for building lean, consistent, cross-platform Native applications. It leverages the rendering and widget power of `gogpu/ui` as its foundation, and wraps it with an opinionated, developer-first platform offering CLI tooling, native OS integration, and hot-reload.

## Installation & Setup

1. **Clone the repository and install dependencies:**
   ```bash
   cd path/to/godel
   go mod tidy
   ```

2. **Install the `godel` CLI tool:**
   To use the `godel` command globally, install it to your Go bin path:
   ```bash
   go install ./cmd/godel
   ```
   *(Ensure `$(go env GOPATH)/bin` is in your system `$PATH`)*

## Running the Examples

You can run the built-in examples directly without the `godel` CLI:

```bash
# Run the Hello World example
go run examples/hello-world/main.go
```

## Using the Gödel CLI

Once installed, the CLI provides the developer platform layer:

```bash
# Initialize a new project
godel init my-app

# Run the project in development mode (with hot reloading)
godel dev

# Build the project for release
godel build --release
```

## Testing & Benchmarking

*(Note: Test and benchmark suites are currently being built out.)*

To run unit tests across the framework:
```bash
go test ./...
```

To run benchmarks (measuring rendering latency and memory):
```bash
go test -bench=. -benchmem ./...
```
