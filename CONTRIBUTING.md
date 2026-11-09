# Contributing

Copyright (c) 2026 Nardeo.

Use [private vulnerability reporting](https://github.com/Nardeo-org/nardeo-backup/security/advisories/new) for security issues. Follow [SECURITY.md](SECURITY.md) before sharing sensitive information.

Start each task with [AGENTS.md](AGENTS.md) and the linked project context. When using an assistant or editor that does not load that file automatically, attach it explicitly. Keep tool-specific instructions as references to the shared file so contributors use the same project rules.

## Development

Use GNU Make 3.81 or newer and a current stable Go release for development and security scanning. Keep application code compatible with Go 1.22. The default unit tests use a controlled `pg_dump` executable and do not require a live database.

```sh
make fmt
make test
make build
make check
make security
```

The [Makefile](Makefile) is the shared command definition for contributors, agents and CI. Run `make help` for the available targets. `make build` writes `bin/nardeo-backup`; use `BINARY=bin/backup` to choose another path or `GO=/path/to/go` to select a Go executable.

Build metadata uses `VERSION=dev` and the current Git commit by default. Release tooling can call `make build VERSION=v0.1.0 COMMIT=abc123def456`. These values populate the CLI's `--version` output through `main.version` and `main.commit` linker variables. Source archives without `.git` use commit `unknown`; builds made directly with Go retain `dev`/`unknown` unless those linker variables are supplied. Additional linker options can be passed with `LDFLAGS`.
`make check` runs formatting verification, race-enabled unit tests, vet (including integration packages), build and the PostgreSQL round trip, in order, stopping at the first failure. It requires Bash and a running Docker daemon. Use `make test` for the unit suite without Docker, and `make fmt-check` for a read-only formatting check. `make security` mirrors the separate vulnerability-scanning workflow and requires a current stable Go toolchain. CodeQL and dependency review remain separate GitHub checks.

The [.editorconfig](.editorconfig) defines UTF-8, LF line endings, final newlines, tabs for Go/Make files and two-space indentation for other text files. Markdown retains intentional trailing whitespace. Go source formatting is enforced by `gofmt` through `make fmt-check`.

Release tooling is pinned in [.tool-versions](.tool-versions). With that GoReleaser version installed, `make release-check` validates the configuration and `make release-snapshot` builds all platform archives without publishing. The required stable-Go CI job runs the same preview. Follow [the release guide](docs/releases.md) for the `v0.1.0` preparation and tag workflow; release notes use the repository's Git history and all release artifacts retain Nardeo licensing.

If you change PostgreSQL subprocess behavior, verify backup and restore against a disposable PostgreSQL database. Use synthetic data and keep credentials and dumps out of commits.

### PostgreSQL integration test

With Go, GNU Make, Bash and a running Docker daemon using Linux amd64 or arm64 containers:

```sh
make test-integration
```

The runner uses `make build` for the CLI and compiles the integration test binary, starts a disposable PostgreSQL 18 container without external networking or published ports, and runs the tests inside it with matching `pg_dump` and `psql` clients. It uses generated temporary credentials and removes the container, its database storage and build artifacts on exit.

The test creates synthetic rows, invokes the CLI using both `PG*` fields and `DATABASE_URL`, streams each gzip backup into a separate restore database, and checks all rows and the restored identity sequence. Fixtures include Unicode, quotes, tabs, newlines and NULL values. Integration tests use the `integration` build tag and run through this wrapper; they do not use a developer's database connection settings. CI runs the wrapper inside both required Go-version jobs, so a failed round trip blocks merging.

The happy paths use an explicit 30-second backup limit. A separate case holds an exclusive table lock and verifies that `--timeout` cancels a blocked `pg_dump`, reports failure and removes partial output. The runner's overall test deadline remains the outer safety limit.

## Pull requests

Create a short-lived branch or fork and open a pull request against `main`. Explain the problem, resulting behavior and validation. Add regression tests for changes to observable behavior, and update usage documentation when necessary.

Required checks must pass on an up-to-date branch. Resolve review conversations and obtain the required maintainer approval before merging. New commits can invalidate earlier approvals. See [Repository policy](.github/REPOSITORY_POLICY.md) for the complete merge requirements.

Use [CODE_REVIEW.md](CODE_REVIEW.md) for the project review checklist. Keep [architecture](docs/architecture.md), [compatibility](BACKWARD_COMPATIBILITY.md) and [decision records](docs/decisions/) current when they are affected. Substantial requirements and cross-session task notes can use the templates in [.ai/](.ai/README.md).

All source code, comments, logs, CLI text and documentation must be written in English. Keep copyright notices under Nardeo. Contributions are accepted under the repository's [MIT License](LICENSE).

## Workflow and dependency changes

Pin every third-party GitHub Action to its full commit SHA and add a version comment. Keep workflow permissions narrow, use GitHub-hosted runners, and avoid executing untrusted pull request code in privileged events such as `pull_request_target`.

Dependabot opens updates for Go modules and GitHub Actions. Dependency review blocks newly introduced dependencies with known vulnerabilities rated moderate or higher. `govulncheck` checks reachable Go vulnerabilities, and CodeQL analyzes Go code and Actions workflows. Review dependency and workflow updates before merging them.
