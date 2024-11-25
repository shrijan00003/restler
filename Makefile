# Default target
.DEFAULT_GOAL := help

# Print help message
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  help                Show this help message"
	@echo "  build               Compile the Go project"
	@echo "  test                Run Go tests"
	@echo "  run                 Run the Go project"
	@echo "  restler             Execute RESTler tests"
	@echo "  clean               Clean up build artifacts"
	@echo "  other               Execute other bash commands"

# Build Go project
.PHONY: build
build:
	@echo "Building Go project..."
	go build ./...

# Run Go tests
.PHONY: test
test:
	@echo "Running Go tests..."
	go test ./...

# Run Go project
.PHONY: run
run:
	@echo "Running Go project..."
	go run bin/main.go

# Execute RESTler tests
.PHONY: restler
restler:
	@echo "Executing RESTler tests..."
	restler p col1/col2/col3/article/something.post.yaml

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up build artifacts..."
	go clean
	@rm -f *.exe
	@rm -rf build/

# Execute other bash commands
.PHONY: other
other:
	@echo "Executing other bash commands..."
	@echo "You can add other specific commands here"

# Optional additional commands
.PHONY: custom
custom:
	@echo "Running custom commands..."
    # Add custom commands here


.PHONY: flow
flow:
	go run bin/main.go p col1/sample
	go run bin/main.go p col1/sample/sample-1
