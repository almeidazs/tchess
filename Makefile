BINARY  := tchess
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
GOFLAGS := -trimpath

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary into ./bin
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/tchess

.PHONY: run
run: build ## Build and start a game
	./bin/$(BINARY)

.PHONY: install
install: ## Install into $GOBIN
	go install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./cmd/tchess

.PHONY: test
test: ## Run every test with the race detector
	go test -race ./...

.PHONY: test-short
test-short: ## Run tests with shallower perft depths
	go test -short ./...

.PHONY: bench
bench: ## Benchmark the move generator
	go test -run '^$$' -bench . -benchmem ./chess/ ./tui/

.PHONY: perft
perft: ## Count perft(6) from the initial position (119,060,324 nodes)
	go run ./cmd/tchess perft 6

.PHONY: cover
cover: ## Test coverage summary
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format the code
	golangci-lint fmt ./...

.PHONY: assets
assets: ## Regenerate the README screenshot and demo GIF
	cd tools/shot && go run . -out ../../assets

.PHONY: snapshot
snapshot: ## Build release archives locally with GoReleaser
	goreleaser release --snapshot --clean

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin dist coverage.out
