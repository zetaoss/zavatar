## Makefile
GOLANGCI_LINT_VERSION := v2.7.2
RCLONE_BIN := /usr/local/bin/rclone
SHELL := /bin/bash

.PHONY: run-local
run-local:
	@rm -rf ./bucket
	@echo "▶ running (local)"
	LOG_LEVEL=debug API_MODE=fake STORAGE_DRIVER=local go run ./cmd/zavatar

.PHONY: run-remote
run-remote:
	@echo "▶ running with .env.remote"
	@set -a; . ./.env.remote; set +a; \
	LOG_LEVEL=debug API_MODE=remote STORAGE_DRIVER=r2 go run ./cmd/zavatar

.PHONY: curl
curl:
	@echo ""
	@echo "=========================================="
	@echo " zavatar local test URLs (curl)"
	@echo "=========================================="
	@echo ""
	@echo "identicon (uid=1, s=40)"
	@echo "  url: http://localhost:8080/u/1?s=40&t=1"
	@curl -s -o /dev/null -D - "http://localhost:8080/u/1?s=40&t=1" | grep -i -E '^HTTP/|^Location:'
	@echo ""
	@echo "letter (uid=2, s=200)"
	@echo "  url: http://localhost:8080/u/2?s=200&t=2"
	@curl -s -o /dev/null -D - "http://localhost:8080/u/2?s=200&t=2" | grep -i -E '^HTTP/|^Location:'
	@echo ""
	@echo "identicon large (uid=3, s=320)"
	@echo "  url: http://localhost:8080/u/3?s=320&t=1"
	@curl -s -o /dev/null -D - "http://localhost:8080/u/3?s=320&t=1" | grep -i -E '^HTTP/|^Location:'
	@echo ""
	@echo "missing user (uid=9999999999, s=40)"
	@echo "  url: http://localhost:8080/u/9999999999?s=40&t=1"
	@curl -s -o /dev/null -D - "http://localhost:8080/u/9999999999?s=40&t=1" | grep -i -E '^HTTP/|^Location:'

.PHONY: lint
lint: lint-install
	@echo "▶ golangci-lint"
	@./bin/golangci-lint run

.PHONY: lint-install
lint-install:
	@mkdir -p ./bin
	@if [ -x "./bin/golangci-lint" ]; then \
		echo "▶ golangci-lint already installed: $$(./bin/golangci-lint version)"; \
	else \
		echo "▶ installing golangci-lint $(GOLANGCI_LINT_VERSION) to ./bin/golangci-lint"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
			| sh -s -- -b ./bin $(GOLANGCI_LINT_VERSION); \
		echo "▶ installed: $$(./bin/golangci-lint version)"; \
	fi

.PHONY: test
test:
	@echo "▶ go test"
	@go test -v ./...

.PHONY: checks
checks: test lint
