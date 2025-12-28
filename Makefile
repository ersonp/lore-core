.PHONY: format lint lint-custom lint-all test test-integration vet build clean check vendor mocks tools build-linter test-linter

# Format all Go files with goimports (excluding vendor and tools)
format:
	find . -name '*.go' -not -path './vendor/*' -not -path './tools/*/vendor/*' | xargs goimports -w

# Run golangci-lint
lint:
	golangci-lint run

# Run custom lore-lint
lint-custom:
	@if [ ! -f tools/lore-lint/bin/lore-lint ]; then \
		echo "Building lore-lint..."; \
		$(MAKE) -C tools/lore-lint build; \
	fi
	tools/lore-lint/bin/lore-lint ./...

# Run all linters
lint-all: lint lint-custom

# Run all tests
test:
	go test -v ./...

# Run integration tests (requires Qdrant running on localhost:6334)
test-integration:
	INTEGRATION_TEST=1 go test -v ./tests/integration/...

# Run go vet
vet:
	go vet ./...

# Build the binary
build:
	go build -o bin/lore ./cmd/lore

# Clean build artifacts
clean:
	rm -rf bin/
	$(MAKE) -C tools/lore-lint clean
	go clean

# Run all checks (format, vet, lint-all, test)
check: format vet lint-all test

# Vendor dependencies
vendor:
	go mod tidy
	go mod vendor

# Generate mocks
mocks:
	mockery --all --output=mocks

# Install development tools
tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	$(MAKE) -C tools/lore-lint build

# Build custom linter
build-linter:
	$(MAKE) -C tools/lore-lint build

# Test custom linter
test-linter:
	$(MAKE) -C tools/lore-lint test
