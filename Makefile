.PHONY: build run test clean

# Project variables
BINARY_NAME=go-web3-api
MAIN_PACKAGE=./cmd/api

# Build the application
build:
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Run the application
run:
	go run $(MAIN_PACKAGE)

# Run tests
test:
	go test -v ./...

# Clean up
clean:
	go clean
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	go mod tidy

# Run with hot reload (requires air - https://github.com/cosmtrek/air)
dev:
	air -c .air.toml
