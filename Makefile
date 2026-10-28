# Makefile for TYFORMS project

# Variables
BACKEND_BINARY := tyforms
BACKEND_ENTRY_POINT := ./cmd/webserver
FRONTEND_DIR := FrontEnd
WWWROOT_DIR := wwwroot

# Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build both backend and frontend"
	@echo "  run             - Build both and run backend"
	@echo "  build-backend   - Build Go backend binary"
	@echo "  build-frontend  - Copy vanilla frontend to wwwroot/"
	@echo "  run-backend     - Run backend server (with air hot reload if available)"
	@echo "  run-backend-prod - Run backend server directly (production)"
	@echo "  dev-backend     - Run backend with air (requires air installation)"
	@echo "  dev-frontend    - Start frontend dev server on :5173 (proxies /api to :5099)"
	@echo "  clean           - Remove built binaries and frontend build artifacts"
	@echo "  clean-backend   - Remove backend binary"
	@echo "  clean-frontend  - Remove frontend build directory (wwwroot/)"
	@echo "  fmt             - Format Go code"
	@echo "  test            - Run Go tests"
	@echo "  deps-backend    - Download Go dependencies"
	@echo "  deps-frontend   - No frontend dependencies (nothing to install)"

# Build both backend and frontend
.PHONY: build
build: build-backend build-frontend

# Build both and run backend
.PHONY: run
run: build run-backend

# Build Go backend
.PHONY: build-backend
build-backend:
	@echo "Building backend binary..."
	go build -tags sqlite_fts5 -o $(BACKEND_BINARY) $(BACKEND_ENTRY_POINT)
	@echo "Backend binary created: $(BACKEND_BINARY)"

# Copy vanilla frontend to the served directory
.PHONY: build-frontend
build-frontend:
	@echo "Copying frontend..."
	@rm -rf $(WWWROOT_DIR)
	@mkdir -p $(WWWROOT_DIR)
	@cp -r $(FRONTEND_DIR)/index.html $(FRONTEND_DIR)/css $(FRONTEND_DIR)/js $(FRONTEND_DIR)/favicon.webp $(WWWROOT_DIR)/
	@echo "Frontend copied to $(WWWROOT_DIR)/"

# Run backend with air (hot reload) if available, otherwise run directly
.PHONY: run-backend
run-backend:
	@if command -v air >/dev/null 2>&1; then \
		echo "Starting backend with air (hot reload)..."; \
		air; \
	else \
		echo "Air not found, starting backend directly..."; \
		$(MAKE) run-backend-prod; \
	fi

# Run backend directly (production)
.PHONY: run-backend-prod
run-backend-prod:
	@echo "Starting backend server..."
	./$(BACKEND_BINARY)

# Run backend with air (requires air installation)
.PHONY: dev-backend
dev-backend:
	@echo "Starting backend with air (hot reload)..."
	air

# Start frontend development server (run the backend alongside it)
.PHONY: dev-frontend
dev-frontend:
	@echo "Starting frontend dev server on :5173 (proxying /api to :5099)..."
	go run ./cmd/devserver

# Clean everything
.PHONY: clean
clean: clean-backend clean-frontend

# Clean backend
.PHONY: clean-backend
clean-backend:
	@echo "Removing backend binary..."
	@rm -f $(BACKEND_BINARY)
	@echo "Cleaning Go cache..."
	@go clean

# Clean frontend
.PHONY: clean-frontend
clean-frontend:
	@echo "Removing frontend build directory..."
	@rm -rf $(WWWROOT_DIR)

# Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run Go tests
.PHONY: test
test:
	@echo "Running Go tests..."
	go test -tags sqlite_fts5 ./...

# Download Go dependencies
.PHONY: deps-backend
deps-backend:
	@echo "Downloading Go dependencies..."
	go mod download
	go mod tidy

# Frontend has no dependencies
.PHONY: deps-frontend
deps-frontend:
	@echo "Frontend has no dependencies."

# Wordle service targets (optional)
.PHONY: build-wordle
build-wordle:
	@echo "Building wordle service..."
	cd wordle && npm install

.PHONY: run-wordle
run-wordle:
	@echo "Starting wordle service..."
	cd wordle && npm start

.PHONY: dev-wordle
dev-wordle:
	@echo "Starting wordle development server..."
	cd wordle && npm run dev