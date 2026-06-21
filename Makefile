.PHONY: build run test clean docker-up docker-down

# Build the application binary
build:
	@echo "Building Shrimpy binary..."
	@go build -o bin/shrimpy ./cmd/shrimpy

# Run the application locally
run:
	@echo "Running Shrimpy locally..."
	@go run cmd/shrimpy/main.go

# Run tests
test:
	@echo "Running unit tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# Start local docker-compose environment
docker-up:
	@echo "Starting local containers..."
	@docker-compose up --build

# Tear down local containers
docker-down:
	@echo "Stopping local containers..."
	@docker-compose down -v
