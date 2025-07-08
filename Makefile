.PHONY: help test test-unit test-e2e test-coverage clean fmt fmt-check vet lint tidy deps check install-tools security dev

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: test-unit test-e2e ## Run all tests

test-unit: ## Run unit tests
	@go test -json ./... | gotestfmt

test-e2e: ## Run end-to-end tests
	@go test -json -run "TestTaskEngineE2E" | gotestfmt

test-coverage: ## Run tests with coverage
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@rm -f coverage.out coverage.html
	@go clean

fmt: ## Format code
	@go fmt ./...

fmt-check: ## Check if code is formatted
	@test -z "$(shell gofmt -s -l . | tee /dev/stderr)" || (echo "Code is not formatted. Run 'make fmt' to fix." && exit 1)

vet: ## Run go vet
	@go vet ./...

lint: ## Run golangci-lint
	@golangci-lint run

tidy: ## Tidy dependencies
	@go mod tidy

deps: ## Download dependencies
	@go mod download

check: fmt vet ## Run code quality checks

security: ## Run security and vulnerability checks
	@echo "Running static security analysis..."
	@gosec ./...
	@echo "Running vulnerability scanning..."
	@govulncheck ./...

install-tools: ## Install development tools
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest

dev: tidy fmt vet lint security test ## Full development workflow 