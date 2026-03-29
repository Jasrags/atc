BINARY    := atc
MODULE    := github.com/Jasrags/atc
GO        := go
GOPATH    := $(shell $(GO) env GOPATH)
GOFLAGS   ?=
LDFLAGS   ?= -s -w

.PHONY: all build run run-tower run-combined dev watch clean test test-race cover cover-html lint vet fmt tidy help

all: fmt vet test build ## Default: format, vet, test, build

build: ## Build the binary
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY) .

run: build ## Build and run (TRACON mode)
	./$(BINARY)

run-tower: build ## Build and run in Tower mode
	./$(BINARY) -role tower

run-combined: build ## Build and run in Combined mode
	./$(BINARY) -role combined

dev: build ## Build and run with developer mode (/ commands)
	./$(BINARY) -dev

watch: ## Rebuild on .go file changes (install entr: brew install entr)
	@command -v entr >/dev/null 2>&1 || { echo "Install entr: brew install entr"; exit 1; }
	@echo "Watching for changes... Run ./$(BINARY) in another terminal."
	@while true; do find . -name '*.go' | entr -dn $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY) .; done

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf tmp
	rm -f coverage.out coverage.html

test: ## Run tests
	$(GO) test ./...

test-race: ## Run tests with race detector
	$(GO) test -race ./...

cover: ## Run tests with coverage summary
	$(GO) test -cover ./...

cover-html: ## Generate HTML coverage report
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

lint: ## Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format all Go files
	gofmt -w .
	goimports -w . 2>/dev/null || true

tidy: ## Tidy and verify module dependencies
	$(GO) mod tidy
	$(GO) mod verify

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
