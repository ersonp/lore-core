.PHONY: build test clean tidy

# Build the linter binary
build:
	go build -o bin/lore-lint .

# Run all tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Update dependencies
tidy:
	go mod tidy
