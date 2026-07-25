.PHONY: all fmt fmt-check lint arch-check test coverage security terraform-check quality-gate setup-hooks

LINT_BIN ?= $(shell which golangci-lint 2>/dev/null || echo "$(HOME)/go/bin/golangci-lint")

all: quality-gate

## Format Go and Terraform files
fmt:
	@echo "🎨 Formatting Go code..."
	@gofmt -s -w .
	@echo "🎨 Formatting Terraform configuration..."
	@terraform fmt terraform

## Check formatting without making changes
fmt-check:
	@echo "🔍 Checking Go formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Unformatted Go files found:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "🔍 Checking Terraform formatting..."
	@terraform fmt -check terraform

## Run golangci-lint static analysis
lint:
	@echo "🧹 Running golangci-lint static analysis..."
	@$(LINT_BIN) run

## Verify Hexagonal Architecture compliance
arch-check:
	@echo "🏗️  Enforcing Hexagonal Architecture boundaries..."
	@go test -v architecture_test.go

## Run all unit tests
test:
	@echo "🧪 Running unit test suite..."
	@go test -v ./...

## Verify code coverage thresholds
coverage:
	@chmod +x scripts/check_coverage.sh
	@./scripts/check_coverage.sh 35.0 65.0

## Run security analysis
security:
	@echo "🔒 Running security checks..."
	@go vet ./...
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "ℹ️  govulncheck not installed locally; security linting handled by golangci-lint (gosec)"; \
	fi

## Validate Terraform infrastructure specs
terraform-check:
	@echo "📐 Validating Terraform specifications..."
	@cd terraform && terraform init -backend=false >/dev/null && terraform validate

## Run complete quality gate pipeline locally
quality-gate: fmt-check arch-check lint test coverage terraform-check
	@echo ""
	@echo "✨ All Quality Gates Passed Successfully! ✨"

## Install Git Pre-Commit Hook
setup-hooks:
	@chmod +x scripts/setup_hooks.sh
	@./scripts/setup_hooks.sh
