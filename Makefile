.PHONY: help lint test build clean tidy fmt

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage report saved to coverage.out"

build: ## Build all packages with strict checks
	@echo "Building packages with strict checks..."
	go build -v -race -trimpath -buildvcs=true -a ./...

build-release: ## Build optimized for release
	@echo "Building optimized for release..."
	go build -v -ldflags="-s -w" -trimpath -buildvcs=true ./...

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
	@echo "Setting up development environment..."
	@command -v goimports >/dev/null 2>&1 || { echo "Installing goimports..."; go install golang.org/x/tools/cmd/goimports@latest; }
	@command -v lefthook >/dev/null 2>&1 || { echo "Installing lefthook..."; go install github.com/evilmartians/lefthook@latest; }
	@lefthook install
	@echo "Development environment ready!"

hooks-install: ## Install git hooks
	@echo "Installing git hooks..."
	@lefthook install
	@echo "Git hooks installed!"

hooks-uninstall: ## Uninstall git hooks
	@echo "Uninstalling git hooks..."
	@lefthook uninstall
	@echo "Git hooks uninstalled!"

hooks-run: ## Run all hooks manually
	@echo "Running all hooks..."
	@lefthook run pre-commit

# Release workflow  
pre-release: fmt lint test build ## Pre-release checks
	@echo "Ready for release!"