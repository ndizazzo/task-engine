.PHONY: help test test-unit test-unit-ci test-integration test-integration-ci test-coverage clean fmt fmt-check vet lint tidy deps check install-tools security dev test-coverage-ci

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: test-unit ## Run all tests

# Unit tests with live progress and JSON capture
test-unit: ## Run unit tests with live progress
	@go test -json -run "^Test.*" ./... | tee test-unit.json | gotestfmt

# Unit tests with JUnit XML output and live progress for CI
test-unit-ci: ## Run unit tests with JUnit XML output and live progress for CI
	@go test -json -run "^Test.*" ./... | tee test-unit.json | gotestfmt
	@gotestsum --junitfile=junit-unit.xml --format=testname --raw-command -- cat test-unit.json

# Integration tests with live progress and JSON capture
test-integration: ## Run integration tests with live progress
	@go test -json -race ./... | tee test-integration.json | gotestfmt

# Integration tests with JUnit XML output and live progress for CI
test-integration-ci: ## Run integration tests with JUnit XML output and live progress for CI
	@go test -json -race ./... | tee test-integration.json | gotestfmt
	@gotestsum --junitfile=junit-integration.xml --format=testname --raw-command -- cat test-integration.json

test-ci: test-unit-ci ## Run all tests with JUnit XML output for CI

test-coverage: ## Run tests with coverage
	@go test -v -race -coverprofile=coverage.out ./...

# Coverage tests with live progress and JSON capture
test-coverage-ci: ## Run tests with coverage and live progress for CI
	@go test -json -race -coverprofile=coverage.out ./... | tee test-coverage.json | gotestfmt

clean: ## Clean build artifacts
	@rm -f *coverage.out junit-unit.xml junit-integration.xml test-unit.json test-integration.json test-e2e.json test-coverage.log test-coverage.json security-scan.log vulnerability-scan.log
	@rm -f coverage.out coverage.html coverage.txt
	@rm -f test-*.json test-*.log test-*.xml
	@rm -f *-scan.log *-coverage.log *-test.log
	@rm -f junit-*.xml
	@rm -f coverage-*.out security-*.log
	@rm -f *.log *.json *.xml *.out
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
	@gosec -exclude=G304,G115 ./... | tee security-scan.log
	@echo "Running vulnerability scanning..."
	@govulncheck ./... | tee vulnerability-scan.log

install-tools: ## Install development tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	@go install gotest.tools/gotestsum@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest

ci: tidy fmt-check vet lint security test-ci ## Run all CI checks and tests
dev: tidy fmt vet lint security test ## Full development workflow 