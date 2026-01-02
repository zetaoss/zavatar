GOLANGCI_LINT_VERSION := v2.7.2
RCLONE_BIN := /usr/local/bin/rclone
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
	go run ./cmd/zavatar
	
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

.PHONY: r2-purge
r2-purge: aws-install
	@echo "▶ loading .env.r2"
	@set -euo pipefail; \
	set -a; . ./.env.r2; set +a; \
	: "$${R2_ACCOUNT_ID:?missing R2_ACCOUNT_ID}"; \
	: "$${R2_BUCKET:?missing R2_BUCKET}"; \
	: "$${R2_ACCESS_KEY_ID:?missing R2_ACCESS_KEY_ID}"; \
	: "$${R2_SECRET_ACCESS_KEY:?missing R2_SECRET_ACCESS_KEY}"; \
	export AWS_ACCESS_KEY_ID="$${R2_ACCESS_KEY_ID}"; \
	export AWS_SECRET_ACCESS_KEY="$${R2_SECRET_ACCESS_KEY}"; \
	export AWS_DEFAULT_REGION=auto; \
	ENDPOINT="https://$${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"; \
	echo "▶ purging R2 bucket: $${R2_BUCKET} (prefix=$${R2_PREFIX:-/})"; \
	if [ -n "$${R2_PREFIX:-}" ]; then \
		aws s3 rm "s3://$${R2_BUCKET}/$${R2_PREFIX}" \
			--recursive \
			--endpoint-url "$$ENDPOINT"; \
	else \
		aws s3 rm "s3://$${R2_BUCKET}" \
			--recursive \
			--endpoint-url "$$ENDPOINT"; \
	fi; \
	echo "✅ done"


.PHONY: aws-install
aws-install:
	@command -v aws >/dev/null 2>&1 || { \
		echo "❌ aws cli not installed"; \
		echo "   install: sudo apt install awscli  (or brew install awscli)"; \
		exit 1; \
	}