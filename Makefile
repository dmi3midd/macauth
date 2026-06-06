# Simple Makefile for a Go project

# Setup the application
setup:
	@chmod +x setup.sh
	@./setup.sh

# Build the application
build:
	@echo "Building..."
	@go build -o main ./cmd/api

# Run the application
run:
	@go run ./cmd/api

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
# Locate air binary path
AIR_BIN := $(shell which air 2>/dev/null || echo "$(shell go env GOPATH)/bin/air")

watch:
	@if [ -x "$(AIR_BIN)" ]; then \
		echo "Starting air for live reload..."; \
		$(AIR_BIN); \
	else \
		echo "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] "; \
		read choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			echo "Installing air..."; \
			go install github.com/air-verse/air@latest; \
			echo "Starting air..."; \
			$(AIR_BIN); \
		else \
			echo "You chose not to install air. Exiting..."; \
			exit 1; \
		fi; \
	fi

.PHONY: setup build run clean watch docker-up docker-down

# Docker — app only (uses external postgres, set host in config.yaml)
docker-run:
	@docker compose up --build -d

# Docker — app + postgres (set host: postgres in config.yaml)
docker-run-all:
	@docker compose --profile db up --build -d

docker-down:
	@docker compose --profile db down

docker-logs:
	@docker compose logs -f app
