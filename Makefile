.PHONY: all fmt lint build test cov clean deps dev demo-all
.PHONY: fe-init fe-deps fe-typecheck fe-lint fe-test fe-start fe-build fe-web fe-ios fe-android fe-clean

# Go parameters - configure SDK path
# Override with: make GO_SDK=/path/to/go/bin build
GO_SDK ?= $(USERPROFILE)/sdk/go1.25.6/bin
GOCMD = $(GO_SDK)/go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOFMT = $(GO_SDK)/gofmt
GOVET = $(GOCMD) vet
BINARY_NAME = globe-expedition-journal
COVERAGE_THRESHOLD = 70

# Directories
CMD_DIR = ./cmd/server
PKG_DIR = ./...
FRONTEND_DIR = ./frontend

all: fmt lint test build

## deps: Download and tidy dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

## fmt: Format code with gofmt
fmt:
	$(GOFMT) -s -w .

## lint: Run golangci-lint or go vet
lint:
	$(GOVET) $(PKG_DIR)

## build: Build the binary
build:
	$(GOBUILD) -o bin/$(BINARY_NAME).exe $(CMD_DIR)

## test: Run unit tests
test:
	$(GOTEST) -v $(PKG_DIR)

## cov: Run tests with coverage
cov:
	$(GOTEST) -coverprofile=coverage.out $(PKG_DIR)
	$(GOCMD) tool cover -func=coverage.out

## cov-html: Generate HTML coverage report
cov-html: cov
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## run: Run the server
run: build
	./bin/$(BINARY_NAME).exe

## demo: Build and run for demo (seeds data, starts server)
demo: build
	@echo "Starting Globe Expedition Journal demo..."
	@echo "Backend: http://localhost:8080"
	@echo "Health check: http://localhost:8080/api/v1/health"
	@echo "Countries API: http://localhost:8080/api/v1/countries"
	@echo ""
	./bin/$(BINARY_NAME).exe

## demo-all: Start backend and frontend together for demo
demo-all: build
	@echo "Starting Globe Expedition Journal (full stack demo)..."
	@echo "Backend:  http://localhost:8080"
	@echo "Frontend: http://localhost:8081"
	@echo ""
	@echo "Press Ctrl+C to stop both servers"
	@echo ""
	./bin/$(BINARY_NAME).exe & cd $(FRONTEND_DIR) && npx expo start --web

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Backend Targets:"
	@echo "  all       - fmt, lint, test, build (default)"
	@echo "  deps      - Download dependencies"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  build     - Build binary"
	@echo "  test      - Run tests"
	@echo "  cov       - Run tests with coverage"
	@echo "  clean     - Remove build artifacts"
	@echo "  run       - Build and run server"
	@echo "  demo      - Run backend in demo mode"
	@echo "  demo-all  - Run backend + frontend together"
	@echo ""
	@echo "Frontend Targets:"
	@echo "  fe-init   - Initialize Expo project"
	@echo "  fe-deps   - Install frontend dependencies"
	@echo "  fe-typecheck - TypeScript type checking"
	@echo "  fe-lint   - Lint frontend code"
	@echo "  fe-test   - Run frontend tests"
	@echo "  fe-start  - Start Expo dev server"
	@echo "  fe-web    - Start web version"
	@echo "  fe-ios    - Start iOS simulator"
	@echo "  fe-android- Start Android emulator"
	@echo "  fe-build  - Build production bundle"
	@echo "  fe-clean  - Clean frontend build artifacts"

# ============================================
# Frontend (React Native + Expo) targets
# ============================================

## fe-init: Initialize new Expo project with TypeScript
fe-init:
	@if [ ! -d "$(FRONTEND_DIR)" ]; then \
		npx create-expo-app@latest $(FRONTEND_DIR) --template expo-template-blank-typescript; \
		cd $(FRONTEND_DIR) && npx expo install expo-router react-native-screens react-native-safe-area-context; \
	else \
		echo "Frontend directory already exists"; \
	fi

## fe-deps: Install frontend dependencies
fe-deps:
	cd $(FRONTEND_DIR) && npm install --legacy-peer-deps

## fe-typecheck: Run TypeScript type checking
fe-typecheck:
	cd $(FRONTEND_DIR) && npm run typecheck

## fe-lint: Lint frontend code
fe-lint:
	cd $(FRONTEND_DIR) && npm run lint 2>/dev/null || npx eslint . --ext .ts,.tsx

## fe-test: Run frontend tests
fe-test:
	cd $(FRONTEND_DIR) && npm test -- --watchAll=false

## fe-start: Start Expo development server
fe-start:
	cd $(FRONTEND_DIR) && npx expo start

## fe-web: Start web version
fe-web:
	cd $(FRONTEND_DIR) && npx expo start --web

## fe-ios: Start iOS simulator
fe-ios:
	cd $(FRONTEND_DIR) && npx expo start --ios

## fe-android: Start Android emulator
fe-android:
	cd $(FRONTEND_DIR) && npx expo start --android

## fe-build: Build production bundle
fe-build:
	cd $(FRONTEND_DIR) && npx expo export

## fe-clean: Clean frontend build artifacts
fe-clean:
	cd $(FRONTEND_DIR) && rm -rf node_modules .expo dist
