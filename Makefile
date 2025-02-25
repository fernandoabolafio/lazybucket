.PHONY: build run clean test

# Build the application
build:
	go build -o gbuckets ./cmd/gbuckets

# Run the application (requires GOOGLE_CLOUD_PROJECT environment variable)
run: build
	./gbuckets

# Run with a specific project ID
run-with-project: build
	@if [ -z "$(PROJECT_ID)" ]; then \
		echo "Error: PROJECT_ID is not set. Use: make run-with-project PROJECT_ID=your-project-id"; \
		exit 1; \
	fi
	./gbuckets --project=$(PROJECT_ID)

# Clean build artifacts
clean:
	rm -f gbuckets

# Run tests
test:
	go test ./...

# Install the application
install:
	go install ./cmd/gbuckets

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o gbuckets-darwin-amd64 ./cmd/gbuckets
	GOOS=darwin GOARCH=arm64 go build -o gbuckets-darwin-arm64 ./cmd/gbuckets
	GOOS=linux GOARCH=amd64 go build -o gbuckets-linux-amd64 ./cmd/gbuckets
	GOOS=windows GOARCH=amd64 go build -o gbuckets-windows-amd64.exe ./cmd/gbuckets

# Default target
default: build 