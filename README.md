# Nardeo Backup

A small Go command-line tool for streaming PostgreSQL backups into timestamped, gzip-compressed SQL files. Built with Cobra and the standard library's structured logger, `log/slog`.

Copyright (c) 2026 Nardeo. Licensed under the [MIT License](LICENSE).

## Requirements

- Go 1.22 or newer to build from source.
- GNU Make 3.81 or newer for the shared build and validation commands.
- PostgreSQL client tools with `pg_dump` available in `PATH`.
- A PostgreSQL account with permission to read the database being backed up.
- An output filesystem supporting hard links for atomic publication of completed backups.

Use a `pg_dump` version compatible with your server. A client older than the server's major version cannot dump that server. See the [PostgreSQL documentation](https://www.postgresql.org/docs/current/app-pgdump.html).

## Build

```sh
git clone https://github.com/nardeo-org/nardeo-backup.git
cd nardeo-backup
make build
./bin/nardeo-backup --help
./bin/nardeo-backup --version
```

The Go module path is `github.com/nardeo/nardeo-backup`; the repository is hosted in the `nardeo-org` organization. Build from a checkout using the commands above.

The executable can also be named `backup`:

```sh
make build BINARY=bin/backup
```

`make build` embeds version `dev` and the current 12-character Git commit. Supply a release version with `make build VERSION=v0.1.0`; `COMMIT` can also be supplied by release tooling. Without a Git checkout, the commit defaults to `unknown`. Direct Go builds use `dev` and `unknown` unless linker metadata is supplied.

```text
nardeo-backup v0.1.0 (commit abc123def456)
```

## Configure

Copy `.env.example` to `.env`, fill in your connection settings, then export them into your shell. The tool reads environment variables; it does not automatically load `.env` files.

```sh
cp .env.example .env
# Edit .env with your database credentials before loading it.
set -a
. ./.env
set +a
./bin/nardeo-backup pg
```

When sourcing `.env`, use valid shell syntax and quote values containing spaces or shell metacharacters. Only source files you trust.

You can instead provide a single connection URL:

```sh
export DATABASE_URL='postgresql://backup:change-me@localhost:5432/app?sslmode=require'
./bin/nardeo-backup pg --output-dir ./backups
```

Prefer `DATABASE_URL` in the environment over a credential-bearing CLI argument to avoid placing credentials in shell history. URI usernames and passwords must be percent-encoded when they contain reserved characters.

| Variable | Purpose | Default |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL URI; takes precedence over the individual connection fields | Unset |
| `PGHOST` | Database hostname or Unix socket directory | `localhost` |
| `PGPORT` | Database port | `5432` |
| `PGUSER` | Database user | Required without a URL |
| `PGPASSWORD` | Database password | Unset |
| `PGDATABASE` | Database name | Required without a URL |
| `OUTPUT_DIR` | Destination directory | `./backups` |
| `NO_COLOR` | Set to a nonempty value to disable banner colors | Unset |

Without a URL, the host, port, user and database identify the connection. Passwords are supported through `PGPASSWORD` or URL credentials. Passwordless authentication and libpq password files such as `.pgpass` are also supported; `pg_dump` never prompts interactively. URL mode ignores inherited `PGHOST`, `PGHOSTADDR`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE` and `PGSERVICE` values. Other libpq settings, including TLS options and `PGPASSFILE`, are inherited. See [libpq environment variables](https://www.postgresql.org/docs/current/libpq-envars.html).

## Usage

```sh
./bin/nardeo-backup
./bin/nardeo-backup --help
./bin/nardeo-backup pg --help
./bin/nardeo-backup pg --output-dir ./backups --verbose
./bin/nardeo-backup pg --output-dir ./backups --timeout 15m
./bin/nardeo-backup pg --db-url 'postgresql://backup@localhost:5432/app'
```

| Flag | Purpose |
| --- | --- |
| `--output-dir` | Override `OUTPUT_DIR` |
| `--db-url` | Override `DATABASE_URL` |
| `--verbose` | Enable debug logs |
| `--timeout` | Maximum backup duration, such as `30s`, `15m` or `1h`; `0` disables the limit (default) |
| `--version` | Print version and commit without requiring database configuration |

Root and help commands print the Nardeo Backup banner. Backup operations write structured text logs to stderr, including the output path and elapsed time. Errors return a nonzero exit code.

A positive timeout creates a deadline for the backup context and cancels `pg_dump` when it expires. The command reports a timeout, returns a nonzero exit status and removes its temporary backup file during normal cleanup. Cleanup may take additional time after the deadline. Negative durations are rejected; a zero timeout preserves the unlimited default. Interrupt signals still cancel the operation independently.

Each successful backup produces a unique file such as:

```text
backups/postgres_20261109T120000.123456789Z-123456789.sql.gz
```

`pg_dump` output is compressed as it arrives, without buffering the database in memory. New directories use mode `0700`; backup files use mode `0600` on systems supporting Unix permissions. A completed dump is published only after gzip finalization and file synchronization. Interrupted or failed backups remove their temporary `.partial` files during normal cleanup; forced termination such as `SIGKILL` can leave a partial file behind. Existing backups are never overwritten.

Missing client tools, connection errors, filesystem errors and cancellation are reported with context. Passwords are removed from the child process's command arguments and redacted from captured diagnostics. Passwords remain available to the child through its environment.

## Restore

Restore into a prepared database using `psql`, since the backup contains plain SQL:

```sh
gzip -dc ./backups/postgres_TIMESTAMP-RANDOM.sql.gz | psql --no-password --dbname="$DATABASE_URL" --set=ON_ERROR_STOP=1
```

The URL in this example must identify the restore destination. A dump covers one database; cluster-wide roles and tablespaces require separate management.

## Development

Start with [AGENTS.md](AGENTS.md) for shared project context, [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow and [SECURITY.md](SECURITY.md) for private vulnerability reporting and operational security guidance. The current design is described in [Architecture](docs/architecture.md); task specifications and handoffs use [.ai/](.ai/README.md).

```sh
make test
make check
make security
```

`make test` runs the unit suite without Docker. `make check` is the same ordered formatting, test, vet, build and PostgreSQL integration gate used in CI; it requires Docker. Run `make security` with current stable Go for the vulnerability scan, and `make help` to list all shared commands. The [.editorconfig](.editorconfig) keeps editor whitespace settings consistent; `make fmt` applies Go formatting.

Unit and subprocess tests exercise configuration, CLI behavior, compression, subprocess failures, cancellation, private output files and credential handling. They use a controlled `pg_dump` executable and do not require a running database. GitHub Actions runs the checks on Go 1.22 and the current stable Go release.

CI also verifies a real PostgreSQL 18 backup and restore with both connection modes, comparing fixture rows and the restored identity sequence. A locked-table case checks timeout cancellation and partial-file cleanup. Run the same isolated check locally with Go, GNU Make, Bash and Docker:

```sh
make test-integration
```

See [the integration test guide](CONTRIBUTING.md#postgresql-integration-test) for environment and cleanup details.

## Releases

GoReleaser prepares Linux and macOS archives for amd64 and arm64, a SHA-256 checksum manifest, and release notes from Git history. The first planned version is `v0.1.0`. Version-tag workflows create a draft after validating the tagged `main` commit; release previews are built in the required CI checks.

Use `make release-snapshot` for a local preview with the tool versions pinned in [.tool-versions](.tool-versions). See [the release guide](docs/releases.md) for preparing tags, reproducing artifacts and publishing the draft.

## Layout

```text
cmd/nardeo-backup/   CLI entry point and Cobra commands
internal/config/    Environment configuration and flag binding
internal/backup/    Streaming PostgreSQL backup engine
internal/ui/        Banner, terminal colors and structured logging
```
