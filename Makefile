GO       ?= go
BIN      := bin/fizzbuzz
FUZZTIME ?= 30s
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
REVISION ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
LDFLAGS  := -s -w -X main.version=$(VERSION) -X main.revision=$(REVISION)

.DEFAULT_GOAL := help
.PHONY: help build run test cover fuzz lint fmt tidy docker clean

help: ## Show available targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*## "}{printf "  %-8s %s\n",$$1,$$2}'

build: ## Build the server binary into bin/
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/server

run: ## Run the server locally
	$(GO) run ./cmd/server

test: ## Run all tests with the race detector
	$(GO) test -race -count=1 ./...

cover: ## Run tests with a coverage summary
	$(GO) test -race -covermode=atomic -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

fuzz: ## Fuzz Generate for FUZZTIME (default 30s)
	$(GO) test ./internal/fizzbuzz -run='^$$' -fuzz=FuzzGenerate -fuzztime=$(FUZZTIME)

lint: ## go vet + golangci-lint
	$(GO) vet ./...
	golangci-lint run ./...

fmt: ## Format the code
	golangci-lint fmt ./...

tidy: ## go mod tidy + verify
	$(GO) mod tidy && $(GO) mod verify

docker: ## Build the Docker image
	docker build --build-arg VERSION=$(VERSION) --build-arg REVISION=$(REVISION) --build-arg CREATED=$$(date -u +%Y-%m-%dT%H:%M:%SZ) -t go-fizzbuzz:$(VERSION) .

clean: ## Remove build artefacts
	rm -rf bin coverage.out
