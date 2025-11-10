VERSION  := 0.1.0
TARGET   := newrelic_exporter

include Makefile.COMMON

# Test targets
.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-short
test-short:
	go test -short ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem ./...

# Code quality targets
.PHONY: fmt
fmt:
	gofmt -w -s .

.PHONY: fmt-check
fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: fmt-check vet

# Combined check target
.PHONY: check
check: lint test

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make              - Build the binary (default)"
	@echo "  make test         - Run all tests with race detection"
	@echo "  make test-coverage- Run tests and generate coverage report"
	@echo "  make test-short   - Run tests in short mode"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make bench        - Run benchmarks"
	@echo "  make fmt          - Format code with gofmt"
	@echo "  make fmt-check    - Check if code is formatted"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run fmt-check and vet"
	@echo "  make check        - Run lint and test"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make archive      - Create release archive"