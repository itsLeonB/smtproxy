.PHONY: help \
smtp \
smtp-hot \
lint \
test \
test-verbose \
test-coverage \
test-coverage-html \
test-clean \
build \
install-pre-push-hook \
uninstall-pre-push-hook

help:
	@echo "Makefile commands:"
	@echo "  make smtp                    - Start the SMTP server"
	@echo "  make smtp-hot                - Start the SMTP server with hot reload (requires air)"
	@echo "  make lint                    - Run golangci-lint on the codebase"
	@echo "  make test                    - Run all tests"
	@echo "  make test-verbose            - Run all tests with verbose output"
	@echo "  make test-coverage           - Run all tests with coverage report"
	@echo "  make test-coverage-html      - Run all tests and generate HTML coverage report"
	@echo "  make test-clean              - Clean test cache and run tests"
	@echo "  make build                   - Build smtp server for production"
	@echo "  make install-pre-push-hook   - Install git pre-push hook for linting and testing"
	@echo "  make uninstall-pre-push-hook - Uninstall git pre-push hook"

smtp:
	go run ./cmd/smtp

smtp-hot:
	@echo "🚀 Starting SMTP server with hot reload..."
	air --build.cmd "go build -o bin/smtp ./cmd/smtp" --build.bin "./bin/smtp"

lint:
	golangci-lint run ./...

test:
	@echo "Running all tests..."
	go test ./internal/...; \

test-verbose:
	@echo "Running all tests with verbose output..."
	go test -v ./internal/...; \

test-coverage:
	@echo "Running all tests with coverage report..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./internal/...; \

test-coverage-html:
	@echo "Running all tests and generating HTML coverage report..."
	go test -v -coverprofile=coverage.out ./internal/... && \
	go tool cover -html=coverage.out -o coverage.html && \
	echo "Coverage report generated: coverage.html"; \

test-clean:
	@echo "Cleaning test cache and running tests..."
	go clean -testcache && go test -v ./internal/...; \

build:
	@echo "Building SMTP server..."
	CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false -ldflags='-w -s' -o bin/smtp ./cmd/smtp
	@echo "Build success! Binary is located at bin/smtp"

install-pre-push-hook:
	@echo "Installing pre-push git hook..."
	@mkdir -p .git/hooks
	@cp scripts/pre-push .git/hooks/pre-push
	@echo "Pre-push hook installed successfully!"

uninstall-pre-push-hook:
	@echo "Uninstalling pre-push git hook..."
	@rm -f .git/hooks/pre-push
	@echo "Pre-push hook uninstalled successfully!"
