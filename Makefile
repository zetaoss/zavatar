GOLANGCI_LINT_VERSION := v2.7.2
SHELL := /bin/bash

.PHONY: all help
all: help

help:
	@echo ""
	@echo "zavatar Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make run-sim    - Run local simulator (SIM_MODE=1)"
	@echo "  make test-sim   - Smoke test via curl (requires run-sim running)"
	@echo "  make lint       - Run golangci-lint (auto install if missing)"
	@echo ""

.PHONY: run-dev
run-dev:
	@echo "▶ loading .env.example"
	@set -a; . ./.env.example; set +a; \
	echo "▶ running (dev)..."; \
	go run ./cmd/zavatar -dev

.PHONY: run-r2
run-r2:
	@echo "▶ loading .env.r2"
	@set -a; . ./.env.r2; set +a; \
	echo "▶ running (r2)..."; \
	go run ./cmd/zavatar

.PHONY: curl
curl:
	@rm -rf ./data

	@echo ""
	@echo "▶ letter avatar (uid=1)"
	@curl -s -D- "http://localhost:8080/u/1?s=40&v=4" -o /dev/null

	@echo ""
	@echo "▶ identicon avatar (uid=2)"
	@curl -s -D- "http://localhost:8080/u/2?s=200&v=4" -o /dev/null

	@echo ""
	@echo "▶ gravatar redirect (uid=3)"
	@curl -s -D- "http://localhost:8080/u/3?s=40&v=4" -o /dev/null

.PHONY: lint lint-install
lint: lint-install
	@echo "▶ golangci-lint"
	@./bin/golangci-lint run

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
