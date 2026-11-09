# Contributing

Copyright (c) 2026 Nardeo.

Use [private vulnerability reporting](https://github.com/Nardeo-org/nardeo-backup/security/advisories/new) for security issues. Follow [SECURITY.md](SECURITY.md) before sharing sensitive information.

Start each task with [AGENTS.md](AGENTS.md) and the linked project context. When using an assistant or editor that does not load that file automatically, attach it explicitly. Keep tool-specific instructions as references to the shared file so contributors use the same project rules.

## Development

Use a current stable Go release for development and security scanning. Keep application code compatible with Go 1.22. Tests use a controlled `pg_dump` executable and do not require a live database.

```sh
go mod download
gofmt -w cmd internal
go test -race ./...
go vet ./...
go build ./cmd/nardeo-backup
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

If you change PostgreSQL subprocess behavior, verify backup and restore against a disposable PostgreSQL database. Use synthetic data and keep credentials and dumps out of commits.

## Pull requests

Create a short-lived branch or fork and open a pull request against `main`. Explain the problem, resulting behavior and validation. Add regression tests for changes to observable behavior, and update usage documentation when necessary.

Required checks must pass on an up-to-date branch. Resolve review conversations and obtain the required maintainer approval before merging. New commits can invalidate earlier approvals. See [Repository policy](.github/REPOSITORY_POLICY.md) for the complete merge requirements.

Use [CODE_REVIEW.md](CODE_REVIEW.md) for the project review checklist. Keep [architecture](docs/architecture.md), [compatibility](BACKWARD_COMPATIBILITY.md) and [decision records](docs/decisions/) current when they are affected. Substantial requirements and cross-session task notes can use the templates in [.ai/](.ai/README.md).

All source code, comments, logs, CLI text and documentation must be written in English. Keep copyright notices under Nardeo. Contributions are accepted under the repository's [MIT License](LICENSE).

## Workflow and dependency changes

Pin every third-party GitHub Action to its full commit SHA and add a version comment. Keep workflow permissions narrow, use GitHub-hosted runners, and avoid executing untrusted pull request code in privileged events such as `pull_request_target`.

Dependabot opens updates for Go modules and GitHub Actions. Dependency review blocks newly introduced dependencies with known vulnerabilities rated moderate or higher. `govulncheck` checks reachable Go vulnerabilities, and CodeQL analyzes Go code and Actions workflows. Review dependency and workflow updates before merging them.
