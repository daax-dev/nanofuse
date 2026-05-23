.PHONY: help build test clean run-local build-image push-image deploy

# Configuration
IMAGE_NAME ?= ghcr.io/daax-dev/nanofuse/todo-app
IMAGE_TAG ?= latest
BACKEND_BIN ?= backend/bin/todo-server

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build backend binary locally
	@echo "Building backend..."
	cd backend && go build -o bin/todo-server ./cmd/server
	@echo "Backend built successfully: $(BACKEND_BIN)"

test: ## Run backend tests
	@echo "Running tests..."
	cd backend && go test -v -race -cover ./...

test-integration: build ## Run integration tests
	@echo "Running integration tests..."
	cd tests/integration && go test -v ./...

run-local: build ## Run server locally for development
	@echo "Starting local server..."
	mkdir -p /tmp/todo-dev
	$(BACKEND_BIN) -db-path /tmp/todo-dev/todos.db -debug

build-image: ## Build container image
	@echo "Building container image: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker build -f docker/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Image built successfully"

push-image: build-image ## Push container image to registry
	@echo "Pushing image to registry..."
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Image pushed successfully"

deploy: push-image ## Deploy to NanoFuse VM
	@echo "Deploying to NanoFuse..."
	# TODO: Add nanofuse deployment commands
	@echo "Deployment command not yet implemented"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf backend/bin
	rm -f /tmp/todo-dev/todos.db*
	@echo "Clean complete"

# Development helpers
dev-db-inspect: ## Inspect development database
	@echo "Inspecting development database..."
	@if [ -f /tmp/todo-dev/todos.db ]; then \
		sqlite3 /tmp/todo-dev/todos.db ".tables"; \
	else \
		echo "No development database found"; \
	fi

dev-logs: ## Show development server logs
	@tail -f /tmp/todo-server.log 2>/dev/null || echo "No log file found"
