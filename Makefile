# MCP-RedTeam Makefile

# Variables
BINARY_NAME=mcp-redteam
GO=go
GOFLAGS=-mod=vendor
LDFLAGS=-ldflags "-s -w"
INSTALL_DIR=/usr/local/bin

# Default target
all: build

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod vendor

# Run tests
test:
	$(GO) test -v ./...

# Build the binary
build: deps
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)

# Install the binary
install: build
	@if [ ! -d $(INSTALL_DIR) ]; then \
		sudo mkdir -p $(INSTALL_DIR); \
	fi
	sudo install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

# Install CLI dependencies for macOS using Homebrew
install-macosx-commands:
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "Error: Homebrew is not installed. Please install it first: https://brew.sh/"; \
		exit 1; \
	fi
	brew install amass
	brew install projectdiscovery/tap/httpx
	brew install projectdiscovery/tap/dnsx
	brew install projectdiscovery/tap/katana
	brew install projectdiscovery/tap/nuclei
	brew install semgrep
	brew install shodan

# Install CLI dependencies for Linux using apt
install-linux-commands:
	@if ! command -v apt-get >/dev/null 2>&1; then \
		echo "Error: apt-get is not available. This script only supports Debian-based distributions."; \
		exit 1; \
	fi
	sudo apt-get update
	sudo apt-get install -y amass
	# Install ProjectDiscovery tools
	sudo apt-get install -y httpx dnsx katana nuclei
	# Install Semgrep
	sudo apt-get install -y semgrep
	# Install Shodan CLI
	sudo apt-get install -y python3-pip
	sudo pip3 install shodan

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf vendor/

# Run the binary
run: build
	./$(BINARY_NAME)

# Help target
help:
	@echo "Available targets:"
	@echo "  all                    - Build the binary (default)"
	@echo "  deps                   - Install Go dependencies"
	@echo "  test                   - Run tests"
	@echo "  build                  - Build the binary"
	@echo "  install                - Install the binary to $(INSTALL_DIR)"
	@echo "  install-macosx-commands - Install CLI dependencies using Homebrew (macOS)"
	@echo "  install-linux-commands  - Install CLI dependencies using apt (Linux)"
	@echo "  clean                  - Clean build artifacts"
	@echo "  run                    - Build and run the binary"
	@echo "  help                   - Show this help message"

.PHONY: all deps test build install install-macosx-commands install-linux-commands clean run help 