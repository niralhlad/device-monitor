VERSION=0.2.0
APP_NAME=device-monitor
CMD_PATH=./cmd/server
SIMULATOR ?= ./device-simulator-mac-arm64
BINARY_NAME=bin/$(APP_NAME)v$(VERSION)
RESULTS_FILE=results.txt

.PHONY: help run test test-no-cache fmt vet build clean simulator

help:
	@echo "Available commands:"
	@echo "  make run            Start the API server"
	@echo "  make test           Run all tests"
	@echo "  make test-no-cache  Run all tests without cache"
	@echo "  make fmt            Format Go code"
	@echo "  make vet            Run go vet"
	@echo "  make simulator      Run the device simulator"
	@echo "  make clean          Remove build artifacts"
	@echo "  make build          Build the server binary"

run:
	go run $(CMD_PATH)

test:
	go test ./...

test-no-cache:
	go test -count=1 ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -o $(BINARY_NAME) $(CMD_PATH)

clean:
	rm -rf bin
	rm -f $(RESULTS_FILE)

simulator:
	./$(SIMULATOR)