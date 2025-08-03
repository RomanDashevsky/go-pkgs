.PHONY: help lint test build clean tidy fmt

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run --no-config

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

build: ## Build all packages
	@echo "Building packages..."
	go build ./...

clean: ## Clean build cache
	@echo "Cleaning..."
	go clean -cache -modcache -testcache

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

check: fmt lint test build ## Run all checks (format, lint, test, build)
	@echo "All checks passed!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

# Development workflow
dev-setup: deps ## Setup development environment
	@echo "Development environment ready!"

# Release workflow  
pre-release: fmt lint test build ## Pre-release checks
	@echo "Ready for release!"