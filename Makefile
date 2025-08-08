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

# CI/CD workflow
ci: fmt lint test build security-check ## Run CI checks
	@echo "CI checks passed!"

security-check: ## Run security checks
	@echo "Running security checks..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...
	@echo "Security checks passed!"

integration-test: ## Run integration tests (requires services)
	@echo "Running integration tests..."
	@if [ "$$INTEGRATION_TESTS" = "true" ]; then \
		go test -v -race -tags=integration ./...; \
	else \
		echo "Skipping integration tests (set INTEGRATION_TESTS=true to run)"; \
	fi

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	@echo "Profile saved to cpu.prof"

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./...
	@echo "Profile saved to mem.prof"

# Release workflow  
pre-release: fmt lint test build security-check ## Pre-release checks
	@echo "Ready for release!"

# Documentation
docs-serve: ## Serve documentation locally
	@echo "Serving documentation..."
	@command -v godoc >/dev/null 2>&1 || { echo "Installing godoc..."; go install golang.org/x/tools/cmd/godoc@latest; }
	@echo "Documentation available at http://localhost:6060/pkg/github.com/rdashevsky/go-pkgs/"
	godoc -http=:6060

# Docker support
docker-build: ## Build Docker image for testing
	@echo "Building Docker image..."
	docker build -f .github/Dockerfile -t go-pkgs:test .

docker-test: ## Run tests in Docker container
	@echo "Running tests in Docker..."
	docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine go test -v ./...

# Tools installation
install-tools: ## Install all required development tools
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/evilmartians/lefthook@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "All tools installed!"

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated!"

mod-graph: ## Show module dependency graph
	@echo "Module dependency graph:"
	go mod graph

mod-why: ## Show why packages are needed
	@echo "Checking why packages are needed..."
	@read -p "Enter package name: " package; go mod why $$package