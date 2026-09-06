BINARY  := settings
CMD     := ./cmd/settings
BIN_DIR := bin

GO      ?= go

# Version and metadata: derived from git, with fallbacks when there is no repo
# or no tags.
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

# Reproducible build: no system paths, no cgo.
BUILD_ENV   := CGO_ENABLED=0
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# `go install` destination: GOBIN if set, otherwise GOPATH/bin.
GOBIN_DIR := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN_DIR),)
GOBIN_DIR := $(shell $(GO) env GOPATH)/bin
endif

.DEFAULT_GOAL := help

## dev: run the TUI directly, leaving no binary on disk
.PHONY: dev
dev:
	$(GO) run $(CMD)

## build: compile the binary into bin/settings
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(BUILD_ENV) $(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)
	@echo "-> $(BIN_DIR)/$(BINARY) ($(VERSION))"

## install: install settings on the PATH (GOBIN, or GOPATH/bin)
.PHONY: install
install:
	$(BUILD_ENV) $(GO) install $(BUILD_FLAGS) $(CMD)
	@echo "-> $(GOBIN_DIR)/$(BINARY) ($(VERSION))"

## theme: write an example theme to the user's config directory
.PHONY: theme
theme:
	$(GO) run $(CMD) -write-theme

## uninstall: remove settings from the PATH (config in ~/.config/settings-cli/ is left alone)
.PHONY: uninstall
uninstall:
	@if [ -f "$(GOBIN_DIR)/$(BINARY)" ]; then \
		rm -f "$(GOBIN_DIR)/$(BINARY)"; \
		echo "removed $(GOBIN_DIR)/$(BINARY)"; \
	else \
		echo "nothing installed at $(GOBIN_DIR)"; \
	fi
	@echo "to drop the config:  rm -rf ~/.config/settings-cli"

## test: run the tests with the race detector
.PHONY: test
test:
	$(GO) test -race ./...

## vet: standard static analysis
.PHONY: vet
vet:
	$(GO) vet ./...

## lint: golangci-lint (must be installed)
.PHONY: lint
lint:
	golangci-lint run

## tidy: sync go.mod / go.sum
.PHONY: tidy
tidy:
	$(GO) mod tidy

## check: what CI validates -- formatting, vet and tests
.PHONY: check
check: vet test
	@test -z "$$(gofmt -l .)" || { echo "gofmt pending in:"; gofmt -l .; exit 1; }
	@echo "check ok"

## snapshot: local release build with GoReleaser, without publishing
.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --clean

## clean: remove build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) dist

## help: show this help
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | \
	    awk -F': ' '{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
