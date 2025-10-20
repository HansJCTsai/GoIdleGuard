# Makefile – Build, test and package for multiple platforms
BIN_DIR := bin
CONFIG := config.yaml

.PHONY: all build linux windows macos ui daemon test clean

all: test build

# Build for host OS
build: ui daemon

# Build UI
ui:
	@echo "→ Building UI (host OS)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/app-ui ./cmd/main
	@echo "   → $(BIN_DIR)/app-ui"

# Build Daemon
daemon:
	@echo "→ Building Daemon (host OS)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/app-daemon ./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon created."

# ----------------------------------------
# Debug targets
# ----------------------------------------

# build-debug: compile the daemon with debug information
#   -gcflags "all=-N -l" disables optimizations and inlining for easier single-stepping
debug-build:
	@echo "→ Building daemon with debug info…"
	@mkdir -p $(BIN_DIR)
	@go build -gcflags "all=-N -l" \
		-o $(BIN_DIR)/app-daemon-debug \
		./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon-debug created."

# debug: run the compiled binary under Delve in interactive mode
# passes the generated config.yaml as an argument
debug-config: debug-build
	@echo "🐛 Launching Delve in interactive mode…"
	@dlv exec $(BIN_DIR)/app-daemon-debug -- -config=$(BIN_DIR)/config.yaml

# debug-headless: start Delve in headless mode for remote attachment
debug-headless:
	@echo "🐛 Launching Delve in headless mode"
	@dlv debug github.com/HanksJCTsai/goidleguard/cmd/daemon/ -- \
		-config=$(BIN_DIR)/config.yaml \
		--log

# Run tests
test:
	@echo "→ Running tests..."
	@go test ./...
	@echo "   ✔ All tests passed."

config-windows:
	@echo "→ Installing $(CONFIG) to $(BIN_DIR)..."
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
	@cmd /C "copy /Y "$(CONFIG)" "$(BIN_DIR)"
	@echo "   → $(BIN_DIR)\\config.yaml created."

# --------------------
# Cross‑platform targets
# --------------------

# Linux amd64
linux: linux-ui linux-daemon config

linux-ui:
	@echo "→ Building UI for Linux/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/app-ui-linux ./cmd/main
	@echo "   → $(BIN_DIR)/app-ui-linux"

linux-daemon:
	@echo "→ Building Daemon for Linux/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/app-daemon-linux ./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon-linux"

# Windows amd64
windows: windows-ui windows-daemon config-windows

windows-ui: config-windows
	@echo "→ Building UI for Windows/amd64…"
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
	@cmd /C "set GOOS=windows&& set GOARCH=amd64&& go build -o $(BIN_DIR)/app-ui.exe ./cmd/main"
	@echo "   → $(BIN_DIR)/app-ui.exe"

windows-daemon: config-windows
	@echo "→ Building Daemon for Windows/amd64..."
	@if not exist "$(BIN_DIR)" mkdir -p "$(BIN_DIR)"
	@cmd /C "set GOOS=windows&& set GOARCH=amd64&& go build -o $(BIN_DIR)/app-daemon.exe ./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon.exe"

# macOS amd64
macos: macos-ui macos-daemon config

macos-ui:
	@echo "→ Building UI for macOS/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/app-ui-darwin ./cmd/main
	@echo "   → $(BIN_DIR)/app-ui-darwin"

macos-daemon:
	@echo "→ Building Daemon for macOS/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/app-daemon-darwin ./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon-darwin"

# --------------------
# Clean
# --------------------
clean:
	@echo "→ Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "   ✔ Removed $(BIN_DIR)/"
