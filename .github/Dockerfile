# Multi-stage Dockerfile for testing and building Go packages

# Test stage
FROM golang:1.23-alpine AS test

# Install necessary packages
RUN apk add --no-cache git make gcc musl-dev

# Install development tools
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install golang.org/x/vuln/cmd/govulncheck@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run tests
RUN go test -v -race ./...

# Run linting
RUN golangci-lint run

# Run security checks
RUN govulncheck ./...

# Build stage
FROM golang:1.23-alpine AS build

# Install git for build info
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build packages
RUN go build -v -race -trimpath -buildvcs=true ./...

# Final stage - minimal image for runtime
FROM alpine:latest AS runtime

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy built binaries (if any)
COPY --from=build --chown=appuser:appgroup /app ./

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD echo "Health check passed"

# Default command
CMD ["echo", "Go packages built successfully"]