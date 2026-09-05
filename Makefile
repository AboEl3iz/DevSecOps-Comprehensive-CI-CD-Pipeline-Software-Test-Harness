# DevSecOps Comprehensive Makefile

.PHONY: all dev-test dev-integration dev-lint sec-scan container-build ci-verify clean help

all: ci-verify

help:
	@echo "Available DevSecOps Makefile targets:"
	@echo "  make dev-test          - Run Go unit tests with code coverage"
	@echo "  make dev-integration   - Run Go integration test suite"
	@echo "  make dev-lint          - Run golangci-lint on app codebase"
	@echo "  make sec-scan          - Run Gitleaks secret scanner and Gosec SAST"
	@echo "  make container-build   - Build container image locally"
	@echo "  make ci-verify         - Run complete local DevSecOps pipeline verification"
	@echo "  make clean             - Clean temporary artifacts and coverage reports"

dev-test:
	@echo "==> Running unit tests..."
	cd app && go test -v -race -coverprofile=coverage.out ./...

dev-integration:
	@echo "==> Running integration tests..."
	cd app && go test -v -run TestIntegration_FullFlow ./...

dev-lint:
	@echo "==> Running golangci-lint..."
	cd app && golangci-lint run

sec-scan:
	@echo "==> Running secret & SAST scanning..."
	@bash scripts/test-action-pinning.sh
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks detect --config config/.gitleaks.toml; fi
	@if command -v gosec >/dev/null 2>&1; then gosec ./app/...; fi

container-build:
	@echo "==> Building container image..."
	docker build -t security-ci-app:local ./app

ci-verify:
	@bash scripts/run-local-devsecops.sh

clean:
	@rm -f app/coverage.out /tmp/grype-test.json /tmp/sig.json
	@echo "Clean complete."
