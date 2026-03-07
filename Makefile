.PHONY: build install build-all build-linux build-windows build-mac clean tidy

BINARY    := aiswitch
BUILD_DIR := ./bin
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION)"

# ── local build (current OS/arch) ────────────────────────────────────────────

build:
	@mkdir -p $(BUILD_DIR)
	GOTOOLCHAIN=auto go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo "✓ Built $(BUILD_DIR)/$(BINARY)"

install: build
	@cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "✓ Installed /usr/local/bin/$(BINARY)"
	@echo ""
	@echo "Add shell integration to your ~/.zshrc or ~/.bashrc:"
	@echo '    eval "$$(aiswitch shell-init)"'
	@echo ""
	@echo "PowerShell — add to \$$PROFILE:"
	@echo '    Invoke-Expression (aiswitch shell-init --shell powershell | Out-String)'

# ── cross-platform release builds ────────────────────────────────────────────

build-all: build-mac build-linux build-windows
	@echo "✓ All platform binaries in $(BUILD_DIR)/"

build-mac:
	@mkdir -p $(BUILD_DIR)
	GOTOOLCHAIN=auto GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64  .
	GOTOOLCHAIN=auto GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64  .
	@echo "✓ macOS (intel + apple silicon)"

build-linux:
	@mkdir -p $(BUILD_DIR)
	GOTOOLCHAIN=auto GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64   .
	GOTOOLCHAIN=auto GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64   .
	@echo "✓ Linux (amd64 + arm64)"

build-windows:
	@mkdir -p $(BUILD_DIR)
	GOTOOLCHAIN=auto GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe .
	@echo "✓ Windows (amd64)"

# ── utilities ─────────────────────────────────────────────────────────────────

tidy:
	GOTOOLCHAIN=auto go mod tidy

clean:
	rm -rf $(BUILD_DIR)

run:
	GOTOOLCHAIN=auto go run . $(ARGS)
