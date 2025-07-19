.PHONY: help test test-unit test-e2e test-coverage clean fmt fmt-check vet lint tidy deps check install-tools security dev

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: test-unit test-e2e ## Run all tests

# Base test execution - generates JSON output
test-unit-json: ## Run unit tests and save JSON output
	@go test -json -run "^Test.*" -skip "TestTaskEngineE2E" ./... > test-unit.json

test-e2e-json: ## Run e2e tests and save JSON output  
	@go test -json -run "TestTaskEngineE2E" > test-e2e.json

# Local development formatters (human-readable)
test-unit: test-unit-json ## Run unit tests with human-readable output
	@cat test-unit.json | gotestfmt

test-e2e: test-e2e-json ## Run e2e tests with human-readable output
	@cat test-e2e.json | gotestfmt

# CI formatters (JUnit XML)
test-unit-ci: test-unit-json ## Run unit tests with JUnit XML output for CI
	@gotestsum --junitfile=junit-unit.xml --format=testname --raw-command -- cat test-unit.json

test-e2e-ci: test-e2e-json ## Run e2e tests with JUnit XML output for CI
	@gotestsum --junitfile=junit-e2e.xml --format=testname --raw-command -- cat test-e2e.json

test-ci: test-unit-ci test-e2e-ci ## Run all tests with JUnit XML output for CI

test-coverage: ## Run tests with coverage
	@go test -v -race -coverprofile=coverage.out ./...

clean: ## Clean build artifacts
	@rm -f coverage.out junit-unit.xml junit-e2e.xml test-unit.json test-e2e.json
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
	@gosec -exclude=G304 ./...
	@echo "Running vulnerability scanning..."
	@govulncheck ./...

install-tools: ## Install development tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	@go install gotest.tools/gotestsum@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest

dev: tidy fmt vet lint security test ## Full development workflow 