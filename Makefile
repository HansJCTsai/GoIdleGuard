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
	@copy /Y "$(CONFIG)" "$(BIN_DIR)\"
	@echo "   → $(BIN_DIR)/config.yaml created."

config-macos:
	@echo "→ Installing $(CONFIG) to $(BIN_DIR)..."
	@if [ ! -d "$(BIN_DIR)" ]; then \
		echo "   → $(BIN_DIR) not found, creating..."; \
		mkdir -p "$(BIN_DIR)"; \
	fi
	@cp "$(CONFIG)" "$(BIN_DIR)/config.yaml"
	@echo "   → $(BIN_DIR)/config.yaml created."
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
windows: clean windows-ui windows-daemon config-windows

windows-ui: clean config-windows
	@echo "→ Building UI for Windows/amd64…"
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
	@set GOOS=windows&& set GOARCH=amd64&& go build -o "$(BIN_DIR)/app-ui.exe" ./cmd/gui
	@echo "   → $(BIN_DIR)/app-ui.exe"

windows-daemon: clean config-windows
	@echo "→ Building Daemon for Windows/amd64..."
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
	@set GOOS=windows&& set GOARCH=amd64&& go build -ldflags -H=windowsgui -o "$(BIN_DIR)/app-daemon.exe" ./cmd/daemon
	@echo "  → $(BIN_DIR)/app-daemon.exe"

# macOS amd64
macos: clean macos-ui macos-daemon config

macos-ui: clean config-macos
	@echo "→ Building UI for macOS/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/app-ui-darwin ./cmd/gui
	@echo "   → $(BIN_DIR)/app-ui-darwin"

macos-daemon: clean config-macos
	@echo "→ Building Daemon for macOS/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/app-daemon-darwin ./cmd/daemon
	@echo "   → $(BIN_DIR)/app-daemon-darwin"

# 定義 MacOS App 名稱
APP_NAME=GoIdleGuard
APP_BUNDLE=bin/$(APP_NAME).app
CONTENTS=$(APP_BUNDLE)/Contents
MACOS_DIR=$(CONTENTS)/MacOS
RESOURCES=$(CONTENTS)/Resources
# --- macOS App Packaging ---
macos-app: macos-daemon
	@echo "Packaging $(APP_NAME).app for macOS..."
	
	# 1. 建立目錄結構
	@mkdir -p "$(MACOS_DIR)"
	@mkdir -p "$(RESOURCES)"
	
	# 2. 複製執行檔與設定檔
	@cp "bin/app-daemon-darwin" "$(MACOS_DIR)/$(APP_NAME)"
	@cp "config.yaml" "$(MACOS_DIR)/"
	
	# 3. 建立 Info.plist
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > "$(CONTENTS)/Info.plist"
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$(CONTENTS)/Info.plist"
	@echo '<plist version="1.0">' >> "$(CONTENTS)/Info.plist"
	@echo '<dict>' >> "$(CONTENTS)/Info.plist"
	@echo '    <key>CFBundleExecutable</key>' >> "$(CONTENTS)/Info.plist"
	@echo '    <string>$(APP_NAME)</string>' >> "$(CONTENTS)/Info.plist"
	@echo '    <key>CFBundleIconFile</key>' >> "$(CONTENTS)/Info.plist"
	@echo '    <string>icon.icns</string>' >> "$(CONTENTS)/Info.plist"
	@echo '    <key>CFBundleIdentifier</key>' >> "$(CONTENTS)/Info.plist"
	@echo '    <string>com.hanks.goidleguard</string>' >> "$(CONTENTS)/Info.plist"
	@echo '    <key>LSUIElement</key>' >> "$(CONTENTS)/Info.plist"
	@echo '    <true/>' >> "$(CONTENTS)/Info.plist"
	@echo '</dict>' >> "$(CONTENTS)/Info.plist"
	@echo '</plist>' >> "$(CONTENTS)/Info.plist"
	
	# 4. (選用) 設定權限
	@chmod +x "$(MACOS_DIR)/$(APP_NAME)"
	
	@echo "App Bundle created at: $(APP_BUNDLE)"
	@echo "You can now double-click $(APP_NAME).app to run silently!"
	
# --------------------
# Clean
# --------------------
ifeq ($(OS),Windows_NT)
    RM = if exist bin rd /s /q $(BIN_DIR)
else
    RM = rm -rf $(BIN_DIR)
endif

clean:
	@echo "→ Cleaning build artifacts..."
	@$(RM)
	@echo "   ✔ Removed $(BIN_DIR)/"