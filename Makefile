.PHONY: build run clean test

# Build the application
build:
	go build -o lazybucket ./cmd/lazybucket

# Run the application (requires GOOGLE_CLOUD_PROJECT environment variable)
run: build
	./lazybucket

# Run with a specific project ID
run-with-project: build
	@if [ -z "$(PROJECT_ID)" ]; then \
		echo "Error: PROJECT_ID is not set. Use: make run-with-project PROJECT_ID=your-project-id"; \
		exit 1; \
	fi
	./lazybucket --project=$(PROJECT_ID)

# Clean build artifacts
clean:
	rm -f lazybucket

# Run tests
test:
	go test ./...

# Install the application
install:
	go install ./cmd/lazybucket

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o lazybucket-darwin-amd64 ./cmd/lazybucket
	GOOS=darwin GOARCH=arm64 go build -o lazybucket-darwin-arm64 ./cmd/lazybucket
	GOOS=linux GOARCH=amd64 go build -o lazybucket-linux-amd64 ./cmd/lazybucket
	GOOS=windows GOARCH=amd64 go build -o lazybucket-windows-amd64.exe ./cmd/lazybucket

# Default target
default: build 