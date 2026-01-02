GOLANGCI_LINT_VERSION := v2.7.2
RCLONE_BIN := /usr/local/bin/rclone
SHELL := /bin/bash

.PHONY: run-dev
run-dev:
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
