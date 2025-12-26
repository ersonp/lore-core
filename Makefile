.PHONY: format lint test test-integration vet build clean check

# Format all Go files with goimports (excluding vendor)
format:
	find . -name '*.go' -not -path './vendor/*' | xargs goimports -w

# Run golangci-lint
lint:
	golangci-lint run

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
	go clean

# Run all checks (format, vet, lint, test)
check: format vet lint test

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
