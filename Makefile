.PHONY: build run test clean fmt lint vet

# Build the application
build:
	go build -o bin/app ./cmd/app

# Run the application
run:
	go run ./cmd/app

# Run all tests
test:
	go test ./... -v -count=1

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./... || true

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/ tmp/

# Tidy dependencies
tidy:
	go mod tidy
	go mod verify
