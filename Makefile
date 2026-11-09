# Copyright (c) 2026 Nardeo.

.DEFAULT_GOAL := help

GO ?= go
BINARY ?= bin/nardeo-backup
GOVULNCHECK_VERSION := v1.9.0
export GO

.PHONY: help build test check fmt fmt-check vet test-integration security

help:
	@printf '%s\n' \
		'make build             Build bin/nardeo-backup' \
		'make test              Run unit and subprocess tests with race detection' \
		'make check             Verify formatting, tests, vet, build and integration (requires Docker)' \
		'make fmt               Format Go files' \
		'make fmt-check         Verify Go formatting without changing files' \
		'make vet               Vet all packages, including integration tests' \
		'make test-integration  Run the isolated PostgreSQL backup and restore test' \
		'make security          Check reachable Go vulnerabilities (use current stable Go)'

build:
	@mkdir -p "$$(dirname -- "$(BINARY)")"
	"$(GO)" build -trimpath -o "$(BINARY)" ./cmd/nardeo-backup

test:
	"$(GO)" test -race ./...

fmt:
	gofmt -w .

fmt-check:
	@set -eu; files="$$(gofmt -l .)"; \
		if [ -n "$$files" ]; then \
			printf '%s\n' 'Go files need formatting. Run make fmt:' "$$files"; \
			exit 1; \
		fi

vet:
	"$(GO)" vet -tags=integration ./...

test-integration:
	bash tests/integration/run.sh

security:
	"$(GO)" run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

check:
	$(MAKE) --no-print-directory fmt-check
	$(MAKE) --no-print-directory test
	$(MAKE) --no-print-directory vet
	$(MAKE) --no-print-directory build
	$(MAKE) --no-print-directory test-integration
