# Project configuration
PROJECT_NAME := qbittal
DB_CONTAINER := qdb
DB_USER := kalairen
DB_HOST := localhost
DB_NAME := qdb
IMAGE_NAME := $(PROJECT_NAME):latest

# Default target
.DEFAULT_GOAL := help

# Helpers
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Commands
.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: dev
dev: ## Run the app in development mode using Air
	air

.PHONY: db
db: ## Connect to the local development database
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -h $(DB_HOST) -d $(DB_NAME)

.PHONY: build
build: ## Build the Docker image
	cd client && npm install && npm run build && cd ..
	docker buildx build -t $(IMAGE_NAME) . --load
	docker save $(IMAGE_NAME) > $(PROJECT_NAME).tar

.PHONY: run
run: ## Run the app (non-development mode)
	go run .

.PHONY: lint
lint: ## Run linters (requires golangci-lint)
	golangci-lint run

.PHONY: clean
clean: ## Clean up build artifacts
	go clean -modcache
	rm -rf ./bin ./dist
	docker buildx prune -f

.PHONY: deps
deps: ## Download and tidy dependencies
	go mod tidy

.PHONY: fmt
fmt: ## Format the code
	go fmt ./...

.PHONY: vet
vet: ## Analyze code for potential issues
	go vet ./...

# Add this to your existing Makefile
.PHONY: coverage
coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: ci
ci: test lint vet fmt coverage ## Run all checks (tests, lint, vet, format)

.PHONY: staging
staging: ## Deploy to staging environment
	scp $(PROJECT_NAME).tar docker-compose_local.yml kalairen@zimbian.sh:~/.services/qbittal
