GOLANGCI_LINT_VERSION := v2.7.2
RCLONE_BIN := /usr/local/bin/rclone
SHELL := /bin/bash

.PHONY: run-dev
run-dev:
	@rm -rf ./data
	@echo "▶ running"
	go run ./cmd/zavatar

.PHONY: run-dev-r2
run-dev-r2:
	@echo "▶ running with .env.r2"
	@set -a; . ./.env.r2; set +a; \
	go run ./cmd/zavatar

.PHONY: run-dev-mysql
run-dev-mysql:
	@echo "▶ running with .env.mysql"
	@set -a; . ./.env.mysql; set +a; \
	go run ./cmd/zavatar

.PHONY: url
url:
	@echo ""
	@echo "=========================================="
	@echo " zavatar local test URLs"
	@echo " (open in browser)"
	@echo "=========================================="
	@echo ""
	@echo "letter (uid=1, s=40)               http://localhost:8080/u/1?s=40"
	@echo "identicon (uid=2, s=200->norm)     http://localhost:8080/u/2?s=200"
	@echo "gravatar (uid=3, s=40)             http://localhost:8080/u/3?s=40"
	@echo "letter large (uid=1, s=320)        http://localhost:8080/u/1?s=320"
	@echo ""
	@echo "preview identicon (uid=1, t=2)     http://localhost:8080/u/1?s=40&t=2"
	@echo "preview gravatar (uid=3, t=3)      http://localhost:8080/u/3?s=80&t=3"
	@echo "preview gravatar->letter fallback  http://localhost:8080/u/1?s=80&t=3"

PURGE_KEY = secret

.PHONY: purge
purge:
	@echo ""
	@echo "=========================================="
	@echo " zavatar purge test"
	@echo "=========================================="
	@echo ""
	@echo "purge official avatars for uid=1"
	@echo ""
	@curl -X POST -H "X-Purge-Key: $(PURGE_KEY)" http://localhost:8080/internal/purge/u/1
	@echo ""

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

.PHONY: test-integration
test-integration:
	@echo "▶ go test (integration)"
	@go test -v -tags=integration ./...

.PHONY: checks
checks: test lint
	@echo "▶ all checks passed"
