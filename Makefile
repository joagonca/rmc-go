.PHONY: all build test clean help

# Binary name
BINARY_NAME=rmc

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Main package path
MAIN_PACKAGE=./cmd/rmc-go

# Test files
TEST_DIR=tests
TEST_FILES=$(wildcard $(TEST_DIR)/*.rm)
TEST_OUTPUT_DIR=test_output

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✓ Build complete: $(BINARY_NAME)"

# Run tests with test files
test: build
	@echo "Running tests with .rm files..."
	@mkdir -p $(TEST_OUTPUT_DIR)
	@echo ""
	@echo "Testing SVG export:"
	@echo "==================="
	@for file in $(TEST_FILES); do \
		basename=$$(basename $$file .rm); \
		echo "  Testing $$basename.rm → $$basename.svg"; \
		./$(BINARY_NAME) $$file -o $(TEST_OUTPUT_DIR)/$$basename.svg; \
		if [ -f $(TEST_OUTPUT_DIR)/$$basename.svg ]; then \
			size=$$(ls -lh $(TEST_OUTPUT_DIR)/$$basename.svg | awk '{print $$5}'); \
			echo "    ✓ Generated ($$size)"; \
		else \
			echo "    ✗ Failed"; \
			exit 1; \
		fi; \
	done
	@echo ""
	@echo "Testing PDF export:"
	@echo "==================="
	@for file in $(TEST_FILES); do \
		basename=$$(basename $$file .rm); \
		echo "  Testing $$basename.rm → $$basename.pdf"; \
		./$(BINARY_NAME) $$file -o $(TEST_OUTPUT_DIR)/$$basename.pdf 2>&1; \
		if [ -f $(TEST_OUTPUT_DIR)/$$basename.pdf ]; then \
			size=$$(ls -lh $(TEST_OUTPUT_DIR)/$$basename.pdf | awk '{print $$5}'); \
			echo "    ✓ Generated ($$size)"; \
		else \
			echo "    ⚠ Failed (Inkscape may not be installed)"; \
		fi; \
	done
	@echo ""
	@echo "Test outputs saved to: $(TEST_OUTPUT_DIR)/"
	@echo "✓ All tests completed"

# Run Go unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./...

# Clean build artifacts and test outputs
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf $(TEST_OUTPUT_DIR)
	@echo "✓ Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOGET) -v ./...
	@echo "✓ Dependencies installed"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build      - Build the $(BINARY_NAME) binary"
	@echo "  make test       - Run integration tests with .rm files"
	@echo "  make test-unit  - Run Go unit tests"
	@echo "  make clean      - Remove binary and test outputs"
	@echo "  make deps       - Install Go dependencies"
	@echo "  make all        - Build the binary (default)"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Test files in $(TEST_DIR)/:"
	@for file in $(TEST_FILES); do \
		echo "  - $$(basename $$file)"; \
	done
