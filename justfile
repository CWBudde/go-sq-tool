# Build the CLI
build:
    go build -v -o go-sq-tool ./...

# Run all tests
test:
    go test -v -race -count=1 ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem -run=^$ ./...

# Run linters
lint:
    golangci-lint run

# Run linters and fix issues
lint-fix:
    golangci-lint run --fix

# Format code using treefmt
fmt:
    treefmt . --allow-missing-formatter

# Check if code is formatted
fmt-check:
    treefmt --allow-missing-formatter --fail-on-change

# Generate coverage report
cover:
    go test -coverprofile=coverage.txt -covermode=atomic ./...
    go tool cover -html=coverage.txt -o coverage.html

# Clean build artifacts
clean:
    rm -f coverage.txt coverage.html go-sq-tool

# Build and serve the WASM demo
web-demo:
    if [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/; else cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/; fi
    GOOS=js GOARCH=wasm go build -buildvcs=false -o web/sqdecoder.wasm .
    python3 -m http.server --directory web 8080

# Run all checks (test, lint, coverage)
check: test lint cover

# Default target
default: build
