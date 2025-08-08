.PHONY: help build test clean install run run-file fmt lint docs build-linux build-darwin build-windows build-all

# Default target
.DEFAULT_GOAL := help

help:           ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build:          ## Build the application
	go build -o nomad-event-logger .

test:           ## Run tests
	go test ./...

test-coverage:   ## Run tests with coverage
	go test -cover ./...

clean:          ## Clean build artifacts
	rm -f nomad-event-logger

install:        ## Install the application
	go install .

run:            ## Run the application with default settings
	./nomad-event-logger start

EVENT_TYPES ?= allocation
run-file:       ## Run the application with file sink
	./nomad-event-logger start --sinks stdout,file --file-path /tmp/nomad-events.json --event-types $(EVENT_TYPES)

fmt:            ## Format code
	go fmt ./...

lint:           ## Lint code
	golangci-lint run

docs:           ## Generate documentation
	godoc -http=:6060

build-linux:    ## Build for Linux (amd64)
	GOOS=linux GOARCH=amd64 go build -o nomad-event-logger-linux-amd64 .

build-darwin:   ## Build for macOS (amd64)
	GOOS=darwin GOARCH=amd64 go build -o nomad-event-logger-darwin-amd64 .

build-windows:  ## Build for Windows (amd64)
	GOOS=windows GOARCH=amd64 go build -o nomad-event-logger-windows-amd64.exe .

build-all: build-linux build-darwin build-windows ## Build for all platforms
