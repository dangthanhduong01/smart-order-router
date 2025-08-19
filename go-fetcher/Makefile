# Uniswap Pool Fetcher Makefile

.PHONY: help build test clean run-example run-tests install-deps

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the project"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-example   - Run example program"
	@echo "  install-deps  - Install dependencies"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"

# Build the project
build:
	@echo "Building uniswap-pool-fetcher..."
	go build -o bin/pool-fetcher .
	@echo "Build complete! Binary available at bin/pool-fetcher"

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete!"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover ./...
	@echo "Tests with coverage complete!"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./test/
	@echo "Benchmarks complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "Clean complete!"

# Run example program
run-example:
	@echo "Running example program..."
	go run examples/example.go

# Run tests only
run-tests:
	@echo "Running tests only..."
	go test -v ./test/

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting complete!"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting complete!"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t uniswap-pool-fetcher .
	@echo "Docker image built!"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -it uniswap-pool-fetcher

# Quick demo
demo:
	@echo "Running quick demo..."
	@echo "Fetching V3 pools from Mainnet..."
	./bin/pool-fetcher -chain 1 -protocol v3 -top 5
	@echo "Demo complete!"

# Performance test
perf-test:
	@echo "Running performance test..."
	./bin/pool-fetcher -chain 1 -protocol all -verbose
	@echo "Performance test complete!"

# All chains test
test-all-chains:
	@echo "Testing all supported chains..."
	@for chain in 1 10 42161 137 8453 56 43114 42220; do \
		echo "Testing chain $$chain..."; \
		./bin/pool-fetcher -chain $$chain -protocol v3 -top 1 || echo "Chain $$chain failed"; \
	done
	@echo "All chains test complete!"

# Development setup
dev-setup: install-deps build
	@echo "Development setup complete!"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/pool-fetcher .
	@echo "Production build complete!"

# Install to system
install: build
	@echo "Installing to system..."
	sudo cp bin/pool-fetcher /usr/local/bin/
	@echo "Installation complete!"

# Uninstall from system
uninstall:
	@echo "Uninstalling from system..."
	sudo rm -f /usr/local/bin/pool-fetcher
	@echo "Uninstallation complete!"

# Show help for CLI
cli-help:
	@echo "CLI Usage Examples:"
	@echo ""
	@echo "Fetch all pools from Mainnet:"
	@echo "  ./bin/pool-fetcher -chain 1 -protocol all"
	@echo ""
	@echo "Fetch V3 pools from Optimism:"
	@echo "  ./bin/pool-fetcher -chain 10 -protocol v3"
	@echo ""
	@echo "Fetch V2 pools from Arbitrum and save to file:"
	@echo "  ./bin/pool-fetcher -chain 42161 -protocol v2 -output arbitrum_v2_pools.json"
	@echo ""
	@echo "Show top 20 V4 pools from Base with verbose logging:"
	@echo "  ./bin/pool-fetcher -chain 8453 -protocol v4 -top 20 -verbose"
	@echo ""
	@echo "For more help:"
	@echo "  ./bin/pool-fetcher -h"
